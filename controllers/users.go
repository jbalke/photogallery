package controllers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"lenslocked.com/context"
	"lenslocked.com/email"
	"lenslocked.com/models"
	"lenslocked.com/rand"
	"lenslocked.com/views"
)

const (
	emailTextBody = `Hi %s!
	
	Welcome to Lenslocked.com, we hope you enjoy our site!
	
	Best Wishes
	John`

	emailHTMLBody = `<p>Hi %s!</p>
	<p>Welcome to <a href="https://www.lenslocked.com">Lenslocked.com</a>!</p>
	<p>Best Wishes<br>John</p>`

	emailSubject = "Welcome to Lenslocked.com!"
)

// NewUsers is used to create a new Users controller.
// This function will panic if the templates are not parsed
// correctly so should only be used during initial setup.
func NewUsers(us models.UserService, mc email.Client) *Users {
	return &Users{
		NewView:   views.NewView("bootstrap", "users/new"),
		LoginView: views.NewView("bootstrap", "users/login"),
		us:        us,
		mc:        mc,
	}
}

type Users struct {
	NewView   *views.View
	LoginView *views.View
	us        models.UserService
	mc        email.Client
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

	//emailText := fmt.Sprintf(emailTextBody, user.Name)
	//emailHTML := fmt.Sprintf(emailHTMLBody, user.Name)

	// Send welcome email
	// u.mc.Send(user.Name, user.Email, emailSubject, emailText, emailHTML)

	alert := views.Alert{
		Level:   views.AlertLvlSuccess,
		Message: "Welcome to LensLocked.com!",
	}
	//http.Redirect(w, r, "/galleries", http.StatusFound)
	views.RedirectAlert(w, r, "/galleries", http.StatusFound, alert)
}

type LoginForm struct {
	Email    string `schema:"email"`
	Password string `schema:"password"`
	Ref      string `schema:"ref"`
}

// LoginWithRef renders the login view and parses the "ref" url param that is the url
// the user requested before being redirected to login by our requireuser middleware.
//
// GET /login
func (u *Users) LoginWithRef(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form LoginForm

	vd.Yield = &form
	err := ParseURLParams(r, &form)
	log.Println(err)
	u.LoginView.Render(w, r, vd)
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
	var url string
	if form.Ref != "" {
		url = form.Ref
	} else {
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
