package handlers

import (
	"ah-follow-modules/models"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo/v4"
	"net/http"
	"net/url"
	"strconv"
)

// this struct is used to return the data to the datatable editor in the correct format
type datatableTask struct {
	Data []interface{} `json:"data"`
}

// this function adds task to the database
func (db *MyConfigurations) AddTask(c echo.Context) error {
	taskToSave := models.Task{
		Description: c.FormValue("description"), // gets the value of the description from the form that was submitted
	}
	db.GormDB.Create(&taskToSave) // saves the task to the database

	linkFiles(db, &c, taskToSave.ID)

	addFollowersUsers(c, db, taskToSave)

	//addWorkingOnUsers(c, db, int(taskToSave.ID), userID)

	// gets the users and people from the database
	db.GormDB.Preload("FollowingUsers").Preload("WorkingOnUsers").Find(&taskToSave, taskToSave.ID)

	addFlashMessage("success", "تم انشاء التكليف", &c)
	return c.JSON(http.StatusOK, map[string]string{
		"status":  "success",
		"message": "تم انشاء التكليف",
	})
}

// this function edits an existing task
func (db *MyConfigurations) EditTask(c echo.Context) error {
	userID, classification := getUserStatus(&c) // gets the user status (id, classification)
	username, _ := getUsernameAndClassification(&c)
	taskID, err := strconv.Atoi(c.FormValue("id")) // gets the ID of the requested task to edit
	if err != nil {                                // if an error occurred parsing the ID, it may be malicious request
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "Invalid Request",
		})
	}

	description := c.FormValue("description")  // gets the value of the description
	finalAction := c.FormValue("final_action") // gets the value of the final_action

	var task models.Task
	db.GormDB.First(&task, taskID) // load the required task from the database using the ID
	taskLink := hostDomain + "?" + task.Hash

	if classification == 1 { // if the logged in user is an admin, then he could change the description of the task
		db.GormDB.Model(&task).Update("description", description)
	}
	if finalAction != task.FinalAction.String { // if the final_action given by the user is different than that in the database
		var admins []models.User
		db.GormDB.Find(&admins, "classification = 1")
		if finalAction == "" { // if final_action string is empty, then the final_action was deleted
			// set the final_action to null and there's no need to mark the task as unseen
			db.GormDB.Model(&task).Updates(map[string]interface{}{"final_action": nil, "seen": true})
			// send a notification to the admin informing him
			//sendNotification("تم الغاء الاجراء النهائي للتكليف", classification, db)
		} else {
			// change the final_action and mark the task as unseen
			db.GormDB.Model(&task).Updates(map[string]interface{}{"final_action": finalAction, "seen": false})
			// send a notification to the admin informing him
			//pushNotificationLink := "/?hash=" + task.Hash
			var adminsIDs = make([]uint, len(admins))
			for _, admin := range admins {
				//sendNotification("تم تعديل الاجراء النهائي للتكليف", admin.ID, db, pushNotificationLink)
				if admin.EmailNotifications {
					sendEmailNotification(&admin, taskLink, username, "تم تعديل الاجراء النهائي للتكليف")
				}
				if admin.PhoneNotifications {
					sendPhoneNotification(&admin, taskLink, username, "تم تعديل الاجراء النهائي للتكليف")
				}
				adminsIDs = append(adminsIDs, admin.ID)
				models.AddNotificationToDatabase(db.GormDB, admin.ID, task.ID, 0, 2, "تعديل الاجراء النهائي")
			}
			gatherUsersToSendNotifications(db, adminsIDs, "تعديل الاجراء النهائي", &task)
		}
	}

	linkFiles(db, &c, task.ID)
	// only the admin has the privileges to assign or delete users to the task
	var followersIDs []uint
	if classification == 1 {
		followersIDs = append(followersIDs, addFollowersUsers(c, db, task)...)
		//sendNotification("تم تعديل التكليف", classification, db) // send notifications to the users telling them that the task was edited
	}

	var workingOnIDs []uint
	workingOnIDs = append(workingOnIDs, addWorkingOnUsers(c, db, &task, classification, userID)...)

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

	addFlashMessage("success", "تم تعديل التكليف", &c)
	return c.JSON(http.StatusOK, map[string]string{
		"status":  "success",
		"message": "تم تعديل التكليف",
	})
	//return redirectWithFlashMessage("success", "تم تعديل التكليف", "/", &c)
}

func addFollowersUsers(c echo.Context, db *MyConfigurations, taskToSave models.Task) []uint {
	username, _ := getUsernameAndClassification(&c)
	totalUsers, _ := strconv.Atoi(c.FormValue("totalUsers"))
	// gets the value of the total users that were assigned to finish that task
	var users = make([]uint, totalUsers)
	taskLink := hostDomain + "?hash=" + taskToSave.Hash
	for i := 0; i < totalUsers; i++ { // loop for the number of the users to add and notify them
		id := c.FormValue("following_users_" + strconv.Itoa(i)) // get the ID of each user
		if id == "" {
			continue
		}
		uid, _ := strconv.ParseUint(id, 10, 64)
		isNew := models.CreateFollowingUserTask(db.GormDB, taskToSave.ID, uint(uid)) // creates a FollowingUserTask to the database
		users = append(users, uint(uid))                                             // append the id to the users array

		if isNew {
			var user models.User
			db.GormDB.Find(&user, uid)
			//sendNotification("تم اضافه تكليف جديد", uint(uid), db, taskLink)
			if user.EmailNotifications {
				sendEmailNotification(&user, taskLink, username, "تم اضافه تكليف جديد")
			}
			if user.PhoneNotifications {
				sendPhoneNotification(&user, taskLink, username, "تكليف جديد")
			}
			models.AddNotificationToDatabase(db.GormDB, user.ID, taskToSave.ID, 0, 1, "تكليف جديد")
		}
	}
	gatherUsersToSendNotifications(db, users, "تكليف جديد", &taskToSave)
	return users
}

func addWorkingOnUsers(c echo.Context, db *MyConfigurations, task *models.Task, classification int, followerID uint) []uint {
	username, _ := getUsernameAndClassification(&c)
	taskLink := hostDomain + "?hash=" + task.Hash
	totalWorkingOnPeople, _ := strconv.Atoi(c.FormValue("totalWorkingOnUsers"))
	var ids = make([]uint, totalWorkingOnPeople)

	// gets the total number of people that should be called to take an action
	for i := 0; i < totalWorkingOnPeople; i++ { // loop over the people to add them
		var userTask models.WorkingOnUserTask
		userID := c.FormValue("people_user_id_" + strconv.Itoa(i)) // get the userTask's userID
		uid, _ := strconv.ParseUint(userID, 10, 64)

		action := c.FormValue("people_action_" + strconv.Itoa(i))                                     // get the userTask's action
		finalResponse, _ := strconv.ParseBool(c.FormValue("people_finalResponse_" + strconv.Itoa(i))) // get the boolean indicating weather or not it is a final action
		notes := c.FormValue("people_notes_" + strconv.Itoa(i))
		if userID == "" { // if no userID is given then continue
			continue
		}
		db.GormDB.Where("user_id = ? AND task_id = ?", userID, task.ID).Find(&userTask) // try to get the userTask with the same userID and task id
		var user models.User
		db.GormDB.Find(&user, userID)
		var id uint
		if userTask.ID == 0 { // if not found create one
			if classification == 2 {
				id = models.CreateWorkingOnUserTask(db.GormDB, task.ID, uint(uid), action, finalResponse, followerID)
				//sendNotification("تم اضافه تكليف جديد", uint(uid), db, taskLink)
				if user.EmailNotifications {
					sendEmailNotification(&user, taskLink, username, "تم اضافه تكليف جديد")
				}
				if user.PhoneNotifications {
					sendPhoneNotification(&user, taskLink, username, "تكليف جديد")
				}
				models.AddNotificationToDatabase(db.GormDB, user.ID, task.ID, 0, 1, "تكليف جديد")
				gatherUsersToSendNotifications(db, []uint{user.ID}, "تكليف جديد", task)
			}
		} else { // if found edit his data
			if classification == 2 {
				userTask.ActionTaken = action
				userTask.FinalResponse = finalResponse
			} else if classification == 3 {
				userTask.Notes = notes
				db.GormDB.Model(models.FollowingUserTask{}).Where("task_id = ? AND user_id = ?", task.ID, userTask.FollowerID).
					Updates(map[string]interface{}{"new_from_working_on_user": true})
				var followerUser models.User
				db.GormDB.Find(&followerUser, userTask.FollowerID)
				//sendNotification("تم اضافه رد على التكليف من القائم به", userTask.FollowerID, db, taskLink)
				if user.EmailNotifications {
					sendEmailNotification(&followerUser, taskLink, username, "تم اضافه رد على التكليف من القائم به")
				}
				if user.PhoneNotifications {
					sendPhoneNotification(&followerUser, taskLink, username, "استجابه")
				}
				models.AddNotificationToDatabase(db.GormDB, userTask.FollowerID, task.ID, user.ID, 3, "استجابه من القائم به")
				gatherUsersToSendNotifications(db, []uint{userTask.FollowerID}, "تم اضافه رد على التكليف من القائم به", task)
			}
			db.GormDB.Save(&userTask)
			id = userTask.UserID
		}
		ids = append(ids, id) // add the id to the peopleIDs
	}
	return ids
}

// this function deletes a task from the database
func (db *MyConfigurations) RemoveTask(c echo.Context) error {
	hash := c.FormValue("hash") // gets the hash of the task to delete
	var task models.Task
	db.GormDB.Where("hash = ?", hash).First(&task) // gets the task from the database
	if task.ID != 0 {
		//task.DeleteChildren(db.GormDB)              // delete any UserTasks assigned to it
		db.GormDB.Delete(&task) // delete the task
	}
	if checkIfRequestFromMobileDevice(c) {
		return c.JSON(http.StatusOK, echo.Map{
			"status":  "success",
			"message": "تم الغاء التكليف",
		})
	}
	return redirectWithFlashMessage("success", "تم الغاء التكليف", "/", &c)
}

// this function changes the seen value of the user on a task
func (db *MyConfigurations) ChangeUserSeen(c echo.Context) error {
	seen, _ := strconv.ParseBool(c.FormValue("seen"))
	taskID := c.FormValue("task_id") // gets the value of task_id
	userID := c.FormValue("user_id") // gets the value of user_id
	isFollower, _ := strconv.ParseBool(c.FormValue("is_follower"))
	if isFollower {
		if seen {
			db.GormDB.Model(models.FollowingUserTask{}).Where("task_id = ? AND user_id = ?", taskID, userID).
				Updates(map[string]interface{}{"seen": true, "marked_as_unseen": false,
					"new_from_minister": false, "new_from_working_on_user": false})
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
func (db *MyConfigurations) ChangeTaskSeen(c echo.Context) error {
	seen := c.FormValue("seen")
	taskID := c.FormValue("task_id") // gets the value of task_id
	db.GormDB.Model(models.Task{}).Where("id = ?", taskID).Update("seen", seen)
	return nil
}

// this function gets the parameters of the datatable to send it to `GetPaginatedTasks` function
func (db *MyConfigurations) GetTasks(c echo.Context) error {
	userID, classification := getUserStatus(&c) // gets the value of userID and classification
	hash := c.QueryParam("hash")
	var tasks []models.Task
	var totalNumberOfRowsInDatabase, totalNumberOfRowsAfterFilter int
	q := c.Request().URL.Query() // gets the URL Query as a map
	draw, _ := strconv.Atoi(q["draw"][0])
	if hash != "" {
		tasks, totalNumberOfRowsInDatabase, totalNumberOfRowsAfterFilter = models.GetTask(hash, db.GormDB, classification, userID)
		dt := generateDTOutput(tasks, totalNumberOfRowsInDatabase, totalNumberOfRowsAfterFilter, draw)
		return c.JSONPretty(http.StatusOK, dt, " ")
	}
	start, length, direction, sortedColumnName, descriptionSearch, sentToSearch, minDateSearch, maxDateSearch, retrieveType :=
		getFilterData(q, classification)

	tasks, totalNumberOfRowsInDatabase, totalNumberOfRowsAfterFilter =
		models.GetPaginatedTasks(db.GormDB, start, length,
			sortedColumnName, direction, descriptionSearch, sentToSearch, minDateSearch, maxDateSearch,
			retrieveType, classification, userID)

	dt := generateDTOutput(tasks, totalNumberOfRowsInDatabase, totalNumberOfRowsAfterFilter, draw)

	return c.JSONPretty(http.StatusOK, dt, " ")
}

func getFilterData(q url.Values, classification int) (int, int, string, string, string, string, string, string, string) {
	start, _ := strconv.Atoi(q["start"][0])
	// the start point of the current data set
	length, _ := strconv.Atoi(q["length"][0])
	// number of records to display (page size)
	sortedColumnNumber, _ := strconv.Atoi(q["order[0][column]"][0])
	// column to which ordering should be applied
	direction := q["order[0][dir]"][0]
	// ordering direction for this column
	sprintf := fmt.Sprintf("columns[%d][name]", sortedColumnNumber)
	// gets the name of the sorted column (not the numer)
	sortedColumnName := q[sprintf][0]
	descriptionSearch, sentToSearch, minDateSearch, maxDateSearch, retrieveType := getUserFilterData(q, classification)
	return start, length, direction, sortedColumnName, descriptionSearch, sentToSearch, minDateSearch, maxDateSearch, retrieveType
}

func getUserFilterData(q url.Values, classification int) (string, string, string, string, string) {
	descriptionSearch := q["description"][0]
	// the value of the description search
	sentToSearch := ""
	if classification != 3 {
		sentToSearch = q["sent_to"][0] // the value of the sent_to search
	}
	minDateSearch := q["min_date"][0]
	// the value of min_date search
	maxDateSearch := q["max_date"][0]
	// the value of max_date search
	retrieveType := q["retrieve"][0]
	// the value of the retrieve type
	return descriptionSearch, sentToSearch, minDateSearch, maxDateSearch, retrieveType
}

func generateDTOutput(tasks []models.Task, totalNumberOfRowsInDatabase, totalNumberOfRowsAfterFilter int, draw int) dtOutput {
	dt := dtOutput{
		Draw:            draw,
		RecordsTotal:    totalNumberOfRowsInDatabase,
		RecordsFiltered: totalNumberOfRowsAfterFilter,
		Data:            tasks,
	}
	return dt
}

func (db *MyConfigurations) showTask(c echo.Context) error {
	userID, classification := getUserStatus(&c)
	hash := c.Param("hash")
	username, stringClassification := getUsernameAndClassification(&c)
	var task models.Task
	title := "تكليف جديد"
	buttonText := "تعديل"
	formUrl := "/tasks/edit"
	if hash == "new" {
		if classification != 1 {
			return redirectWithFlashMessage("failure", "ليس لديك الصلاحيه لهذه الصفحه", "/", &c)
		}
		buttonText = "حفظ"
		formUrl = "/tasks/add"
	} else {
		models.PreloadFollowingAndWorkingOnUsers(classification, db.GormDB, userID).
			Preload("Files", func(db *gorm.DB) *gorm.DB {
				return db.Select("id, created_at, updated_at, deleted_at, task_id, content_type, hash, file_name, user_id, extension").
					Order("created_at ASC")
			}).Preload("Files.User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, username")
		}).
			Find(&task, "hash = ?", hash)
		title = "تعديل التكليف"
		db.markTaskAsSeen(classification, task.ID, userID)
	}
	var followingUsers []models.User
	var workingOnUsers []models.User
	// get the followingUsers ordered by the [order] column
	db.GormDB.Preload("FollowingUserTasks").Order("[order] ASC").Find(&followingUsers, "classification = 2")
	// get the workingOnUsers ordered by the [order] column
	db.GormDB.Preload("WorkingOnUserTasks").Order("[order] ASC").Find(&workingOnUsers, "classification = 3")

	for fileNumber, file := range task.Files {
		task.Files[fileNumber].FileDisplay = file.CreatedAt.String()[0:10] + "  رقم:  " + strconv.Itoa(fileNumber+1) +
			" الاسم " + file.FileName + "." + file.Extension
		fmt.Println(task.Files[fileNumber].FileDisplay)
	}

	isClickedFromNotificationPanel := c.Request().Header.Get("from-notification")
	if isClickedFromNotificationPanel != "" {
		isFromNotification, err := strconv.ParseBool(isClickedFromNotificationPanel)
		if isFromNotification && (err == nil) {
			models.MarkNotificationAsClicked(db.GormDB, userID, task.ID)
		}
	}

	valuesToReturn := echo.Map{
		"Task":                 task,
		"buttonText":           buttonText,
		"classification":       classification,
		"title":                title,
		"username":             username,
		"stringClassification": stringClassification,
		"followingUsers":       followingUsers,
		"workingOnUsers":       workingOnUsers,
		"formUrl":              formUrl,
	}
	if checkIfRequestFromMobileDevice(c) {
		return c.JSON(http.StatusOK, valuesToReturn)
	}
	return c.Render(http.StatusOK, "create-edit-task.html", valuesToReturn)
}

func (db *MyConfigurations) markTaskAsSeen(classification int, taskID uint, userID uint) {
	if classification == 1 {
		db.GormDB.Model(models.Task{}).Where("id = ?", taskID).Update("seen", true)
	} else if classification == 2 {
		db.GormDB.Model(models.FollowingUserTask{}).Where("task_id = ? AND user_id = ?", taskID, userID).
			Updates(map[string]interface{}{"seen": true, "marked_as_unseen": false,
				"new_from_minister": false, "new_from_working_on_user": false})
	} else if classification == 3 {
		db.GormDB.Model(models.WorkingOnUserTask{}).Where("task_id = ? AND user_id = ?", taskID, userID).
			Updates(map[string]interface{}{"seen": true, "marked_as_unseen": false})
	}
}
func (db *MyConfigurations) markTaskAsUnseen(c echo.Context) error {
	userID, classification := getUserStatus(&c)
	taskID := c.FormValue("task_id") // gets the value of user_id
	if taskID, _ := strconv.Atoi(taskID); taskID == 0 {
		return c.JSON(http.StatusOK, echo.Map{
			"status":  "failure",
			"message": "يجب حفظ التكليف قبل اعتباره جديد",
		})
	}
	if classification == 1 {
		db.GormDB.Model(models.Task{}).Where("id = ?", taskID).Update("seen", false)
	} else if classification == 2 {
		db.GormDB.Model(models.FollowingUserTask{}).Where("task_id = ? AND user_id = ?", taskID, userID).
			Updates(map[string]interface{}{"seen": true, "marked_as_unseen": true,
				"new_from_minister": false, "new_from_working_on_user": false})
	} else if classification == 3 {
		db.GormDB.Model(models.WorkingOnUserTask{}).Where("task_id = ? AND user_id = ?", taskID, userID).
			Updates(map[string]interface{}{"seen": true, "marked_as_unseen": true})
	}
	return c.JSON(http.StatusOK, echo.Map{
		"status":  "success",
		"message": "تم اعتبار التكليف كأنه جديد",
	})
}

// struct to return the datatable rows in the correct format
type dtOutput struct {
	Draw            int                    `json:"draw"`
	RecordsTotal    int                    `json:"recordsTotal"`
	RecordsFiltered int                    `json:"recordsFiltered"`
	Data            []models.Task          `json:"data"`
	Files           map[string]interface{} `json:"files"`
}

func gatherUsersToSendNotifications(db *MyConfigurations, users []uint, notificationTitle string, task *models.Task) {
	var userDevices []string // get all registered devices for a specific user
	db.GormDB.Table("device_tokens").Where("user_id IN (?) AND deleted_at IS NULL", users).Pluck("token", &userDevices)
	//usersTokens = append(usersTokens, userDevices...)
	sendFirebaseNotificationToMultipleUsers(db.MessagingClient, userDevices, notificationTitle, task)
}
