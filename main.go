package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	_ "./src"
	pro "./src/proxy"
	_ "github.com/iris-contrib/middleware/cors"
	_ "github.com/kataras/iris"
)

var (
	/*sql相关*/
	sqlType    = "mysql"
	sqlFullURL = "gmlmaster:123456@tcp(39.106.135.11:32306)/GMLResource?charset=utf8"
)

func main() {
	loadResource()
	return
	fmt.Println("GMLSearcher=====>启动中")
	runLoopChan := make(chan int)
	// app := iris.New();
	// app.Get("test",func(ctx iris.Context){
	// 	ctx.Write([]byte("测试成功"))
	// })
	// fmt.Println("GMLSearcher=====>启动成功")
	// app.Run(iris.Addr("0.0.0.0:65535"));
	sqlPro := pro.NewSQL(sqlType, sqlFullURL)
	sqlPro.OnLinkComplete = func() {
		log.Println("数据库连接成功")
		//初始化资源加载器
		resLoader := &pro.Loader{SQL: sqlPro}
		resLoader.Initial("http://www.9ku.com", "/", "./music/9ku/", pro.AnalysisHandler_9Ku, pro.SaveResourceListToSQL_9k)
		resLoader.Start() //开始加载
	}
	go sqlPro.Start()
	//lm := src.New();

	<-runLoopChan
	fmt.Println("GMLSearcher=====>停止")

}

//加载资源
// @param name 资源名称
// @param path 资源下载地址
// @param m_type 资源类型 如：.mp3 .mp4 html
// @param des 资源描述 如：作者xxx
func loadResource() {

	url := "http://isure.stream.qqmusic.qq.com/C400000V8En93R3Dvd.m4a?tb=20&guid=7068150205&vkey=2254E708B03F7201143D8F3AB56A66B845898AD4B1D17E7AFB3B55D0588B1E8DB5353F0B66482BF41093F4873D575F0CFA3D50C73EB7B21C&uin=3803&fromtag=66"
	//url := "https://y.qq.com/n/yqq/toplist/4.html#stat=y_new.top.pop.logout"
	req, err := http.NewRequest("GET", url, nil)
	//req.Header.Set("content-type", "application/json;charset=utf-8")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.131 Safari/537.36")

	if nil != err {
		fmt.Println("资源加载错误,", err.Error())
	}
	timeo := time.Minute * 5                   //默认超时时间为5分钟
	httpClient := &http.Client{Timeout: timeo} //new一个请求器， 设置超时时间为30秒
	resp, err := httpClient.Do(req)
	if err != nil {
		// handle error
		fmt.Println("资源 请求资源出错:", err.Error())
		return
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
		fmt.Println("资源内容读取失败:", err.Error())
		return
	}
	// fmt.Println(string(body))

	//写本地
	strPath, fileerr := pro.SaveFileToLocal("./music/", "test.m4a", body)
	if fileerr != nil {
		fmt.Println("fileerr保存错误:", fileerr)
	} else {
		fmt.Println("文件存储路径", strPath)
	}
}

func makeRequest(url string) (*http.Request, error) {
	var req *http.Request
	var err error
	req, err = http.NewRequest("get", url, nil)
	if err == nil {
		//req.Header = map[string][]string{
		//"Content-Type": {"application/text; charset=utf-8"},
		//"User-Agent":   {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.108 Safari/537.36"}}
		//http://isure.stream.qqmusic.qq.com/C400000V8En93R3Dvd.m4a?guid=7068150205&vkey=2254E708B03F7201143D8F3AB56A66B845898AD4B1D17E7AFB3B55D0588B1E8DB5353F0B66482BF41093F4873D575F0CFA3D50C73EB7B21C&uin=3803&fromtag=66
		//http://isure.stream.qqmusic.qq.com/C400003UAhhG2Bm3Nq.m4a?guid=7068150205&vkey=FAC247B2EC11AD113757FE6951780103064D677ADBA543BAE87B710CC027352FA0F2A62E6265F371B802C473DDD6F27E05B7A94600CBBED2&uin=3803&fromtag=66
		req.Header.Set("content-type", "application/json;charset=utf-8")
		req.Header.Set("user-agent", "'User-Agent': 'Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.131 Safari/537.36'")
		// req.AddCookie(&http.Cookie{Name: "pgv_pvi", Value: "2986213376"})
		// req.AddCookie(&http.Cookie{Name: "RK", Value: "0EiA3v/tcz"})
		// req.AddCookie(&http.Cookie{Name: "ptcz", Value: "049eda3257e45bdad88cb5ee990f2b29823734209ee3a517fb43adafdb6326f4"})
		// req.AddCookie(&http.Cookie{Name: "tvfe_boss_uuid", Value: "7f8e2ca64e64f297"})
		// req.AddCookie(&http.Cookie{Name: "pgv_pvid", Value: "7068150205"})
		// req.AddCookie(&http.Cookie{Name: "pgv_info", Value: "ssid=s5381139750"})
		// req.AddCookie(&http.Cookie{Name: "_qpsvr_localtk", Value: "0.709412766972437"})
		// req.AddCookie(&http.Cookie{Name: "pgv_si", Value: "s2784992256"})
		// req.AddCookie(&http.Cookie{Name: "wxuin", Value: "o1152921504788623067"})
		// req.AddCookie(&http.Cookie{Name: "qm_keyst", Value: "9B2AEFF153BA720F467AC5CED17FDE6D93BB7BDEC0953034E9BEE457D40FA84F"})
		// req.AddCookie(&http.Cookie{Name: "wxopenid", Value: "opCFJw44yIRx5N4tfMSUWybq0JA0"})
		// req.AddCookie(&http.Cookie{Name: "wxrefresh_token", Value: "27_inGi_cGIQeBqfO1A6QF_uYuhSF4_5KJGn5Bi86qdoABJzew71YnA7rlvqeHtBhxNp7BXEUuoh_EkXgPR9hzC02Io9M0eh1gHglJNh9MnJ9Q"})
		// req.AddCookie(&http.Cookie{Name: "wxunionid", Value: "oqFLxsr99H6U_96z3LKAk8Q2hRlg"})
		// req.AddCookie(&http.Cookie{Name: "login_type", Value: "2"})
		// req.AddCookie(&http.Cookie{Name: "qqmusic_fromtag", Value: "66"})
	}

	return req, err
}
