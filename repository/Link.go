package repository

import (
	"gorm.io/gorm"
	"searchproject/utils/errmsg"
)

type Link struct {
	// 设置 Fid 为外键
	Favorite Favorite `gorm:"foreignkey:Fid"`
	gorm.Model
	Title string `gorm:"type:varchar(100);not null" json:"title"`
	// Fid 将 Link 与 Favorite 关联起来
	Fid      int    `gorm:"type:int;not null" json:"fid"`
	Content  string `gorm:"type:varchar(100)" json:"content"`
	Img      string `gorm:"type:varchar(100)" json:"img"`
	Username string `gorm:"varchar(20);not null" json:"username"`
}

// 查询文章是否存在
func CheckLink(title string, username string) int {
	var link Link
	link.ID = 0
	db.Select("id").Where("title = ? and username = ?", title, username).First(&link)
	if link.ID > 0 {
		return errmsg.ERROR_LINKNAME_USED
	}
	return errmsg.SUCCESS
}

// 创建文章
func CreateLink(data *Link) int {
	err = db.Create(&data).Error
	if err != nil {
		return errmsg.ERROR
	}
	return errmsg.SUCCESS
}

// 根据分类获取文章列表
func GetLinkByFavorite(id int, username string) []Link {
	var linkByFavo []Link
	//err := db.Preload("Category").Limit(pageSize).Offset((pageNum - 1) * pageSize).Where("cid = ?", id).Find(&linkByFavo).Error
	err := db.Preload("Favorite").Where("fid = ? and username = ?", id, username).Find(&linkByFavo).Error

	if err != nil {
		return nil
	}
	return linkByFavo
}

// 获取单个文章
func GetLinkInfo(id int, username string) Link {
	var link Link
	err := db.Preload("Favorite").Where("id = ? and username = ?", id, username).First(&link).Error
	if err != nil {
		return link
	}

	return link
}

// 获取文章列表
func GetLinks(username string) []Link {
	var links []Link
	//err = db.Preload("Category").Limit(pageSize).Offset((pageNum -1 ) * pageSize).Find(&links).Error
	err = db.Preload("Favorite").Where("username = ?", username).Find(&links).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil
	}
	return links
}

// 编辑文章
func EditLink(id int, username string, data *Link) int {
	var link Link

	err = db.Model(&link).Where("id = ? and username = ?", id, username).Update("title", data.Title).Update("fid",
		data.Fid).Update("content", data.Content).Update("img", data.Img).Error
	if err != nil {
		return errmsg.ERROR
	}
	return errmsg.SUCCESS

}

// 删除文章
func DeleteLink(id int, username string) int {
	var link Link
	err = db.Where("id = ? and username = ?", id, username).Delete(&link).Error
	if err != nil {
		return errmsg.ERROR
	}
	return errmsg.SUCCESS
}
