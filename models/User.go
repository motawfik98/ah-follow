package models

import (
	"github.com/jinzhu/gorm"
)

type User struct {
	gorm.Model
	Username      string `form:"username" gorm:"unique_index"`
	Password      string `form:"password"`
	Hash          string
	Order         int
	Admin         bool            `gorm:"default:0;not null"`
	Tasks         []*UserTask     `gorm:"PRELOAD:false"`
	Subscriptions []*Subscription `gorm:"PRELOAD:false"`
}

func (user *User) AfterCreate(scope *gorm.Scope) error {
	ID := int(user.ID)
	admin := false
	if ID == 1 {
		admin = true
	}
	hash := generateHash(ID)
	scope.DB().Model(user).Updates(User{Hash: hash, Order: ID, Admin: admin})
	return nil
}

func GetAllUsernames(db *gorm.DB) []string {
	var usernames []string
	db.Table("users").Order("[order] ASC").Pluck("username", &usernames)
	return usernames
}
