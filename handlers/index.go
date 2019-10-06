package handlers

import (
	"ah-follow-modules/models"
	"github.com/labstack/echo/v4"
	"net/http"
)

// this function serves the index page of the program
func (db *MyDB) index(c echo.Context) error {
	userID, classification := getUserStatus(&c) // get the user ID and the classification int from the cookie that is stored
	username, stringClassification := getUsernameAndClassification(&c)
	status, message := getFlashMessages(&c) // gets the flash message and status if there was any
	var followingUsers []models.User
	var workingOnUsers []models.User
	// get the followingUsers ordered by the [order] column
	db.GormDB.Preload("FollowingUserTasks").Order("[order] ASC").Find(&followingUsers, "classification = 2")
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
		"followingUsers":       followingUsers, // pass the followingUsers array
		"workingOnUsers":       workingOnUsers, // pass the workingOnUsers array
		"classification":       classification, // pass the classification variable to use in JS
		"userID":               userID,         // pass the userID
		"username":             username,       // pass the username
		"stringClassification": stringClassification,
		"hashExist":            hashExist,
	})
}

// removes the cache of the dataTables editor file to be able to change it and the user won't feel any difference
// (as the trial ends each 15 days)
func serveDataTablesEditorFile(context echo.Context) error {
	context.Response().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
	context.Response().Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
	context.Response().Header().Set("Expires", "0")                                         // Proxies.
	return context.File("static/js/dataTables.editor.js")
}
