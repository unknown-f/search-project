package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"searchproject/repository"
	"searchproject/utils/errmsg"
	"strconv"
)

// 添加链接
func AddLink(c *gin.Context) {
	var data repository.Link
	// 绑定数据模型
	c.ShouldBindJSON(&data)
	// 获取 username
	username, _ := c.MustGet("username").(string)
	// 判断文章标题是否存在
	code := repository.CheckLink(data.Title, username)
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
	id, _ := strconv.Atoi(c.Param("id"))

	// 获取 username
	username, _ := c.MustGet("username").(string)
	data := repository.GetLinkByFavorite(id, username)
	code := errmsg.SUCCESS

	c.JSON(http.StatusOK, gin.H{
		"status":  code,
		"data":    data,
		"message": errmsg.GetErrMsg(code),
	})
}

// 查询单个链接
func GetLinkInfo(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	// 获取 username
	username, _ := c.MustGet("username").(string)
	data := repository.GetLinkInfo(id, username)
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

// 编辑链接
func EditLink(c *gin.Context) {
	var data repository.Link
	id, _ := strconv.Atoi(c.Param("id"))
	// 获取 username
	username, _ := c.MustGet("username").(string)
	c.ShouldBindJSON(&data)
	code := repository.CheckLink(data.Title, username)
	if code == errmsg.SUCCESS {
		repository.EditLink(id, username, &data)
	}
	if code == errmsg.ERROR_LINKNAME_USED {
		c.Abort()
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  code,
		"data":    data,
		"message": errmsg.GetErrMsg(code),
	})
}

// 删除链接
func DeleteLink(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	// 获取 username
	username, _ := c.MustGet("username").(string)
	code := repository.DeleteLink(id, username)

	c.JSON(http.StatusOK, gin.H{
		"status":  code,
		"message": errmsg.GetErrMsg(code),
	})
}
