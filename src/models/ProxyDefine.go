package models;


import(
	"github.com/go-xorm/core"
	"database/sql"
)
/**
所有委托的接口相关的定义，请定义到这里
*/

type SQLInterface interface{
	/**
	开始链接sql
	*/
	Start();

	/**
	停止链接sql
	*/
	Stop();

	/**
	执行sql语句， insert delete update
	*/
	Exec(str string) (res sql.Result, err error)

	/**
	执行sql查询语句， select
	*/
	Query(str string)(res []map[string][]byte, err error)

	/**
	执行sql查询语句， select
	*/
	QueryInterface(str string)(res []map[string]interface{}, err error)
	
	/**
	设置Debug Log等级
	*/
	SetLogLevel(arg core.LogLevel)
}