package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *Config) routes() http.Handler {
	mux := chi.NewRouter()

	// set up middleware
	mux.Use(middleware.Recoverer)
	mux.Use(app.SessionLoad)

	// set up routes
	mux.Get("/", app.HomePage)
	mux.Get("/login", app.LoginPage)
	mux.Post("/login", app.PostLoginPage)
	mux.Get("/logout", app.Logout)
	mux.Get("/register", app.RegisterPage)
	mux.Post("/register", app.PostRegisterPage)
	mux.Get("/activate", app.ActivateAccount)

	// mux.Get("/test-email", func(w http.ResponseWriter, r *http.Request) {
	// 	app.InfoLog.Println("Sending test email")
	// 	m := Mail{
	// 		Domain:      "localhost",
	// 		Host:        "localhost",
	// 		Port:        1025,
	// 		Encryption:  "none",
	// 		FromAddress: "info@mycompany.com",
	// 		FromName:    "Test",
	// 		ErrorChan:   make(chan error),
	// 	}

	// 	msg := Message{
	// 		To:       "me@here.com",
	// 		Subject:  "Test Email",
	// 		Data:     "Hello World",
	// 		Template: "mail",
	// 	}
	// 	app.InfoLog.Println("Sending email")
	// 	m.sendMail(msg, m.ErrorChan)
	// })

	return mux
}
