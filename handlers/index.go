package handlers

import (
	"github.com/labstack/echo"
	"net/http"
)

func (db *MyDB) index(c echo.Context) error {

	status, message := getFlashMessages(&c)
	return c.Render(http.StatusOK, "index.html", echo.Map{
		"title":   "الرأيسيه",
		"status":  status,
		"message": message,
	})
}
