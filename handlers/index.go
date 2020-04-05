package handlers

import (
	"ah-follow-modules/models"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
	"time"
)

// this function serves the index page of the program
func (db *MyConfigurations) index(c echo.Context) error {
	userID, classification := getUserStatus(&c) // get the user ID and the classification int from the cookie that is stored
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
	_ = db.saveDeviceToken(c)
	var user models.User
	db.GormDB.First(&user, userID)

	returnedValues := echo.Map{
		"title":                "الرئيسية",     // sets the title of the page
		"status":               status,         // pass the status of the flash message
		"message":              message,        // pass the message
		"workingOnUsers":       workingOnUsers, // pass the workingOnUsers array
		"classification":       classification, // pass the classification variable to use in JS
		"username":             username,       // pass the username
		"stringClassification": stringClassification,
		"hashExist":            hashExist,
		"email":                user.Email,
		"validEmail":           user.ValidEmail,
	}
	if checkIfRequestFromMobileDevice(c) {
		return c.JSON(http.StatusOK, returnedValues)
	}
	return c.Render(http.StatusOK, "index.html", returnedValues)
}

func (db *MyConfigurations) saveDeviceToken(c echo.Context) error {
	userID, _ := getUserStatus(&c) // get the user ID and the classification int from the cookie that is stored

	stringToken := c.Request().Header.Get("fcm-token")
	deviceToken := models.DeviceToken{
		Token:  stringToken,
		UserID: userID,
	}
	if stringToken != "" {
		if err := db.GormDB.Create(&deviceToken).Error; err != nil {
			db.GormDB.Table("device_tokens").Where("token = ?",
				stringToken).Updates(map[string]interface{}{
				"user_id": userID, "deleted_at": nil, "updated_at": time.Now()})
		}
	}
	return nil
}

func (db *MyConfigurations) GetRecentNotifications(c echo.Context) error {
	userID, _ := getUserStatus(&c) // get the user ID and the classification int from the cookie that is stored
	start, err := strconv.Atoi(c.QueryParam("start"))
	if err != nil {
		return c.JSON(http.StatusPartialContent, echo.Map{})
	}
	length, err := strconv.Atoi(c.QueryParam("length"))
	if err != nil {
		return c.JSON(http.StatusPartialContent, echo.Map{})
	}

	recentNotifications := models.GetRecentNotifications(db.GormDB, userID, start, length)
	countOfNonDismissedNotifications := models.GetNumberOfNonDismissedNotifications(db.GormDB, userID)
	return c.JSON(http.StatusOK, echo.Map{
		"recentNotification":   recentNotifications,
		"countOfNotifications": countOfNonDismissedNotifications,
	})

}

func (db *MyConfigurations) GetRecentNotificationsCount(c echo.Context) error {
	userID, _ := getUserStatus(&c) // get the user ID and the classification int from the cookie that is stored

	countOfNonDismissedNotifications := models.GetNumberOfNonDismissedNotifications(db.GormDB, userID)
	return c.JSON(http.StatusOK, echo.Map{
		"countOfNotifications": countOfNonDismissedNotifications,
	})
}
