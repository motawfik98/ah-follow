package models

import "github.com/jinzhu/gorm"

type UserTask struct {
	TaskID      uint `gorm:"primary_key;auto_increment:false;type:int" json:"task_id"`
	Task        *Task
	UserID      uint   `gorm:"primary_key;auto_increment:false;type:int" json:"user_id"`
	User        *User  `json:"user"`
	ActionTaken string `gorm:"task_id" json:"action_taken"`
}

func CreateUserTask(db *gorm.DB, taskID, userID uint) {
	personTask := UserTask{
		TaskID: taskID,
		UserID: userID,
	}
	db.Create(&personTask)
}