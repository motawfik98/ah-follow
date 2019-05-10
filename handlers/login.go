package handlers

import (
	"../models"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

type MyDB struct {
	GormDB *gorm.DB
}

func (db *MyDB) showLoginPage(c echo.Context) error {
	status, message := getFlashMessages(&c)
	return c.Render(http.StatusOK, "login.html", echo.Map{
		"status":     status,
		"message":    message,
		"title":      "تسجيل دخول",
		"hideNavBar": true,
	})
}

func (db *MyDB) performLogin(c echo.Context) error {
	var loginData, user models.User
	_ = c.Bind(&loginData)
	db.GormDB.Where("username = ?",
		loginData.Username).First(&user)
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginData.Password))
	if user.ID == 0 || err != nil {
		sess := getSession("flash", &c)
		sess.AddFlash("failure", "status")
		sess.AddFlash("يانات الدخول ليست صحيحه", "message")
		_ = sess.Save(c.Request(), c.Response())
		return c.Redirect(http.StatusFound, "/login")
	} else {
		err := addSession(&c, user.ID)
		if err != nil {
			return err
		}
		return c.Redirect(http.StatusFound, "/")
	}
}

func addSession(context *echo.Context, id uint) error {
	sess := getSession("authorization", context)
	sess.Values["user_id"] = id
	return sess.Save((*context).Request(), (*context).Response())
}

func logout(c echo.Context) error {
	sess, _ := session.Get("authorization", c)
	deleteSession(sess, c)
	return c.Redirect(http.StatusFound, "/login")
}
