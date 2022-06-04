package repository

import (
	"searchproject/utils/errmsg"

	"gorm.io/gorm"
)

type Link struct {
	gorm.Model
	Title        string `gorm:"type:varchar(100);not null" json:"title"`
	Content      string `gorm:"type:varchar(500)" json:"content"`
	Username     string `gorm:"varchar(20);not null" json:"username"`
	Favoritename string `gorm:"type:varchar(50);not null" json:"favoritename"`
}

// 查询链接是否存在
func CheckLink(title string, username string) int {
	var link Link
	link.ID = 0
	db.Select("id").Where("title = ? and username = ?", title, username).First(&link)
	if link.ID > 0 {
		return errmsg.ERROR_LINKNAME_USED
	}
	return errmsg.SUCCESS
}

// 创建链接
func CreateLink(data *Link) int {
	err = db.Create(&data).Error
	if err != nil {
		return errmsg.ERROR
	}
	return errmsg.SUCCESS
}

// 根据收藏夹获取链接
func GetLinkByFavorite(favoritename string, username string) []Link {
	var linkByFavo []Link
	err := db.Where("favoritename = ? and username = ?", favoritename, username).Find(&linkByFavo).Error

	if err != nil {
		return nil
	}
	return linkByFavo
}

// 获取单个链接
func GetLinkInfo(id int, username string) Link {
	var link Link
	err := db.Where("id = ? and username = ?", id, username).First(&link).Error
	if err != nil {
		return link
	}

	return link
}

// 获取链接列表
func GetLinks(username string) []Link {
	var links []Link
	err = db.Where("username = ?", username).Find(&links).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil
	}
	return links
}

// 编辑链接
func EditLink(id int, username string, data *Link) int {
	var link Link

	err = db.Model(&link).Where("id = ? and username = ?", id, username).Update("title", data.Title).Update("favoritename",
		data.Favoritename).Update("content", data.Content).Error
	if err != nil {
		return errmsg.ERROR
	}
	return errmsg.SUCCESS

}

// 删除链接
func DeleteLink(title string, username string) int {
	var link Link
	err = db.Where("username = ? and title = ?", username, title).Delete(&link).Error
	if err != nil {
		return errmsg.ERROR
	}
	return errmsg.SUCCESS
}
