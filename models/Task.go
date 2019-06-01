package models

import (
	"database/sql"
	"github.com/jinzhu/gorm"
)

// this struct holds the data needed for the task
type Task struct {
	gorm.Model
	Description string         `gorm:"not null;type:nvarchar(1024)" form:"description" json:"description"`
	Users       []*UserTask    `gorm:"PRELOAD:false" json:"users"`
	People      []Person       `gorm:"PRELOAD:false" json:"people"`
	FinalAction sql.NullString `json:"final_action" gorm:"default: null;type:nvarchar(1024)"`
	Seen        bool           `gorm:"default:1;not null" json:"seen"`
	Hash        string
}

func (task *Task) AfterCreate(scope *gorm.Scope) error {
	ID := int(task.ID)
	hash := generateHash(ID)
	scope.DB().Model(task).Updates(Task{Hash: hash})
	return nil
}

// this function removes all the UserTasks that were assigned to that task
func (task *Task) DeleteChildren(db *gorm.DB) {
	for i := 0; i < len(task.Users); i++ {
		db.Delete(task.Users[i])
	}
}

// this function takes the search parameters (datatables parameters) and return the corresponding data
func GetAllTasks(db *gorm.DB, offset int, limit int, sortedColumn, direction,
	descriptionSearch, sentToSearch, minDateSearch, maxDateSearch, retrieveType string, admin bool, userID uint) ([]Task, int, int) {

	sortedColumn = "tasks." + sortedColumn // set the name of the column that the end user is sorting with
	var tasks []Task
	var totalNumberOfRowsInDatabase int
	if !admin {
		// join the user_tasks table to get only the tasks that were assigned to the logged in user
		db = db.Table("tasks").Joins("JOIN user_tasks ON user_tasks.task_id = tasks.id")
		db = db.Where("user_tasks.user_id = ?", userID)
	}
	db.Model(&Task{}).Count(&totalNumberOfRowsInDatabase) // gets the total number of records available for that specific user (or admin)
	if retrieveType == "replied" {                        // if the end user searches by the tasks that DOES HAVE final action
		db = db.Where("final_action IS NOT NULL")
	} else if retrieveType == "nonReplied" { // if the end user searches by the tasks that DOES NOT HAVE final action
		db = db.Where("final_action IS NULL")
	} else if retrieveType == "unseen" { // if the end user searches by the tasks that HE HAS NOT SEEN
		if admin { // if the end user was an admin
			db = db.Where("seen = 0 AND final_action IS NOT NULL") // get all the tasks that has a `final_action` and he has not seen yes
		} else { // if not
			db = db.Where("user_tasks.seen = 0") // get all the tasks that he has not seen weather or not it has a `final_action`
		}
	} else if retrieveType == "seen" { // if the end user searches by the tasks that HE HAS SEEN BEFORE
		if admin { // if the end user was an admin
			db = db.Where("seen = 1 AND final_action IS NOT NULL") // get all the tasks that has a `final_action` and he has not seen yes
		} else { // if not
			db = db.Where("user_tasks.seen = 1") // get all the tasks that he has not seen weather or not it has a `final_action`
		}
	} else if retrieveType == "notRepliedByAll" { // if the end user wants to get the tasks that was not replied by all the people it was assigned to
		var ids []int
		// gets the `ids` of the tasks that there is any person who's final response is equal to 0
		db.Table("tasks").Select("tasks.id").
			Joins("JOIN people ON tasks.id = people.task_id").
			Where("final_response = 0 AND people.deleted_at IS NULL").
			Group("tasks.id").Pluck("tasks.id", &ids)
		// gets all the tasks where its id is found in the ids array
		db = db.Where("tasks.id IN (?)", ids)
	}
	db = db.Preload("Users").Preload("People")
	if descriptionSearch != "" { // if the end user entered data in the description search
		descriptionSearch = "%" + descriptionSearch + "%" // search by the entered value with % before and after to match any
		db = db.Where("description LIKE ?", descriptionSearch)
	}
	if sentToSearch != "" { // if the end user entered data in the sent to search
		var ids []int
		sentToSearch = "%" + sentToSearch + "%" // add % before and after to match any
		// gets the ids of the tasks which have the search value in their People array
		db.Table("tasks").Select("DISTINCT tasks.id").
			Joins("JOIN people ON tasks.id = people.task_id").
			Where("people.name LIKE ?", sentToSearch).Pluck("tasks.id", &ids)
		// gets all the tasks where its id is found in the ids array
		db = db.Where("tasks.id IN (?)", ids)
	}
	if minDateSearch != "" {
		minDateSearch = minDateSearch + " 00:00:00.0000000 +02:00" // add correct formatting to the date search
		db = db.Where("created_at >= ?", minDateSearch)
	}
	if maxDateSearch != "" {
		maxDateSearch = maxDateSearch + " 00:00:00.0000000 +02:00" // add correct formatting to the date search
		db = db.Where("created_at <= ?", maxDateSearch)
	}
	// gets the total number of records after applying all the filtering
	var totalNumberOfRowsAfterFilter int
	db.Find(&tasks).Count(&totalNumberOfRowsAfterFilter)
	if sortedColumn != "created_at" { // if the user doesn't sort by the created_at column
		// order by the unseen first then the seen tasks
		if admin {
			db = db.Order("tasks.seen")
		} else {
			db = db.Order("user_tasks.seen")
		}
	}
	// adds the offset and limit to apply pagination
	db.Offset(offset).Limit(limit).Order(sortedColumn + " " + direction).Find(&tasks)
	return tasks, totalNumberOfRowsInDatabase, totalNumberOfRowsAfterFilter
}
