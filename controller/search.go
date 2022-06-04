package controller

import (
	"fmt"
	"searchproject/repository"

	"github.com/gin-gonic/gin"
)

func Search(c *gin.Context) {
	time := c.Param("time")
	text := c.Param("text")
	srlttest := repository.SearchRltToDoc(repository.Search(text, 5))
	data := repository.SearchRespond{
		SearchTime: time,
		SearchText: text,
		ReturnRes:  srlttest,
	}
	fmt.Println("return:", data)
	c.JSON(200, data)
}
