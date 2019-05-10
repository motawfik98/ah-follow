package handlers

import (
	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
	"net/http"
)

func ensureNotLoggedIn(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, _ := session.Get("authorization", c)
		if sess.Values["user_id"] == nil {
			return next(c)
		}
		sess = getSession("flash", &c)
		sess.AddFlash("failure", "status")
		sess.AddFlash("يجب تسجيل الخروج", "message")
		_ = sess.Save(c.Request(), c.Response())
		return c.Redirect(http.StatusFound, "/")
	}
}

func ensureLoggedIn(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, _ := session.Get("authorization", c)
		if sess.Values["user_id"] != nil {
			return next(c)
		}
		sess = getSession("flash", &c)
		sess.AddFlash("failure", "status")
		sess.AddFlash("يجب تسجيل الدخول", "message")
		_ = sess.Save(c.Request(), c.Response())
		return c.Redirect(http.StatusFound, "/login")
	}
}

func getSession(sessionName string, c *echo.Context) *sessions.Session {
	sess, _ := session.Get(sessionName, *c)
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
	}
	return sess
}

func deleteSession(sess *sessions.Session, c echo.Context) {
	sess.Options = &sessions.Options{
		MaxAge: -1,
	}
	_ = sessions.Save(c.Request(), c.Response())
}

func getFlashMessages(c *echo.Context) (string, string) {
	sess, _ := session.Get("flash", *c)
	var status, message string
	if i := sess.Flashes("status"); len(i) > 0 {
		status = i[0].(string)
		message = sess.Flashes("message")[0].(string)
	}

	deleteSession(sess, *c)
	return status, message
}
