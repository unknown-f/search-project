package repository

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"searchproject/utils"
	"time"
)

var db *gorm.DB
var err error

func InitDB() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		utils.DbUser,
		utils.DbPassword,
		utils.DbHost,
		utils.DbPort,
		utils.DbName,
	)
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		fmt.Println("连接数据库失败：", err)
	}

	err := db.AutoMigrate(&User{}, &Favorite{}, &Link{})
	if err != nil {
		return
	}

	DB, _ := db.DB()
	DB.SetMaxIdleConns(10)
	DB.SetMaxOpenConns(200)
	DB.SetConnMaxLifetime(100 * time.Second)
}
