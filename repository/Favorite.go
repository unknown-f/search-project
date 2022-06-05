package repository

import (
	"fmt"
	"searchproject/utils/errmsg"

	"gorm.io/gorm"
)

type Favorite struct {
	ID       uint   `gorm:"primary_key;auto_increment" json:"id"`
	Name     string `gorm:"type:varchar(50);not null" json:"name"`
	Username string `gorm:"varchar(20);not null" json:"username"`
}

// 查询收藏夹是否存在
func CheckFavorite(name string, username string) int {
	var favo Favorite
	favo.ID = 0
	db.Select("id").Where("name = ? and username = ?", name, username).First(&favo)
	if favo.ID > 0 {
		fmt.Printf("Favorite名字已存在， ID: %d\n", favo.ID)
		return errmsg.ERROR_FAVORITENAME_USED
	}
	return errmsg.SUCCESS
}

// 添加收藏夹
func CreateFavorite(data *Favorite) int {
	err := db.Create(&data).Error
	if err != nil {
		return errmsg.ERROR
	}
	return errmsg.SUCCESS
}

// 查询收藏夹列表
func GetFavorites(username string) []Favorite {
	var favos []Favorite
	//err = db.Limit(pageSize).Offset((pageNum - 1) * pageSize).Find(&cates).Error
	err = db.Where("username = ?", username).Find(&favos).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil
	}
	return favos
}

// 编辑收藏夹信息
func EditFavorite(oldname string, username string, data *Favorite) int {
	var favo Favorite
	favo.ID = 0
	db.Select("id").Where("name = ? and username = ?", oldname, username).First(&favo)
	if favo.ID == 0 {
		fmt.Printf("未找到目标Favorite\n")
		return errmsg.ERROR
	}
	err = db.Model(&favo).Where("id = ?", favo.ID).Update("name", data.Name).Error
	if err != nil {
		return errmsg.ERROR
	}
	return errmsg.SUCCESS
}

// 删除收藏夹
func DeleteFavorite(name string, username string) int {
	var favo Favorite
	err = db.Where("name = ? and username = ?", name, username).Delete(&favo).Error
	if err != nil {
		return errmsg.ERROR
	}
	return errmsg.SUCCESS
}
