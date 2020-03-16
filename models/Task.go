package models

import (
	"database/sql"
	"github.com/jinzhu/gorm"
)

// this struct holds the data needed for the task
type Task struct {
	gorm.Model
	Description    string         `gorm:"not null;type:nvarchar(1024)" form:"description" json:"description"`
	FinalAction    sql.NullString `json:"final_action" gorm:"default: null;type:nvarchar(1024)"`
	Files          []File         `json:"files" gorm:"PRELOAD:false"`
	Seen           bool           `gorm:"default:1;not null" json:"seen"`
	Hash           string
	FollowingUsers []*FollowingUserTask `json:"following_users"`
	WorkingOnUsers []*WorkingOnUserTask `json:"workingOn_users"`
	SeenStatus     string               `gorm:"-" json:"seen_status"`
}

func (task *Task) AfterCreate(scope *gorm.Scope) error {
	ID := int(task.ID)
	hash := generateHash(ID)
	scope.DB().Model(task).Updates(Task{Hash: hash})
	return nil
}

func searchByRetrieveType(db *gorm.DB, retrieveType string, classification int) *gorm.DB {
	if retrieveType == "replied" { // if the end user searches by the tasks that DOES HAVE final action
		db = db.Where("final_action IS NOT NULL")
	} else if retrieveType == "nonReplied" { // if the end user searches by the tasks that DOES NOT HAVE final action
		db = db.Where("final_action IS NULL")
	} else if retrieveType == "unseen" { // if the end user searches by the tasks that HE HAS NOT SEEN
		if classification == 1 { // if the end user was an admin
			db = db.Where("seen = 0 AND final_action IS NOT NULL") // get all the tasks that has a `final_action` and he has not seen yes
		} else { // if not
			db = db.Where("(user_tasks.seen = 0 OR user_tasks.marked_as_unseen = 1)") // get all the tasks that he has not seen weather or not it has a `final_action`
		}
	} else if retrieveType == "seen" { // if the end user searches by the tasks that HE HAS SEEN BEFORE
		if classification == 1 { // if the end user was an admin
			db = db.Where("seen = 1 AND final_action IS NOT NULL") // get all the tasks that has a `final_action` and he has not seen yes
		} else { // if not
			db = db.Where("user_tasks.seen = 1 AND user_tasks.marked_as_unseen = 0") // get all the tasks that he has not seen weather or not it has a `final_action`
		}
	} else if retrieveType == "notRepliedByAll" { // if the end user wants to get the tasks that was not replied by all the workingOnUsers it was assigned to
		var ids []int
		// gets the `ids` of the tasks that there is any person who's final response is equal to 0
		db.Table("tasks").Select("tasks.id").
			Joins("JOIN working_on_user_tasks wout ON tasks.id = wout.task_id").
			Where("wout.final_response = 0 AND wout.deleted_at IS NULL").
			Group("tasks.id").Pluck("tasks.id", &ids)
		// gets all the tasks where its id is found in the ids array
		db = db.Where("tasks.id IN (?)", ids)
	} else if retrieveType == "notFinished" {
		var ids []int
		db.Table("tasks").Select("tasks.id").
			Where("user_tasks.final_response = 0 AND user_tasks.deleted_at IS NULL").
			Group("tasks.id").Pluck("tasks.id", &ids)
		db = db.Where("tasks.id IN (?)", ids)
	} else if retrieveType == "newFromWorkingOnUsers" {
		var ids []int
		db.Table("tasks").Select("tasks.id").
			Where("user_tasks.new_from_working_on_user = 1 AND user_tasks.deleted_at IS NULL").
			Group("tasks.id").Pluck("tasks.id", &ids)
		db = db.Where("tasks.id IN (?)", ids)
	}
	return db
}

func filterByFields(db *gorm.DB, descriptionSearch, sentToSearch, minDateSearch, maxDateSearch string) *gorm.DB {
	if descriptionSearch != "" { // if the end user entered data in the description search
		descriptionSearch = "%" + descriptionSearch + "%" // search by the entered value with % before and after to match any
		db = db.Where("description LIKE ?", descriptionSearch)
	}
	if sentToSearch != "" { // if the end user entered data in the sent to search
		var ids []int
		// gets the ids of the tasks which have the search value in their People array
		db.Table("tasks").Select("DISTINCT tasks.id").
			Joins("JOIN working_on_user_tasks wout ON tasks.id = wout.task_id").
			Joins("JOIN users ON users.id = wout.user_id").
			Where("users.hash = ?", sentToSearch).Pluck("tasks.id", &ids)
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
	return db
}

// this function takes the search parameters (datatables parameters) and return the corresponding data
func GetPaginatedTasks(db *gorm.DB, offset int, limit int, sortedColumn, direction,
	descriptionSearch, sentToSearch, minDateSearch, maxDateSearch, retrieveType string,
	classification int, userID uint) ([]Task, int, int) {

	db, totalNumberOfRowsInDatabase, totalNumberOfRowsAfterFilter :=
		filterTasks(sortedColumn, db, classification, userID, retrieveType,
			descriptionSearch, sentToSearch, minDateSearch, maxDateSearch)
	// adds the offset and limit to apply pagination
	var tasks []Task
	db.Offset(offset).Limit(limit).Order(sortedColumn + " " + direction).Find(&tasks)

	evaluateSeenStatus(&tasks, classification, userID)

	return tasks, totalNumberOfRowsInDatabase, totalNumberOfRowsAfterFilter
}

func filterTasks(sortedColumn string, db *gorm.DB, classification int, userID uint, retrieveType string, descriptionSearch string, sentToSearch string, minDateSearch string, maxDateSearch string) (*gorm.DB, int, int) {

	sortedColumn = "tasks." + sortedColumn
	// set the name of the column that the end user is sorting with
	var tasks []Task
	totalNumberOfRowsInDatabase, db := getTotalNumberOfRecordsInDatabase(classification, db, userID)
	db = searchByRetrieveType(db, retrieveType, classification)
	db = PreloadFollowingAndWorkingOnUsers(classification, db, userID)

	db = filterByFields(db, descriptionSearch, sentToSearch, minDateSearch, maxDateSearch)
	// gets the total number of records after applying all the filtering
	var totalNumberOfRowsAfterFilter int
	db.Find(&tasks).Count(&totalNumberOfRowsAfterFilter)
	if sortedColumn != "created_at" { // if the user doesn't sort by the created_at column
		// order by the unseen first then the seen tasks
		if classification == 1 {
			db = db.Order("tasks.seen")
		} else {
			db = db.Order("user_tasks.seen")
		}
	}
	return db, totalNumberOfRowsInDatabase, totalNumberOfRowsAfterFilter
}

func GetAllTasks(db *gorm.DB, sortedColumn, direction, descriptionSearch, sentToSearch,
	minDateSearch, maxDateSearch, retrieveType string, classification int, userID uint) ([]Task, int, int) {

	db, totalNumberOfRowsInDatabase, totalNumberOfRowsAfterFilter :=
		filterTasks(sortedColumn, db, classification, userID, retrieveType,
			descriptionSearch, sentToSearch, minDateSearch, maxDateSearch)
	// adds the offset and limit to apply pagination
	var tasks []Task
	db.Order(sortedColumn + " " + direction).Find(&tasks)
	return tasks, totalNumberOfRowsInDatabase, totalNumberOfRowsAfterFilter
}

func PreloadFollowingAndWorkingOnUsers(classification int, db *gorm.DB, userID uint) *gorm.DB {
	if classification == 3 {
		db = db.Preload("WorkingOnUsers", func(db *gorm.DB) *gorm.DB {
			return db.Where("working_on_user_tasks.user_id = ?", userID)
		})
		db = db.Preload("FollowingUsers", func(db *gorm.DB) *gorm.DB {
			return db.Where("following_user_tasks.user_id IN (?)",
				db.Table("working_on_user_tasks").Select("working_on_user_tasks.follower_id").
					Where("working_on_user_tasks.user_id = ? AND working_on_user_tasks.deleted_at IS NULL", userID).QueryExpr())
		})
	} else {
		db = db.Preload("FollowingUsers").Preload("WorkingOnUsers.User")
	}
	return db
}

func getTotalNumberOfRecordsInDatabase(classification int, db *gorm.DB, userID uint) (int, *gorm.DB) {
	var totalNumberOfRowsInDatabase int
	if classification == 2 {
		// join the user_tasks table to get only the tasks that were assigned to the logged in user
		db = db.Table("tasks").Joins("JOIN following_user_tasks user_tasks " +
			"ON user_tasks.task_id = tasks.id")
		db = db.Where("user_tasks.user_id = ?", userID)
	} else if classification == 3 {
		// join the user_tasks table to get only the tasks that were assigned to the logged in user
		db = db.Table("tasks").Joins("JOIN working_on_user_tasks user_tasks " +
			"ON user_tasks.task_id = tasks.id")
		db = db.Where("user_tasks.user_id = ?", userID)
	}
	db.Model(&Task{}).Count(&totalNumberOfRowsInDatabase)
	// gets the total number of records available for that specific user (or admin)
	return totalNumberOfRowsInDatabase, db
}

func GetTask(hash string, db *gorm.DB, classification int, userID uint) ([]Task, int, int) {
	var tasks []Task

	db = PreloadFollowingAndWorkingOnUsers(classification, db, userID)
	db.Find(&tasks, "hash = ?", hash)
	totalNumberOfRecords, _ := getTotalNumberOfRecordsInDatabase(classification, db, userID)
	evaluateSeenStatus(&tasks, classification, userID)
	return tasks, totalNumberOfRecords, 1
}

// this function adds the CSS class that should be added to each row in the table
func evaluateSeenStatus(tasks *[]Task, classification int, userID uint) {
	for index, task := range *tasks {
		if classification == 1 && !task.Seen {
			(*tasks)[index].SeenStatus = "unseen"
		} else if classification == 2 {
			for followingUserIndex, _ := range task.FollowingUsers {
				if task.FollowingUsers[followingUserIndex].UserID == userID {
					if !task.FollowingUsers[followingUserIndex].Seen || task.FollowingUsers[followingUserIndex].NewFromMinister {
						(*tasks)[index].SeenStatus = "unseen"
					} else if task.FollowingUsers[followingUserIndex].NewFromWorkingOnUser {
						(*tasks)[index].SeenStatus = "new_from_working_on_user"
					} else if task.FollowingUsers[followingUserIndex].MarkedAsUnseen {
						(*tasks)[index].SeenStatus = "marked_as_unseen"
					}
				}
			}
		} else if classification == 3 {
			for workingOnUserIndex, _ := range task.WorkingOnUsers {
				if task.WorkingOnUsers[workingOnUserIndex].UserID == userID {
					if !task.WorkingOnUsers[workingOnUserIndex].Seen {
						(*tasks)[index].SeenStatus = "unseen"
					} else if task.WorkingOnUsers[workingOnUserIndex].MarkedAsUnseen {
						(*tasks)[index].SeenStatus = "marked_as_unseen"
					}
				}
			}
		}

	}
}
