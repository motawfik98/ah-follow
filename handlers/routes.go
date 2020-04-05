package handlers

import (
	"github.com/labstack/echo/v4"
)

func InitializeRoutes(e *echo.Echo, myConfigurations *MyConfigurations) {
	e.HTTPErrorHandler = customHTTPErrorHandler

	e.GET("/login", myConfigurations.showLoginPage, ensureNotLoggedIn)
	e.POST("/login", myConfigurations.performLogin, ensureNotLoggedIn)
	e.GET("/signup", showSignUpPage, ensureNotLoggedIn)
	e.POST("/signup", myConfigurations.performSignUp, ensureNotLoggedIn)
	e.GET("/reset-password", myConfigurations.showResetPasswordUpPage, ensureNotLoggedIn)
	e.POST("/reset-password-by-email", myConfigurations.resetPasswordByEmail, ensureNotLoggedIn)
	e.POST("/reset-password", myConfigurations.performResetPassword, ensureLoggedIn)

	e.GET("/logout", myConfigurations.logout, ensureLoggedIn)
	e.GET("/", myConfigurations.index, ensureLoggedIn)
	e.POST("/save-token", myConfigurations.saveDeviceToken, ensureLoggedIn)

	tasks := e.Group("/tasks", ensureLoggedIn)

	tasks.POST("/add", myConfigurations.AddTask, ensureLoggedIn, ensureAdmin)
	tasks.POST("/remove", myConfigurations.RemoveTask, ensureLoggedIn, ensureAdmin)

	tasks.POST("/edit", myConfigurations.EditTask)
	tasks.GET("/getData", myConfigurations.GetTasks)

	tasks.GET("/file/:hash", myConfigurations.showFile)
	tasks.GET("/task/:hash/:time-stamp", myConfigurations.showTask)
	tasks.POST("/task/mark-as-unseen", myConfigurations.markTaskAsUnseen)

	notifications := e.Group("/notifications")
	notifications.GET("/get-recent", myConfigurations.GetRecentNotifications, ensureLoggedIn)
	notifications.GET("/get-count", myConfigurations.GetRecentNotificationsCount, ensureLoggedIn)
	notifications.GET("/dismiss-all", myConfigurations.dismissNotifications, ensureLoggedIn)
	notifications.POST("/remove", myConfigurations.removeNotification, ensureLoggedIn)
	notifications.POST("/register", myConfigurations.registerClientToNotify, ensureLoggedInWithoutFlashMessage)
	e.GET("/service-worker.js", serveServiceWorkerFile, ensureLoggedInWithoutFlashMessage)

	e.GET("/user-settings", myConfigurations.showSettingsPage, ensureLoggedIn)

	//e.GET("/send-verification-code", myConfigurations.sendVerificationCode, ensureLoggedIn)
	//e.POST("/change-phone-number", myConfigurations.changePhoneNumber, ensureLoggedIn)
	//e.POST("/verify-phone-number", myConfigurations.verifyPhoneNumber, ensureLoggedIn)

	e.GET("/send-verification-link", myConfigurations.sendVerificationLink, ensureLoggedIn)
	e.POST("/change-email", myConfigurations.changeEmail, ensureLoggedIn)
	e.GET("/verify-email", myConfigurations.verifyEmail)

	e.GET("/email-reset-password", myConfigurations.showResetPasswordByEmailPage, ensureNotLoggedIn)
	e.POST("/email-reset-password", myConfigurations.performResetPasswordByEmail, ensureNotLoggedIn)
	e.GET("/change-password", myConfigurations.showResetPasswordPage, ensureLoggedIn)

	e.GET("/change-notifications", myConfigurations.changeNotifications, ensureLoggedIn)

	//e.GET("/send-trial-notification", sendFirebaseNotification)

	//e.GET("/generate-pdf", myConfigurations.generatePDF, ensureLoggedIn)
}
