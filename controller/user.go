package controller

import (
	"net/http"
	"searchproject/repository"
	"searchproject/utils/errmsg"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 查询用户名是否存在
func UserExist(c *gin.Context) {

}

// 添加用户
func AddUser(c *gin.Context) {
	var user repository.User
	// 绑定数据模型
	_ = c.ShouldBindJSON(&user)
	// 判断用户是否存在，不存在则添加
	code := repository.CheckUser(user.Username)
	if code == errmsg.SUCCESS {
		repository.CreateUser(&user)
	}
	if code == errmsg.ERROR_USERNAME_USED {
		code = errmsg.ERROR_USERNAME_USED
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  code,
		"data":    user,
		"message": errmsg.GetErrMsg(code),
	})
}

// 查询用户列表
func GetUsers(c *gin.Context) {
	users := repository.GetUsers()
	code := errmsg.SUCCESS

	c.JSON(http.StatusOK, gin.H{
		"status":  code,
		"data":    users,
		"message": errmsg.GetErrMsg(code),
	})
}

// 编辑用户
func EditUser(c *gin.Context) {
	var user repository.User
	id, _ := strconv.Atoi(c.Param("id"))
	_ = c.ShouldBindJSON(&user)
	code := repository.CheckUser(user.Username)
	if code == errmsg.SUCCESS {
		repository.EditUser(id, &user)
	}
	// 如果用户存在不执行
	if code == errmsg.ERROR_USERNAME_USED {
		c.Abort()
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  code,
		"message": errmsg.GetErrMsg(code),
	})
}

// 删除用户
func DeleteUser(c *gin.Context) {
	Username := c.Param("Username")
	code := repository.DeleteUser(Username)

	c.JSON(http.StatusOK, gin.H{
		"status":  code,
		"message": errmsg.GetErrMsg(code),
	})
}
