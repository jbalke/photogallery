package views

import (
	"html/template"
	"net/http"
	"path/filepath"
)

const (
	TemplateDir = "views/"
	LayoutDir   = "views/layouts/"
	TemplateExt = ".gohtml"
)

func NewView(layout string, files ...string) *View {
	addTemplatePathAndExt(files)
	files = append(files, layoutFiles()...)

	t, err := template.ParseFiles(files...)
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
	if err := v.Render(w, nil); err != nil {
		panic(err)
	}
}

func (v *View) Render(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "text/html")
	// If a Data type is not passed in, parse data as a Data object.
	switch data.(type) {
	case Data:
		// do nothing
	default:
		data = Data{
			Yield: data,
		}

	}
	return v.Template.ExecuteTemplate(w, v.Layout, data)
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
