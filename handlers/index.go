package handlers

import (
	"../models"
	"github.com/labstack/echo"
	"net/http"
)

func (db *MyDB) index(c echo.Context) error {
	sess := getSession("authorization", &c)
	userID := sess.Values["user_id"].(uint)
	admin := sess.Values["isAdmin"].(bool)
	var username string
	status, message := getFlashMessages(&c)
	var users []models.User
	db.GormDB.Preload("Tasks").Order("[order] ASC").Find(&users)
	for _, element := range users {
		if element.ID == userID {
			username = element.Username
		}
	}
	users = users[1:]
	return c.Render(http.StatusOK, "index.html", echo.Map{
		"title":    "الرئيسية",
		"status":   status,
		"message":  message,
		"users":    users,
		"admin":    admin,
		"userID":   userID,
		"username": username,
	})
}

func serveDataTablesEditorFile(context echo.Context) error {
	context.Response().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
	context.Response().Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
	context.Response().Header().Set("Expires", "0")                                         // Proxies.
	return context.File("static/js/dataTables.editor.js")
}
