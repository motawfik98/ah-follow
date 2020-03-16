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

	tasks.POST("/edit", db.EditTask)
	tasks.GET("/getData", db.GetTasks)

	tasks.GET("/file/:hash", db.showFile)
	tasks.GET("/task/:hash/:time-stamp", db.showTask)
	tasks.POST("/task/mark-as-unseen", db.markTaskAsUnseen)

	notifications := e.Group("/notifications")
	notifications.POST("/register", db.registerClientToNotify, ensureLoggedInWithoutFlashMessage)
	e.GET("/service-worker.js", serveServiceWorkerFile, ensureLoggedInWithoutFlashMessage)

	e.GET("/user-settings", db.showSettingsPage, ensureLoggedIn)

	//e.GET("/send-verification-code", db.sendVerificationCode, ensureLoggedIn)
	//e.POST("/change-phone-number", db.changePhoneNumber, ensureLoggedIn)
	//e.POST("/verify-phone-number", db.verifyPhoneNumber, ensureLoggedIn)

	e.GET("/send-verification-link", db.sendVerificationLink, ensureLoggedIn)
	e.POST("/change-email", db.changeEmail, ensureLoggedIn)
	e.GET("/verify-email", db.verifyEmail, ensureLoggedIn)

	e.GET("/email-reset-password", db.showResetPasswordByEmailPage, ensureNotLoggedIn)
	e.POST("/email-reset-password", db.performResetPasswordByEmail, ensureNotLoggedIn)

	e.GET("/change-notifications", db.changeNotifications, ensureLoggedIn)

	//e.GET("/generate-pdf", db.generatePDF, ensureLoggedIn)
}
