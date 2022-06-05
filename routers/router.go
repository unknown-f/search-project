package routers

import (
	"log"
	"net/http"
	"searchproject/controller"
	"searchproject/middleware"
	"searchproject/utils"

	"github.com/gin-gonic/gin"
)

func InitRouter() error {

	gin.SetMode(utils.AppMode)
	r := gin.Default()
	//r := gin.Default()

	r.Use(Cors())
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
		authFavoRouter.DELETE("/:name", controller.DeleteFavorite)
	}

	authlinkRouter := r.Group("link")
	authlinkRouter.Use(middleware.JwtToken())
	{
		// link 模块路由
		authlinkRouter.POST("/add", controller.AddLink)
		authlinkRouter.GET("/list/:favoritename", controller.GetLinkByFavo)
		authlinkRouter.GET("/info/:id", controller.GetLinkInfo)
		authlinkRouter.GET("/links", controller.GetLinks)
		authlinkRouter.PUT("/:id", controller.EditLink)
		authlinkRouter.DELETE("/:title", controller.DeleteLink)
	}

	searchRouter := r.Group("search")
	{
		//search 搜索路由
		searchRouter.POST("/hotdoc", controller.SearchTopNDoc)
		searchRouter.POST("/hotkeyword", controller.SearchTopNKeyword)
		searchRouter.GET("/:time/:text", controller.Search)
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
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin") //请求头部
		if origin != "" {
			//接收客户端发送的origin （重要！）
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			//服务器支持的所有跨域请求的方法
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE,UPDATE")
			//允许跨域设置可以返回其他子段，可以自定义字段
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Length, X-CSRF-Token, Token,session,X_Requested_With,Accept, Origin, Host, Connection, Accept-Encoding, Accept-Language,DNT, X-CustomHeader, Keep-Alive, User-Agent, X-Requested-With, If-Modified-Since, Cache-Control, Content-Type, Pragma")
			// 允许浏览器（客户端）可以解析的头部 （重要）
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers")
			//设置缓存时间
			c.Header("Access-Control-Max-Age", "172800")
			//允许客户端传递校验信息比如 cookie (重要)
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		//允许类型校验
		if method == "OPTIONS" {
			c.JSON(http.StatusOK, "ok!")
		}

		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic info is: %v", err)
			}
		}()
		c.Next()
	}
}
