package models
import(
	"time"
	"fmt"
)
var(
	//数据库同步时间
	syncSQLTime time.Duration = time.Second * 120;
)

type MD5Key string;
type Resource struct{
	MD5 MD5Key;//资源标识
	Name string;//名称
	Path string;//资源下载地址
	M_type string;//资源类型 如：.mp3 .mp4
	Des string;//描述
}
/**
加载器定义
*/
type Loader struct{
	//主站域名
	BaseHost string
	//主页地址
	IndexPage string
	//待加载的资源Key数组
	WaitReqHostArr []MD5Key
	//已完成加载的资源KEY数组
	LoadedReqHostArr []MD5Key
	//加载失败的资源KEY数组
	FaildReqHostArr []MD5Key
	//资源字典
	ResourceMap []map[MD5Key]Resource;

	resChan chan int
	LoadChan chan int
	SQL SQLInterface
}
type LoaderInterface interface{
	//初始化
	Initial(_BaseHost string,_IndexPage string);

	//加载资源
	// @param name 资源名称
	// @param path 资源下载地址
	// @param m_type 资源类型 如：.mp3 .mp4
	// @param des 资源描述 如：作者xxx
	LoadResource(name string,path string,m_type string,des string)
}

func(l *Loader)Initial(_BaseHost string,_IndexPage string){
	l.BaseHost = _BaseHost;
	l.IndexPage = _IndexPage;
	l.WaitReqHostArr = []MD5Key{};
	l.LoadedReqHostArr = []MD5Key{};
	l.FaildReqHostArr = []MD5Key{};
	l.resChan = make(chan int,1);
	l.resChan <- 1;
	l.LoadChan = make(chan int,1);
	l.LoadChan <- 1;

	go l.runloopSyncSQL();//启动数据库信息，同步机制
} 

/**
开始
*/
func(l *Loader)Start(){
	go l.runloopLoadURL();
}

/**
停止，并释放
*/
func(l *Loader)StopAndDestroy(){
	_,isOK := <- l.resChan
	if true == isOK{
		close(l.resChan)
	}
}

/**
循环加载资源
*/
func(l *Loader)runloopLoadURL(){
	_,isOk := <- l.resChan;
	for isOk == true{
		
	}
}

/**
定时同步信息到数据库
*/
func(l *Loader)runloopSyncSQL(){
	_,isOk := <- l.resChan;
	for isOk == true{

		//将LoadedReqHostArr写入数据库

		//清空LoadedReqHostArr

		l.resChan <- 1;
		time.Sleep(syncSQLTime)
	}
}


type MusicLoader Loader;

func(l *MusicLoader)test(){

}