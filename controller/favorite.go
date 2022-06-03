package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"searchproject/repository"
	"searchproject/utils/errmsg"
	"strconv"
)

// 添加分类
func AddFavorite(c *gin.Context) {
	var data repository.Favorite
	_ = c.ShouldBindJSON(&data)
	data.Username, _ = c.MustGet("username").(string)
	code := repository.CheckFavorite(data.Name)
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

// 查询分类列表
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

// 编辑分类
func EditFavorite(c *gin.Context) {
	var data repository.Favorite
	id, _ := strconv.Atoi(c.Param("id"))
	c.ShouldBindJSON(&data)
	username, _ := c.MustGet("username").(string)
	code := repository.CheckFavorite(data.Name)
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

// 删除分类
func DeleteFavorite(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	username, _ := c.MustGet("username").(string)
	code := repository.DeleteFavorite(id, username)

	c.JSON(http.StatusOK, gin.H{
		"ststus":  code,
		"message": errmsg.GetErrMsg(code),
	})
}
