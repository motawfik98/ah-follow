package handlers

import (
	"../models"
	"crypto/rand"
	"fmt"
	"github.com/labstack/echo"
	"math/big"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf16"
)

// this function serves the userSettings page
func (db *MyDB) showSettingsPage(c echo.Context) error {
	userID, _ := getUserStatus(&c)
	username, stringClassification := getUsernameAndClassification(&c)
	var user models.User
	db.GormDB.First(&user, userID)
	status, message := getFlashMessages(&c) // gets the flash message and status if there was any
	phoneNumber := user.PhoneNumber
	var hidePhoneVerification, hideEmailVerification bool
	if len(phoneNumber) == 0 || user.ValidPhoneNumber {
		hidePhoneVerification = true
	}
	email := user.Email
	if len(email) == 0 || user.ValidEmail {
		hideEmailVerification = true
	}

	return c.Render(http.StatusOK, "user-settings.html", echo.Map{
		"status":                status,                // pass the status of the flash message
		"message":               message,               // pass the message
		"phoneNumber":           phoneNumber,           // pass the phone number
		"hidePhoneVerification": hidePhoneVerification, // pass if the number is activated or not
		"email":                 email,                 // pass the email
		"hideEmailVerification": hideEmailVerification, // pass if the email is activated or not
		"title":                 "بيانات المستخدم",     // the title of the page
		"buttonText":            "حفظ",                 // the action button text the should be displayed to the user
		"formAction":            "/save-settings",      // the URL that the form should be submitted to
		"username":              username,              // pass the username
		"stringClassification":  stringClassification,
		"activatedPhoneNumber":  user.ValidPhoneNumber,
		"activatedEmail":        user.ValidEmail,
		"phoneNotifications":    user.PhoneNotifications,
		"emailNotifications":    user.EmailNotifications,
	})
}

// this function changes the user's phone number
func (db *MyDB) changePhoneNumber(c echo.Context) error {
	userID, _ := getUserStatus(&c)
	var user models.User
	db.GormDB.First(&user, userID)
	phoneNumber := c.FormValue("phoneNumber") // gets the submitted phone number
	match, _ := regexp.MatchString("[0-9]{11}", phoneNumber)
	if !match { // verify that the number contains only 11 digits
		return c.JSON(http.StatusOK, map[string]string{ // return error to be displayed for the user
			"status":  "failure",
			"message": "يجب ان يكون الرقم مكون من 11 رقم",
		})
	}
	if user.PhoneNumber == phoneNumber { // checks that the number actually changed
		return c.JSON(http.StatusOK, map[string]string{ // inform the user that he entered an already saved number
			"status":  "failure",
			"message": "قم بتغيير الرقم قبل الضغط على الزر",
		})
	}
	user.PhoneNumber = phoneNumber                  // change the user's number
	user.ValidPhoneNumber = false                   // make the new number unverified
	db.GormDB.Save(user)                            // update the user
	_ = db.sendVerificationCode(c)                  // sends a verification code for the new number
	return c.JSON(http.StatusOK, map[string]string{ // inform the user that a message has been sent to him
		"status":  "success",
		"message": "لقد تم ارسال كود التفعيل الخاص بك برجاء الانتظار",
	})
}

// this function sends a verification code for the logged in user
func (db *MyDB) sendVerificationCode(c echo.Context) error {
	userID, _ := getUserStatus(&c)
	var user models.User
	db.GormDB.First(&user, userID)
	if user.ValidPhoneNumber { // checks that the number is not verified (as no point in sending a verification code for a verified number)
		return c.JSON(http.StatusOK, map[string]string{ // inform the user that the number is verified
			"status":  "failure",
			"message": "لقد تم تفعيل هذا الرقم من قبل! سوف يتم تجاهل طلبك",
		})
	}
	code, _ := getRandNum() // gets a random number
	otp := models.OTP{      // create a OneTimePassword (OTP)
		UserID:           userID,
		PhoneNumber:      user.PhoneNumber,
		VerificationCode: code,
		Type:             "phone number",
	}
	db.GormDB.Create(&otp)
	message := generateUTF16Message("كود التفعيل الخاص بالهاتف هو " + code) // gets the UTF-16 encoding
	sendMessage(user.PhoneNumber, message)                                  // send the message
	return nil
}

func generateUTF16Message(message string) string {
	ss := utf16.Encode([]rune(message))
	var string16Builder strings.Builder
	for _, s := range ss { // make sure that each character is encoded with 4 digits
		string16 := strconv.FormatInt(int64(s), 16)
		if len(string16) == 1 {
			string16Builder.WriteString("000" + string16)
		} else if len(string16) == 2 {
			string16Builder.WriteString("00" + string16)
		} else if len(string16) == 3 {
			string16Builder.WriteString("0" + string16)
		} else {
			string16Builder.WriteString(string16)
		}
	}
	return string16Builder.String()
}

// getRandNum returns a random number of size four
func getRandNum() (string, error) {
	nBig, e := rand.Int(rand.Reader, big.NewInt(8999))
	if e != nil {
		return "", e
	}
	return strconv.FormatInt(nBig.Int64()+1000, 10), nil
}

func sendMessage(phoneNumber, message string) {
	apiSerial := "5d1b9537bdcb3" // api for SMSBulko
	username := "tawfik"
	password := "tawfik"
	requestApi := fmt.Sprintf("http://smsbulko.com/smsportal/user_api.php?api_ser=%s&username=%s&password=%s"+
		"&request=5&unicode=1&country=65&to=%s&msg=%s&sender=SMBULKO&hlr=0", apiSerial, username, password, phoneNumber, message) // format the string to get the required URL
	_, _ = http.Post(requestApi, "", nil) // make the post request
}

func (db *MyDB) verifyPhoneNumber(c echo.Context) error {
	userID, _ := getUserStatus(&c)
	var user models.User
	var otp models.OTP
	db.GormDB.First(&user, userID)
	verificationCode := c.FormValue("verificationCode")                                                                                  // gets the submitted verification code
	db.GormDB.Where("user_id = ? AND phone_number = ? AND verification_code = ?", userID, user.PhoneNumber, verificationCode).Find(&otp) // get all the not deleted OTP from the database for that specific user, number, and verification code

	var status, returnMessage string
	if otp.ID != 0 { // if the number was successfully verified
		db.GormDB.Model(&user).Update("valid_phone_number", true)
		db.GormDB.Model(&otp).Update("used", true)
		db.GormDB.Where("user_id = ? AND phone_number = ?", userID, user.PhoneNumber).Delete(models.OTP{})
		addFlashMessage("success", "تم تفعيل رقم الهاتف", &c) // inform the user that the number is verified
		status = "success"
	} else {
		status = "failure" // inform the user that the verification code that he's entered is not correct
		returnMessage = "عفوا الرقم اللذي ادخلته غير صحيح"
	}

	return c.JSON(http.StatusOK, map[string]string{ // send the status and message to be displayed
		"status":  status,
		"message": returnMessage,
	})
}

func (db *MyDB) changeNotifications(c echo.Context) error {
	userID, _ := getUserStatus(&c)
	var user models.User
	db.GormDB.Find(&user, userID)
	notificationsType := c.FormValue("type")
	notifications, _ := strconv.ParseBool(c.FormValue("notifications"))

	if notificationsType == "phone" {
		if notifications && !user.ValidPhoneNumber {
			addFlashMessage("failure", "يجب تفعيل رقم الهاتف لأستقبال الاشعارات", &c)
			return c.JSON(http.StatusOK, map[string]string{
				"status": "reload",
			})
		}
		db.GormDB.Model(&user).Updates(map[string]interface{}{"phone_notifications": notifications})
		addFlashMessage("success", "لقد تم تنفيذ طلبك بنجاح", &c)
		return c.JSON(http.StatusOK, map[string]string{
			"status": "reload",
		})
	} else if notificationsType == "email" {
		if notifications && !user.ValidEmail {
			addFlashMessage("failure", "يجب تفعيل البريد الالكتروني لأستقبال الاشعارات", &c)
			return c.JSON(http.StatusOK, map[string]string{
				"status": "reload",
			})
		}
		db.GormDB.Model(&user).Updates(map[string]interface{}{"email_notifications": notifications})
		addFlashMessage("success", "لقد تم تنفيذ طلبك بنجاح", &c)
		return c.JSON(http.StatusOK, map[string]string{
			"status": "reload",
		})
	}
	return nil
}
