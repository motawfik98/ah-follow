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
		SentTo:      c.FormValue("data[sent_to]"),
		FollowedBy:  c.FormValue("data[followed_by]"),
		ActionTaken: c.FormValue("data[action_taken]"),
	}
	db.GormDB.Create(&taskToSave)
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
		SentTo:      c.FormValue("data[sent_to]"),
		FollowedBy:  c.FormValue("data[followed_by]"),
		ActionTaken: c.FormValue("data[action_taken]"),
	}
	var task models.Task
	db.GormDB.First(&task, id)
	db.GormDB.Model(&task).Updates(updatedValues)
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
	db.GormDB.Delete(models.Task{}, "id = ?", id)
	return c.JSON(http.StatusOK, models.Task{})
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
	searchValue := q["search[value]"][0]
	fmt.Println(direction, sprintf, searchValue, sortedColumnName)
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
