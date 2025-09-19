package controllers

import (
	"clothes/models"
	"clothes/views"
	"context"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5"
)

func GetServerMux() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		views.RenderPage("home", w, views.PageData{Title: "Home Page"})
	})

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	mux.HandleFunc("GET /clothes", func(w http.ResponseWriter, r *http.Request) {
		rows, err := models.GetDb().Query(context.Background(), "SELECT * FROM api.get_base_items()")
		if err != nil {
			http.Error(w, "Error querying database", http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		data, err := pgx.CollectRows(rows, pgx.RowToMap)
		if err != nil {
			http.Error(w, "Error collecting rows", http.StatusInternalServerError)
			return
		}
		fmt.Println(data[0]["get_base_items"])
		views.RenderPage("browse", w, views.PageData{Title: "Clothes", Data: data[0]["get_base_items"]})
	})

	mux.HandleFunc("GET /item/{id}", func(w http.ResponseWriter, r *http.Request) {
		views.RenderPage("item", w, views.PageData{Title: "Item Detail"})
	})

	return gzipMiddleware(loggingMiddleware(mux))
}
