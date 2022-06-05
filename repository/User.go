package repository

import (
	"encoding/base64"
	"fmt"
	"log"
	"searchproject/utils/errmsg"

	"golang.org/x/crypto/scrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `gorm:"varchar(20);not null" json:"username"`
	Password string `gorm:"varchar(20);not null" json:"password"`
	Role     int    `gorm:"type:int" json:"role"`
}

// 查询用户是否存在
func CheckUser(name string) (code int) {
	var users User
	users.ID = 0
	db.Select("id").Where("username = ?", name).First(&users)
	fmt.Println(users)
	if users.ID > 0 {
		return errmsg.ERROR_USERNAME_USED
	}
	return errmsg.SUCCESS
}

// 新增用户
func CreateUser(data *User) int {
	// 密码加密
	data.Password = ScryptPw(data.Password)
	// 写入数据库
	err := db.Create(&data).Error
	if err != nil {
		return errmsg.ERROR
	}
	return errmsg.SUCCESS
}

// 查询用户列表
func GetUsers() []User {
	var users []User
	err := db.Find(&users).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil
	}
	return users
}

// 编辑用户
func EditUser(id int, data *User) int {
	var user User
	err := db.Model(&user).Where("id = ?", id).Update("username", data.Username).Error
	if err != nil {
		return errmsg.ERROR
	}

	err = db.Model(&user).Update("role", data.Role).Error
	if err != nil {
		return errmsg.ERROR
	}
	return errmsg.SUCCESS
}

// 删除用户
func DeleteUser(Username string) int {
	var user User
	err := db.Where("Username = ?", Username).Delete(&user).Error
	if err != nil {
		return errmsg.ERROR
	}
	return errmsg.SUCCESS
}

// 密码加密
func ScryptPw(passwd string) string {
	const KeyLen = 10
	salt := make([]byte, 8)
	salt = []byte{12, 32, 4, 6, 66, 22, 222, 11}
	HashPw, err := scrypt.Key([]byte(passwd), salt, 16384, 8, 1, KeyLen)
	if err != nil {
		log.Fatal(err)
	}
	fpw := base64.StdEncoding.EncodeToString(HashPw)
	return fpw
}

// 登录验证
func CheckLogin(username, password string) int {
	var user User
	db.Where("username = ?", username).First(&user)

	if user.ID == 0 {
		return errmsg.ERROR_USER_NOT_EXIST
	}

	if ScryptPw(password) != user.Password {
		return errmsg.ERROR_PASSWORD_WRONG
	}
	//TODO 还需要认证是否同一用户
	return errmsg.SUCCESS
}
