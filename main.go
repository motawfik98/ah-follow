package main

import (
	"./configurations"
	"./handlers"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/middleware"
	"golang.org/x/crypto/acme/autocert"
	"html/template"
	"io"
	"os"
)

// Define the template registry struct
type TemplateRegistry struct {
	templates *template.Template
}

// Implement e.Renderer interface
func (t *TemplateRegistry) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	_ = os.Setenv("administrator_password", "Nuccma6246V55")
	db, _ := configurations.InitDB()

	e := echo.New()
	e.Pre(middleware.HTTPSRedirect())
	//e.AutoTLSManager.HostPolicy = autocert.HostWhitelist("ahtawfik.redirectme.net")
	e.AutoTLSManager.Cache = autocert.DirCache(".cache")
	e.Static("/", "static")
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("secret"))))
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Renderer = &TemplateRegistry{
		templates: template.Must(template.ParseGlob("static/*.html")),
	}

	myDb := handlers.MyDB{GormDB: db}
	handlers.InitializeRoutes(e, &myDb)

	e.Logger.Fatal(e.StartAutoTLS(":443"))
	//e.Logger.Fatal(e.Start(":8081"))
}
