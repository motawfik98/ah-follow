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
	status, message := getFlashMessages(&c)
	var users []models.User
	db.GormDB.Preload("Tasks").Find(&users)
	return c.Render(http.StatusOK, "index.html", echo.Map{
		"title":   "الرأيسيه",
		"status":  status,
		"message": message,
		"users":   users,
		"admin":   admin,
		"userID":  userID,
	})
}
