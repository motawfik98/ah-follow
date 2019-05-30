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

	totalUsers, _ := strconv.Atoi(c.FormValue("data[totalUsers]")) // gets the value of the total users that were assigned to finish that task
	var users []uint
	for i := 0; i < totalUsers; i++ { // loop for the number of the users to add and notify them
		id := c.FormValue("data[users_" + strconv.Itoa(i) + "]") // get the ID of each user
		uid, _ := strconv.ParseUint(id, 10, 64)
		models.CreateUserTask(db.GormDB, taskToSave.ID, uint(uid)) // creates a UserTask to the database
		users = append(users, uint(uid))                           // append the id to the users array
	}
	_, isAdmin := getUserStatus(&c)                      // gets the user status (id, admin)
	sendNotification("تم اضافه تكليف جديد", isAdmin, db) // send notification (sent from `isAdmin` variable)

	totalPeople, _ := strconv.Atoi(c.FormValue("data[totalPeople]")) // gets the total number of people that should be called to take an action
	for i := 0; i < totalPeople; i++ {                               // loop for the number of the users to add them
		finalResponse, _ := strconv.ParseBool(c.FormValue("data[people_finalResponse_" + strconv.Itoa(i) + "]")) // gets weather or not the response is final
		// create a Person passing person_name, person_action, taskID, and weather or not it is a final response
		models.CreatePerson(db.GormDB, c.FormValue("data[people_name_"+strconv.Itoa(i)+"]"),
			c.FormValue("data[people_action_"+strconv.Itoa(i)+"]"), taskToSave.ID, finalResponse)
	}

	// gets the users and people from the database
	db.GormDB.Preload("Users").Preload("People").Find(&taskToSave, taskToSave.ID)

	// make a new array that should contain datatablesTask struct and return it to the datatable editor
	dataArray := make([]interface{}, 1)
	dataArray[0] = taskToSave
	datatableTask := datatableTask{dataArray}
	return c.JSONPretty(http.StatusOK, datatableTask, " ")
}

// this function edits an existing task
func (db *MyDB) EditTask(c echo.Context) error {
	_, isAdmin := getUserStatus(&c)                // gets the user status (id, admin)
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
	if isAdmin {                   // if the logged in user is an admin, then he could change the description of the task
		db.GormDB.Model(&task).UpdateColumn("description", description)
	}
	if finalAction != task.FinalAction.String { // if the final_action given by the user is different than that in the database
		var adminsIDs []uint
		db.GormDB.Model(&models.User{}).Where("admin = 1").Pluck("id", &adminsIDs)
		if finalAction == "" { // if final_action string is empty, then the final_action was deleted
			// set the final_action to null and there's no need to mark the task as unseen
			db.GormDB.Model(&task).Updates(map[string]interface{}{"final_action": nil, "seen": true})
			// send a notification to the admin informing him
			sendNotification("تم الغاء الاجراء النهائي للتكليف", isAdmin, db)
		} else {
			// change the final_action and mark the task as unseen
			db.GormDB.Model(&task).Updates(map[string]interface{}{"final_action": finalAction, "seen": false})
			// send a notification to the admin informing him
			sendNotification("تم تعديل الاجراء النهائي للتكليف", isAdmin, db)
		}
	}

	// only the admin has the privileges to assign or delete users to the task
	if isAdmin {
		totalUsers, _ := strconv.Atoi(c.FormValue("data[totalUsers]")) // gets the number of the totalUsers assigned to the task
		var ids []uint
		for i := 0; i < totalUsers; i++ { // loop over the users to add them
			var userTask models.UserTask

			id := c.FormValue("data[users_" + strconv.Itoa(i) + "]") // gets the id of the user
			uid, _ := strconv.ParseUint(id, 10, 64)
			ids = append(ids, uint(uid))                                                // adds the id to the ids array
			db.GormDB.Where("task_id = ? AND user_id = ?", taskID, uid).Find(&userTask) // search for a UserTask where it has the same task_id and user_id
			if userTask.TaskID == 0 && userTask.UserID == 0 {                           // if not found then create a new one
				models.CreateUserTask(db.GormDB, uint(taskID), uint(uid))
			} // if found then no action is taken
		}
		if ids == nil { // checks if the ids array is empty
			ids = []uint{0} // add a dummy id (0) to avoid unexpected behaviors
		}
		db.GormDB.Delete(models.UserTask{}, "task_id = ? AND user_id NOT IN (?)", taskID, ids) // delete any assigned user that his ID is not in the ids array
		sendNotification("تم تعديل التكليف", isAdmin, db)                                      // send notifications to the users telling them that the task was edited
	}

	totalPeople, _ := strconv.Atoi(c.FormValue("data[totalPeople]")) // gets the total number of people that should be called to take an action
	var peopleIDs []int
	for i := 0; i < totalPeople; i++ { // loop over the people to add them
		var person models.Person
		name := c.FormValue("data[people_name_" + strconv.Itoa(i) + "]")                                         // get the person's name
		action := c.FormValue("data[people_action_" + strconv.Itoa(i) + "]")                                     // get the person's action
		finalResponse, _ := strconv.ParseBool(c.FormValue("data[people_finalResponse_" + strconv.Itoa(i) + "]")) // get the boolean indicating weather or not it is a final action
		if name == "" {                                                                                          // if no name is given then continue
			continue
		}
		db.GormDB.Where("name = ? AND task_id = ?", name, taskID).Find(&person) // try to get the person with the same name and task id
		var personID int
		if person.ID == 0 { // if not found create one
			personID = models.CreatePerson(db.GormDB, name, action, uint(taskID), finalResponse)
		} else { // if found edit his data
			person.ActionTaken = action
			person.FinalResponse = finalResponse
			db.GormDB.Save(&person)
			personID, _ = strconv.Atoi(c.FormValue("data[people_id_" + strconv.Itoa(i) + "]"))
		}
		peopleIDs = append(peopleIDs, personID) // add the personID to the peopleIDs
	}
	if peopleIDs == nil { // checks if the ids array is empty
		peopleIDs = []int{0} // add a dummy id (0) to avoid unexpected behaviors
	}
	if isAdmin { // since only the admin has the privileges to delete people
		db.GormDB.Delete(models.Person{}, "task_id = ? AND id NOT IN (?)", taskID, peopleIDs)
	}

	dataArray := make([]interface{}, 1)
	dataArray[0] = task
	datatableTask := datatableTask{dataArray}
	return c.JSONPretty(http.StatusOK, datatableTask, " ")
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
	db.GormDB.First(&task, id)                  // gets the task from the database
	task.DeleteChildren(db.GormDB)              // delete any UserTasks assigned to it
	db.GormDB.Delete(&task)                     // delete the task
	return c.JSON(http.StatusOK, models.Task{}) // return empty struct to the datatable editor
}

// this function changes the seen value of the user on a task
func (db *MyDB) ChangePersonSeen(c echo.Context) error {
	seen := c.FormValue("seen")
	taskID := c.FormValue("task_id") // gets the value of task_id
	userID := c.FormValue("user_id") // gets the value of user_id
	db.GormDB.Model(models.UserTask{}).Where("task_id = ? AND user_id = ?", taskID, userID).Update("seen", seen)
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
	q := c.Request().URL.Query() // gets the URL Query as a map
	draw, _ := strconv.Atoi(q["draw"][0])
	start, _ := strconv.Atoi(q["start"][0])                         // the start point of the current data set
	length, _ := strconv.Atoi(q["length"][0])                       // number of records to display (page size)
	sortedColumnNumber, _ := strconv.Atoi(q["order[0][column]"][0]) // column to which ordering should be applied
	direction := q["order[0][dir]"][0]                              // ordering direction for this column
	sprintf := fmt.Sprintf("columns[%d][name]", sortedColumnNumber) // gets the name of the sorted column (not the numer)
	sortedColumnName := q[sprintf][0]
	descriptionSearch := q["description"][0] // the value of the description search
	sentToSearch := q["sent_to"][0]          // the value of the sent_to search
	minDateSearch := q["min_date"][0]        // the value of min_date search
	maxDateSearch := q["max_date"][0]        // the value of max_date search
	retrieveType := q["retrieve"][0]         // the value of the retrieve type
	userID, admin := getUserStatus(&c)       // gets the value of userID and admin
	tasks, totalNumberOfRowsInDatabase, totalNumberOfRowsAfterFilter := models.GetAllTasks(db.GormDB, start, length,
		sortedColumnName, direction, descriptionSearch, sentToSearch, minDateSearch, maxDateSearch, retrieveType, admin, userID)
	dt := dtOutput{
		Draw:            draw,
		RecordsTotal:    totalNumberOfRowsInDatabase,
		RecordsFiltered: totalNumberOfRowsAfterFilter,
		Data:            tasks,
	}
	return c.JSONPretty(http.StatusOK, dt, " ")
}

// struct to return the datatable rows in the correct format
type dtOutput struct {
	Draw            int           `json:"draw"`
	RecordsTotal    int           `json:"recordsTotal"`
	RecordsFiltered int           `json:"recordsFiltered"`
	Data            []models.Task `json:"data"`
}
