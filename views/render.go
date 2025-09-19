package views

import (
	"log/slog"
	"net/http"
	"path/filepath"
	"text/template"
)

type PageData struct {
	Title string
	Data  any
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

	if err := tmpl.Execute(w, pageData); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}
