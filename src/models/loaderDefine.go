package models

type MD5Key string
type Resource struct {
	MD5       MD5Key //资源标识
	Name      string //名称
	Path      string //资源下载地址
	M_type    string //资源类型 如：.mp3 .mp4 html
	Des       string //描述
	Save_Path string //存储位置
}
