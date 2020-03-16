package handlers

import (
	"ah-follow-modules/models"
	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/labstack/echo/v4"
	"strconv"
)

func (db *MyDB) generatePDF(c echo.Context) error {
	wkhtmltopdf.SetPath("C:/Program Files/wkhtmltopdf/bin/wkhtmltopdf.exe")

	r := NewRequestPdf("")

	userID, classification := getUserStatus(&c) // gets the value of userID and classification
	q := c.Request().URL.Query()                // gets the URL Query as a map
	descriptionSearch, sentToSearch, minDateSearch, maxDateSearch, retrieveType := getUserFilterData(q, classification)
	sortedColumnName := q["sort_column"][0]
	direction := q["sort_direction"][0]
	collapsed, _ := strconv.ParseBool(q["collapsed"][0])

	tasks, totalNumberOfRowsInDatabase, totalNumberOfRowsAfterFilter := models.GetAllTasks(db.GormDB, sortedColumnName,
		direction, descriptionSearch, sentToSearch, minDateSearch, maxDateSearch, retrieveType, classification, userID)
	//html template path
	templatePath := "static/reports/pdf-report.html"
	templateName := "pdf-report.html"
	reportDescription := "هذا التقرير خاص ب" + generateReportDescription(retrieveType, classification)

	//html template data
	templateData := struct {
		Collapsed                    bool
		Description                  string
		Tasks                        []models.Task
		TotalNumberOfRowsInDatabase  int
		TotalNumberOfRowsAfterFilter int
	}{
		Collapsed:                    collapsed,
		Description:                  reportDescription,
		Tasks:                        tasks,
		TotalNumberOfRowsInDatabase:  totalNumberOfRowsInDatabase,
		TotalNumberOfRowsAfterFilter: totalNumberOfRowsAfterFilter,
	}

	if err := r.ParseTemplate(templateName, templatePath, templateData); err == nil {
		pdfg, _ := r.GeneratePDF()
		c.Response().Header().Set("Content-Type", "application/pdf")
		c.Response().Header().Set("content-disposition", "inline;filename=")
		c.Response().Header().Set("Cache-control", "must-revalidate, post-check=0, pre-check=0")
		_, _ = c.Response().Write(pdfg.Bytes())
		c.Response().Flush()
	}
	return nil
}

func generateReportDescription(retrieveType string, classification int) string {
	if retrieveType == "unseen" && classification == 1 {
		return "التكليفات الجديده التي لها اجراء نهائي"
	}
	if retrieveType == "seen" && classification == 1 {
		return "التكليفات اللتي تمت رؤيتها من قبل ولها اجراء نهائي"
	}
	if retrieveType == "replied" {
		return "جميع التكليفات التي لها اجراء نهائي"
	}
	if retrieveType == "unseen" {
		return "التكليفات الجديده"
	}
	if retrieveType == "seen" {
		return "التكليفات اللتي تمت رؤيتها من قبل"
	}
	if retrieveType == "all" {
		return "جميع التكليفات"
	}
	if retrieveType == "notRepliedByAll" {
		return "التكليفات التي لم يرد عليها كل القائمين به"
	}
	if retrieveType == "nonReplied" {
		return "التكليفات التي ليس لها اجراء نهائي"
	}
	if retrieveType == "newFromWorkingOnUsers" {
		return "التكليفات الجديده من القائمين به"
	}
	if retrieveType == "notFinished" {
		return "التكليفات التي لم تعتبر اجراء نهائي"
	}
	return ""
}
