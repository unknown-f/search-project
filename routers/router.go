package routers

import (
	"github.com/gin-gonic/gin"
	"searchproject/controller"
	"searchproject/middleware"
	"searchproject/utils"
)

func InitRouter(r *gin.Engine) {
	gin.SetMode(utils.AppMode)
	//r := gin.Default()

	// 使用 jwt 验证
	authUserRouter := r.Group("user")
	authUserRouter.Use(middleware.JwtToken())
	{
		authUserRouter.PUT("/:id", controller.EditUser)
		authUserRouter.DELETE("/:id", controller.DeleteUser)
	}

	userRouter := r.Group("user")
	{
		// user模块路由
		userRouter.POST("/add", controller.AddUser)
		userRouter.GET("users", controller.GetUsers)

		userRouter.POST("login", controller.Login)
	}

	authFavoRouter := r.Group("favo")
	authFavoRouter.Use(middleware.JwtToken())
	{
		// favorite模块路由
		authFavoRouter.POST("/add", controller.AddFavorite)
		authFavoRouter.GET("favos", controller.GetFavorites)
		authFavoRouter.PUT("/:id", controller.EditFavorite)
		authFavoRouter.DELETE("/:id", controller.DeleteFavorite)
	}

	authlinkRouter := r.Group("link")
	authlinkRouter.Use(middleware.JwtToken())
	{
		// link 模块路由
		authlinkRouter.POST("/add", controller.AddLink)
		authlinkRouter.GET("/list/:id", controller.GetLinkByFavo)
		authlinkRouter.GET("/info/:id", controller.GetLinkInfo)
		authlinkRouter.GET("/links", controller.GetLinks)
		authlinkRouter.PUT("/:id", controller.EditLink)
		authlinkRouter.DELETE("/:id", controller.DeleteLink)
	}

	//err := r.Run(utils.HTTPPort)
	//if err != nil {
	//	return
	//}
}
