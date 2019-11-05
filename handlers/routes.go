package handlers

import (
	"github.com/labstack/echo/v4"
)

func InitializeRoutes(e *echo.Echo, db *MyDB) {

	e.GET("/login", db.showLoginPage, ensureNotLoggedIn)
	e.POST("/login", db.performLogin, ensureNotLoggedIn)
	e.GET("/signup", showSignUpPage, ensureNotLoggedIn)
	e.POST("/signup", db.performSignUp, ensureNotLoggedIn)
	e.GET("/reset-password", db.showResetPasswordUpPage, ensureNotLoggedIn)
	e.GET("/reset-password-by-email", db.resetPasswordByEmail, ensureNotLoggedIn)
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
	notifications.POST("/register", db.registerClientToNotify, ensureLoggedInWithoutFlashMessage)
	e.GET("/service-worker.js", serveServiceWorkerFile, ensureLoggedInWithoutFlashMessage)
	e.GET("/js/dataTables.editor.js", serveDataTablesEditorFile, ensureLoggedInWithoutFlashMessage)

	e.GET("/user-settings", db.showSettingsPage, ensureLoggedIn)

	e.GET("/send-verification-code", db.sendVerificationCode, ensureLoggedIn)
	e.POST("/change-phone-number", db.changePhoneNumber, ensureLoggedIn)
	e.POST("/verify-phone-number", db.verifyPhoneNumber, ensureLoggedIn)

	e.GET("/send-verification-link", db.sendVerificationLink, ensureLoggedIn)
	e.POST("/change-email", db.changeEmail, ensureLoggedIn)
	e.GET("/verify-email", db.verifyEmail, ensureLoggedIn)

	e.GET("/email-reset-password", db.showResetPasswordByEmailPage, ensureNotLoggedIn)
	e.POST("/email-reset-password", db.performResetPasswordByEmail, ensureNotLoggedIn)

	e.GET("/change-notifications", db.changeNotifications, ensureLoggedIn)

	e.GET("/generate-pdf", db.generatePDF, ensureLoggedIn)
}
