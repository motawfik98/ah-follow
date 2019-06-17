package models

import "github.com/jinzhu/gorm"

// this struct stores the data needed to apply the `webpush` package and API
type Subscription struct {
	gorm.Model
	Hash           string
	Endpoint       string `json:"endpoint" gorm:"UNIQUE_INDEX"`
	Auth           string `json:"auth"`
	P256dh         string `json:"p256dh"`
	UserID         uint   `json:"user_id"`
	Classification int    `json:"classification"`
}

// this function generates the hash then update the Subscription created
func (subscription *Subscription) AfterCreate(scope *gorm.Scope) error {
	ID := int(subscription.ID)
	hash := generateHash(ID)
	scope.DB().Model(subscription).Updates(Person{Hash: hash})
	return nil
}
