package controllers

import (
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
