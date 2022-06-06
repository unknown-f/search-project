package repository

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/yanyiwu/gojieba"
	"gopkg.in/mgo.v2/bson"
)

var Jbfc *gojieba.Jieba
var mgomutex sync.Mutex

type Doc struct {
	ID     int    `json:"ID" bson:"ID"`
	ImgUrl string `json:"ImgUrl" bson:"ImgUrl"`
	Text   string `json:"Text" bson:"Text"`
}

//倒排索引，由关键词查文档索引
type Keyword struct {
	Name    string `json:"Name" bson:"Name"`
	DocList []int  `json:"DocList" bson:"DocList"`
}

//查询的结果:文档及得分
type SearchRlt struct {
	Grade float32 `json:"Grade" bson:"Grade"`
	DocID int     `json:"DocID" bson:"DocID"`
}

type SearchRespond struct {
	SearchTime  string   `json:"SearchTime" bson:"SearchTime"`
	SearchText  string   `json:"SearchText" bson:"SearchText"`
	ReturnRes   []Doc    `json:"ReturnRes" bson:"ReturnRes"`
	RelatedInfo []string `json:"RelatedInfo" bson:"RelatedInfo"`
}

//***结构体排序方法Start***
type SearchRltC []SearchRlt

func (s SearchRltC) Len() int {
	return len(s)
}

func (s SearchRltC) Less(i, j int) bool {
	return s[i].Grade > s[j].Grade
}

func (s SearchRltC) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

//***结构体排序方法End***

func InitJieba() {
	Jbfc = gojieba.NewJieba()
}

func SearchTopNDoc(NDoc int64) []Doc {
	var docrlt []Doc
	hotdoc, _ := redisdb.ZRevRangeWithScores("hotdoc", 0, NDoc).Result()
	fmt.Println("hotdoc", hotdoc)
	for _, docid := range hotdoc {
		docidint, _ := strconv.Atoi(docid.Member.(string))
		if docidint == 0 {
			continue
		}
		docrlt = append(docrlt, SearchOneRltToDoc(docidint))
	}
	return docrlt
}

func SearchTopNKeyword(NWord int64) []string {
	var hotkeywords []string
	hotkeyword, _ := redisdb.ZRevRangeWithScores("hotkeyword", 0, NWord).Result()
	fmt.Println("hotkeyword", hotkeyword)
	for _, docid := range hotkeyword {
		hotkeywords = append(hotkeywords, docid.Member.(string))
	}
	return hotkeywords
}

//根据检索到的文档ID，查询得到文档

func SearchOneRltToDoc(docid int) Doc {
	var docrltp Doc
	redisrlt, errs := redisdb.Get("doc:" + strconv.Itoa(docid)).Result()
	if errs != nil {
		var redisbyte []byte
		c_indextodoc.Find(bson.M{"ID": docid}).One(&docrltp)
		redisbyte, errs = json.Marshal(docrltp)
		if errs != nil {
			fmt.Println(errs)
		} else {
			redisdb.Set("doc:"+strconv.Itoa(docid), string(redisbyte), time.Hour*24)
		}
	} else {
		fmt.Println("read from redis", docid)
		redisdb.ZIncrBy("hotdoc", 1, strconv.Itoa(docid)).Result()
		errs = json.Unmarshal([]byte(redisrlt), &docrltp)
		if errs != nil {
			fmt.Println(errs)
		}
	}
	return docrltp
}

func SearchRltToDoc(rlt []SearchRlt) []Doc {
	var docrlt []Doc
	for _, rltp := range rlt {
		docrlt = append(docrlt, SearchOneRltToDoc(rltp.DocID))
	}
	return docrlt
}

func SearchRelatedInfo(Mvkey string) []string {
	keysearchrltp := SearchKeyword(Mvkey)
	var relatedinfo []string
	for _, docid := range keysearchrltp.DocList {
		docinfo := SearchOneRltToDoc(docid)
		RText := []rune(docinfo.Text)
		rlen := len(RText)
		keylen := len([]rune(Mvkey))
		kindex := strings.Index(docinfo.Text, Mvkey)
		if kindex > 0 {
			prefix := []byte(docinfo.Text)[:kindex]
			rs := []rune(string(prefix))
			kindex = len(rs)
		}
		if kindex+keylen+4 < rlen {
			relatedinfo = append(relatedinfo, string(RText[kindex:kindex+4+keylen]))

		} else {
			if kindex > 4 {
				relatedinfo = append(relatedinfo, string(RText[kindex-4:kindex+keylen]))
			}
		}
	}
	return relatedinfo
}

func SearchKeyword(key string) Keyword {
	var keysearchrltp Keyword
	redisrlt, errread := redisdb.Get("keyword:" + key).Result()
	if errread != nil {
		var redisbyte []byte
		fmt.Println("read", errread)
		errread = c_keytoindx.Find(bson.M{"Name": key}).One(&keysearchrltp)
		redisbyte, errread = json.Marshal(keysearchrltp)
		if errread != nil {
			fmt.Println(errread)
		} else {
			redisdb.Set("keyword:"+key, string(redisbyte), time.Hour*24).Result()
		}
	} else {
		errread = json.Unmarshal([]byte(redisrlt), &keysearchrltp)
		if errread != nil {
			fmt.Println(errread)
		}
	}
	return keysearchrltp
}

//text待查询的文本，maxnumofrlt返回的文档ID的数量的上限
func Search(text string, maxnumofrlt int, mrelatedinfo int) ([]SearchRlt, []string) {
	var srlt SearchRltC
	var rinfo []string
	var mvkey string
	keyword_tfidf := make(map[int]float32)
	keyword_fre := make(map[string]float32)

	words := CutWords(text, Jbfc)
	//查询每个关键词对应的DocList，并聚合到一起
	for _, value := range words {
		keysearchrltp := SearchKeyword(value)
		redisdb.ZIncrBy("hotkeyword", 1, value).Result()
		keyword_fre[value] += 1 / float32(len(keysearchrltp.DocList))
		for _, docid := range keysearchrltp.DocList {
			keyword_tfidf[docid] += 1 / float32(len(keysearchrltp.DocList))
		}
	}
	if len(keyword_tfidf) != 0 {
		mvkey = words[0]
		for key, _ := range keyword_fre {
			if keyword_fre[key] > keyword_fre[mvkey] {
				mvkey = key
			}
		}
		for key, grade := range keyword_tfidf {
			srlt = append(srlt, SearchRlt{Grade: grade, DocID: key})
		}
		sort.Sort(srlt)
		rinfo = SearchRelatedInfo(mvkey)
		if len(rinfo) > mrelatedinfo {
			rinfo = rinfo[0:mrelatedinfo]
		}
		if len(srlt) > maxnumofrlt {
			srlt = srlt[0:maxnumofrlt]
		}
	} else {
		srlt = []SearchRlt{}
	}

	fmt.Println(srlt)
	return srlt, rinfo
}

//返回文本的分词结果
func CutWords(doctext string, jbfc *gojieba.Jieba) []string {
	words := jbfc.CutForSearch(doctext, true)
	//words := jbfc.Cut(doctext, true)
	fmt.Println(words)
	return words
}
func ReadCutAndWrite(jbfc *gojieba.Jieba) error {
	//startTime := time.Now()
	fileName := "./data/wukong50k_release.csv"
	// fileName := "./data/test.csv"
	fs, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer fs.Close()

	r := csv.NewReader(fs)
	//针对大文件，一行一行的读取文件
	doc_id := 0
	for {
		row, errread := r.Read()
		fmt.Println("row", row)
		//if errread == errread {
		//	fmt.Println(time.Now().Sub(startTime))
		//}
		if len(row) == 0 {
			return nil
		}
		//目的是跳过第一行的标题
		if doc_id == 0 {
			doc_id += 1
			continue
		}
		newdoc := Doc{
			ID:     latestDocid,
			ImgUrl: row[0],
			Text:   row[1],
		}
		fmt.Println(newdoc)
		if errread != nil && errread != io.EOF {
			return errread
		}
		c_indextodoc.Insert(newdoc)
		words := CutWords(row[1], jbfc)
		//" "要不要处理一下
		for _, value := range words {
			c_keytoindx.Upsert(bson.M{"Name": value}, bson.M{"$push": bson.M{"DocList": latestDocid}})
		}
		latestDocid += 1
	}

}
func CutAndWriteOnce(doctext string) error {
	mgomutex.Lock()
	newdoc := Doc{
		ID:     latestDocid + 1,
		ImgUrl: "",
		Text:   doctext,
	}
	err := c_indextodoc.Insert(newdoc)
	if err != nil {
		return err
	}
	words := CutWords(doctext, Jbfc)
	for _, value := range words {
		_, err = c_keytoindx.Upsert(bson.M{"Name": value}, bson.M{"$push": bson.M{"DocList": latestDocid + 1}})
		if err != nil {
			return err
		}
	}
	latestDocid += 1
	mgomutex.Unlock()
	return nil
}
