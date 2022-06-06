package controller

import (
	"fmt"
	"searchproject/repository"
	"strconv"

	"github.com/gin-gonic/gin"
)

func Search(c *gin.Context) {
	time := c.Param("time")
	text := c.Param("text")
	doc, relatedinfo := repository.Search(text, 20, 5)
	srlttest := repository.SearchRltToDoc(doc)
	data := repository.SearchRespond{
		SearchTime:  time,
		SearchText:  text,
		ReturnRes:   srlttest,
		RelatedInfo: relatedinfo,
	}
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
	err := repository.CutAndWriteOnce(newtest["test"].(string))
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
