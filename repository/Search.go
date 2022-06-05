package repository

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/yanyiwu/gojieba"
	"gopkg.in/mgo.v2/bson"
)

var Jbfc *gojieba.Jieba

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
	SearchTime string `json:"SearchTime" bson:"SearchTime"`
	SearchText string `json:"SearchText" bson:"SearchText"`
	ReturnRes  []Doc  `json:"ReturnRes" bson:"ReturnRes"`
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

//text待查询的文本，maxnumofrlt返回的文档ID的数量的上限
func Search(text string, maxnumofrlt int) []SearchRlt {
	var keysearchrltp Keyword
	var srlt SearchRltC
	keyword_ifidf := make(map[int]float32)
	words := CutWords(text, Jbfc)
	//查询每个关键词对应的DocList，并聚合到一起
	for _, value := range words {
		redisrlt, errread := redisdb.Get("keyword:" + value).Result()
		if errread != nil {
			var redisbyte []byte
			fmt.Println("read1", errread)
			errread = c_keytoindx.Find(bson.M{"Name": value}).One(&keysearchrltp)
			redisbyte, errread = json.Marshal(keysearchrltp)
			if errread != nil {
				fmt.Println(errread)
			} else {
				redisdb.Set("keyword:"+value, string(redisbyte), time.Hour*24).Result()
			}
		} else {
			redisdb.ZIncrBy("hotkeyword", 1, value).Result()
			errread = json.Unmarshal([]byte(redisrlt), &keysearchrltp)
			if errread != nil {
				fmt.Println(errread)
			}
			//redisdb.ZAdd("hotkeyword", redis.Z{Score: 1, Member: "keyword:" + value})
		}
		for _, docid := range keysearchrltp.DocList {
			keyword_ifidf[docid] += 1 / float32(len(keysearchrltp.DocList))
		}
	}
	if len(keyword_ifidf) != 0 {
		for key, grade := range keyword_ifidf {
			srlt = append(srlt, SearchRlt{Grade: grade, DocID: key})
		}
		sort.Sort(srlt)
		if len(srlt) > maxnumofrlt {
			srlt = srlt[0:maxnumofrlt]
		}
	} else {
		srlt = []SearchRlt{}
	}
	fmt.Println(srlt)
	return srlt
}

//返回文本的分词结果
func CutWords(doctext string, jbfc *gojieba.Jieba) []string {
	words := jbfc.CutForSearch(doctext, true)
	fmt.Println(words)
	return words
}
func ReadCutAndWrite(jbfc *gojieba.Jieba) error {
	startTime := time.Now()
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
		if errread == errread {
			fmt.Println(time.Now().Sub(startTime))
		}
		if len(row) == 0 {
			return nil
		}
		//目的是跳过第一行的标题
		if doc_id == 0 {
			doc_id += 1
			continue
		}
		newdoc := Doc{
			ID:     doc_id,
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
			c_keytoindx.Upsert(bson.M{"Name": value}, bson.M{"$push": bson.M{"DocList": doc_id}})
		}
		doc_id += 1
	}

}
