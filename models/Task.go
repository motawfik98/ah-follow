package models

import "github.com/jinzhu/gorm"

type Task struct {
	gorm.Model
	Description string `gorm:"not null" form:"description"`
	SentTo      string `gorm:"not null" form:"sentTo"`
	FollowedBy  string `form: "followedBy"`
	ActionTaken string `form: "actionTaken"`
	Hash        string
}

func (task *Task) AfterCreate(scope *gorm.Scope) error {
	ID := int(task.ID)
	hash := generateHash(ID)
	scope.DB().Model(task).Update("Hash", hash)
	return nil
}
