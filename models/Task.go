package models

import (
	"github.com/jinzhu/gorm"
)

type Task struct {
	//DT_RowId	string `json:"DT_RowId"`
	gorm.Model
	Description string `gorm:"not null" form:"description" json:"description"`
	SentTo      string `gorm:"not null" form:"sent_to" json:"sent_to"`
	FollowedBy  string `form:"followed_by" json:"followed_by"`
	ActionTaken string `form:"action_taken" json:"action_taken"`
	Hash        string
}

func (task *Task) AfterCreate(scope *gorm.Scope) error {
	ID := int(task.ID)
	hash := generateHash(ID)
	scope.DB().Model(task).Updates(Task{Hash: hash})
	return nil
}

func GetAllTasks(db *gorm.DB, offset int, limit int, sortedColumn string, direction string) ([]Task, int) {
	var tasks []Task
	db.Offset(offset).Limit(limit).Order(sortedColumn + " " + direction).Find(&tasks)
	var totalNumberOfRows int
	db.Model(&Task{}).Count(&totalNumberOfRows)
	return tasks, totalNumberOfRows
}
