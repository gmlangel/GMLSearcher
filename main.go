package main;

import(
	"fmt"
	_ "github.com/kataras/iris"
	_"github.com/iris-contrib/middleware/cors"
	_ "./src"
	pro "./src/proxy"
	"log"
	m "./src/models"
)

var(
	/*sql相关*/
	sqlType = "mysql";
	sqlFullURL = "gmlmaster:123456@tcp(39.106.135.11:32306)/GMLResource?charset=utf8";
)

func main(){
	fmt.Println("GMLSearcher=====>启动中")
	runLoopChan := make(chan int);
	// app := iris.New();
	// app.Get("test",func(ctx iris.Context){
	// 	ctx.Write([]byte("测试成功"))
	// })
	// fmt.Println("GMLSearcher=====>启动成功")
	// app.Run(iris.Addr("0.0.0.0:65535"));
	sqlPro := pro.NewSQL(sqlType,sqlFullURL)
	sqlPro.OnLinkComplete = func(){
		log.Println("数据库连接成功");
	}
	go sqlPro.Start();
	//lm := src.New();
	
	<- runLoopChan
	fmt.Println("GMLSearcher=====>停止")
}