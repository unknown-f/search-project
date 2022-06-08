package controller

import (
	"fmt"
	"searchproject/repository"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

var mgomutex sync.Mutex

func Search(c *gin.Context) {
	var doc []repository.SearchRlt
	var relatedinfo []string
	var data repository.SearchRespond
	c.BindJSON(&data)
	searchinfo := strings.Split(data.SearchText, " -")
	fmt.Println("SSS", searchinfo)
	if len(searchinfo) > 1 {
		doc, relatedinfo = repository.Search(searchinfo[0], searchinfo[1:], 20, 5)
	} else {
		doc, relatedinfo = repository.Search(searchinfo[0], []string{}, 20, 5)
	}
	srlttest := repository.SearchRltToDoc(doc)
	data.RelatedInfo = relatedinfo
	data.ReturnRes = srlttest
	fmt.Println("return:", data)
	c.JSON(200, data)
}

func SearchTopNDoc(c *gin.Context) {
	num, err := strconv.ParseInt(c.Param("num"), 10, 64)
	if err != nil {
		c.JSON(406, "请求的数量不是整数")
	} else {
		c.JSON(200, repository.SearchTopNDoc(num))
	}
}

func AddNewTextDoc(c *gin.Context) {
	newtest := make(map[string]interface{})
	c.BindJSON(&newtest)
	mgomutex.Lock()
	err := repository.CutAndWriteOnce(newtest["text"].(string))
	fmt.Println(newtest["text"].(string))
	mgomutex.Unlock()
	if err != nil {
		c.JSON(500, err.Error())
	} else {
		c.JSON(200, "Ok")
	}
}

func SearchTopNKeyword(c *gin.Context) {
	num, err := strconv.ParseInt(c.Param("num"), 10, 64)
	if err != nil {
		c.JSON(406, "请求的数量不是整数")
	} else {
		c.JSON(200, repository.SearchTopNKeyword(num))
	}
}
