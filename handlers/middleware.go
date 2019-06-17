package handlers

import (
	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
	"net/http"
)

// this function ensures that the request is coming from non-logged in user
func ensureNotLoggedIn(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, _ := session.Get("authorization", c) // gets the session with name `authorization`
		if sess.Values["user_id"] == nil {         // if the value of `user_id` is null then there's no logged in user
			return next(c) // continue the request
		}
		sess = getSession("flash", &c)               // gets the session with value `flash`
		sess.AddFlash("failure", "status")           // add a key value pairs of (status, failure)
		sess.AddFlash("يجب تسجيل الخروج", "message") // add a key value pair of (message, ...)
		_ = sess.Save(c.Request(), c.Response())     // save the `flash` session
		return c.Redirect(http.StatusFound, "/")     // redirect to home page, and there show the user the flash message
	}
}

// this function ensures that the request is coming from logged in user
func ensureLoggedIn(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, _ := session.Get("authorization", c) // gets the session with name `authorization`
		if sess.Values["user_id"] != nil {         // if the value of `user_id` is not null then there's logged in user
			return next(c) // continue the request
		}
		sess = getSession("flash", &c)                // gets the session with value `flash`
		sess.AddFlash("failure", "status")            // add a key value pairs of (status, failure)
		sess.AddFlash("يجب تسجيل الدخول", "message")  // add a key value pair of (message, ...)
		_ = sess.Save(c.Request(), c.Response())      // save the `flash` session
		return c.Redirect(http.StatusFound, "/login") // redirect to login page, and there show the user the flash message
	}
}

// this function ensures that the request is coming from an admin user
func ensureAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		_, classification := getUserStatus(&c) // gets the classification value from the session that was added
		if classification == 1 {
			return next(c) // the user is an admin, continue the request
		}
		sess := getSession("flash", &c)                                    // gets the session with value `flash`
		sess.AddFlash("failure", "status")                                 // add a key value pairs of (status, failure)
		sess.AddFlash("عفوا, ليس لديك الصلاحيه لأتمام العمليه", "message") // add a key value pair of (message, ...)
		_ = sess.Save(c.Request(), c.Response())                           // save the `flash` session
		return c.Redirect(http.StatusFound, "/")                           // redirect to home page, and there show the user the flash message
	}
}

// this function gets the session with the given name
func getSession(sessionName string, c *echo.Context) *sessions.Session {
	sess, _ := session.Get(sessionName, *c)
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
	}
	return sess
}

// this function returns the user_id and the classification values that were stored in the session cookie
func getUserStatus(c *echo.Context) (uint, int) {
	sess := getSession("authorization", c)
	return sess.Values["user_id"].(uint), sess.Values["classification"].(int)
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
