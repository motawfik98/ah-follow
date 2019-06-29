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
	databaseError := db.Create(&personTask).GetErrors()
	if len(databaseError) > 0 {
		db.Table("following_user_tasks").Where("task_id = ? AND user_id = ?", taskID, userID).Update("deleted_at", nil)
	}
}

func CreateWorkingOnUserTask(db *gorm.DB, taskID, userID uint, action string, finalResponse bool, followerID uint) uint {
	personTask := WorkingOnUserTask{
		UserTask:      UserTask{TaskID: taskID, UserID: userID},
		ActionTaken:   action,
		FinalResponse: finalResponse,
		FollowerID:    followerID,
	}
	databaseErrors := db.Create(&personTask).GetErrors()
	if len(databaseErrors) > 0 {
		db.Table("working_on_user_tasks").Where("task_id = ? AND user_id = ?", taskID, userID).
			Updates(map[string]interface{}{"deleted_at": nil, "action_taken": action, "final_response": finalResponse, "follower_id": followerID})
	}
	return personTask.UserID
}
