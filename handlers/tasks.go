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
		FollowedBy:  c.FormValue("data[followed_by]"),
		People:      []*models.PersonTask{},
	}
	db.GormDB.Create(&taskToSave)

	totalPeople, _ := strconv.Atoi(c.FormValue("data[totalPeople]"))
	//db.GormDB.Preload("People")
	for i := 0; i < totalPeople; i++ {
		id := c.FormValue("data[" + strconv.Itoa(i) + "]")
		uid, _ := strconv.ParseUint(id, 10, 64)
		personTask := models.PersonTask{
			TaskID: taskToSave.ID,
			Task:   &taskToSave,
			UserID: uint(uid),
		}
		db.GormDB.Create(&personTask)
	}

	db.GormDB.Preload("People").Find(&taskToSave, taskToSave.ID)

	dataArray := make([]interface{}, 1)
	dataArray[0] = taskToSave
	datatableTask := datatableTask{dataArray}
	return c.JSONPretty(http.StatusOK, datatableTask, " ")
}

func (db *MyDB) EditTask(c echo.Context) error {
	id, err := strconv.Atoi(c.FormValue("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "Invalid Request",
		})
	}
	updatedValues := models.Task{
		Description: c.FormValue("data[description]"),
		FollowedBy:  c.FormValue("data[followed_by]"),
	}
	var task models.Task
	db.GormDB.First(&task, id)
	db.GormDB.Model(&task).Updates(updatedValues)

	totalPeople, _ := strconv.Atoi(c.FormValue("data[totalPeople]"))

	for i := 1; i <= totalPeople; i++ {
		var person models.Person
		name := "data[name_" + strconv.Itoa(i) + "_repeat]"
		action := "data[action_" + strconv.Itoa(i) + "_repeat]"
		db.GormDB.Where("name = ? AND task_id = ?",
			c.FormValue(name), id).Find(&person)
		if person.ID == 0 {
			models.CreatePerson(db.GormDB, c.FormValue(name), c.FormValue(action), id)
		} else {

			db.GormDB.Save(&person)
		}
	}

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
	followedBySearch := q["followed_by"][0]
	minDateSearch := q["min_date"][0]
	maxDateSearch := q["max_date"][0]
	tasks, totalNumberOfRowsInDatabase, totalNumberOfRowsAfterFilter := models.GetAllTasks(db.GormDB, start, length,
		sortedColumnName, direction, descriptionSearch, followedBySearch, minDateSearch, maxDateSearch)
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
