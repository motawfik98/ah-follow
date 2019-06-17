package handlers

import (
	"../models"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"os"
	"strconv"
)

type MyDB struct {
	GormDB *gorm.DB
}

// this function serves the login page
func (db *MyDB) showLoginPage(c echo.Context) error {
	status, message := getFlashMessages(&c) // gets the flash message and status if there was any
	users := models.GetAllUsers(db.GormDB)  // gets all the users that are in the database
	return c.Render(http.StatusOK, "login.html", echo.Map{
		"status":     status,       // pass the status of the flash message
		"message":    message,      // pass the message
		"title":      "تسجيل دخول", // the title of the page
		"hideNavBar": true,         // boolean to indicate weather or not the NavBar should be displayed
		"users":      users,        // pass the users array to display to the user
	})
}

// this function performs the login logic
func (db *MyDB) performLogin(c echo.Context) error {
	var loginData, user models.User
	_ = c.Bind(&loginData)                                                                  // gets the form data from the context and binds it to the `loginData` struct
	db.GormDB.First(&user, "username = ?", loginData.Username)                              // gets the user from the database where his username is equal to the entered username
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginData.Password)) // compare the hashed password that is stored in the database with the hashed version of the password that the user entered
	administratorPassword := os.Getenv("administrator_password")                            // gets the administrator password that could login to any account
	// checks if the user ID is 0 (which means that no user was found with that username)
	// checks that err is not null (which means that the hashed password is the same of the hashed version of the user entered password)
	// makes sure that the password that the user entered is not the administrator password
	if user.ID == 0 || (err != nil && loginData.Password != administratorPassword) {
		// redirect to /login and add a failure flash message
		return redirectWithFlashMessage("failure", "بيانات الدخول ليست صحيحه", "/login", &c)
	} else {
		// login successfully, add cookie to browser
		err := addSession(&c, user.ID, user.Classification)
		if err != nil {
			return err
		}
		// redirect the user to the index page
		return c.Redirect(http.StatusFound, "/")
	}
}

// this function serves the signUp page
func showSignUpPage(c echo.Context) error {
	status, message := getFlashMessages(&c) // gets the flash message and status if there was any
	return c.Render(http.StatusOK, "signup.html", echo.Map{
		"status":                   status,                                                                 // pass the status of the flash message
		"message":                  message,                                                                // pass the message
		"title":                    "مستخدم جديد",                                                          // the title of the page
		"hideNavBar":               true,                                                                   // boolean to indicate weather or not the NavBar should be displayed
		"buttonText":               "تسجيل مستخدم جديد",                                                    // the action button text the should be displayed to the user
		"formAction":               "/signup",                                                              // the URL that the form should be submitted to
		"adminPasswordPlaceholder": "كلمه السر الخاصه بالوزير",                                             // string that should be shown in the admin password input placeholder
		"adminPasswordHelp":        "هذا الحقل خاص بالوزير, ولا يمكن ان تضيف مستخدم جديد الا بالرجوع اليه", // some helper text for the admin password field
		"isSignUp":                 true,
	})
}

// this function performs the signUp logic
func (db *MyDB) performSignUp(c echo.Context) error {
	username := c.FormValue("username")                              // gets the username from the form submitted data
	password := c.FormValue("password")                              // gets the password from the form submitted data
	classification, _ := strconv.Atoi(c.FormValue("classification")) // gets the classification from the form submitted data
	passwordVerify := c.FormValue("passwordVerify")                  // gets the password verification from the form submitted data
	adminPassword := c.FormValue("adminPassword")                    // gets the admin's password (or administrator's password) from the form submitted data
	if password != passwordVerify {                                  // checks that the password is equal to the password verification
		// if not, redirect to /signup with failure flash message
		return redirectWithFlashMessage("failure", "كلمه السر ليست متطابقه", "/signup", &c)
	}
	var admin models.User
	db.GormDB.First(&admin, "classification = 1")                // gets from the database where `admin` column is set to one
	administratorPassword := os.Getenv("administrator_password") // gets the administrator's password
	// checks if the adminPassword field is equal to the administrator's password OR its hash is equal to the one stored in the database
	if !(adminPassword == administratorPassword || bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(adminPassword)) == nil) {
		// if not, redirect the user to /signup with a failure flash message
		return redirectWithFlashMessage("failure", "كلمه السر الخاصه ليست صحيحه", "/signup", &c)
	}
	// all conditions are met and we're ready to store the user to the database
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), 10) // hash the password that the user entered
	// set the username and hashed password that the user entered to the `user` struct to save into the database
	user := models.User{Username: username, Password: string(hashedPassword), Classification: classification}
	databaseError := db.GormDB.Create(&user).GetErrors() // try saving the `user` struct
	if len(databaseError) > 0 {                          // checks for database errors
		// if found, then it mainly will be because of the unique key index of the username
		// redirect the user to /signup with failure flash message
		return redirectWithFlashMessage("failure", "تم تسجيل هذا المستخدم من قبل", "/signup", &c)
	}
	// if we reached here, then the user is successfully signed up and he's ready to sign in
	err := addSession(&c, user.ID, user.Classification) // add cookie to browser
	if err != nil {
		return err
	}
	return c.Redirect(http.StatusFound, "/") // redirect the user to the home page
}

// this function serves the resetPassword page
func (db *MyDB) showResetPasswordUpPage(c echo.Context) error {
	status, message := getFlashMessages(&c) // gets the flash message and status if there was any
	users := models.GetAllUsers(db.GormDB)  // gets all the users that are in the database
	return c.Render(http.StatusOK, "signup.html", echo.Map{
		"status":                   status,                                                                        // pass the status of the flash message
		"message":                  message,                                                                       // pass the message
		"title":                    "تغيير كلمه السر",                                                             // the title of the page
		"hideNavBar":               true,                                                                          // boolean to indicate weather or not the NavBar should be displayed
		"users":                    users,                                                                         // pass the users array to display to the user
		"buttonText":               "تغيير كلمه السر",                                                             // the action button text the should be displayed to the user
		"formAction":               "/reset-password",                                                             // the URL that the form should be submitted to
		"adminPasswordPlaceholder": "كلمه السر الخاصه بالوزير (او القديمه)",                                       // string that should be shown in the admin password input placeholder
		"adminPasswordHelp":        "يجب ادخال كلمه السر الخاصه بالوزير (او كلمه السر القديمه) لتتمكن من تغييرها", // some helper text for the admin password field
	})
}

func (db *MyDB) performResetPassword(c echo.Context) error {
	username := c.FormValue("username")             // gets the username from the form submitted data
	password := c.FormValue("password")             // gets the password from the form submitted data
	passwordVerify := c.FormValue("passwordVerify") // gets the password verification from the form submitted data
	adminPassword := c.FormValue("adminPassword")   // gets the admin's password (or administrator's password) from the form submitted data
	if password != passwordVerify {                 // checks that the password is equal to the password verification
		// if not, redirect to /signup with failure flash message
		return redirectWithFlashMessage("failure", "كلمه السر ليست متطابقه", "/reset-password", &c)
	}
	var admin, user models.User
	db.GormDB.First(&admin, 1)
	db.GormDB.Where("username = ?", username).First(&user) // gets the user where the username is equal to the entered username by the end user
	administratorPassword := os.Getenv("administrator_password")
	if !(adminPassword == administratorPassword ||
		bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(adminPassword)) == nil ||
		bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(adminPassword)) == nil) {
		return redirectWithFlashMessage("failure", "كلمه السر الخاصه ليست صحيحه", "/reset-password", &c)
	}
	// all conditions are met and we're ready to change the user's password
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), 10) // hash the password that the user entered
	user.Password = string(hashedPassword)                                 // sets the password to the hashed password
	db.GormDB.Save(&user)                                                  // update the user in the database
	err := addSession(&c, user.ID, user.Classification)                    // add a cookie to the browser to log the user in
	if err != nil {
		return err
	}
	return c.Redirect(http.StatusFound, "/") // redirect to the index page
}

// this function redirect the user to a url with flash status and message
func redirectWithFlashMessage(status string, message string, url string, c *echo.Context) error {
	sess := getSession("flash", c)
	sess.AddFlash(status, "status")
	sess.AddFlash(message, "message")
	_ = sess.Save((*c).Request(), (*c).Response())
	return (*c).Redirect(http.StatusFound, url)
}

// this function adds a cookie to the browser, adding in it the user_id and weather or not he's an admin
func addSession(context *echo.Context, id uint, classification int) error {
	sess := getSession("authorization", context)
	sess.Values["user_id"] = id
	sess.Values["classification"] = classification
	return sess.Save((*context).Request(), (*context).Response())
}

// this function removes the cookie that was added to the browser and log the user out
func logout(c echo.Context) error {
	sess, _ := session.Get("authorization", c)
	deleteSession(sess, c)
	return c.Redirect(http.StatusFound, "/login")
}
