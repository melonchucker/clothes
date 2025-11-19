package views

import (
	"log/slog"
	"net/http"
	"path/filepath"
	"text/template"
)

type UserDetails struct {
	Email           string
	IsAuthenticated bool
}

type PageData struct {
	UserDetails UserDetails
	Title       string
	Data        any
}

const layoutFile string = "views/base-page.gohtml"

func RenderPage(page string, w http.ResponseWriter, pageData PageData) {
	pageFile := filepath.Join("views/pages", page+".gohtml")

	templateFiles := []string{layoutFile, pageFile}

	widgetFiles, err := filepath.Glob("views/widgets/*.gohtml")
	if err != nil {
		slog.Error("Error finding widget templates:", slog.Any("err", err))
		http.Error(w, "Error loading templates", http.StatusInternalServerError)
		return
	}
	templateFiles = append(templateFiles, widgetFiles...)

	tmpl, err := template.ParseFiles(templateFiles...)
	if err != nil {
		slog.Error("Error loading templates:", slog.Any("err", err))
		http.Error(w, "Error loading templates", http.StatusInternalServerError)
		return
	}

	funcs := template.FuncMap{
		"repeat": func(n int) []struct{} {
			return make([]struct{}, n)
		},
	}

	tmpl = tmpl.Funcs(funcs)

	if err := tmpl.Execute(w, pageData); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

func RenderWidget(widget string, w http.ResponseWriter, data any) {
	widgetFile := "views/widgets/widgets.gohtml"

	tmpl, err := template.ParseFiles(widgetFile)
	if err != nil {
		slog.Error("Error loading widget template:", slog.Any("err", err))
		http.Error(w, "Error loading widget template", http.StatusInternalServerError)
		return
	}
	if err := tmpl.ExecuteTemplate(w, widget, data); err != nil {
		slog.Error("Error rendering widget template:", slog.Any("err", err))
		http.Error(w, "Error rendering widget template", http.StatusInternalServerError)
	}
}
