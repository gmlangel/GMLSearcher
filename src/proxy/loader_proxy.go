package proxy

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"fmt"

	m "../models"
)

var (
	//并行下载数
	downLoadThreads int = 10
	//数据库同步时间
	syncSQLTime   time.Duration            = time.Second * 120
	reqTimeOutMap map[string]time.Duration = map[string]time.Duration{
		".htm":  time.Second * 30,
		".html": time.Second * 30,
		".mp3":  time.Minute * 5,
		".m4a":  time.Minute * 5,
		".mp4":  time.Minute * 60}
)

/**
加载器定义
*/
type Loader struct {
	//主站域名
	BaseHost string
	//主页地址
	IndexPage string
	//待加载的资源Key数组
	WaitReqHostArr []m.MD5Key
	//已完成加载的资源KEY数组
	LoadedReqHostArr []m.MD5Key
	//加载失败的资源KEY数组
	FaildReqHostArr []m.MD5Key
	//资源字典
	ResourceMap map[m.MD5Key]*m.Resource
	//本地存储目录
	LocalDirectoryPath string

	resChan                      chan int
	LoadChan                     chan int //同一时间的资源加载请求并发数
	SQL                          m.SQLInterface
	gloger                       *log.Logger
	glogerFile                   *os.File
	resourceRecordLoger          *log.Logger
	resourceRecordLogerFile      *os.File //存已经加载完毕的所有 多媒体资源的列表
	AnalysisHandler              func([]byte, *Loader, *m.Resource)
	SaveResourceListToSQLHandler func(*Loader)
	ResourceListStr              string
}

// type LoaderInterface interface {
// 	//初始化
// 	Initial(_BaseHost string, _IndexPage string)

// 	//加载资源
// 	// @param name 资源名称
// 	// @param path 资源下载地址
// 	// @param m_type 资源类型 如：.mp3 .mp4
// 	// @param des 资源描述 如：作者xxx
// 	LoadResource(name string, path string, m_type string, des string)
// }

func (l *Loader) Initial(_BaseHost string, _IndexPage string, _baseSavePath string, _analysisHandler func([]byte, *Loader, *m.Resource), _saveResourceListToSQLHandler func(*Loader)) {
	l.BaseHost = _BaseHost
	l.IndexPage = _IndexPage
	l.AnalysisHandler = _analysisHandler
	l.SaveResourceListToSQLHandler = _saveResourceListToSQLHandler
	l.LocalDirectoryPath = _baseSavePath

	l.LoadedReqHostArr = []m.MD5Key{}
	l.FaildReqHostArr = []m.MD5Key{}
	l.resChan = make(chan int, 1)
	l.resChan <- 1
	l.LoadChan = make(chan int, downLoadThreads)
	for i := 0; i < downLoadThreads; i++ {
		l.LoadChan <- 1
	}
	//默认将_BaseHost地址，作为第一个资源，填充到资源列表
	urlPath := fmt.Sprintf("%s%s", _BaseHost, _IndexPage)
	md5 := m.MD5Key(MakeMD5(urlPath))
	res := &m.Resource{MD5: md5, Name: _IndexPage, Path: urlPath, M_type: ".html"}
	l.ResourceMap = map[m.MD5Key]*m.Resource{md5: res}
	l.WaitReqHostArr = []m.MD5Key{md5}
	//初始化日志服务
	tlog, tf, _ := MakeLogger(_baseSavePath, "loaderlog")
	l.gloger = tlog
	l.glogerFile = tf
	tlog2, tf2, _ := MakeLogger(_baseSavePath, "resourceList")
	l.resourceRecordLoger = tlog2
	l.resourceRecordLogerFile = tf2
	// //读取已经加载过的资源列表   临时屏蔽， 原因是每次从 几MB的数据中检索 字符传是否存在，开销太大
	// tmps, err3 := ioutil.ReadAll(tf2)
	// if nil == err3 {
	// 	l.ResourceListStr = string(tmps)
	// }

	go l.runloopSyncSQL() //启动数据库信息，同步机制
}

/**
开始
*/
func (l *Loader) Start() {
	go l.runloopLoadURL()
}

/**
停止，并释放
*/
func (l *Loader) StopAndDestroy() {
	_, isOK := <-l.resChan
	if true == isOK {
		close(l.resChan)
	}

	_, isOK2 := <-l.LoadChan
	if true == isOK2 {
		close(l.LoadChan)
	}

	l.ResourceListStr = ""
	l.WaitReqHostArr = nil
	l.LoadedReqHostArr = nil
	l.FaildReqHostArr = nil
	l.SQL = nil
	l.AnalysisHandler = nil

	//清理日志相关
	l.gloger = nil
	l.resourceRecordLoger = nil
	l.glogerFile.Close()
	l.glogerFile = nil
	l.resourceRecordLogerFile.Close()
	l.resourceRecordLogerFile = nil
}

/**
循环加载资源
*/
func (l *Loader) runloopLoadURL() {
	j := 0
	for {
		_, isOk2 := <-l.resChan
		if false == isOk2 {
			break
		}
		j = len(l.LoadChan) //取可用下载线程数
		for i, v := range l.WaitReqHostArr {
			if item, isContains := l.ResourceMap[v]; isContains == true && i < j {
				fmt.Println("准备下载资源", item.Name, item.Path)
				l.resourceRecordLoger.Println(item.MD5, item.Path)
				go l.loadResource(item)
			}
		}
		if j > len(l.WaitReqHostArr) {
			l.WaitReqHostArr = l.WaitReqHostArr[0:0]
		} else {
			l.WaitReqHostArr = l.WaitReqHostArr[j:len(l.WaitReqHostArr)] //移除已下载
		}
		l.resChan <- 1
	}
}

//加载资源
// @param name 资源名称
// @param path 资源下载地址
// @param m_type 资源类型 如：.mp3 .mp4 html
// @param des 资源描述 如：作者xxx
func (l *Loader) loadResource(arg *m.Resource) {
	_, isOk := <-l.LoadChan
	if isOk == false {
		return
	}
	url := arg.Path
	req, err := http.NewRequest("GET", url, nil)
	if nil != err {
		l.gloger.Println("资源加载错误,", err.Error())
	}
	timeo := time.Second * 30 //默认超时时间为30秒
	if t, isContains := reqTimeOutMap[arg.M_type]; isContains == true {
		timeo = t //设置超时时间为指定类型对应的时间
	}
	fmt.Println("设置超时时间", arg.M_type, "=", timeo)
	httpClient := &http.Client{Timeout: timeo} //new一个请求器， 设置超时时间为30秒
	resp, err := httpClient.Do(req)
	if err != nil {
		// handle error
		l.gloger.Println("资源", arg.Name, " 请求资源出错:", err.Error())
		l.LoadChan <- 1
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
		l.gloger.Println("资源", arg.Name, " 内容读取失败:", err.Error())
		l.LoadChan <- 1
		return
	}
	_, isOk2 := <-l.resChan
	if false == isOk2 {
		l.LoadChan <- 1
		return
	}
	//判断类型，做相应处理
	if nil != l.AnalysisHandler {
		l.AnalysisHandler(body, l, arg)
	}
	l.resChan <- 1
	l.LoadChan <- 1
}

/**
定时同步信息到数据库
*/
func (l *Loader) runloopSyncSQL() {

	for {
		if _, isOk := <-l.LoadChan; isOk == false {
			break
		}
		//写入数据库
		if l.SaveResourceListToSQLHandler != nil {
			l.SaveResourceListToSQLHandler(l)
		}
		//清空LoadedReqHostArr
		l.LoadedReqHostArr = l.LoadedReqHostArr[0:0]
		l.LoadChan <- 1
		time.Sleep(syncSQLTime)
	}
}
