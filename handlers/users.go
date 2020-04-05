package handlers

import (
	"ah-follow-modules/models"
	"firebase.google.com/go/messaging"
	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type MyConfigurations struct {
	GormDB          *gorm.DB
	MessagingClient *messaging.Client // to pass the client variable to send notifications
}

// this function serves the login page
func (db *MyConfigurations) showLoginPage(c echo.Context) error {
	status, message := getFlashMessages(&c) // gets the flash message and status if there was any
	users := models.GetAllUsers(db.GormDB)  // gets all the users that are in the database
	valuesToReturn := echo.Map{
		"status":     status,       // pass the status of the flash message
		"message":    message,      // pass the message
		"title":      "تسجيل دخول", // the title of the page
		"hideNavBar": true,         // boolean to indicate weather or not the NavBar should be displayed
		"users":      users,        // pass the users array to display to the user
	}
	if checkIfRequestFromMobileDevice(c) {
		return c.JSON(http.StatusOK, &valuesToReturn)
	}
	return c.Render(http.StatusOK, "login.html", &valuesToReturn)
}

// this function performs the login logic
func (db *MyConfigurations) performLogin(c echo.Context) error {
	var loginData, user models.User
	_ = c.Bind(&loginData)                                                                  // gets the form data from the context and binds it to the `loginData` struct
	db.GormDB.First(&user, "username = ?", loginData.Username)                              // gets the user from the database where his username is equal to the entered username
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginData.Password)) // compare the hashed password that is stored in the database with the hashed version of the password that the user entered
	// checks if the user ID is 0 (which means that no user was found with that username)
	// checks that err is not null (which means that the hashed password is the same of the hashed version of the user entered password)
	// makes sure that the password that the user entered is not the administrator password
	if user.ID == 0 || (err != nil && loginData.Password != administratorPassword) {
		if checkIfRequestFromMobileDevice(c) {
			return c.JSON(http.StatusBadRequest, echo.Map{
				"message": "بيانات الدخول ليست صحيحه",
			})
		}
		// redirect to /login and add a failure flash message
		return redirectWithFlashMessage("failure", "بيانات الدخول ليست صحيحه", "/login", &c)
	} else {
		token, err := createToken(user.ID, user.Classification, user.Username)
		if err != nil {
			return err
		}
		if checkIfRequestFromMobileDevice(c) {
			return c.JSON(http.StatusOK, echo.Map{
				"securityToken": token,
				"url":           "/",
			})
		}
		c.SetCookie(&http.Cookie{
			Name:    "Authorization",
			Value:   token,
			Expires: time.Now().Add(time.Hour * 24 * 30),
		})
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
func (db *MyConfigurations) performSignUp(c echo.Context) error {
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
	token, err := createToken(user.ID, user.Classification, user.Username) // add cookie to browser
	if err != nil {
		return err
	}
	if checkIfRequestFromMobileDevice(c) {
		return c.JSON(http.StatusOK, echo.Map{})
	}
	c.SetCookie(&http.Cookie{
		Name:  "Authorization",
		Value: token,
	})
	return c.Redirect(http.StatusFound, "/") // redirect the user to the home page
}

// this function serves the resetPassword page
func (db *MyConfigurations) showResetPasswordUpPage(c echo.Context) error {
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

func (db *MyConfigurations) showResetPasswordPage(c echo.Context) error {
	_, classification := getUserStatus(&c) // get the user ID and the classification int from the cookie that is stored
	username, stringClassification := getUsernameAndClassification(&c)
	status, message := getFlashMessages(&c) // gets the flash message and status if there was any
	formAction := "/reset-password"

	return c.Render(http.StatusOK, "reset-password.html", echo.Map{
		"title":                "تغيير كلمه السر", // sets the title of the page
		"status":               status,            // pass the status of the flash message
		"message":              message,           // pass the message
		"hideNavBar":           false,             // boolean to indicate weather or not the NavBar should be displayed
		"formAction":           formAction,        // the URL that the form should be submitted to
		"classification":       classification,
		"username":             username,
		"stringClassification": stringClassification,
	})

}

// this function serves the resetPassword page
func (db *MyConfigurations) showResetPasswordByEmailPage(c echo.Context) error {
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
func (db *MyConfigurations) performResetPasswordByEmail(c echo.Context) error {
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
	db.GormDB.Model(&models.User{}).Where("id = ?", otp.UserID).Update("password", string(hashedPassword))

	return redirectWithFlashMessage("success", "تم تغيير كلمه السر", "/login", &c)
}

func (db *MyConfigurations) performResetPassword(c echo.Context) error {
	userID, _ := getUserStatus(&c) // get the user ID and the classification int from the cookie that is stored

	password := strings.TrimSpace(c.FormValue("password"))             // gets the password from the form submitted data
	passwordVerify := strings.TrimSpace(c.FormValue("passwordVerify")) // gets the password verification from the form submitted data
	oldPassword := strings.TrimSpace(c.FormValue("oldPassword"))       // gets the admin's password (or administrator's password) from the form submitted data
	if len(password) == 0 {
		return redirectWithFlashMessage("failure", "يجب تحديد كلمه السر", "/change-password", &c)
	}
	if password != passwordVerify { // checks that the password is equal to the password verification
		// if not, redirect to /reset-password with failure flash message
		return redirectWithFlashMessage("failure", "كلمه السر ليست متطابقه", "/change-password", &c)
	}
	var admin, user models.User
	db.GormDB.First(&admin, 1)
	db.GormDB.First(&user, userID) // gets the user where the username is equal to the entered username by the end user

	if !(oldPassword == administratorPassword ||
		bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(oldPassword)) == nil ||
		bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)) == nil) {
		return redirectWithFlashMessage("failure", "كلمه السر القديمه ليست صحيحه", "/change-password", &c)
	}
	// all conditions are met and we're ready to change the user's password
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), 10) // hash the password that the user entered
	user.Password = string(hashedPassword)                                 // sets the password to the hashed password
	db.GormDB.Save(&user)                                                  // update the user in the database
	return redirectWithFlashMessage("success", "تم تغيير كلمه السر بنجاح", "/", &c)
}

func (db *MyConfigurations) resetPasswordByEmail(c echo.Context) error {
	email := strings.TrimSpace(c.FormValue("email"))
	if len(email) == 0 {
		if checkIfRequestFromMobileDevice(c) {
			return c.JSON(http.StatusNotAcceptable, echo.Map{
				"message": "يجب تحديد البريد الالكتروني",
			})
		}
		addFlashMessage("failure", "يجب تحديد البريد الالكتروني", &c)
		return c.JSON(http.StatusOK, map[string]string{
			"status": "reload",
		})
	}
	var user models.User
	db.GormDB.First(&user, "email = ?", email)
	if user.ID == 0 {
		if checkIfRequestFromMobileDevice(c) {
			return c.JSON(http.StatusNotAcceptable, echo.Map{
				"message": "لم نتمكن من ايجاد البريد الالكتروني الخاص بك. تأكد من صحته واعد المحاوله مره اخرى",
			})
		}
		addFormData([]string{"email"}, []string{email}, &c)
		addFlashMessage("failure", "تأكد من صحه البريد الالكتروني", &c)
		return c.JSON(http.StatusOK, map[string]string{
			"status": "reload",
		})
	}
	if user.ValidEmail == false {
		if checkIfRequestFromMobileDevice(c) {
			return c.JSON(http.StatusNotAcceptable, echo.Map{
				"message": "عفوا انت لم تقم بتفعيل البريد الالكتروني الخاص بك. برجاء التواصل مع المسؤول",
			})
		}
		addFormData([]string{"email"}, []string{email}, &c)
		addFlashMessage("failure", "عفوا انت لم تقم بتفعيل البريد الالكتروني الخاص بك/ برجاء التواصل مع المسؤول", &c)
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
	if checkIfRequestFromMobileDevice(c) {
		return c.JSON(http.StatusOK, echo.Map{
			"message": "تم ارسال بريد الكتروني بالتعليمات لتغيير كلمه السر",
		})
	}
	addFlashMessage("success", "تم ارسال بريد الكتروني بالتعليمات لتغيير كلمه السر", &c)
	return c.JSON(http.StatusOK, map[string]string{
		"status": "reload",
	})
}

func redirectWithFormData(names, values []string, url, status, message string, c *echo.Context) error {
	if checkIfRequestFromMobileDevice(*c) {
		return (*c).JSON(http.StatusNotAcceptable, echo.Map{
			"status":  status,
			"message": message,
		})
	}
	addFlashMessage(status, message, c)
	addFormData(names, values, c)
	return (*c).Redirect(http.StatusFound, url)
}

// this function redirect the user to a url with flash status and message
func redirectWithFlashMessage(status string, message string, url string, c *echo.Context) error {
	context := *c
	if checkIfRequestFromMobileDevice(context) {
		if status == "success" {
			return context.JSON(http.StatusOK, echo.Map{
				"message": message,
			})
		} else {
			return context.JSON(http.StatusNotAcceptable, echo.Map{
				"message": message,
			})
		}
	}
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

func createToken(id uint, classification int, username string) (string, error) {
	// Create token
	token := jwt.New(jwt.SigningMethodHS256)
	// Set claims
	// This is the information which frontend can use
	// The backend can also decode the token and get user_id etc.
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = id
	claims["classification"] = classification
	claims["username"] = username
	if classification == 1 {
		claims["stringClassification"] = "الوزير"
	} else if classification == 2 {
		claims["stringClassification"] = "متابع"
	} else {
		claims["stringClassification"] = "قائم به"
	}
	claims["exp"] = time.Now().Add(time.Hour * 24 * 30).Unix()

	// Generate encoded token and send it as response.
	// The signing string should be secret (a generated UUID works too)
	t, err := token.SignedString([]byte("very-secret-key-to-encode-tokens"))
	if err != nil {
		return "", err
	}
	return t, nil

}

// this function removes the cookie that was added to the browser and log the user out
func (db *MyConfigurations) logout(c echo.Context) error {
	fcmToken := c.Request().Header.Get("fcm-token")
	db.GormDB.Where("token = ?", fcmToken).Delete(&models.DeviceToken{})
	c.SetCookie(&http.Cookie{
		Name:    "Authorization",
		Value:   "",
		Expires: time.Now(),
		MaxAge:  0,
	})
	return c.Redirect(http.StatusFound, "/login")
}
