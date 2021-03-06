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
//从redis中获取热度最高的文档
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
//从redis中获取热度最高的关键词
func SearchTopNKeyword(NWord int64) []string {
	var hotkeywords []string
	hotkeyword, _ := redisdb.ZRevRangeWithScores("hotkeyword", 0, NWord).Result()
	fmt.Println("hotkeyword", hotkeyword)
	for _, docid := range hotkeyword {
		hotkeywords = append(hotkeywords, docid.Member.(string))
	}
	return hotkeywords
}

//根据文档ID，查询得到文档
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

//根据文档ID，组合搜索结果
func SearchRltToDoc(rlt []SearchRlt) []Doc {
	var docrlt []Doc
	for _, rltp := range rlt {
		docrlt = append(docrlt, SearchOneRltToDoc(rltp.DocID))
	}
	return docrlt
}
//根据文档中最重要的关键词，进行相关搜索的推荐
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
		if kindex+keylen+6 < rlen {
			relatedinfo = append(relatedinfo, string(RText[kindex:kindex+6+keylen]))

		} else {
			if kindex > 6 {
				relatedinfo = append(relatedinfo, string(RText[kindex-6:kindex+keylen]))
			}
		}
	}
	return relatedinfo
}
//根据关键词查询相应的文档序号，如果缓存中不存在就写入缓存
func SearchKeyword(key string) Keyword {
	var keysearchrltp Keyword
	redisrlt, errread := redisdb.Get("keyword:" + key).Result()
	if errread != nil {
		var redisbyte []byte
		fmt.Println("read", errread)
		fmt.Println("readkeyword", errread)
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

//text待查询的文本，filtration被过滤的关键词，maxnumofrlt返回的文档ID的数量的上限，mrelatedinfo返回的相关搜索信息的数量的上线
func Search(text string, filtration []string, maxnumofrlt int, mrelatedinfo int) ([]SearchRlt, []string) {
	var srlt SearchRltC
	var rinfo []string
	var mvkey string
	keyword_tfidf := make(map[int]float32)//存储所有可能相关的文档的tf-idf
	keyword_fre := make(map[string]float32)//存储每个关键词出现的频数
	keyword_filter := make(map[int]struct{})//记录要过滤的关键词
	for _, ff := range filtration {
		wordsfilter := CutWords(ff, Jbfc)
		for _, fv := range wordsfilter {
			keysearchfilter := SearchKeyword(fv)
			for _, docid := range keysearchfilter.DocList {
				keyword_filter[docid] = struct{}{}//通过hashmap记录过滤词
			}
		}
	}
	words := CutWords(text, Jbfc)
	for _, value := range words {
		keysearchrltp := SearchKeyword(value)
		redisdb.ZIncrBy("hotkeyword", 1, value).Result()//关键词每被搜索一次即在redis中更新关键词的热度
		keyword_fre[value] += 1 / float32(len(keysearchrltp.DocList))
		for _, docid := range keysearchrltp.DocList {
			_, ok := keyword_filter[docid]
			if ok == true {
				continue
			}
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
	//words := jbfc.Cut(doctext, true) 精准分词模式
	fmt.Println(words)
	return words
}
//读悟空图文对数据集->分词->建立倒排索引
func ReadCutAndWrite(jbfc *gojieba.Jieba) error {
	fileName := "./data/wukong50k_release.csv"
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
		latestDocid += 1
		for _, value := range words {//分词，建立倒排索引
			c_keytoindx.Upsert(bson.M{"Name": value}, bson.M{"$push": bson.M{"DocList": latestDocid}})
		}

	}

}
//分词并建立倒排索引，用于用户上传的新的文档
func CutAndWriteOnce(doctext string) error {
	var redisbyte []byte
	newdoc := Doc{
		ID:     latestDocid + 1,
		ImgUrl: "",
		Text:   doctext,
	}
	fmt.Println(newdoc)
	err := c_indextodoc.Insert(newdoc)
	if err != nil {
		return err
	}
	redisbyte, err = json.Marshal(newdoc)
	if err != nil {
		fmt.Println(err)
	} else {
		redisdb.Set("doc:"+strconv.Itoa(latestDocid+1), string(redisbyte), time.Hour*24)//用户上传新文档时直接存一份缓存
	}
	words := CutWords(doctext, Jbfc)
	for _, value := range words {//分词，建立倒排索引
		fmt.Println(value, latestDocid+1)
		_, err = c_keytoindx.Upsert(bson.M{"Name": value}, bson.M{"$push": bson.M{"DocList": latestDocid + 1}})
		//引入新文档导致部分关键词的倒排索引发生变化，采用先更新数据库再删除缓存的方法保证缓存与数据库的一致性
		redisdb.Del("keyword:" + value)
		if err != nil {
			return err
		}
	}
	latestDocid += 1
	return nil
}
