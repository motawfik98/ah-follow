package models

import (
	"github.com/jinzhu/gorm"
)

type User struct {
	gorm.Model
	Username           string `form:"username" gorm:"unique_index" json:"username"`
	Password           string `form:"password"`
	Hash               string
	Order              int
	Classification     int                  `gorm:"NOT NULL" json:"classification"`
	FollowingUserTasks []*FollowingUserTask `gorm:"PRELOAD:false"`
	WorkingOnUserTasks []*WorkingOnUserTask `gorm:"PRELOAD:false"`
	Subscriptions      []*Subscription      `gorm:"PRELOAD:false"`
	PhoneNumber        string
	ValidPhoneNumber   bool `gorm:"default:0"`
	PhoneNotifications bool `gorm:"default:0"`
	Email              string
	ValidEmail         bool `gorm:"default:0"`
	EmailNotifications bool `gorm:"default:0"`
}

// this function updates the Hash and Admin column of the user after create
func (user *User) AfterCreate(scope *gorm.Scope) error {
	ID := int(user.ID)
	hash := generateHash(ID)
	scope.DB().Model(user).Updates(User{Hash: hash, Order: ID})
	return nil
}

// this function gets all the users that are in the database
func GetAllUsers(db *gorm.DB) []User {
	var users []User
	db.Table("users").Select("username, classification").
		Order("[order] ASC").Scan(&users)
	return users
}
