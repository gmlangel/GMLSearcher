package main

import (
	"fmt"
	"log"

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
