package controllers

import (
	"lenslocked.com/views"
)

func NewStatic() *Static {
	return &Static{
		Home:    views.NewView("bootstrap", "views/static/home.gohtml"),
		Contact: views.NewView("bootstrap", "views/static/contact.gohtml"),
	}
}

type Static struct {
	HomeView    *views.View
	ContactView *views.View
}
