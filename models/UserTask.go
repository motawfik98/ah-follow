package models

import "github.com/jinzhu/gorm"

type UserTask struct {
	gorm.Model
	TaskID         uint `gorm:"unique_index:idx_user_task" json:"task_id"`
	Task           *Task
	UserID         uint  `gorm:"unique_index:idx_user_task" json:"user_id"`
	User           *User `json:"user"`
	Seen           bool  `json:"seen" gorm:"default:0;not null"`
	MarkedAsUnseen bool  `json:"marked_as_unseen" gorm:"default:0; not null"`
}

type FollowingUserTask struct {
	UserTask
}

type WorkingOnUserTask struct {
	UserTask
	ActionTaken   string `json:"action_taken"`
	FinalResponse bool   `json:"final_response" gorm:"default:0"`
	FollowerID    uint   `json:"follower_id"`
	Notes         string `json:"notes"`
}

func CreateFollowingUserTask(db *gorm.DB, taskID, userID uint) {
	personTask := FollowingUserTask{
		UserTask: UserTask{TaskID: taskID, UserID: userID},
	}
	db.Create(&personTask)
}

func CreateWorkingOnUserTask(db *gorm.DB, taskID, userID uint, action string, finalResponse bool, followerID uint) uint {
	personTask := WorkingOnUserTask{
		UserTask:      UserTask{TaskID: taskID, UserID: userID},
		ActionTaken:   action,
		FinalResponse: finalResponse,
		FollowerID:    followerID,
	}
	db.Create(&personTask)
	return personTask.UserID
}
