package handlers

import (
	"../models"
	"fmt"
	"github.com/labstack/echo"
	"net/http"
	"strconv"
)

type datatableTask struct {
	Data []interface{} `json:"data"`
}

func (db *MyDB) AddTask(c echo.Context) error {
	taskToSave := models.Task{
		Description: c.FormValue("data[description]"),
	}
	db.GormDB.Create(&taskToSave)

	totalUsers, _ := strconv.Atoi(c.FormValue("data[totalUsers]"))
	for i := 0; i < totalUsers; i++ {
		id := c.FormValue("data[users_" + strconv.Itoa(i) + "]")
		uid, _ := strconv.ParseUint(id, 10, 64)
		models.CreateUserTask(db.GormDB, taskToSave.ID, uint(uid))
	}

	totalPeople, _ := strconv.Atoi(c.FormValue("data[totalPeople]"))
	for i := 0; i < totalPeople; i++ {
		models.CreatePerson(db.GormDB, c.FormValue("data[people_name_"+strconv.Itoa(i)+"]"),
			c.FormValue("data[people_action_"+strconv.Itoa(i)+"]"), taskToSave.ID)
	}

	db.GormDB.Preload("Users").Preload("People").Find(&taskToSave, taskToSave.ID)

	dataArray := make([]interface{}, 1)
	dataArray[0] = taskToSave
	datatableTask := datatableTask{dataArray}
	return c.JSONPretty(http.StatusOK, datatableTask, " ")
}

func (db *MyDB) EditTask(c echo.Context) error {
	taskID, err := strconv.Atoi(c.FormValue("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "Invalid Request",
		})
	}
	description := c.FormValue("data[description]")
	finalAction := c.FormValue("data[final_action]")

	var task models.Task
	db.GormDB.First(&task, taskID)
	db.GormDB.Model(&task).UpdateColumn("description", description)
	if finalAction != task.FinalAction {
		db.GormDB.Model(&task).Updates(map[string]interface{}{"final_action": finalAction, "seen": false})
	}

	totalUsers, _ := strconv.Atoi(c.FormValue("data[totalUsers]"))
	var ids []int
	for i := 0; i < totalUsers; i++ {
		var userTask models.UserTask

		id := c.FormValue("data[users_" + strconv.Itoa(i) + "]")
		uid, _ := strconv.ParseUint(id, 10, 64)
		ids = append(ids, int(uid))
		db.GormDB.Where("task_id = ? AND user_id = ?", taskID, uid).Find(&userTask)
		if userTask.TaskID == 0 && userTask.UserID == 0 {
			models.CreateUserTask(db.GormDB, uint(taskID), uint(uid))
		}
	}
	if ids == nil {
		ids = []int{0}
	}
	db.GormDB.Delete(models.UserTask{}, "task_id = ? AND user_id NOT IN (?)", taskID, ids)

	totalPeople, _ := strconv.Atoi(c.FormValue("data[totalPeople]"))
	var peopleIDs []int
	for i := 0; i < totalPeople; i++ {
		var person models.Person
		name := c.FormValue("data[people_name_" + strconv.Itoa(i) + "]")
		action := c.FormValue("data[people_action_" + strconv.Itoa(i) + "]")
		if name == "" {
			continue
		}
		db.GormDB.Where("name = ? AND task_id = ?", name, taskID).Find(&person)
		var personID int
		if person.ID == 0 {
			personID = models.CreatePerson(db.GormDB, name, action, uint(taskID))
		} else {
			person.ActionTaken = action
			db.GormDB.Save(&person)
			personID, _ = strconv.Atoi(c.FormValue("data[people_id_" + strconv.Itoa(i) + "]"))
		}
		peopleIDs = append(peopleIDs, personID)
	}
	if peopleIDs == nil {
		peopleIDs = []int{0}
	}
	db.GormDB.Delete(models.Person{}, "task_id = ? AND id NOT IN (?)", taskID, peopleIDs)

	dataArray := make([]interface{}, 1)
	dataArray[0] = task
	datatableTask := datatableTask{dataArray}
	return c.JSONPretty(http.StatusOK, datatableTask, " ")
}

func (db *MyDB) RemoveTask(c echo.Context) error {
	id, err := strconv.Atoi(c.FormValue("id[]"))
	if err != nil || id == 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "Invalid Request",
		})
	}
	var task models.Task
	db.GormDB.First(&task, id)
	task.DeleteChildren(db.GormDB)
	db.GormDB.Delete(&task)
	return c.JSON(http.StatusOK, models.Task{})
}

func (db *MyDB) RemoveChild(c echo.Context) error {
	personId, _ := strconv.Atoi(c.FormValue("id"))
	db.GormDB.Delete(models.Person{}, "id = ?", personId)
	return nil
}

func (db *MyDB) ChangePersonSeen(c echo.Context) error {
	seen := c.FormValue("seen")
	taskID := c.FormValue("task_id")
	userID := c.FormValue("user_id")
	db.GormDB.Model(models.UserTask{}).Where("task_id = ? AND user_id = ?", taskID, userID).Update("seen", seen)
	return nil
}

func (db *MyDB) ChangeTaskSeen(c echo.Context) error {
	seen := c.FormValue("seen")
	taskID := c.FormValue("task_id")
	db.GormDB.Model(models.Task{}).Where("id = ?", taskID).Update("seen", seen)
	return nil
}

func (db *MyDB) GetTasks(c echo.Context) error {
	q := c.Request().URL.Query()
	draw, _ := strconv.Atoi(q["draw"][0])
	start, _ := strconv.Atoi(q["start"][0])
	length, _ := strconv.Atoi(q["length"][0])
	sortedColumnNumber, _ := strconv.Atoi(q["order[0][column]"][0])
	direction := q["order[0][dir]"][0]
	sprintf := fmt.Sprintf("columns[%d][name]", sortedColumnNumber)
	sortedColumnName := q[sprintf][0]
	descriptionSearch := q["description"][0]
	sentToSearch := q["sent_to"][0]
	minDateSearch := q["min_date"][0]
	maxDateSearch := q["max_date"][0]
	retrieveType := q["retrieve"][0]
	sess := getSession("authorization", &c)
	admin := sess.Values["isAdmin"].(bool)
	userID := sess.Values["user_id"].(uint)
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

type dtOutput struct {
	Draw            int           `json:"draw"`
	RecordsTotal    int           `json:"recordsTotal"`
	RecordsFiltered int           `json:"recordsFiltered"`
	Data            []models.Task `json:"data"`
}
