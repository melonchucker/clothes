package controllers

import (
	"clothes/models"
	"clothes/views"
	"clothes/views/widgets"
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
		rows, err := models.GetDb().Query(r.Context(), "SELECT * FROM api.browse()")
		if err != nil {
			http.Error(w, "Error querying database", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		x, err := models.ApiQuery(r.Context(), "browse", 5)
		if err != nil {
			http.Error(w, "Error querying database", http.StatusInternalServerError)
			return
		}
		fmt.Println(x)

		data, err := pgx.CollectRows(rows, pgx.RowToMap)
		if err != nil {
			http.Error(w, "Error collecting rows", http.StatusInternalServerError)
			return
		}

		cards := []widgets.Card{}
		itemsIface, ok := models.ApiQuery(r.Context(), "browse").([]interface{})
		if !ok {
			http.Error(w, "Unexpected data format from database", http.StatusInternalServerError)
			return
		}

		for _, it := range itemsIface {
			row, ok := it.(map[string]any)
			if !ok {
				continue
			}

			itemName, _ := row["item_name"].(string)
			brandName, _ := row["brand_name"].(string)
			card := widgets.Card{
				Title:    itemName,
				Content:  brandName,
				ImageURL: fmt.Sprintf("/static/images/%v", row["thumbnail_url"]),
				ImageAlt: fmt.Sprintf("%s %s", brandName, itemName),
				Href:     fmt.Sprintf("/item/%v", row["sku"]),
			}

			cards = append(cards, card)
		}
		views.RenderPage("browse", w, views.PageData{Title: "Clothes", Data: cards})
	})

	mux.HandleFunc("GET /item/{id}", func(w http.ResponseWriter, r *http.Request) {
		views.RenderPage("detail", w, views.PageData{Title: "Item Detail", Data: map[string]any{
			"ImageUrl": "/static/images/a-line-mini-dress-sgntr-the-label-neutral-gingham-614--1.jpg",
			"Brand":    "Brand Name",
			"ItemName": "Item Name",
			"Color":    "Color",
			"Rating":   widgets.Rating{Rating: 3.25, Max: 5},
		}})
	})

	return gzipMiddleware(loggingMiddleware(mux))
}
