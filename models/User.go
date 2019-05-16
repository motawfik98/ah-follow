package models

import (
	"github.com/jinzhu/gorm"
)

type User struct {
	gorm.Model
	Username string `form:"username"`
	Password string `form:"password"`
	Hash     string
	Name     string `json:"name"`
	Admin    bool
	Tasks    []*PersonTask `gorm:"PRELOAD:false"`
}

func (user *User) AfterCreate(scope *gorm.Scope) error {
	ID := int(user.ID)
	hash := generateHash(ID)
	scope.DB().Model(user).Update("Hash", hash)
	return nil
}
