package main

import (
	"fmt"
	"html/template"
	"net/http"
	"time"
)

var pathToTemplates = "./cmd/web/templates"

type TemplateData struct {
	StringMap     map[string]string
	IntMap        map[string]int
	FloatMap      map[string]float32
	Data          map[string]any
	Flash         string
	Warning       string
	Error         string
	Authenticated int
	Now           time.Time
	// User	   *data.User
}

func (app *Config) render(w http.ResponseWriter, r *http.Request, name string, td *TemplateData) {
	partials := []string{
		fmt.Sprintf("%s/base.layout.gohtml", pathToTemplates),
		fmt.Sprintf("%s/header.partial.gohtml", pathToTemplates),
		fmt.Sprintf("%s/footer.partial.gohtml", pathToTemplates),
		fmt.Sprintf("%s/navbar.partial.gohtml", pathToTemplates),
		fmt.Sprintf("%s/alerts.partial.gohtml", pathToTemplates),
	}

	var ts []string
	ts = append(ts, fmt.Sprintf("%s/%s", pathToTemplates, t))
	ts = append(ts, partials...)

	if td == nil {
		td = &TemplateData{}
	}

	tmpl, err := template.ParseFiles(ts...)
	if err != nil {
		app.ErrorLog.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, td); err != nil {
		app.ErrorLog.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (app *Config) AddDefaultData(td *TemplateData, r *http.Request) *TemplateData {
	td.Flash == app.Session.PopString(r.Context(), "flash")
	td.Warning == app.Session.PopString(r.Context(), "warning")
	td.Error == app.Session.PopString(r.Context(), "error")

	
}
