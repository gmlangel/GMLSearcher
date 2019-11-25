package proxy

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	m "../models"
)

var (
	getAllSingerInfoInterface string = `https://u.y.qq.com/cgi-bin/musicu.fcg?-=getUCGI025670438061479395&g_tk=5381&loginUin=%s&hostUin=0&format=json&inCharset=utf8&outCharset=utf-8&notice=0&platform=yqq.json&needNewCode=0&data={"comm":{"ct":24,"cv":0},"singerList":{"module":"Music.SingerListServer","method":"get_singer_list","param":{"area":-100,"sex":-100,"genre":-100,"index":-100,"sin":%d,"cur_page":%d}}}`
	getAllSongInfoFromSinger  string = `https://u.y.qq.com/cgi-bin/musicu.fcg?-=getSingerSong3676457483501352&g_tk=5381&loginUin=%s&hostUin=0&format=json&inCharset=utf8&outCharset=utf-8&notice=0&platform=yqq.json&needNewCode=0&data={"comm":{"ct":24,"cv":0},"singerSongList":{"method":"GetSingerSongList","param":{"order":1,"singerMid":"%s","begin":0,"num":%d},"module":"musichall.song_list_server"}}`
	getSongInfo               string = `https://u.y.qq.com/cgi-bin/musicu.fcg?-=getplaysongvkey3359389840108824&g_tk=5381&loginUin=%s&hostUin=0&format=json&inCharset=utf8&outCharset=utf-8&notice=0&platform=yqq.json&needNewCode=0&data={"req_0":{"module":"vkey.GetVkeyServer","method":"CgiGetVkey","param":{"guid":"%s","songmid":["%s"],"songtype":[0],"uin":"%s","loginflag":1,"platform":"20"}},"comm":{"uin":"%s","format":"json","ct":24,"cv":0}}`
)

type Reptile_QQMusic struct {
	singerMap     map[string]string
	uin           string //qq音乐网站，登录前的用户ID
	uid           string //qq音乐网站，登录后的用户ID
	baseSavePath  string
	maxPageID     int //歌手列表页的最大页数
	currentPageID int //当前歌手列表页的id。 1-300之间，具体总页数，去https://y.qq.com/portal/singer_list.html页面看
}

// type Singer struct {
// 	Sid    int    //歌手ID
// 	S_Mid  string //用于查询歌曲的歌手ID
// 	S_Name string //歌手名称
// }

//生成 “分页获取歌手信息的接口”
func (reptile *Reptile_QQMusic) makeGetAllSingerInfoInterface(uin string, _pid int) string {
	return fmt.Sprintf(getAllSingerInfoInterface, uin, (_pid-1)*80, _pid)
}

//生成 “获取歌手对应的歌曲信息接口”
func (reptile *Reptile_QQMusic) makeGetSongInfoBySingerInterface(uin string, singerMid string, songCount int) string {
	return fmt.Sprintf(getAllSongInfoFromSinger, uin, singerMid, songCount)
}

//生成 “获取歌曲信息”
func (reptile *Reptile_QQMusic) makeGetSongInfoBySongMid(uin string, uid string, songMid string) string {
	return fmt.Sprintf(getSongInfo, uin, uid, songMid, uin, uin)
}

func (reptile *Reptile_QQMusic) Init() {
	reptile.currentPageID = 1
	reptile.maxPageID = 1
	reptile.uin = "1152921504788623067"
	reptile.uid = "7068150205"
	reptile.singerMap = map[string]string{}
	//fmt.Println(reptile.makeGetAllSingerInfoInterface(uid))
}

func (reptile *Reptile_QQMusic) Start(sqlPro *SQLProxy, cp int, cpmax int) {
	if cp > -1 {
		reptile.currentPageID = cp
	}

	if cpmax > -1 {
		reptile.maxPageID = cpmax
	}
	reptile.baseSavePath = fmt.Sprintf("./music/QQMusic_%d_%d/", reptile.currentPageID, reptile.maxPageID) //文件存储路径
	//初始化资源加载器
	resLoader := &Loader{SQL: sqlPro}
	var url string
	var res *m.Resource
	var md5 m.MD5Key
	var baseReq []*m.Resource = []*m.Resource{}
	url = reptile.makeGetAllSingerInfoInterface(reptile.uin, reptile.currentPageID)
	md5 = m.MD5Key(MakeMD5(url))
	res = &m.Resource{MD5: md5, Path: url, M_type: "makeGetAllSingerInfoInterface"}
	baseReq = append(baseReq, res)
	//封装，超时时间map
	reqTimeOutMap := map[string]time.Duration{
		".htm":                             time.Second * 30,
		".html":                            time.Second * 30,
		".mp3":                             time.Minute * 5,
		".m4a":                             time.Minute * 5,
		".mp4":                             time.Minute * 60,
		"makeGetSongInfoBySongMid":         time.Second * 60,
		"makeGetSongInfoBySingerInterface": time.Second * 60,
		"makeGetAllSingerInfoInterface":    time.Second * 60}

	resLoader.Initial(baseReq, reptile.baseSavePath, reqTimeOutMap, reptile)
	resLoader.Start() //开始加载
}

func (reptile *Reptile_QQMusic) AnalysisHandler(bts []byte, l *Loader, res *m.Resource) {
	mType := res.M_type
	var k, val string
	var url string
	var nres *m.Resource
	var md5 m.MD5Key
	_ = k
	_ = val
	_ = nres
	_ = url
	_ = md5
	if mType == "makeGetAllSingerInfoInterface" {
		if reptile.currentPageID < reptile.maxPageID {
			reptile.currentPageID++
			//封装下一页歌手的请求
			url = reptile.makeGetAllSingerInfoInterface(reptile.uin, reptile.currentPageID)
			md5 = m.MD5Key(MakeMD5(url))
			nres = &m.Resource{MD5: md5, Path: url, M_type: "makeGetAllSingerInfoInterface"}
			l.AddResourceToLoadQueue(md5, nres)
		}
		//解析json，提取歌手信息，批量生成歌手信息请求连接
		var allSinger map[string]interface{}
		jsonErr := json.Unmarshal(bts, &allSinger)
		if nil != jsonErr {
			l.gloger.Println("资源", res.Name, "Err:", jsonErr.Error())
		} else {
			//取json信息
			if obj, isOk := allSinger["singerList"].(map[string]interface{}); isOk == true {
				if obj2, isOk := obj["data"].(map[string]interface{}); isOk == true {
					if singerlist, isOk := obj2["singerlist"].([]interface{}); isOk == true {
						for _, v := range singerlist {
							if tv, isOk := v.(map[string]interface{}); isOk == true {
								if k, isOK2 := tv["singer_mid"].(string); isOK2 == true {
									if val, isOK3 := tv["singer_name"].(string); isOK3 == true {
										reptile.singerMap[k] = val
										//生成待加载的歌手信息资源
										url = reptile.makeGetSongInfoBySingerInterface(reptile.uin, k, 150) //默认读取该歌手的150首歌
										md5 = m.MD5Key(MakeMD5(url))
										nres = &m.Resource{MD5: md5, Path: url, M_type: "makeGetSongInfoBySingerInterface", Des: val}
										l.AddResourceToLoadQueue(md5, nres)
									}
								}
							}
						}
						return
					}
				}
			}
			l.gloger.Println("资源", res.Name, "Err:singerList字段不存在")
		}
	} else if mType == "makeGetSongInfoBySingerInterface" {
		//解析数据
		var jsonObj map[string]interface{}
		jsonErr := json.Unmarshal(bts, &jsonObj)
		if jsonErr != nil {
			l.gloger.Println("资源", res.Name, "Err:", jsonErr.Error())
		} else {
			if singerSongList, isOk := jsonObj["singerSongList"].(map[string]interface{}); isOk == true {
				if data, isOk := singerSongList["data"].(map[string]interface{}); isOk == true {
					if songList, isOk := data["songList"].([]interface{}); isOk == true {
						singerName := res.Des
						for _, v := range songList {
							if nv, isOk := v.(map[string]interface{}); isOk == true {
								if songInfo, isOk := nv["songInfo"].(map[string]interface{}); isOk == true {
									if k, isOk := songInfo["mid"].(string); isOk == true {
										songName, _ := songInfo["name"].(string)
										//获得了歌曲id
										url = reptile.makeGetSongInfoBySongMid(reptile.uin, reptile.uid, k) //默认读取该歌手的30首歌
										md5 = m.MD5Key(MakeMD5(url))
										nres = &m.Resource{MD5: md5, Path: url, M_type: "makeGetSongInfoBySongMid", Des: singerName, Name: songName}
										l.AddResourceToLoadQueue(md5, nres)
									}
								}
							}
						}
					}
				}
			}
		}
	} else if mType == "makeGetSongInfoBySongMid" {
		//解析歌曲详细信息，分析歌曲真正的下载地址
		var jsonObj map[string]interface{}
		jsonErr := json.Unmarshal(bts, &jsonObj)
		if jsonErr != nil {
			l.gloger.Println("资源", res.Name, "Err:", jsonErr.Error())
		} else {
			if req_0, isOk := jsonObj["req_0"].(map[string]interface{}); isOk == true {
				if data, isOk := req_0["data"].(map[string]interface{}); isOk == true {
					if sip, isOk := data["sip"].([]interface{}); isOk == true && len(sip) > 0 {
						host, _ := sip[len(sip)-1].(string)
						des := res.Des
						songName := res.Name
						if midurlinfo, isOk := data["midurlinfo"].([]interface{}); isOk == true {
							if len(midurlinfo) > 0 {
								v := midurlinfo[0]
								if nv, isOk := v.(map[string]interface{}); isOk == true {
									if purl, isOk := nv["purl"].(string); isOk == true {
										//获得了歌曲播放地址
										purl = strings.Replace(purl, "\\u0026", "&", -1)
										url = host + purl
										md5 = m.MD5Key(MakeMD5(url))
										nres = &m.Resource{MD5: md5, Path: url, M_type: ".m4a", Des: des, Name: songName}
										l.AddResourceToLoadQueue(md5, nres)
									}
								}
							}
						}
					}
				}
			}
		}
	} else if res.M_type == ".m4a" {
		//本地存储
		res.Stat = "err_wuxiao"
		if len(bts) < 256024 {
			//不够 256K证明音乐不完整
			return
		}
		fileName := res.Name + res.M_type
		//保存资源到本地
		filePath, err := SaveFileToLocal(reptile.baseSavePath, fileName, bts)
		if nil != err {
			l.gloger.Println("资源", res.Name, "本地存储出错:", err.Error())
		} else {
			res.Stat = "ok"
			//资源本地存储成功，则改变资源状态
			res.Save_Path = filePath
			//将资源写入LoadedReqHostArr
			l.LoadedReqHostArr = append(l.LoadedReqHostArr, res.MD5)
		}
	}
}

func (reptile *Reptile_QQMusic) SaveResourceListToSQL(l *Loader) {
	sqlstr := "insert `music_0`(`m_name`,`save_path`,`m_type`,`lastUpdate`,`des`) values"
	argstr := "" //;
	gmlformat := "('%s','%s','%s',%d,'%s')"
	timeValue := time.Now().Unix() / 1000
	m_name := ""
	save_path := ""
	des := ""
	//将LoadedReqHostArr写入数据库
	for _, key := range l.LoadedReqHostArr {
		if item, isOk := l.ResourceMap[key]; isOk == true && item.M_type == ".m4a" {
			//字符传的特殊处理，防止字符传中含有特殊字符，造成写库失败
			m_name = EncodeBase64([]byte(item.Name))
			save_path = EncodeBase64([]byte(item.Save_Path))
			des = EncodeBase64([]byte(item.Des))
			//遍历多媒体资源，将之写入数据库
			str := fmt.Sprintf(gmlformat, m_name, save_path, item.M_type, timeValue, des)
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
