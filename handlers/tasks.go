package handlers

import (
	"../models"
	"fmt"
	"github.com/labstack/echo"
	"net/http"
	"strconv"
)

// this struct is used to return the data to the datatable editor in the correct format
type datatableTask struct {
	Data []interface{} `json:"data"`
}

// this function adds task to the database
func (db *MyDB) AddTask(c echo.Context) error {
	taskToSave := models.Task{
		Description: c.FormValue("data[description]"), // gets the value of the description from the form that was submitted
	}
	db.GormDB.Create(&taskToSave) // saves the task to the database

	linkFiles(db, &c, taskToSave.ID)

	addFollowersUsers(c, db, taskToSave)
	_, classification := getUserStatus(&c)                      // gets the user status (id, admin)
	sendNotification("تم اضافه تكليف جديد", classification, db) // send notification (sent from `classification` variable)

	//addWorkingOnUsers(c, db, int(taskToSave.ID), userID)

	// gets the users and people from the database
	db.GormDB.Preload("Users").Find(&taskToSave, taskToSave.ID)

	// make a new array that should contain datatablesTask struct and return it to the datatable editor
	dataArray := make([]interface{}, 1)
	dataArray[0] = taskToSave
	datatableTask := datatableTask{dataArray}
	return c.JSONPretty(http.StatusOK, datatableTask, " ")
}

// this function edits an existing task
func (db *MyDB) EditTask(c echo.Context) error {
	userID, classification := getUserStatus(&c)    // gets the user status (id, classification)
	taskID, err := strconv.Atoi(c.FormValue("id")) // gets the ID of the requested task to edit
	if err != nil {                                // if an error occurred parsing the ID, it may be malicious request
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "Invalid Request",
		})
	}

	description := c.FormValue("data[description]")  // gets the value of the description
	finalAction := c.FormValue("data[final_action]") // gets the value of the final_action

	var task models.Task
	db.GormDB.First(&task, taskID) // load the required task from the database using the ID
	if classification == 1 {       // if the logged in user is an admin, then he could change the description of the task
		db.GormDB.Model(&task).UpdateColumn("description", description)
	}
	if finalAction != task.FinalAction.String { // if the final_action given by the user is different than that in the database
		var adminsIDs []uint
		db.GormDB.Model(&models.User{}).Where("classification = 1").Pluck("id", &adminsIDs)
		if finalAction == "" { // if final_action string is empty, then the final_action was deleted
			// set the final_action to null and there's no need to mark the task as unseen
			db.GormDB.Model(&task).Updates(map[string]interface{}{"final_action": nil, "seen": true})
			// send a notification to the admin informing him
			sendNotification("تم الغاء الاجراء النهائي للتكليف", classification, db)
		} else {
			// change the final_action and mark the task as unseen
			db.GormDB.Model(&task).Updates(map[string]interface{}{"final_action": finalAction, "seen": false})
			// send a notification to the admin informing him
			sendNotification("تم تعديل الاجراء النهائي للتكليف", classification, db)
		}
	}

	linkFiles(db, &c, task.ID)
	// only the admin has the privileges to assign or delete users to the task
	var followersIDs []uint
	if classification == 1 {
		followersIDs = append(followersIDs, addFollowersUsers(c, db, task)...)
		sendNotification("تم تعديل التكليف", classification, db) // send notifications to the users telling them that the task was edited
	}

	var workingOnIDs []uint
	if classification == 2 {
		workingOnIDs = append(workingOnIDs, addWorkingOnUsers(c, db, taskID, userID)...)
	}

	if followersIDs == nil {
		followersIDs = []uint{0}
	}
	if workingOnIDs == nil { // checks if the ids array is empty
		workingOnIDs = []uint{0} // add a dummy id (0) to avoid unexpected behaviors
	}

	if classification == 1 { // since only the admin has the privileges to delete people
		db.GormDB.Delete(models.FollowingUserTask{}, "task_id = ? AND user_id NOT IN (?)", taskID, followersIDs)
	} else if classification == 2 {
		db.GormDB.Delete(models.WorkingOnUserTask{}, "task_id = ? AND user_id NOT IN (?)", taskID, workingOnIDs)
	}

	dataArray := make([]interface{}, 1)
	dataArray[0] = task
	datatableTask := datatableTask{dataArray}
	return c.JSONPretty(http.StatusOK, datatableTask, " ")
}

func addFollowersUsers(c echo.Context, db *MyDB, taskToSave models.Task) []uint {
	totalUsers, _ := strconv.Atoi(c.FormValue("data[totalUsers]"))
	// gets the value of the total users that were assigned to finish that task
	var users []uint
	for i := 0; i < totalUsers; i++ { // loop for the number of the users to add and notify them
		id := c.FormValue("data[following_users_" + strconv.Itoa(i) + "]") // get the ID of each user
		if id == "" {
			continue
		}
		uid, _ := strconv.ParseUint(id, 10, 64)
		models.CreateFollowingUserTask(db.GormDB, taskToSave.ID, uint(uid)) // creates a FollowingUserTask to the database
		users = append(users, uint(uid))                                    // append the id to the users array
	}
	return users
}

func addWorkingOnUsers(c echo.Context, db *MyDB, taskID int, followerID uint) []uint {
	var ids []uint
	totalWorkingOnPeople, _ := strconv.Atoi(c.FormValue("data[totalWorkingOnUsers]"))
	// gets the total number of people that should be called to take an action
	for i := 0; i < totalWorkingOnPeople; i++ { // loop over the people to add them
		var userTask models.WorkingOnUserTask
		userID := c.FormValue("data[people_user_id_" + strconv.Itoa(i) + "]") // get the userTask's userID
		uid, _ := strconv.ParseUint(userID, 10, 64)

		action := c.FormValue("data[people_action_" + strconv.Itoa(i) + "]")                                     // get the userTask's action
		finalResponse, _ := strconv.ParseBool(c.FormValue("data[people_finalResponse_" + strconv.Itoa(i) + "]")) // get the boolean indicating weather or not it is a final action
		notes := c.FormValue("data[people_notes_" + strconv.Itoa(i) + "]")
		if userID == "" { // if no userID is given then continue
			continue
		}
		db.GormDB.Where("user_id = ? AND task_id = ?", userID, taskID).Find(&userTask) // try to get the userTask with the same userID and task id
		var id uint
		if userTask.ID == 0 { // if not found create one
			id = models.CreateWorkingOnUserTask(db.GormDB, uint(taskID), uint(uid), action, finalResponse, followerID)
		} else { // if found edit his data
			userTask.ActionTaken = action
			userTask.FinalResponse = finalResponse
			userTask.Notes = notes
			db.GormDB.Save(&userTask)
			id = userTask.UserID
		}
		ids = append(ids, id) // add the id to the peopleIDs
	}
	return ids
}

// this function deletes a task from the database
func (db *MyDB) RemoveTask(c echo.Context) error {
	id, err := strconv.Atoi(c.FormValue("id[]")) // gets the id of the task to delete
	if err != nil || id == 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "Invalid Request",
		})
	}
	var task models.Task
	db.GormDB.First(&task, id) // gets the task from the database
	//task.DeleteChildren(db.GormDB)              // delete any UserTasks assigned to it
	db.GormDB.Delete(&task)                     // delete the task
	return c.JSON(http.StatusOK, models.Task{}) // return empty struct to the datatable editor
}

// this function changes the seen value of the user on a task
func (db *MyDB) ChangeUserSeen(c echo.Context) error {
	seen, _ := strconv.ParseBool(c.FormValue("seen"))
	taskID := c.FormValue("task_id") // gets the value of task_id
	userID := c.FormValue("user_id") // gets the value of user_id
	isFollower, _ := strconv.ParseBool(c.FormValue("is_follower"))
	if isFollower {
		if seen {
			db.GormDB.Model(models.FollowingUserTask{}).Where("task_id = ? AND user_id = ?", taskID, userID).
				Updates(map[string]interface{}{"seen": true, "marked_as_unseen": false})
		} else {
			db.GormDB.Model(models.FollowingUserTask{}).Where("task_id = ? AND user_id = ?", taskID, userID).
				Update("marked_as_unseen", true)
		}
	} else {
		if seen {
			db.GormDB.Model(models.WorkingOnUserTask{}).Where("task_id = ? AND user_id = ?", taskID, userID).
				Updates(map[string]interface{}{"seen": true, "marked_as_unseen": false})
		} else {
			db.GormDB.Model(models.WorkingOnUserTask{}).Where("task_id = ? AND user_id = ?", taskID, userID).
				Update("marked_as_unseen", true)
		}

	}
	return nil
}

// this function changes the value of the task seen from the admin's account
func (db *MyDB) ChangeTaskSeen(c echo.Context) error {
	seen := c.FormValue("seen")
	taskID := c.FormValue("task_id") // gets the value of task_id
	db.GormDB.Model(models.Task{}).Where("id = ?", taskID).Update("seen", seen)
	return nil
}

// this function gets the parameters of the datatable to send it to `GetAllTasks` function
func (db *MyDB) GetTasks(c echo.Context) error {
	userID, classification := getUserStatus(&c) // gets the value of userID and classification
	q := c.Request().URL.Query()                // gets the URL Query as a map
	draw, _ := strconv.Atoi(q["draw"][0])
	start, _ := strconv.Atoi(q["start"][0])                         // the start point of the current data set
	length, _ := strconv.Atoi(q["length"][0])                       // number of records to display (page size)
	sortedColumnNumber, _ := strconv.Atoi(q["order[0][column]"][0]) // column to which ordering should be applied
	direction := q["order[0][dir]"][0]                              // ordering direction for this column
	sprintf := fmt.Sprintf("columns[%d][name]", sortedColumnNumber) // gets the name of the sorted column (not the numer)
	sortedColumnName := q[sprintf][0]
	descriptionSearch := q["description"][0] // the value of the description search
	sentToSearch := ""
	if classification != 3 {
		sentToSearch = q["sent_to"][0] // the value of the sent_to search
	}
	minDateSearch := q["min_date"][0] // the value of min_date search
	maxDateSearch := q["max_date"][0] // the value of max_date search
	retrieveType := q["retrieve"][0]  // the value of the retrieve type
	tasks, totalNumberOfRowsInDatabase, totalNumberOfRowsAfterFilter, files := models.GetAllTasks(db.GormDB, start, length,
		sortedColumnName, direction, descriptionSearch, sentToSearch, minDateSearch, maxDateSearch, retrieveType, classification, userID)
	dt := dtOutput{
		Draw:            draw,
		RecordsTotal:    totalNumberOfRowsInDatabase,
		RecordsFiltered: totalNumberOfRowsAfterFilter,
		Data:            tasks,
		Files:           files,
	}

	return c.JSONPretty(http.StatusOK, dt, " ")
}

// struct to return the datatable rows in the correct format
type dtOutput struct {
	Draw            int                    `json:"draw"`
	RecordsTotal    int                    `json:"recordsTotal"`
	RecordsFiltered int                    `json:"recordsFiltered"`
	Data            []models.Task          `json:"data"`
	Files           map[string]interface{} `json:"files"`
}
