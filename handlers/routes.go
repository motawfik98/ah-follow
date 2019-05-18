package handlers

import (
	"github.com/labstack/echo"
)

func InitializeRoutes(e *echo.Echo, db *MyDB) {

	e.GET("/login", db.showLoginPage, ensureNotLoggedIn)
	e.POST("/login", db.performLogin, ensureNotLoggedIn)

	e.GET("/logout", logout, ensureLoggedIn)
	e.GET("/", db.index, ensureLoggedIn)

	e.POST("/tasks/add", db.AddTask)
	e.POST("/tasks/edit", db.EditTask)
	e.POST("/tasks/remove", db.RemoveTask)
	e.POST("/tasks/removeChild", db.RemoveChild)
	e.POST("/tasks/seen", db.ChangeSeen)
	e.GET("/getData", db.GetTasks)
}
