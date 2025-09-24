package controllers

import (
	"clothes/models"
	"clothes/views"
	"clothes/views/widgets"
	"fmt"
	"net/http"
)

func GetServerMux() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		views.RenderPage("home", w, views.PageData{Title: "Home Page"})
	})

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	mux.HandleFunc("GET /clothes", func(w http.ResponseWriter, r *http.Request) {
		pageStr := r.URL.Query().Get("page")
		if pageStr == "" {
			pageStr = "1"
		}
		var page int
		_, err := fmt.Sscanf(pageStr, "%d", &page)
		if err != nil || page < 1 {
			http.Error(w, "Invalid page number", http.StatusBadRequest)
			return
		}

		pageSizeStr := r.URL.Query().Get("pageSize")
		if pageSizeStr == "" {
			pageSizeStr = "20"
		}
		var pageSize int
		_, err = fmt.Sscanf(pageSizeStr, "%d", &pageSize)
		if err != nil || pageSize < 1 || pageSize > 100 {
			http.Error(w, "Invalid 'pageSize' number", http.StatusBadRequest)
			return
		}

		items, err := models.ApiQuery[models.Browse](r.Context(), "browse", page, pageSize)
		if err != nil {
			http.Error(w, "Error querying database", http.StatusInternalServerError)
			return
		}

		cards := []widgets.Card{}
		for _, item := range items.Items {
			card := widgets.Card{
				Title:    item.ItemName,
				Content:  item.BrandName,
				ImageURL: fmt.Sprintf("/static/images/%s", item.ThumbnailUrl),
				ImageAlt: fmt.Sprintf("%s %s", item.BrandName, item.ItemName),
				Href:     fmt.Sprintf("/item/%s", item.ItemName),
			}
			cards = append(cards, card)
		}

		data := struct {
			Cards      []widgets.Card
			Pagination widgets.Pageination
		}{
			Cards: cards,
			Pagination: widgets.Pageination{
				CurrentPage: page,
				TotalPages:  items.TotalPages,
				BaseURL:     *r.URL,
			},
		}

		views.RenderPage("browse", w, views.PageData{Title: "Clothes", Data: data})
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
