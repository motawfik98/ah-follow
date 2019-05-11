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

func (db *MyDB) HandleTasks(c echo.Context) error {

	switch c.FormValue("action") {
	case "create":
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
	case "edit":
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
	case "remove":
		id, err := strconv.Atoi(c.FormValue("id[]"))
		if err != nil || id == 0 {
			return c.JSON(http.StatusBadRequest, echo.Map{
				"message": "Invalid Request",
			})
		}
		db.GormDB.Delete(models.Task{}, "id = ?", id)
		return c.JSON(http.StatusOK, models.Task{})
	default:
		return nil
	}
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
	tasks, totalNumberOfRows := models.GetAllTasks(db.GormDB, start, length, sortedColumnName, direction)
	dt := dtOutput{
		Draw:            draw,
		RecordsTotal:    totalNumberOfRows,
		RecordsFiltered: totalNumberOfRows,
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
