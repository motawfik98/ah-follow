package models

import (
	"github.com/jinzhu/gorm"
)

type User struct {
	gorm.Model
	Username string `form:"username"`
	Password string `form:"password"`
	Hash     string
	Order    int `gorm:"AUTO_INCREMENT;type:int"`
	Admin    bool
	Tasks    []*UserTask `gorm:"PRELOAD:false"`
}

func (user *User) AfterCreate(scope *gorm.Scope) error {
	ID := int(user.ID)
	hash := generateHash(ID)
	scope.DB().Model(user).Update("Hash", hash)
	return nil
}

func GetAllUsernames(db *gorm.DB) []string {
	var usernames []string
	db.Table("users").Order("[order] ASC").Pluck("username", &usernames)
	return usernames
}
