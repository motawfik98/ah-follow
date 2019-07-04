package models

import (
	"github.com/jinzhu/gorm"
)

type OTP struct {
	gorm.Model
	UserID           uint
	PhoneNumber      string
	Email            string
	Type             string
	VerificationCode string
	Used             bool `gorm:"default:0"`
}
