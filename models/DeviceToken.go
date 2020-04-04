package models

import "github.com/jinzhu/gorm"

type DeviceToken struct {
	gorm.Model
	Hash   string
	Token  string `json:"token" gorm:"UNIQUE_INDEX"`
	UserID uint   `json:"user_id"`
}

// this function generates the hash then update the Subscription created
func (deviceToken *DeviceToken) AfterCreate(scope *gorm.Scope) error {
	ID := int(deviceToken.ID)
	hash := generateHash(ID)
	scope.DB().Model(deviceToken).Updates(Person{Hash: hash})
	return nil
}
