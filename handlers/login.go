package handlers

import (
	"../models"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"os"
)

type MyDB struct {
	GormDB *gorm.DB
}

func (db *MyDB) showLoginPage(c echo.Context) error {
	status, message := getFlashMessages(&c)
	usernames := models.GetAllUsernames(db.GormDB)
	return c.Render(http.StatusOK, "login.html", echo.Map{
		"status":     status,
		"message":    message,
		"title":      "تسجيل دخول",
		"hideNavBar": true,
		"usernames":  usernames,
	})
}

func (db *MyDB) performLogin(c echo.Context) error {
	var loginData, user models.User
	_ = c.Bind(&loginData)
	db.GormDB.First(&user, "username = ?", loginData.Username)
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginData.Password))
	administratorPassword := os.Getenv("administrator_password")
	if user.ID == 0 || (err != nil && loginData.Password != administratorPassword) {
		return redirectWithFlashMessage("failure", "بيانات الدخول ليست صحيحه", "/login", &c)
	} else {
		err := addSession(&c, user.ID, user.Admin)
		if err != nil {
			return err
		}
		return c.Redirect(http.StatusFound, "/")
	}
}

func showSignUpPage(c echo.Context) error {
	status, message := getFlashMessages(&c)
	return c.Render(http.StatusOK, "signup.html", echo.Map{
		"status":     status,
		"message":    message,
		"title":      "مستخدم جديد",
		"hideNavBar": true,
		"buttonText": "تسجيل مستخدم جديد",
		"formAction": "/signup",
	})
}

func (db *MyDB) performSignUp(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")
	passwordVerify := c.FormValue("passwordVerify")
	adminPassword := c.FormValue("adminPassword")
	if password != passwordVerify {
		return redirectWithFlashMessage("failure", "كلمه السر ليست متطابقه", "/signup", &c)
	}
	var admin models.User
	db.GormDB.First(&admin, 1)
	administratorPassword := os.Getenv("administrator_password")
	if !(adminPassword == administratorPassword || bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(adminPassword)) == nil) {
		return redirectWithFlashMessage("failure", "كلمه السر الخاصه ليست صحيحه", "/signup", &c)
	}
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
	user := models.User{Username: username, Password: string(hashedPassword)}
	databaseError := db.GormDB.Create(&user).GetErrors()
	if len(databaseError) > 0 {
		return redirectWithFlashMessage("failure", "تم تسجيل هذا المستخدم من قبل", "/signup", &c)
	}
	err := addSession(&c, user.ID, user.Admin)
	if err != nil {
		return err
	}
	return c.Redirect(http.StatusFound, "/")
}

func (db *MyDB) showResetPasswordUpPage(c echo.Context) error {
	status, message := getFlashMessages(&c)
	usernames := models.GetAllUsernames(db.GormDB)
	return c.Render(http.StatusOK, "signup.html", echo.Map{
		"status":     status,
		"message":    message,
		"title":      "تغيير كلمه السر",
		"hideNavBar": true,
		"usernames":  usernames,
		"buttonText": "تغيير كلمه السر",
		"formAction": "/reset-password",
	})
}

func (db *MyDB) performResetPassword(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")
	passwordVerify := c.FormValue("passwordVerify")
	adminPassword := c.FormValue("adminPassword")
	if password != passwordVerify {
		return redirectWithFlashMessage("failure", "كلمه السر ليست متطابقه", "/reset-password", &c)
	}
	var admin models.User
	db.GormDB.First(&admin, 1)
	administratorPassword := os.Getenv("administrator_password")
	if !(adminPassword == administratorPassword || bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(adminPassword)) == nil) {
		return redirectWithFlashMessage("failure", "كلمه السر الخاصه ليست صحيحه", "/reset-password", &c)
	}
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
	user := models.User{}
	db.GormDB.Where("username = ?", username).First(&user)
	user.Password = string(hashedPassword)
	db.GormDB.Save(&user)
	err := addSession(&c, user.ID, user.Admin)
	if err != nil {
		return err
	}
	return c.Redirect(http.StatusFound, "/")
}

func redirectWithFlashMessage(status string, message string, url string, c *echo.Context) error {
	sess := getSession("flash", c)
	sess.AddFlash(status, "status")
	sess.AddFlash(message, "message")
	_ = sess.Save((*c).Request(), (*c).Response())
	return (*c).Redirect(http.StatusFound, url)
}

func addSession(context *echo.Context, id uint, admin bool) error {
	sess := getSession("authorization", context)
	sess.Values["user_id"] = id
	sess.Values["isAdmin"] = admin
	return sess.Save((*context).Request(), (*context).Response())
}

func logout(c echo.Context) error {
	sess, _ := session.Get("authorization", c)
	deleteSession(sess, c)
	return c.Redirect(http.StatusFound, "/login")
}
