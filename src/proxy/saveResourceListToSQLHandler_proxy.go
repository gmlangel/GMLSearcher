package proxy

import (
	"fmt"
	"time"
)

func SaveResourceListToSQL_9k(l *Loader) {
	sqlstr := "insert `music_0`(`m_name`,`source_path`,`save_path`,`m_type`,`lastUpdate`,`des`) values"
	argstr := "" //;
	gmlformat := "('%s','%s','%s','%s',%d,'%s')"
	timeValue := time.Now().Unix() / 1000
	//将LoadedReqHostArr写入数据库
	for _, key := range l.LoadedReqHostArr {
		if item, isOk := l.ResourceMap[key]; isOk == true && item.M_type != ".htm" && item.M_type != ".html" {
			//遍历多媒体资源，将之写入数据库
			str := fmt.Sprintf(gmlformat, item.Name, item.Path, item.Save_Path, item.M_type, timeValue, item.Des)
			if argstr == "" {
				argstr = str
			} else {
				argstr = argstr + "," + str
			}
		}
	}
	if argstr != "" {
		//写库
		sqlstr += argstr
		_, err := l.SQL.Query(sqlstr)
		if nil != err {
			l.gloger.Println("数据库写入失败")
		}
	}
}
