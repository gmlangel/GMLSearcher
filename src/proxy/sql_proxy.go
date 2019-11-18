package proxy
import(
	"fmt"
	"github.com/go-xorm/xorm"
	"github.com/go-xorm/core"
	"database/sql"
	"time"
	_ "github.com/go-sql-driver/mysql"
)

/**
新建SQL
*/
func NewSQL(_SQLType,_DBFullURL string)*SQLProxy  {
	return &SQLProxy{
		SQLType:_SQLType,
		DBFullURL:_DBFullURL,
		MaxIdleConns:50,
		MaxOpenConns:50,
		LogLevel:core.LOG_WARNING,
		SqlHeartOffset:time.Second * 30};
}

type SQLProxy struct{
	SQLType string //数据库类型
	DBFullURL string //"gmlmaster:123456@tcp(39.106.135.11:32306)/GMLPlanDB?charset=utf8"
	MaxIdleConns int //设置连接池的空闲数大小
	MaxOpenConns int //设置最大打开连接数
	LogLevel core.LogLevel//日志级别
	SqlHeartOffset time.Duration//心跳间隔
	sqlEngine *xorm.Engine//数据库引擎
	IsConnected bool//是否已经连接成功
	OnLinkComplete func()
}

func (sp *SQLProxy)sqlHeart(){
	for{
		sp.sqlEngine.Ping();
		time.Sleep(sp.SqlHeartOffset);
	}
}

/*
启动sql链接数据库
*/
func (sp *SQLProxy)Start(){
	//user:password@tcp(localhost:5555)/dbname?tls=skip-verify&autocommit=true
	var sqlErr error;
	sp.sqlEngine,sqlErr = xorm.NewEngine(sp.SQLType,sp.DBFullURL);
	if sqlErr == nil{
		sp.sqlEngine.Logger().SetLevel(sp.LogLevel);//控制台打印sql日志
		sp.sqlEngine.SetMaxIdleConns(sp.MaxIdleConns);//设置连接池的空闲数大小
		sp.sqlEngine.SetMaxOpenConns(sp.MaxOpenConns);//设置最大打开连接数
		sp.IsConnected = true;
		if nil != sp.OnLinkComplete{
			sp.OnLinkComplete();
			sp.OnLinkComplete = nil;
		}
		_,err := sp.sqlEngine.DBMetas();
		if err != nil{
			fmt.Printf("错误信息%v",err);
		}
		//fmt.Printf("表数据%v",sqlinfo);
		//维持sql长连接
		go sp.sqlHeart();
	}else{
		fmt.Printf("\n数据库连接错误:%v\n",sqlErr);
	}
}

/**
关闭数据库连接
*/
func (sp *SQLProxy)Stop(){
	if sp.IsConnected == true{
		sp.sqlEngine.Close();
	}
}

func (sp *SQLProxy)Exec(str string) (res sql.Result, err error){
	res,err = sp.sqlEngine.Exec(str);
	if err != nil{
		fmt.Println("sql语句执行错误:",str,"错误原因:",err);
	}
	return res,err;
}

func (sp *SQLProxy)Query(str string)(res []map[string][]byte, err error){
	res,err = sp.sqlEngine.Query(str);
	if err != nil{
		fmt.Println("sql查询语句执行失败:",str,"错误原因:",err);
	}
	return res,err;
}

func (sp *SQLProxy)QueryInterface(str string)(res []map[string]interface{}, err error){
	res,err = sp.sqlEngine.QueryInterface(str);
	if err != nil{
		fmt.Println("sql查询语句执行失败:",str,"错误原因:",err);
	}
	return res,err;
}

func (sp *SQLProxy)SetLogLevel(arg core.LogLevel){
	sp.LogLevel = arg;
	sp.sqlEngine.Logger().SetLevel(arg);
}