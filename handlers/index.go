package handlers

import (
	"ah-follow-modules/models"
	"github.com/labstack/echo/v4"
	"net/http"
)

// this function serves the index page of the program
func (db *MyDB) index(c echo.Context) error {
	_, classification := getUserStatus(&c) // get the user ID and the classification int from the cookie that is stored
	username, stringClassification := getUsernameAndClassification(&c)
	status, message := getFlashMessages(&c) // gets the flash message and status if there was any
	var workingOnUsers []models.User
	// get the workingOnUsers ordered by the [order] column
	db.GormDB.Preload("WorkingOnUserTasks").Order("[order] ASC").Find(&workingOnUsers, "classification = 3")

	hash := c.QueryParam("hash")
	hashExist := true
	if hash == "" {
		hashExist = false
	}
	return c.Render(http.StatusOK, "index.html", echo.Map{
		"title":                "الرئيسية",     // sets the title of the page
		"status":               status,         // pass the status of the flash message
		"message":              message,        // pass the message
		"workingOnUsers":       workingOnUsers, // pass the workingOnUsers array
		"classification":       classification, // pass the classification variable to use in JS
		"username":             username,       // pass the username
		"stringClassification": stringClassification,
		"hashExist":            hashExist,
	})
}
