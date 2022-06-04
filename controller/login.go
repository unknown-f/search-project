package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"searchproject/middleware"
	"searchproject/repository"
	"searchproject/utils/errmsg"
)

func Login(c *gin.Context) {
	var data repository.User
	c.ShouldBindJSON(&data)

	var token string
	var code int
	code = repository.CheckLogin(data.Username, data.Password)

	if code == errmsg.SUCCESS {
		token, code = middleware.SetToken(data.Username)
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  code,
		"token":   token,
		"message": errmsg.GetErrMsg(code),
	})
}
