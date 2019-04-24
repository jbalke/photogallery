package views

import (
	"bytes"
	"errors"
	"html/template"
	"io"
	"log"
	"net/http"
	"path/filepath"

	"github.com/gorilla/csrf"

	"lenslocked.com/context"
)

const (
	TemplateDir = "views/"
	LayoutDir   = "views/layouts/"
	TemplateExt = ".gohtml"
)

func NewView(layout string, files ...string) *View {
	addTemplatePathAndExt(files)
	files = append(files, layoutFiles()...)

	// Define csrfField function stub so template compiles
	// but update at Render with required bits
	t, err := template.New("").Funcs(template.FuncMap{
		"csrfField": func() (template.HTML, error) {
			return "", errors.New("csrfField is not implemented")
		},
	}).ParseFiles(files...)
	if err != nil {
		panic(err)
	}

	return &View{Template: t, Layout: layout}
}

type View struct {
	Template *template.Template
	Layout   string
}

func (v *View) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	v.Render(w, r, nil)
}

func (v *View) Render(w http.ResponseWriter, r *http.Request, data interface{}) {
	w.Header().Set("Content-Type", "text/html")
	// If a Data type is not passed in, parse data as a Data object.
	var vd Data
	switch d := data.(type) {
	case Data:
		vd = d
	default:
		vd = Data{
			Yield: data,
		}
	}
	vd.User = context.User(r.Context())
	// write to buf to capture any errors and copy to w if all is ok.
	var buf bytes.Buffer
	csrfField := csrf.TemplateField(r)
	tpl := v.Template.Funcs(template.FuncMap{
		"csrfField": func() template.HTML {
			return csrfField
		},
	})
	if err := tpl.ExecuteTemplate(&buf, v.Layout, vd); err != nil {
		log.Println(err)
		http.Error(w, "Something went wrong. If the problem persists, please email support@lenslocked.com", http.StatusInternalServerError)
		return
	}
	io.Copy(w, &buf)
}

// layoutFiles returns a slice of strings representing the layout files used in this application.
func layoutFiles() []string {
	files, err := filepath.Glob(LayoutDir + "*" + TemplateExt)
	if err != nil {
		panic(err)
	}

	return files
}

// addTemplatePathAndExt takes a slice of strings
// representing file paths for templates and it prepends
// the TemplateDir directory and appends TemplateExt to each string in the slice
func addTemplatePathAndExt(files []string) {
	for i, f := range files {
		files[i] = TemplateDir + f + TemplateExt
	}
}

// addTemplatePath takes a slice of strings
// representing file paths for templates and it prepends
// the TemplateDir directory to each string in the slice
func addTemplatePath(files []string) {
	for i, f := range files {
		files[i] = TemplateDir + f
	}
}

// addTemplateExt takes a slice of strings
// representing file paths for templates and it appends
// the TemplateExt to each string in the slice
func addTemplateExt(files []string) {
	for i, f := range files {
		files[i] = f + TemplateExt
	}
}
