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
	srlttest := repository.SearchRltToDoc(repository.Search(text, 20))
	data := repository.SearchRespond{
		SearchTime: time,
		SearchText: text,
		ReturnRes:  srlttest,
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

func SearchTopNKeyword(c *gin.Context) {
	num, err := strconv.ParseInt(c.Param("num"), 10, 64)
	if err != nil {
		c.JSON(406, "请求的数量不是整数")
	} else {
		c.JSON(200, repository.SearchTopNKeyword(num))
	}
}
