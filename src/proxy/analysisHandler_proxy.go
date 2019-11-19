package proxy

import (
	"regexp"
	"strings"

	m "../models"
)

var (
	//A_TagReg, err = regexp.Compile(`<a([\s]+|[\s]+[^<>]+[\s]+)href=(\"([^<>"\']*)\"|\'([^<>"\']*)\')[^<>]*>`)
	A_TagReg, _             = regexp.Compile(`<a .*?>.*?</a>`)
	A_TagHRefReg, _         = regexp.Compile(`(?i:href)[\s]*?=[\s]*?[\"\'].*?[\"\']`)
	A_TagTitleReg, _        = regexp.Compile(`(?i:title)[\s]*?=[\s]*?[\"\'].*?[\"\']`)
	Content_Reg, _          = regexp.Compile(`<.*?>`)
	DeleteYinHao_Reg, _     = regexp.Compile(`[\"\'].*?[\"\']`)
	minSize             int = 1024 * 512
)

/**
生成A标签
*/
func makeATag(bts []byte) *m.Tag_A {
	result := &m.Tag_A{}
	if nil == DeleteYinHao_Reg {
		return result
	}
	if A_TagHRefReg != nil {
		result.Href = string(A_TagHRefReg.Find(bts))
		result.Href = DeleteYinHao_Reg.FindString(result.Href)
		result.Href = strings.ReplaceAll(result.Href, "\"", "")
		result.Href = strings.ReplaceAll(result.Href, "'", "")
	}
	if A_TagTitleReg != nil {
		result.Title = string(A_TagTitleReg.Find(bts))
		if result.Title != "" {
			//取引号中间的内容
			result.Title = DeleteYinHao_Reg.FindString(result.Title)
			result.Title = strings.ReplaceAll(result.Title, "\"", "")
			result.Title = strings.ReplaceAll(result.Title, "'", "")
		} else if result.Title == "" && Content_Reg != nil {
			//没有title就取innerText
			result.Title = string(bts)
			arr := Content_Reg.FindAllString(result.Title, -1)
			for _, v := range arr {
				result.Title = strings.Replace(result.Title, v, "", 1)
			}
		}
	}
	return result
}

/**
http://www.9ku.com/
的资源解析器
*/
func AnalysisHandler_9Ku(bts []byte, l *Loader, res *m.Resource) {
	if ".html" == res.M_type || ".htm" == res.M_type {
		//解析网页,取出所有<a>标签
		if A_TagReg != nil {
			byts := A_TagReg.FindAll(bts, -1)
			var aTag *m.Tag_A
			var tmphref string
			var canDownload bool = false
			for _, v := range byts {
				aTag = makeATag(v)
				if aTag.Href != "" && aTag.Title != "" {
					//有链接，有名称,开始区分资源类型
					tmphref = aTag.Href
					//去除连接中原有的域名，类似于将http://love.9ku.com/go/20060805/ 转为 /go/20060805/
					tmphref = strings.ReplaceAll(tmphref, l.BaseHost, "")
					if strings.Contains(tmphref, ":/") == true {
						//其它网站的连接，不需要去请求
					} else if strings.Contains(tmphref, ".htm") == true || strings.Contains(tmphref, "/") == true {
						resource := &m.Resource{}
						resource.Name = aTag.Title
						resource.MD5 = m.MD5Key(MakeMD5(tmphref))
						canDownload = false
						if strings.Index(tmphref, "/play/") == 0 {
							canDownload = true
							//为m4a音频资源，可以直接下载
							tmphref = strings.ReplaceAll(tmphref, ".htm", ".m4a") //获取真正的下载地址
							resource.Path = strings.ReplaceAll(tmphref, "/play/", "http://mp3.9ku.com/m4a/")
							resource.M_type = ".m4a"
							resource.Des = ""
						} else if strings.Contains(tmphref, "/play/") == false {
							canDownload = true
							tmphref = l.BaseHost + tmphref //拼装url请求地址
							//网页连接，需要让爬虫继续检索
							resource.Path = tmphref
							resource.M_type = ".htm"
							resource.Des = ""
						}
						if _, isContains := l.ResourceMap[resource.MD5]; isContains == false && canDownload == true {
							//如果该资源是未下载过的，则将该资源加入到WaitLoad等待列表
							l.ResourceMap[resource.MD5] = resource
							l.WaitReqHostArr = append(l.WaitReqHostArr, resource.MD5)
						}
					}
				}
			}
		}
	} else {
		if len(bts) < minSize {
			//不够 512K证明音乐不完整
			return
		}
		fileName := res.Name + res.M_type
		//保存资源到本地
		filePath, err := SaveFileToLocal(l.LocalDirectoryPath, fileName, bts)
		if nil != err {
			l.gloger.Println("资源", res.Name, "本地存储出错:", err.Error())
		} else {
			//资源本地存储成功，则改变资源状态
			res.Save_Path = filePath
			//将资源写入LoadedReqHostArr
			l.LoadedReqHostArr = append(l.LoadedReqHostArr, res.MD5)
		}
	}
}
