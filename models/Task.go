package models

import (
	"github.com/jinzhu/gorm"
)

type Task struct {
	gorm.Model
	Description string `gorm:"not null" form:"description" json:"description"`
	FollowedBy  string `form:"followed_by" json:"followed_by"`
	People      []Person
	Hash        string
}

func (task *Task) AfterCreate(scope *gorm.Scope) error {
	ID := int(task.ID)
	hash := generateHash(ID)
	scope.DB().Model(task).Updates(Task{Hash: hash})
	return nil
}

func GetAllTasks(db *gorm.DB, offset int, limit int, sortedColumn string, direction string,
	descriptionSearch string, followedBySearch string, minDateSearch string, maxDateSearch string) ([]Task, int, int) {
	var tasks []Task
	if descriptionSearch != "" {
		descriptionSearch = "%" + descriptionSearch + "%"
		db = db.Where("description LIKE ?", descriptionSearch)
	}
	if followedBySearch != "" {
		followedBySearch = "%" + followedBySearch + "%"
		db = db.Where("followed_by LIKE ?", followedBySearch)
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
	var totalNumberOfRowsInDatabase int
	db.Model(&Task{}).Count(&totalNumberOfRowsInDatabase)
	return tasks, totalNumberOfRowsInDatabase, totalNumberOfRowsAfterFilter
}
