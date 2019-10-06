package handlers

import (
	"ah-follow-modules/models"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strconv"
	"strings"
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
	// checks if the user ID is 0 (which means that no user was found with that username)
	// checks that err is not null (which means that the hashed password is the same of the hashed version of the user entered password)
	// makes sure that the password that the user entered is not the administrator password
	if user.ID == 0 || (err != nil && loginData.Password != administratorPassword) {
		// redirect to /login and add a failure flash message
		return redirectWithFlashMessage("failure", "بيانات الدخول ليست صحيحه", "/login", &c)
	} else {
		// login successfully, add cookie to browser
		err := addSession(&c, user.ID, user.Classification, user.Username)
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
	formFields := getFormData(&c, []string{"username", "classification"})
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
		"username":                 formFields["username"],
		"classification":           formFields["classification"],
	})
}

// this function performs the signUp logic
func (db *MyDB) performSignUp(c echo.Context) error {
	username := strings.TrimSpace(c.FormValue("username"))             // gets the username from the form submitted data
	password := strings.TrimSpace(c.FormValue("password"))             // gets the password from the form submitted data
	classification, _ := strconv.Atoi(c.FormValue("classification"))   // gets the classification from the form submitted data
	passwordVerify := strings.TrimSpace(c.FormValue("passwordVerify")) // gets the password verification from the form submitted data
	adminPassword := strings.TrimSpace(c.FormValue("adminPassword"))   // gets the admin's password (or administrator's password) from the form submitted data
	if len(username) == 0 || len(password) == 0 {
		return redirectWithFormData([]string{"username", "classification"}, []string{username, string(classification)},
			"/signup", "failure", "يجب تحديد اسم المستخدم وكلمه السر", &c)
	}
	if password != passwordVerify { // checks that the password is equal to the password verification
		// if not, redirect to /signup with failure flash message
		return redirectWithFormData([]string{"username", "classification"}, []string{username, string(classification)},
			"/signup", "failure", "كلمه السر ليست متطابقه", &c)
	}
	var admin models.User
	db.GormDB.First(&admin, "classification = 1") // gets from the database where `admin` column is set to one
	// checks if the adminPassword field is equal to the administrator's password OR its hash is equal to the one stored in the database
	if !(adminPassword == administratorPassword || bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(adminPassword)) == nil) {
		// if not, redirect the user to /signup with a failure flash message
		return redirectWithFormData([]string{"username", "classification"}, []string{username, string(classification)},
			"/signup", "failure", "كلمه السر الخاصه ليست صحيحه", &c)
	}
	// all conditions are met and we're ready to store the user to the database
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), 10) // hash the password that the user entered
	// set the username and hashed password that the user entered to the `user` struct to save into the database
	user := models.User{Username: username, Password: string(hashedPassword), Classification: classification}
	databaseError := db.GormDB.Create(&user).GetErrors() // try saving the `user` struct
	if len(databaseError) > 0 {                          // checks for database errors
		// if found, then it mainly will be because of the unique key index of the username
		// redirect the user to /signup with failure flash message
		return redirectWithFormData([]string{"username", "classification"}, []string{username, string(classification)},
			"/signup", "failure", "تم تسجيل هذا المستخدم من قبل", &c)
	}
	// if we reached here, then the user is successfully signed up and he's ready to sign in
	err := addSession(&c, user.ID, user.Classification, user.Username) // add cookie to browser
	if err != nil {
		return err
	}
	return c.Redirect(http.StatusFound, "/") // redirect the user to the home page
}

// this function serves the resetPassword page
func (db *MyDB) showResetPasswordUpPage(c echo.Context) error {
	status, message := getFlashMessages(&c) // gets the flash message and status if there was any
	users := models.GetAllUsers(db.GormDB)  // gets all the users that are in the database
	formFields := getFormData(&c, []string{"username"})
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
		"username":                 formFields["username"],
		"resetPage":                true,
	})
}

// this function serves the resetPassword page
func (db *MyDB) showResetPasswordByEmailPage(c echo.Context) error {
	status, message := getFlashMessages(&c) // gets the flash message and status if there was any
	email := c.QueryParam("email")
	hash := c.QueryParam("hash")
	var otp models.OTP
	db.GormDB.Table("otps").First(&otp, "email = ? AND verification_code = ?", email, hash)
	if otp.UserID == 0 {
		return redirectWithFlashMessage("failure", "حدث خطأ ما برجاء اعاده المحاوله مره اخرى (او قم بأرسال البريد الالكتروني مره اخرى)", "/login", &c)
	}
	formAction := "/email-reset-password?email=" + email + "&hash=" + hash
	return c.Render(http.StatusOK, "reset-password-by-email.html", echo.Map{
		"status":     status,     // pass the status of the flash message
		"message":    message,    // pass the message
		"hideNavBar": true,       // boolean to indicate weather or not the NavBar should be displayed
		"formAction": formAction, // the URL that the form should be submitted to
	})
}

// this function serves the resetPassword page
func (db *MyDB) performResetPasswordByEmail(c echo.Context) error {
	email := c.QueryParam("email")
	hash := c.QueryParam("hash")
	url := "/email-reset-password?email=" + email + "&hash=" + hash
	var otp models.OTP

	password := strings.TrimSpace(c.FormValue("password"))
	passwordVerify := strings.TrimSpace(c.FormValue("passwordVerify"))

	if len(password) == 0 {
		return redirectWithFlashMessage("failure", "يجب تحديد كيمه السر الجديده", url, &c)
	}
	if password != passwordVerify { // checks that the password is equal to the password verification
		// if not, redirect to /reset-password-by-email with failure flash message
		return redirectWithFlashMessage("failure", "كلمه السر ليست متطابقه", url, &c)
	}

	db.GormDB.Table("otps").First(&otp, "email = ? AND verification_code = ?", email, hash)
	otp.Used = true
	db.GormDB.Save(&otp)
	db.GormDB.Where("email = ? AND user_id = ?", email, otp.UserID).Delete(models.OTP{})
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), 10) // hash the password that the user entered
	db.GormDB.Model(&models.User{}).UpdateColumn("password", string(hashedPassword))

	return redirectWithFlashMessage("success", "تم تغيير كلمه السر", "/login", &c)
}

func (db *MyDB) performResetPassword(c echo.Context) error {
	username := strings.TrimSpace(c.FormValue("username"))             // gets the username from the form submitted data
	password := strings.TrimSpace(c.FormValue("password"))             // gets the password from the form submitted data
	passwordVerify := strings.TrimSpace(c.FormValue("passwordVerify")) // gets the password verification from the form submitted data
	adminPassword := strings.TrimSpace(c.FormValue("adminPassword"))   // gets the admin's password (or administrator's password) from the form submitted data
	if len(username) == 0 || len(password) == 0 {
		return redirectWithFormData([]string{"username"}, []string{username},
			"/reset-password", "failure", "يجب تحديد اسم المستخدم وكلمه السر", &c)
	}
	if password != passwordVerify { // checks that the password is equal to the password verification
		// if not, redirect to /reset-password with failure flash message
		return redirectWithFormData([]string{"username"}, []string{username},
			"/reset-password", "failure", "كلمه السر ليست متطابقه", &c)
	}
	var admin, user models.User
	db.GormDB.First(&admin, 1)
	db.GormDB.Where("username = ?", username).First(&user) // gets the user where the username is equal to the entered username by the end user
	if user.ID == 0 {
		return redirectWithFormData([]string{"username"}, []string{username},
			"/reset-password", "failure", "عفوا لم نتمكن من ايجاد المستخدم- برجاء التأكد من اسم المستخدم واعاده المحاوله", &c)
	}
	if !(adminPassword == administratorPassword ||
		bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(adminPassword)) == nil ||
		bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(adminPassword)) == nil) {
		return redirectWithFormData([]string{"username"}, []string{username},
			"/reset-password", "failure", "كلمه السر الخاصه ليست صحيحه", &c)
	}
	// all conditions are met and we're ready to change the user's password
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), 10) // hash the password that the user entered
	user.Password = string(hashedPassword)                                 // sets the password to the hashed password
	db.GormDB.Save(&user)                                                  // update the user in the database
	err := addSession(&c, user.ID, user.Classification, user.Username)     // add a cookie to the browser to log the user in
	if err != nil {
		return err
	}
	return redirectWithFlashMessage("success", "تم تغيير كلمه السر بنجاح", "/", &c)
}

func (db *MyDB) resetPasswordByEmail(c echo.Context) error {
	username := strings.TrimSpace(c.FormValue("username"))
	if len(username) == 0 {
		addFlashMessage("failure", "يجب تحديد اسم المستخدم", &c)
		return c.JSON(http.StatusOK, map[string]string{
			"status": "reload",
		})
	}
	var user models.User
	db.GormDB.First(&user, "username = ?", username)
	if user.ID == 0 {
		addFormData([]string{"username"}, []string{username}, &c)
		addFlashMessage("failure", "تأكد من صحه اسم المستخدم", &c)
		return c.JSON(http.StatusOK, map[string]string{
			"status": "reload",
		})
	}
	if user.ValidEmail == false {
		addFormData([]string{"username"}, []string{username}, &c)
		addFlashMessage("failure", "عفوا انت لم تقم بتفعيل البريد الالكتروني الخاص بك/ برجاء التواصل مع السؤول", &c)
		return c.JSON(http.StatusOK, map[string]string{
			"status": "reload",
		})
	}
	emailHash, link := generateHashAndPasswordResetLink(user.Email)
	otp := models.OTP{
		UserID:           user.ID,
		Email:            user.Email,
		VerificationCode: emailHash,
		Type:             "email-password-reset",
	}
	db.GormDB.Create(&otp)
	sendResetLink(&user, link)
	addFlashMessage("success", "تم ارسال بريد الكتروني بالتعليمات لتغيير كلمه السر", &c)
	return c.JSON(http.StatusOK, map[string]string{
		"status": "reload",
	})
}

func redirectWithFormData(names, values []string, url, status, message string, c *echo.Context) error {
	addFlashMessage(status, message, c)
	addFormData(names, values, c)
	return (*c).Redirect(http.StatusFound, url)
}

// this function redirect the user to a url with flash status and message
func redirectWithFlashMessage(status string, message string, url string, c *echo.Context) error {
	addFlashMessage(status, message, c)
	return (*c).Redirect(http.StatusFound, url)
}

func addFormData(names, values []string, c *echo.Context) {
	sess := getSession("formData", c)
	for index := range names {
		sess.AddFlash(values[index], names[index])
	}
	_ = sess.Save((*c).Request(), (*c).Response())
}

func addFlashMessage(status, message string, c *echo.Context) {
	sess := getSession("flash", c)
	sess.AddFlash(status, "status")
	sess.AddFlash(message, "message")
	_ = sess.Save((*c).Request(), (*c).Response())
}

// this function adds a cookie to the browser, adding in it the user_id and weather or not he's an admin
func addSession(context *echo.Context, id uint, classification int, username string) error {
	sess := getSession("authorization", context)
	sess.Values["user_id"] = id
	sess.Values["classification"] = classification
	sess.Values["username"] = username
	if classification == 1 {
		sess.Values["stringClassification"] = "الوزير"
	} else if classification == 2 {
		sess.Values["stringClassification"] = "متابع"
	} else {
		sess.Values["stringClassification"] = "قائم به"
	}

	return sess.Save((*context).Request(), (*context).Response())
}

// this function removes the cookie that was added to the browser and log the user out
func logout(c echo.Context) error {
	sess, _ := session.Get("authorization", c)
	deleteSession(sess, c)
	return c.Redirect(http.StatusFound, "/login")
}
