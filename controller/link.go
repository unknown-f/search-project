package controller

import (
	"fmt"
	"net/http"
	"searchproject/repository"
	"searchproject/utils/errmsg"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 添加链接
func AddLink(c *gin.Context) {
	var data repository.Link
	// 绑定数据模型
	c.ShouldBindJSON(&data)
	favoriteid, _ := strconv.Atoi(c.Param("favoriteid"))
	// 获取 username
	username, _ := c.MustGet("username").(string)
	data.Favoriteid = favoriteid
	data.Username = username
	// 判断文章标题是否存在
	code := repository.CheckLink(data.Favoriteid, data.Title)
	fmt.Println("?????", data)
	if code == errmsg.SUCCESS {
		repository.CreateLink(&data)
	}
	if code == errmsg.ERROR_LINKNAME_USED {
		code = errmsg.ERROR_LINKNAME_USED
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  code,
		"data":    data,
		"message": errmsg.GetErrMsg(code),
	})
}

// 根据收藏夹获取链接
func GetLinkByFavo(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("favoriteid"))
	data := repository.GetLinkByFavorite(id)
	code := errmsg.SUCCESS

	c.JSON(http.StatusOK, gin.H{
		"status":  code,
		"data":    data,
		"message": errmsg.GetErrMsg(code),
	})
}

// 查询链接列表
func GetLinks(c *gin.Context) {
	// 获取 username
	username, _ := c.MustGet("username").(string)
	data := repository.GetLinks(username)
	code := errmsg.SUCCESS

	c.JSON(http.StatusOK, gin.H{
		"status":  code,
		"data":    data,
		"message": errmsg.GetErrMsg(code),
	})
}

// 删除链接
func DeleteLink(c *gin.Context) {
	title := c.Param("title")
	// 获取 username
	username, _ := c.MustGet("username").(string)
	code := repository.DeleteLink(title, username)

	c.JSON(http.StatusOK, gin.H{
		"status":  code,
		"message": errmsg.GetErrMsg(code),
	})
}
