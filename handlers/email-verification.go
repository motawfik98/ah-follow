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
		Type:             "email",
	}
	db.GormDB.Create(&otp)
	sendLink(&user, link)
	addFlashMessage("success", "تم ارسال البريد الالكتروني للتفعيل", &c)
	return c.JSON(http.StatusOK, map[string]string{
		"status": "success",
	})
}

func generateHashAndVerificationLink(email string) (string, string) {
	var emailBuilder strings.Builder
	emailBuilder.WriteString("http://localhost:8081/verify-email?email=" + email)
	emailHash := models.GenerateEmailHash(email)
	emailBuilder.WriteString("&hash=" + emailHash)
	return emailHash, emailBuilder.String()
}

func sendLink(user *models.User, verificationLink string) {
	// Configure hermes by setting a theme and your product info
	h := hermes.Hermes{
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
			TroubleText: "لو لم يعمل زر ال{ACTION} قم بالضغط على الرابط التالي",
			// Custom copyright notice
			Copyright: "Copyright © 2019 Eng. Ahmed Tawfik.",
		},
	}

	generatedEmail := hermes.Email{
		Body: hermes.Body{
			Greeting:  "اهلا",
			Signature: "مع تحيات",
			Name:      user.Username,
			Intros: []string{
				"هذا البريد الالكتروني خاص بالتفعيل",
			},
			Actions: []hermes.Action{
				{
					Instructions: "لتفعيل البريد الالكتروني اضغط هنا",
					Button: hermes.Button{
						Color: "#22BC66", // Optional action button color
						Text:  "تفعيل",
						Link:  verificationLink,
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

	m := gomail.NewMessage()
	m.SetHeader("From", "takaleef@gmail.com")
	m.SetHeader("To", user.Email)
	m.SetHeader("Subject", "تفعيل البريد الالكتروني")
	m.SetBody("text/plain", emailText)
	m.AddAlternative("text/html", emailHTML)

	d := gomail.NewDialer("smtp.gmail.com", 465, "motawfik10@gmail.com", "aoulcplxdwgkurzf")
	// Send the email
	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}
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
		addFlashMessage("failure", "تأكد من الدخول بالحساب الصحيح", &c)
		return c.Redirect(http.StatusOK, "/")
	}

	if otp.ID != 0 { // if the number was successfully verified
		return redirectWithFlashMessage("success", "تم تفعيل البريد الالكتروني", "/user-settings", &c)
	} else {
		return redirectWithFlashMessage("failure", "عفوا حدث خطأ ما برجاء المحاوله مره اخري او اعاده ارسال رابط التفعيل", "user-settings", &c)
	}
}
