package main

import "net/http"

// HomePage handles the GET request to /
func (app *Config) HomePage(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "home.page.gohtml", nil)
}

// LoginPage handles the GET request to /login
func (app *Config) LoginPage(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "login.page.gohtml", nil)
}

// PostLoginPage handles the POST request to /login
func (app *Config) PostLoginPage(w http.ResponseWriter, r *http.Request) {
	_ = app.Session.RenewToken(r.Context())

	// parse form post
	err := r.ParseForm()
	if err != nil {
		app.Session.Put(r.Context(), "error", "Invalid form post")
		app.ErrorLog.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// get form values
	email := r.PostForm.Get("email")
	password := r.PostForm.Get("password")

	// authenticate user
	user, err := app.Models.User.GetByEmail(email)
	if err != nil {
		app.Session.Put(r.Context(), "error", "Invalid login credentials")
		app.ErrorLog.Println(err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// check password
	validPassword, err := user.PasswordMatches(password)
	if err != nil {
		app.Session.Put(r.Context(), "error", "Invalid login credentials")
		app.ErrorLog.Println(err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if !validPassword {
		app.Session.Put(r.Context(), "error", "Invalid login credentials")
		app.ErrorLog.Println(err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// add user id to session
	app.Session.Put(r.Context(), "user_id", user.ID)
	app.Session.Put(r.Context(), "user", user)

	// flash successful login
	app.Session.Put(r.Context(), "flash", "Login successful")
	// redirect to home
	http.Redirect(w, r, "/", http.StatusSeeOther)

}

// Logout logs the user out
func (app *Config) Logout(w http.ResponseWriter, r *http.Request) {
	// clean up session
	_ = app.Session.Destroy(r.Context())
	_ = app.Session.RenewToken(r.Context())

	// flash successful logout
	app.Session.Put(r.Context(), "flash", "Logout successful")

	// redirect to home
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// RegistPage displays the registration page
func (app *Config) RegisterPage(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "register.page.gohtml", nil)
}

func (app *Config) PostRegisterPage(w http.ResponseWriter, r *http.Request) {
	// create a new user

	// send an activation email to the user

	// subscribe the user to an account
}

// ActivateAccount activates a user account
func (app *Config) ActivateAccount(w http.ResponseWriter, r *http.Request) {
	// validate url

	// generate an invoice

	// send an email with the plan attachments

	// send an email to the user with the invoice

}
