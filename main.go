package main

import (
	"ah-follow-modules/configurations"
	"ah-follow-modules/handlers"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"html/template"
	"io"
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
	db, _ := configurations.InitDB()
	messagingClient := configurations.CreateMessagingClient()

	e := echo.New()
	//e.Pre(middleware.HTTPSRedirect())
	//e.AutoTLSManager.HostPolicy = autocert.HostWhitelist("ahtawfik.redirectme.net")
	//e.AutoTLSManager.Cache = autocert.DirCache(".cache")
	e.Static("/", "static")
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("secret"))))
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Renderer = &TemplateRegistry{
		templates: template.Must(template.ParseGlob("static/*.html")),
	}

	myConfigurations := handlers.MyConfigurations{
		GormDB:          db,
		MessagingClient: messagingClient,
	}
	handlers.InitializeRoutes(e, &myConfigurations)

	//e.Logger.Fatal(e.StartAutoTLS(":443"))
	e.Logger.Fatal(e.Start(":8085"))
}
