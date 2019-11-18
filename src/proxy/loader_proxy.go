package proxy

import (
	"io/ioutil"
	"log"
	"time"

	"net/http"

	"fmt"

	m "../models"
)

var (
	//数据库同步时间
	syncSQLTime time.Duration = time.Second * 120
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

	resChan             chan int
	LoadChan            chan int //同一时间的资源加载请求并发数
	SQL                 m.SQLInterface
	gloger              *log.Logger
	resourceRecordLoger *log.Logger
	AnalysisHandler     func([]byte, *Loader)
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

func (l *Loader) Initial(_BaseHost string, _IndexPage string, _baseSavePath string) {
	l.BaseHost = _BaseHost
	l.IndexPage = _IndexPage

	l.LoadedReqHostArr = []m.MD5Key{}
	l.FaildReqHostArr = []m.MD5Key{}
	l.resChan = make(chan int, 1)
	l.resChan <- 1
	l.LoadChan = make(chan int, 20)
	for i := 0; i < 20; i++ {
		l.LoadChan <- 1
	}
	//默认将_BaseHost地址，作为第一个资源，填充到资源列表
	urlPath := fmt.Sprintf("%s%s", _BaseHost, _IndexPage)
	md5 := m.MD5Key(MakeMD5(urlPath))
	res := &m.Resource{MD5: md5, Name: _IndexPage, Path: urlPath, M_type: "html"}
	l.ResourceMap = map[m.MD5Key]*m.Resource{md5: res}
	l.WaitReqHostArr = []m.MD5Key{md5}
	//初始化日志服务
	tlog, _ := MakeLogger(_baseSavePath, "loaderlog")
	l.gloger = tlog
	tlog2, _ := MakeLogger(_baseSavePath, "resourceList")
	l.resourceRecordLoger = tlog2

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

}

/**
循环加载资源
*/
func (l *Loader) runloopLoadURL() {
	for {
		_, isOk2 := <-l.resChan
		if false == isOk2 {
			break
		}
		for _, v := range l.WaitReqHostArr {
			if item, isContains := l.ResourceMap[v]; isContains == true {
				l.resourceRecordLoger.Println(item.Path)
				go l.loadResource(item)
			}
		}
		l.WaitReqHostArr = l.WaitReqHostArr[0:0] //清空数组
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
	httpClient := &http.Client{Timeout: time.Second * 30} //new一个请求器， 设置超时时间为30秒
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
		log.Println("资源", arg.Name, " 内容读取失败:", err.Error())
		l.LoadChan <- 1
		return
	}
	_, isOk2 := <-l.resChan
	if false == isOk2 {
		l.LoadChan <- 1
		return
	}
	//将资源写入LoadedReqHostArr
	l.LoadedReqHostArr = append(l.LoadedReqHostArr, arg.MD5)
	l.resChan <- 1
	l.resourceRecordLoger.Println(string(body)) //测试用
	//判断类型，做相应处理
	if nil != l.AnalysisHandler {
		l.AnalysisHandler(body, l)
	}
	l.LoadChan <- 1
}

/**
定时同步信息到数据库
*/
func (l *Loader) runloopSyncSQL() {

	for {
		//将LoadedReqHostArr写入数据库

		//清空LoadedReqHostArr

		time.Sleep(syncSQLTime)
	}
}
