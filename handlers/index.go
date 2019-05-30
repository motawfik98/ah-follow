package handlers

import (
	"../models"
	"github.com/labstack/echo"
	"net/http"
)

// this function serves the index page of the program
func (db *MyDB) index(c echo.Context) error {
	userID, admin := getUserStatus(&c) // get the user ID and the admin bool from the cookie that is stored
	var username string
	status, message := getFlashMessages(&c) // gets the flash message and status if there was any
	var users []models.User
	db.GormDB.Preload("Tasks").Order("[order] ASC").Find(&users) // get the users ordered by the [order] column
	for _, element := range users {                              // sets the `username` variable to the current user
		if element.ID == userID { // checks the user ID
			username = element.Username // gets the username
		}
	}
	users = users[1:] // remove the first element from the users array (to display them in the UserTask section)
	return c.Render(http.StatusOK, "index.html", echo.Map{
		"title":    "الرئيسية", // sets the title of the page
		"status":   status,     // pass the status of the flash message
		"message":  message,    // pass the message
		"users":    users,      // pass the users array
		"admin":    admin,      // pass the isAdmin variable to use in JS
		"userID":   userID,     // pass the userID
		"username": username,   // pass the username
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
