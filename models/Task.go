package models

import (
	"database/sql"
	"github.com/jinzhu/gorm"
)

type Task struct {
	gorm.Model
	Description string         `gorm:"not null" form:"description" json:"description"`
	Users       []*UserTask    `gorm:"PRELOAD:false" json:"users"`
	People      []Person       `gorm:"PRELOAD:false" json:"people"`
	FinalAction sql.NullString `json:"final_action" gorm:"default: null"`
	Seen        bool           `gorm:"default:1;not null" json:"seen"`
	Hash        string
}

func (task *Task) AfterCreate(scope *gorm.Scope) error {
	ID := int(task.ID)
	hash := generateHash(ID)
	scope.DB().Model(task).Updates(Task{Hash: hash})
	return nil
}

func (task *Task) DeleteChildren(db *gorm.DB) {
	for i := 0; i < len(task.Users); i++ {
		db.Delete(task.Users[i])
	}
}

func GetAllTasks(db *gorm.DB, offset int, limit int, sortedColumn, direction,
	descriptionSearch, sentToSearch, minDateSearch, maxDateSearch, retrieveType string, admin bool, userID uint) ([]Task, int, int) {
	var tasks []Task
	var totalNumberOfRowsInDatabase int
	if !admin {
		db = db.Table("tasks").Joins("JOIN user_tasks ON user_tasks.task_id = tasks.id")
		db = db.Where("user_tasks.user_id = ?", userID)
	}
	db.Model(&Task{}).Count(&totalNumberOfRowsInDatabase)
	if retrieveType == "replied" {
		db = db.Where("final_action IS NOT NULL")
	} else if retrieveType == "nonReplied" {
		db = db.Where("final_action IS NULL")
	} else if retrieveType == "unseen" {
		if admin {
			db = db.Where("seen = 0 AND final_action IS NOT NULL")
		} else {
			db = db.Where("user_tasks.seen = 0")
		}
	} else if retrieveType == "seen" {
		if admin {
			db = db.Where("seen = 1 AND final_action IS NOT NULL")
		} else {
			db = db.Where("user_tasks.seen = 1")
		}
	} else if retrieveType == "notRepliedByAll" {
		var ids []int
		db.Table("tasks").Select("tasks.id").
			Joins("JOIN people ON tasks.id = people.task_id").
			Where("final_response = 0").
			Group("tasks.id").Pluck("tasks.id", &ids)
		db = db.Where("tasks.id IN (?)", ids)
	}
	db = db.Preload("Users").Preload("People")
	if descriptionSearch != "" {
		descriptionSearch = "%" + descriptionSearch + "%"
		db = db.Where("description LIKE ?", descriptionSearch)
	}
	if sentToSearch != "" {
		sentToSearch = "%" + sentToSearch + "%"
		db = db.Table("tasks").Joins("JOIN people ON tasks.id = people.task_id")
		db = db.Where("people.name LIKE ?", sentToSearch)
	}
	if minDateSearch != "" {
		minDateSearch = minDateSearch + " 00:00:00.0000000 +02:00"
		db = db.Where("created_at >= ?", minDateSearch)
	}
	if maxDateSearch != "" {
		maxDateSearch = maxDateSearch + " 00:00:00.0000000 +02:00"
		db = db.Where("created_at <= ?", maxDateSearch)
	}
	var totalNumberOfRowsAfterFilter int
	db.Find(&tasks).Count(&totalNumberOfRowsAfterFilter)
	db.Offset(offset).Limit(limit).Order(sortedColumn + " " + direction).Find(&tasks)
	return tasks, totalNumberOfRowsInDatabase, totalNumberOfRowsAfterFilter
}
