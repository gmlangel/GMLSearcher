package models

type MD5Key string
type Resource struct {
	MD5       MD5Key `json:"md5"`  //资源标识
	Name      string `json:"name"` //名称
	Path      string `json:"p"`    //资源下载地址
	M_type    string `json:"t"`    //资源类型 如：.mp3 .mp4 html
	Des       string `json:"d"`    //描述
	Save_Path string `json:"sp"`   //存储位置
	Stat      string `json:"stat"` //是否加载完毕
}

/*对应HTML中的<a>标签*/
type Tag_A struct {
	Href  string
	Title string
}
