package routers

import (
	"searchproject/controller"
	"searchproject/middleware"
	"searchproject/utils"

	"github.com/gin-gonic/gin"
)

func InitRouter() error {

	gin.SetMode(utils.AppMode)
	r := gin.Default()
	//r := gin.Default()

	r.Use(middleware.Cors())
	r.LoadHTMLFiles("./view.html")
	r.StaticFile("/logo.png", "./logo.png")

	authUserRouter := r.Group("user")
	authUserRouter.Use(middleware.JwtToken())
	{
		authUserRouter.PUT("/:id", controller.EditUser)
		authUserRouter.DELETE("/:Username", controller.DeleteUser)
	}

	userRouter := r.Group("user")
	{
		// user模块路由
		userRouter.POST("/add", controller.AddUser)
		userRouter.GET("/users", controller.GetUsers)

		userRouter.POST("/login", controller.Login)
	}

	authFavoRouter := r.Group("favo")
	authFavoRouter.Use(middleware.JwtToken())
	{
		// favorite模块路由
		authFavoRouter.POST("/add", controller.AddFavorite)
		authFavoRouter.GET("/favos", controller.GetFavorites)
		authFavoRouter.PUT("/:id", controller.EditFavorite)
		authFavoRouter.DELETE("/:name", controller.DeleteFavorite)
	}

	authlinkRouter := r.Group("link")
	authlinkRouter.Use(middleware.JwtToken())
	{
		// link 模块路由
		authlinkRouter.POST("/add/:favoriteid", controller.AddLink)
		authlinkRouter.GET("/list/:favoriteid", controller.GetLinkByFavo)
		authlinkRouter.GET("/links", controller.GetLinks)
		authlinkRouter.DELETE("/:title", controller.DeleteLink)
	}

	searchRouter := r.Group("search")
	{
		//search 搜索路由
		searchRouter.GET("/hotdoc/:num", controller.SearchTopNDoc)
		searchRouter.GET("/hotkeyword/:num", controller.SearchTopNKeyword)
		searchRouter.GET("/:time/:text", controller.Search)
		searchRouter.POST("/doc", controller.AddNewTextDoc)
	}

	{
		r.GET("/", func(c *gin.Context) {
			c.HTML(200, "view.html", nil)
		})
	}
	err := r.Run()
	if err != nil {
		return err
	}
	return nil
}
