package models

import (
	"github.com/jinzhu/gorm"
)

type File struct {
	gorm.Model
	Bytes       []byte `json:"bytes" gorm:"type:varbinary(max)"`
	TaskID      uint   `json:"task_id" gorm:"default:null"`
	ContentType string `json:"content_type"`
	Extension   string `json:"extension"`
	FileName    string `json:"file_name"`
	Hash        string `json:"hash"`
	UserID      uint   `json:"user_id"` // to indicate which user uploaded the file
	User        *User
	FileDisplay string `gorm:"-"`
}

func (file *File) AfterCreate(scope *gorm.Scope) error {
	ID := int(file.ID)
	hash := generateHash(ID)
	scope.DB().Model(file).Updates(File{Hash: hash})
	return nil
}
