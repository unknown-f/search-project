package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/yanyiwu/gojieba"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//读取的图文对文档
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
	Grade int `json:"Grade" bson:"Grade"`
	DocID int `json:"DocID" bson:"DocID"`
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

var URL string = "localhost:27017" //mongodb的地址

var c_indextodoc *mgo.Collection //docid与doc的对应关系集合
var c_keytoindx *mgo.Collection  //关键词与docid的对应关系集合

//这两个表用哈希索引应该比较快，这个需要后面对比一下

func main() {
	//配置分词库
	x := gojieba.NewJieba()
	defer x.Free()

	session, err := mgo.Dial(URL) //连接数据库
	if err != nil {
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	db := session.DB("search_project")
	c_indextodoc = db.C("indextosource")
	c_keytoindx = db.C("keytoindex")
	//ReadCutAndWrite(x)
	srlttest := SearchRltToDoc(Search("最帅呀", 5, x))
	fmt.Println(srlttest)
}

//根据检索到的文档ID，查询得到文档
func SearchRltToDoc(rlt []SearchRlt) []Doc {
	var docrlt []Doc
	var docrltp Doc
	if len(rlt) != 0 {
		for _, rltp := range rlt {
			c_indextodoc.Find(bson.M{"ID": rltp.DocID}).One(&docrltp)
			docrlt = append(docrlt, docrltp)
		}
	} else {
		docrlt = []Doc{}
	}
	return docrlt
}

//text待查询的文本，maxnumofrlt返回的文档ID的数量的上限
func Search(text string, maxnumofrlt int, jbfc *gojieba.Jieba) []SearchRlt {
	var keysearchrlt []int
	var keysearchrltp Keyword
	var srlt SearchRltC
	words := CutWords(text, jbfc)
	//查询每个关键词对应的DocList，并聚合到一起
	for _, value := range words {
		errread := c_keytoindx.Find(bson.M{"Name": value}).One(&keysearchrltp)
		if errread != nil {
			fmt.Println(errread)
		} else {
			keysearchrlt = append(keysearchrlt, keysearchrltp.DocList...)
		}
	}
	//对Doclist的集合排序，目的是让相同的ID排在相邻位置
	sort.Ints(keysearchrlt)
	//fmt.Println(keysearchrlt)
	if len(keysearchrlt) != 0 {
		var pindex int = 0
		var p int = keysearchrlt[0]
		//统计每个DocID出现的次数，即该DocID与检索文本的关联度得分
		for index, docid := range keysearchrlt {
			if docid != p {
				srlt = append(srlt, SearchRlt{Grade: index - pindex, DocID: docid})
				p = docid
				pindex = index
			}
		}
		//按得分从高到低排序
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

//读取数据集，倒排索引，写入数据库
func ReadCutAndWrite(jbfc *gojieba.Jieba) error {
	fileName := "./data/wukong50k_release.csv"
	//fileName := "./data/test.csv"
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
