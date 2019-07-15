package handlers

import (
	"../models"
	"github.com/arxdsilva/hermes"
	"github.com/go-gomail/gomail"
	"github.com/goware/emailx"
	"github.com/labstack/echo"
	"net/http"
	"strings"
)

func (db *MyDB) changeEmail(c echo.Context) error {
	userID, _ := getUserStatus(&c)
	var user models.User
	db.GormDB.First(&user, userID)
	email := c.FormValue("email") // gets the submitted email
	err := emailx.Validate(email)
	if err == emailx.ErrInvalidFormat {
		return c.JSON(http.StatusOK, map[string]string{
			"status":  "failure",
			"message": "تأكد من صحه البريد الالكتروني",
		})
	}
	if user.Email == email {
		return c.JSON(http.StatusOK, map[string]string{
			"status":  "failure",
			"message": "تأكد من تغيير البريد الالكتروني قبل الحفظ",
		})
	}
	user.Email = email
	user.ValidEmail = false
	user.EmailNotifications = false
	db.GormDB.Save(&user)
	return db.sendVerificationLink(c)
}

func (db *MyDB) sendVerificationLink(c echo.Context) error {
	userID, _ := getUserStatus(&c)
	var user models.User
	db.GormDB.First(&user, userID)
	if user.ValidEmail { // checks that the number is not verified (as no point in sending a verification code for a verified email)
		return c.JSON(http.StatusOK, map[string]string{ // inform the user that the email is verified
			"status":  "failure",
			"message": "لقد تم تفعيل هذا البريد الالكتروني من قبل! سوف يتم تجاهل طلبك",
		})
	}
	emailHash, link := generateHashAndVerificationLink(user.Email)
	otp := models.OTP{
		UserID:           userID,
		Email:            user.Email,
		VerificationCode: emailHash,
		Type:             "email-verify",
	}
	db.GormDB.Create(&otp)
	sendVerificationLink(&user, link)
	addFlashMessage("success", "تم ارسال البريد الالكتروني للتفعيل", &c)
	return c.JSON(http.StatusOK, map[string]string{
		"status": "success",
	})
}

func generateHashAndVerificationLink(email string) (string, string) {
	var emailBuilder strings.Builder
	emailBuilder.WriteString("http://localhost:8081/verify-email?email=" + email)
	emailHash := models.GenerateEmailHash(email, "verification")
	emailBuilder.WriteString("&hash=" + emailHash)
	return emailHash, emailBuilder.String()
}
func generateHashAndPasswordResetLink(email string) (string, string) {
	var emailBuilder strings.Builder
	emailBuilder.WriteString("http://localhost:8081/email-reset-password?email=" + email)
	emailHash := models.GenerateEmailHash(email, "password-reset")
	emailBuilder.WriteString("&hash=" + emailHash)
	return emailHash, emailBuilder.String()
}

func generateHermesEmail(username, intro, actionInstruction, actionColor, actionText, actionLink string, h hermes.Hermes) (string, string) {
	generatedEmail := hermes.Email{
		Body: hermes.Body{
			Greeting:  "اهلا",
			Signature: "مع تحيات",
			Name:      username,
			Intros: []string{
				intro,
			},
			Actions: []hermes.Action{
				{
					Instructions: actionInstruction,
					Button: hermes.Button{
						Color: actionColor, // Optional action button color
						Text:  actionText,
						Link:  actionLink,
					},
				},
			},
		},
	}
	// Generate an HTML email with the provided contents (for modern clients)
	emailHTML, err := h.GenerateHTML(generatedEmail)
	if err != nil {
		panic(err) // Tip: Handle error with something else than a panic ;)
	}
	// Generate the plaintext version of the e-mail (for clients that do not support xHTML)
	emailText, err := h.GeneratePlainText(generatedEmail)
	if err != nil {
		panic(err) // Tip: Handle error with something else than a panic ;)
	}

	return emailHTML, emailText
}

func generateHermesStruct() hermes.Hermes {
	return hermes.Hermes{
		// Custom text direction
		TextDirection: hermes.TDRightToLeft,
		// Optional Theme
		// Theme: new(Default)
		Product: hermes.Product{
			// Appears in header & footer of e-mails
			Name: "التكاليف الوزاريه",
			Link: "http://localhost:8081",
			// Optional product logo
			Logo: "https://i1.wp.com/doist.com/blog/wp-content/uploads/sites/3/2017/08/Ways-to-add-tasks-to-Todoist-.png?fit=2000%2C1000&quality=85&strip=all&ssl=1",
			// Custom trouble text
			TroubleText: "لو لم يعمل زر \"{ACTION}\" قم بالضغط على الرابط التالي",
			// Custom copyright notice
			Copyright: "Copyright © 2019 Eng. Ahmed Tawfik.",
		},
	}
}

func sendEmail(userEmail, emailHTML, emailText, subject string) {
	m := gomail.NewMessage()
	m.SetHeader("From", "takaleef@gmail.com")
	m.SetHeader("To", userEmail)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", emailText)
	m.AddAlternative("text/html", emailHTML)

	d := gomail.NewDialer("smtp.gmail.com", 465, "motawfik10@gmail.com", "aoulcplxdwgkurzf")
	// Send the email
	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}
}

func sendVerificationLink(user *models.User, verificationLink string) {
	// Configure hermes by setting a theme and your product info
	h := generateHermesStruct()
	emailHTML, emailText := generateHermesEmail(user.Username, "هذا البريد الالكتروني خاص بالتفعيل", "لتفعيل البريد الالكتروني اضغط هنا",
		"#22BC66", "تفعيل", verificationLink, h)
	sendEmail(user.Email, emailHTML, emailText, "تفعيل البريد الالكتروني")
}

func sendResetLink(user *models.User, resetLink string) {
	h := generateHermesStruct()
	emailHTML, emailText := generateHermesEmail(user.Username, "هذا البريد الالكتروني خاص بتغيير كلمه السر", "لتغيير كلمه السر اضغط هنا",
		"#DC4D2F", "تغيير كلمه السر", resetLink, h)
	sendEmail(user.Email, emailHTML, emailText, "تغيير كلمه السر")
}

func sendEmailNotification(user *models.User, taskLink, from, emailBody string) {
	h := generateHermesStruct()
	emailBody = emailBody + " بواسطه " + from
	emailHTML, emailText := generateHermesEmail(user.Username, "هذا البريد الالكتروني خاص بالاشعارات-"+emailBody,
		"لعرض التكليف اضغط هنا", "#0000FF", "عرض التكليف", taskLink, h)
	sendEmail(user.Email, emailHTML, emailText, "اشعار من "+from)
}

func (db *MyDB) verifyEmail(c echo.Context) error {
	userID, _ := getUserStatus(&c)
	var user models.User
	var otp models.OTP
	db.GormDB.First(&user, userID)
	submittedEmail := c.QueryParam("email")
	submittedHash := c.QueryParam("hash")

	db.GormDB.Where("user_id = ? AND email = ? AND verification_code = ?", userID, submittedEmail, submittedHash).First(&otp) // get all the not deleted OTPs from the database for that specific user and email
	if user.Email != submittedEmail {
		return redirectWithFlashMessage("failure", "تأكد من الدخول بالحساب الصحيح", "/", &c)
	}

	if otp.ID != 0 { // if the number was successfully verified
		db.GormDB.Model(&user).Update("valid_email", true)
		db.GormDB.Model(&otp).Update("used", true)
		db.GormDB.Where("user_id = ? AND email = ?", userID, user.Email).Delete(models.OTP{})
		return redirectWithFlashMessage("success", "تم تفعيل البريد الالكتروني", "/user-settings", &c)
	} else {
		return redirectWithFlashMessage("failure", "عفوا حدث خطأ ما برجاء المحاوله مره اخري او اعاده ارسال رابط التفعيل", "/user-settings", &c)
	}
}
