package handlers

import (
	"github.com/labstack/echo"
)

func InitializeRoutes(e *echo.Echo, db *MyDB) {

	e.GET("/login", db.showLoginPage, ensureNotLoggedIn)
	e.POST("/login", db.performLogin, ensureNotLoggedIn)

	e.GET("/logout", logout, ensureLoggedIn)
	e.GET("/", db.index, ensureLoggedIn)

	e.POST("/tasksHandler", db.HandleTasks)
	e.GET("/getData", db.GetTasks)
}
