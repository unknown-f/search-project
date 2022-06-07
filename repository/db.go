package repository

import (
	"database/sql"
	"fmt"
	"searchproject/utils"
	"time"

	"github.com/go-redis/redis"
	"gopkg.in/mgo.v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var db *gorm.DB
var redisdb *redis.Client
var c_indextodoc *mgo.Collection //docid与doc的对应关系集合
var c_keytoindx *mgo.Collection  //关键词与docid的对应关系集合
var latestDocid int

func InitDB() error {
	err := InitMysql()
	if err != nil {
		return err
	}
	err = InitRedis()
	if err != nil {
		return err
	}
	err = InitMongodb()
	if err != nil {
		return err
	}
	return nil
}

func InitRedis() error {
	redisdb = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379", // 指定
		Password: "",
		DB:       0, // redis一共16个库，指定其中一个库即可
	})
	_, err := redisdb.Ping().Result()
	if err != nil {
		return err
	}
	return nil
}

func InitMongodb() error {
	session, err := mgo.Dial("localhost:27017") //连接数据库
	if err != nil {
		return err
	}
	session.SetMode(mgo.Monotonic, true)
	mgodb := session.DB("search_project")
	c_indextodoc = mgodb.C("indextosource")
	c_keytoindx = mgodb.C("keytoindex")
	var lastestdoc Doc
	err = c_indextodoc.Find(nil).Sort("-ID").One(&lastestdoc)
	if err != nil {
		latestDocid = 0
	} else {
		latestDocid = lastestdoc.ID
	}
	latestDocid = lastestdoc.ID
	fmt.Println("latestDocid", latestDocid)
	return nil
}

func InitMysql() error {
	var DB *sql.DB
	var err error
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
		return err
	}
	err = db.AutoMigrate(&User{}, &Favorite{}, &Link{})
	if err != nil {
		return err
	}
	DB, err = db.DB()
	if err != nil {
		return err
	}
	DB.SetMaxIdleConns(10)
	DB.SetMaxOpenConns(200)
	DB.SetConnMaxLifetime(100 * time.Second)
	return nil
}
