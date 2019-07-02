package models

import (
	"github.com/jinzhu/gorm"
)

type OTP struct {
	gorm.Model
	UserID           uint
	PhoneNumber      string
	VerificationCode string
	Used             bool `gorm:"default:0"`
}
