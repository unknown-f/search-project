package controller

import (
	"net/http"
	"searchproject/repository"
	"searchproject/utils/errmsg"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 添加收藏夹
func AddFavorite(c *gin.Context) {
	var data repository.Favorite
	_ = c.ShouldBindJSON(&data)
	data.Username, _ = c.MustGet("username").(string)
	code := repository.CheckFavorite(data.Name, data.Username)
	if code == errmsg.SUCCESS {
		repository.CreateFavorite(&data)
	}
	if code == errmsg.ERROR_FAVORITENAME_USED {
		code = errmsg.ERROR_FAVORITENAME_USED
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  code,
		"data":    data,
		"message": errmsg.GetErrMsg(code),
	})
}

// 查询收藏夹列表
func GetFavorites(c *gin.Context) {
	username, _ := c.MustGet("username").(string)
	data := repository.GetFavorites(username)
	code := errmsg.SUCCESS

	c.JSON(http.StatusOK, gin.H{
		"status":  code,
		"data":    data,
		"message": errmsg.GetErrMsg(code),
	})

}

// 编辑收藏夹
func EditFavorite(c *gin.Context) {
	var data repository.Favorite
	id, _ := strconv.Atoi(c.Param("id"))
	c.ShouldBindJSON(&data)
	username, _ := c.MustGet("username").(string)
	code := repository.CheckFavorite(data.Name, username)
	if code == errmsg.SUCCESS {
		repository.EditFavorite(id, username, &data)
	}
	if code == errmsg.ERROR_FAVORITENAME_USED {
		c.Abort()
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  code,
		"data":    data,
		"message": errmsg.GetErrMsg(code),
	})

}

// 删除收藏夹
func DeleteFavorite(c *gin.Context) {
	//id, _ := strconv.Atoi(c.Param("id"))
	name := c.Param("name")
	username, _ := c.MustGet("username").(string)
	code := repository.DeleteFavorite(name, username)

	c.JSON(http.StatusOK, gin.H{
		"status":  code,
		"message": errmsg.GetErrMsg(code),
	})
}
