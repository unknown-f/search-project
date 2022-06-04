package main

import (
	"searchproject/repository"
	"searchproject/routers"
)

//这两个表用哈希索引应该比较快，这个需要后面对比一下

func main() {
	// 引用数据库
	repository.InitJieba()
	err := repository.InitDB()
	if err != nil {
		panic(err)
	}
	// 用户登录、注册、收藏夹等部分的路由
	routers.InitRouter()
	if err != nil {
		panic(err)
	}
}
