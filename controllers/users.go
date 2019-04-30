package controllers

import (
	"fmt"
	"net/http"
	"time"

	"lenslocked.com/cookies"

	"lenslocked.com/context"
	"lenslocked.com/email"
	"lenslocked.com/models"
	"lenslocked.com/rand"
	"lenslocked.com/views"
)

// NewUsers is used to create a new Users controller.
// This function will panic if the templates are not parsed
// correctly so should only be used during initial setup.
func NewUsers(us models.UserService, mc email.MailClient) *Users {
	return &Users{
		NewView:      views.NewView("bootstrap", "users/new"),
		LoginView:    views.NewView("bootstrap", "users/login"),
		ForgotPwView: views.NewView("bootstrap", "users/forgot_pw"),
		ResetPwView:  views.NewView("bootstrap", "users/reset_pw"),
		us:           us,
		emailer:      mc,
	}
}

type Users struct {
	NewView      *views.View
	LoginView    *views.View
	ForgotPwView *views.View
	ResetPwView  *views.View
	us           models.UserService
	emailer      email.MailClient
}

// New is used to render the signup form.
//
// GET /signup
func (u Users) New(w http.ResponseWriter, r *http.Request) {
	var form SignupForm
	ParseURLParams(r, &form)
	u.NewView.Render(w, r, form)
}

type SignupForm struct {
	Name     string `schema:"name"`
	Email    string `schema:"email"`
	Password string `schema:"password"`
}

// Create is used to process the signup form. This creates a new user account.
//
// POST /signup
func (u *Users) Create(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form SignupForm

	vd.Yield = &form
	if err := ParseForm(r, &form); err != nil {
		vd.SetAlert(err)
		u.NewView.Render(w, r, vd)
		return
	}

	user := models.User{
		Name:     form.Name,
		Email:    form.Email,
		Password: form.Password,
	}

	if err := u.us.Create(&user); err != nil {
		vd.SetAlert(err)
		u.NewView.Render(w, r, vd)
		return
	}

	err := u.signIn(w, &user)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// Send welcome email
	// u.emailer.Welcome(user.Name, user.Email)

	alert := views.Alert{
		Level:   views.AlertLvlSuccess,
		Message: "Welcome to LensLocked.com!",
	}
	views.RedirectAlert(w, r, "/galleries", http.StatusFound, alert)
}

type LoginForm struct {
	Email    string `schema:"email"`
	Password string `schema:"password"`
}

// Login is used to process the login form. This is used to verify a user's email
// and password and then log them in if correct.
//
// POST /login
func (u *Users) Login(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form LoginForm

	vd.Yield = &form
	if err := ParseForm(r, &form); err != nil {
		vd.SetAlert(err)
		u.LoginView.Render(w, r, vd)
		return
	}
	user, err := u.us.Authenticate(form.Email, form.Password)

	if err != nil {
		switch err {
		case models.ErrNotFound:
			vd.AlertError("Invalid email address")
		default:
			vd.SetAlert(err)
		}
		u.LoginView.Render(w, r, vd)
		return
	}

	err = u.signIn(w, user)
	if err != nil {
		vd.SetAlert(err)
		u.LoginView.Render(w, r, vd)
		return
	}
	alert := views.Alert{
		Level:   views.AlertLvlSuccess,
		Message: fmt.Sprintf("Welcome back%s!", " "+user.Name),
	}
	url := cookies.GetRedirect(r)
	cookies.ClearRedirect(w)
	if url == "" {
		url = "/galleries"
	}

	views.RedirectAlert(w, r, url, http.StatusFound, alert)
	//http.Redirect(w, r, "/galleries", http.StatusFound)
}

// Logout is used to delete a user's session cookie and update their remember token
// to prevent stored cookies being valid.
//
// POST /logout
func (u *Users) Logout(w http.ResponseWriter, r *http.Request) {
	cookie := http.Cookie{
		Name:     "remember_token",
		Value:    "",
		Expires:  time.Now(),
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)

	// Invalidate any stored remember_tokens
	user := context.User(r.Context())
	token, _ := rand.RememberToken()
	user.Remember = token
	u.us.UpdateRememberHash(user)

	http.Redirect(w, r, "/", http.StatusFound)
}

// ResetPwForm is used by both initiate and complete steps
type resetPwForm struct {
	Email    string `schema:"email"`
	Token    string `schema:"token"`
	Password string `schema:"password"`
}

// InitiateReset Process the forgot password form from /forgot
//
// POST /forgot
func (u *Users) InitiateReset(w http.ResponseWriter, r *http.Request) {
	// TODO: process frogot pw form and initiate process
	var vd views.Data
	var form resetPwForm

	vd.Yield = &form
	if err := ParseForm(r, &form); err != nil {
		vd.SetAlert(err)
		u.ForgotPwView.Render(w, r, vd)
		return
	}

	token, err := u.us.InitiateReset(form.Email)
	if err != nil {
		vd.SetAlert(err)
		u.ForgotPwView.Render(w, r, vd)
		return
	}

	_ = token

	// Send reset password email
	// err = u.emailer.ResetPw(form.Email, token)
	// if err != nil {
	// 	vd.SetAlert(err)
	// 	u.ForgotPwView.Render(w, r, vd)
	// 	return
	// }

	views.RedirectAlert(w, r, "/reset", http.StatusFound,
		views.AlertSuccess("Instructions for resetting your password have been emailed to you."),
	)
}

// ResetPw displays the reset password form and has a method
// to prefill the form data with the token emailed to the user via
// URL query params.
//
// GET /reset
func (u *Users) ResetPw(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form resetPwForm
	vd.Yield = &form

	if err := ParseURLParams(r, &form); err != nil {
		vd.SetAlert(err)
	}
	u.ResetPwView.Render(w, r, vd)
}

// CompleteReset processes the reset password form
//
// POST /reset
func (u *Users) CompleteReset(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form resetPwForm
	vd.Yield = &form

	if err := ParseForm(r, &form); err != nil {
		vd.SetAlert(err)
		u.ResetPwView.Render(w, r, vd)
		return
	}

	user, err := u.us.CompleteReset(form.Token, form.Password)
	if err != nil {
		vd.SetAlert(err)
		u.ResetPwView.Render(w, r, vd)
		return
	}
	u.signIn(w, user)
	views.RedirectAlert(w, r, "/galleries", http.StatusFound,
		views.AlertSuccess("Your password has been reset successfully!"),
	)
}

// signIn sets the cookie for the user's session
func (u *Users) signIn(w http.ResponseWriter, user *models.User) error {
	if user.Remember == "" {
		token, err := rand.RememberToken()
		if err != nil {
			return err
		}
		user.Remember = token
		err = u.us.UpdateRememberHash(user)
		if err != nil {
			return err
		}
	}
	// need to set cookie before writing to ResponseWriter
	cookie := http.Cookie{
		Name:     "remember_token",
		Value:    user.Remember,
		HttpOnly: true, // prevents non-http access to cookie e.g. from client javascript
	}
	http.SetCookie(w, &cookie)
	return nil
}

// CookieTest is used to display cookies set on the current user
func (u *Users) CookieTest(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("remember_token")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	user, err := u.us.ByRemember(cookie.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, "User is: ", user)
}
