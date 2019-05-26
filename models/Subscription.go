package models

import "github.com/jinzhu/gorm"

type Subscription struct {
	gorm.Model
	Hash     string
	Endpoint string `json:"endpoint" gorm:"UNIQUE_INDEX"`
	Auth     string `json:"auth"`
	P256dh   string `json:"p256dh"`
	UserID   uint   `json:"user_id"`
}

func (subscription *Subscription) AfterCreate(scope *gorm.Scope) error {
	ID := int(subscription.ID)
	hash := generateHash(ID)
	scope.DB().Model(subscription).Updates(Person{Hash: hash})
	return nil
}
