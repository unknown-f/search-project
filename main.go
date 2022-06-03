package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
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

var URL string = "localhost:27017" //mongodb的地址

var c_indextodoc *mgo.Collection //docid与doc的对应关系集合
var c_keytoindx *mgo.Collection  //关键词与docid的对应关系集合
var redisdb *redis.Client

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
	// ReadCutAndWrite(x)
	err = initRedis()
	if err != nil {
		panic(err)
	}
	fmt.Println(redisdb.ZIncrBy("hotdoc", 1, "doc1").Result())
	r := gin.Default() //打开服务器
	r.Use(Cors())

	// 加载 html 文件
	r.LoadHTMLFiles("./view.html")
	r.StaticFile("/logo.png", "./logo.png")

	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "view.html", nil)
	})

	r.GET("/:time/:text", func(c *gin.Context) {
		time := c.Param("time")
		text := c.Param("text")
		srlttest := SearchRltToDoc(Search(text, 5, x))
		data := SearchRespond{
			SearchTime: time,
			SearchText: text,
			ReturnRes:  srlttest,
		}
		fmt.Println("return:", data)
		c.JSON(200, data)
	})
	err = r.Run()
	if err != nil {
		return
	}
}

func initRedis() error {
	redisdb = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379", // 指定
		Password: "",
		DB:       0, // redis一共16个库，指定其中一个库即可
	})
	_, err := redisdb.Ping().Result()
	return err
}

func SearchTopNDoc(NDoc int64) []Doc {
	var docrlt []Doc
	hotdoc, _ := redisdb.ZRevRangeWithScores("hotdoc", 0, NDoc).Result()
	for _, docid := range hotdoc {
		docrlt = append(docrlt, SearchOneRltToDoc(docid.Member.(int)))
	}
	return docrlt
}

func SearchTopNKeyword(NDoc int64) []string {
	var hotkeywords []string
	hotkeyword, _ := redisdb.ZRevRangeWithScores("hotkeyword", 0, NDoc).Result()
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
		redisdb.ZIncrBy("hotdoc", 1, "doc:"+strconv.Itoa(docid)).Result()
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
func Search(text string, maxnumofrlt int, jbfc *gojieba.Jieba) []SearchRlt {
	var keysearchrltp Keyword
	var srlt SearchRltC
	keyword_ifidf := make(map[int]float32)
	words := CutWords(text, jbfc)
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
			redisdb.ZIncrBy("hotkeyword", 1, "keyword:"+value).Result()
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

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin") //请求头部
		if origin != "" {
			//接收客户端发送的origin （重要！）
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			//服务器支持的所有跨域请求的方法
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE,UPDATE")
			//允许跨域设置可以返回其他子段，可以自定义字段
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Length, X-CSRF-Token, Token,session,X_Requested_With,Accept, Origin, Host, Connection, Accept-Encoding, Accept-Language,DNT, X-CustomHeader, Keep-Alive, User-Agent, X-Requested-With, If-Modified-Since, Cache-Control, Content-Type, Pragma")
			// 允许浏览器（客户端）可以解析的头部 （重要）
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers")
			//设置缓存时间
			c.Header("Access-Control-Max-Age", "172800")
			//允许客户端传递校验信息比如 cookie (重要)
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		//允许类型校验
		if method == "OPTIONS" {
			c.JSON(http.StatusOK, "ok!")
		}

		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic info is: %v", err)
			}
		}()
		c.Next()
	}
}

//读取数据集，倒排索引，写入数据库
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
