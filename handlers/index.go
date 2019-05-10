package handlers

import (
	"github.com/labstack/echo"
	"net/http"
)

func index(c echo.Context) error {
	status, message := getFlashMessages(&c)
	return c.Render(http.StatusOK, "index.html", echo.Map{
		"title":   "الرأييسيه",
		"status":  status,
		"message": message,
	})
}
