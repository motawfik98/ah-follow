package handlers

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
)

// this function ensures that the request is coming from logged in user
var ensureLoggedIn = middleware.JWTWithConfig(middleware.JWTConfig{
	ErrorHandler: func(e error) error {
		return raiseNewHTTPError(http.StatusUnauthorized, "true", "/login", "failure", "يجب تسجيل الدخول")
	},
	SigningKey:  []byte("very-secret-key-to-encode-tokens"),
	TokenLookup: "cookie:Authorization",
})

// this function ensures that the request is coming from non-logged in user
func ensureNotLoggedIn(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cookieValue := c.Get("user")
		if cookieValue == nil {
			return next(c)
		}
		return raiseNewHTTPError(http.StatusUnauthorized, "true", "/", "failure", "يجب تسجيل الخروج")
	}
}

func customHTTPErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	withFlash := "true"
	flashStatus := "failure"
	flashMessage := "حدث خطأ ما برجاء المحاوله مره اخرى او التواصل مع المسؤول"
	redirectLink := "/"

	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		switch he.Message.(type) {
		case map[string]string:
			redirectMap := he.Message.(map[string]string)
			withFlash = redirectMap["withFlash"]
			redirectLink = redirectMap["redirectLink"]
			flashStatus = redirectMap["flashStatus"]
			flashMessage = redirectMap["flashMessage"]
		case string:
			flashMessage = he.Message.(string)
		}

	}

	if code != 200 {
		if checkIfRequestFromMobileDevice(c) {
			_ = c.JSON(http.StatusBadRequest, echo.Map{
				"url":     redirectLink,
				"status":  flashStatus,
				"message": flashMessage,
			})
		} else {
			if withFlash == "true" {
				sess := getSession("flash", &c)          // gets the session with value `flash`
				sess.AddFlash(flashStatus, "status")     // add a key value pairs of (status, failure)
				sess.AddFlash(flashMessage, "message")   // add a key value pair of (errorMessage, ...)
				_ = sess.Save(c.Request(), c.Response()) // save the `flash` session
			}
			_ = c.Redirect(http.StatusFound, redirectLink) // redirect to home page, and there show the user the flash errorMessage
		}
	}
}

func ensureLoggedInWithoutFlashMessage(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, _ := session.Get("authorization", c) // gets the session with name `authorization`
		if sess.Values["user_id"] != nil {         // if the value of `user_id` is not null then there's logged in user
			return next(c) // continue the request
		}
		return nil
	}
}

// this function ensures that the request is coming from an admin user
func ensureAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		_, classification := getUserStatus(&c) // gets the classification value from the session that was added
		if classification == 1 {
			return next(c) // the user is an admin, continue the request
		}
		return raiseNewHTTPError(http.StatusUnauthorized, "true", "/", "failure", "عفوا, ليس لديك الصلاحيه لأتمام العمليه")
	}
}

func raiseNewHTTPError(statusCode int, withFlash, redirectLink, flashStatus, flashMessage string) error {
	return echo.NewHTTPError(statusCode, map[string]string{
		"withFlash":    withFlash,
		"redirectLink": redirectLink,
		"flashStatus":  flashStatus,
		"flashMessage": flashMessage,
	})
}

// this function gets the session with the given name
func getSession(sessionName string, c *echo.Context) *sessions.Session {
	sess, _ := session.Get(sessionName, *c)
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   10, // to enforce the browser to logout after the session was closed
		HttpOnly: true,
	}
	return sess
}

// this function gets the session with the given name
func getToken(c *echo.Context) jwt.MapClaims {
	user := (*c).Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	return claims
}

// this function returns the user_id and the classification values that were stored in the session cookie
func getUserStatus(c *echo.Context) (uint, int) {
	sess := getToken(c)
	return uint(sess["user_id"].(float64)), int(sess["classification"].(float64))
}

func getUsernameAndClassification(c *echo.Context) (string, string) {
	sess := getToken(c)
	return sess["username"].(string), sess["stringClassification"].(string)
}

// this function deletes the session cookie from the browser (useful in logout)
func deleteSession(sess *sessions.Session, c echo.Context) {
	sess.Options = &sessions.Options{
		MaxAge: -1,
	}
	_ = sessions.Save(c.Request(), c.Response())
}

// this function returns the status and message of the flash message if found
func getFlashMessages(c *echo.Context) (string, string) {
	sess, _ := session.Get("flash", *c) // gets the session with name `flash`
	var status, message string
	if i := sess.Flashes("status"); len(i) > 0 { // gets the flashes with name `status`
		// if the length of the flashes is > 0, then a flash message was found
		status = i[0].(string)                        // set the status variable to the status that was found
		message = sess.Flashes("message")[0].(string) // set the message variable to the status that was found
	}

	deleteSession(sess, *c) // deletes the session with name `flash` to avoid taking it once more in the future
	return status, message  // returns the status and message
}

// this function returns the status and message of the flash message if found
func getFormData(c *echo.Context, names []string) map[string]string {
	sess, _ := session.Get("formData", *c) // gets the session with name `flash`
	values := make(map[string]string, len(names))
	for _, name := range names {
		flash := sess.Flashes(name)
		if len(flash) > 0 {
			values[name] = flash[0].(string)
		} else {
			values[name] = ""
		}
	}

	deleteSession(sess, *c) // deletes the session with name `formData` to avoid taking it once more in the future
	return values           // returns the status and message
}

// this function checks if the request is coming from a mobile device or a normal browser
// as if it's coming from a mobile device then the JSON response is returned
// while from a normal browser the whole page is required
func checkIfRequestFromMobileDevice(c echo.Context) bool {
	return c.Request().Header.Get("mobile-request") == "true"
}
