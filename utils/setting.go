package utils

import (
	"fmt"
	"gopkg.in/ini.v1"
)

var (
	AppMode  string
	HTTPPort string
	JwtKey   string

	DbType     string
	DbHost     string
	DbPort     string
	DbUser     string
	DbPassword string
	DbName     string
)

func init() {
	file, err := ini.Load("config/config.ini")
	if err != nil {
		fmt.Println("未读取到配置文件，检查文件路径or文件是否存在：", err)
	}
	LoadServer(file)
	LoadDb(file)
}

func LoadServer(file *ini.File) {
	AppMode = file.Section("server").Key("AppMode").MustString("debug")
	HTTPPort = file.Section("server").Key("HTTPPort").MustString(":8080")
	JwtKey = file.Section("server").Key("jwtKey").MustString("qaqaqakkkkkk")
}

func LoadDb(file *ini.File) {
	DbType = file.Section("database").Key("DbType").MustString("mysql")
	DbHost = file.Section("database").Key("DbHost").MustString("127.0.0.1")
	DbPort = file.Section("database").Key("DbPort").MustString("3306")
	DbUser = file.Section("database").Key("DbUser").MustString("root")
	DbPassword = file.Section("database").Key("DbPassword").MustString("236242")
	DbName = file.Section("database").Key("DbName").MustString("search-project")
}
