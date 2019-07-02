package handlers

import (
	"github.com/labstack/echo"
)

func InitializeRoutes(e *echo.Echo, db *MyDB) {

	e.GET("/login", db.showLoginPage, ensureNotLoggedIn)
	e.POST("/login", db.performLogin, ensureNotLoggedIn)
	e.GET("/signup", showSignUpPage, ensureNotLoggedIn)
	e.POST("/signup", db.performSignUp, ensureNotLoggedIn)
	e.GET("/reset-password", db.showResetPasswordUpPage, ensureNotLoggedIn)
	e.POST("/reset-password", db.performResetPassword, ensureNotLoggedIn)

	e.GET("/logout", logout, ensureLoggedIn)
	e.GET("/", db.index, ensureLoggedIn)

	tasks := e.Group("/tasks", ensureLoggedIn)

	tasks.POST("/add", db.AddTask, ensureAdmin)
	tasks.POST("/remove", db.RemoveTask, ensureAdmin)
	tasks.POST("/seen", db.ChangeTaskSeen, ensureAdmin)

	tasks.POST("/edit", db.EditTask)
	tasks.POST("/person/seen", db.ChangeUserSeen)
	tasks.GET("/getData", db.GetTasks)
	tasks.POST("/validate-image", db.validateFile)
	tasks.GET("/file/:hash", db.showFile)

	notifications := e.Group("/notifications")
	notifications.POST("/register", db.registerClientToNotify)
	e.GET("/service-worker.js", serveServiceWorkerFile)
	e.GET("/js/dataTables.editor.js", serveDataTablesEditorFile)

	e.GET("/user-settings", db.showSettingsPage, ensureLoggedIn)

	e.GET("/send-verification-code", db.sendVerificationCode, ensureLoggedIn)
	e.POST("/change-phone-number", db.changePhoneNumber, ensureLoggedIn)
	e.POST("/verify-phone-number", db.verifyPhoneNumber, ensureLoggedIn)

}
