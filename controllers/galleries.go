package controllers

import (
	"fmt"
	"net/http"

	"lenslocked.com/models"
	"lenslocked.com/views"
)

// NewGalleries is used to create a new Galleries controller.
// This function will panic if the templates are not parsed
// correctly so should only be used during initial setup.
func NewGalleries(gs models.GalleryService) *Galleries {
	return &Galleries{
		New: views.NewView("bootstrap", "galleries/new"),
		gs:  gs,
	}
}

type Galleries struct {
	New *views.View
	gs  models.GalleryService
}

type GalleryForm struct {
	Title string `schema:"title"`
}

// Create is used to process the signup form. This creates a new user account.
//
// POST /signup
func (g *Galleries) Create(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form GalleryForm

	if err := ParseForm(r, &form); err != nil {
		vd.SetAlert(err)
		g.New.Render(w, vd)
		return
	}

	gallery := models.Gallery{
		Title: form.Title,
	}

	if err := g.gs.Create(&gallery); err != nil {
		vd.SetAlert(err)
		g.New.Render(w, vd)
		return
	}
	fmt.Fprintln(w, gallery)
}
