package controllers

import (
	"clothes/models"
	"clothes/views"
	"clothes/views/widgets"
	"fmt"
	"net/http"
	"strings"
)

func GetServerMux() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		views.RenderPage("home", w, views.PageData{Title: "Home Page"})
	})

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	mux.HandleFunc("GET /brands", func(w http.ResponseWriter, r *http.Request) {
		brands, err := models.ApiQuery[models.Brands](r.Context(), "brands")
		if err != nil {
			http.Error(w, "Error querying database", http.StatusInternalServerError)
			return
		}

		fmt.Println(brands)

		views.RenderPage("brands", w, views.PageData{Title: "Brands", Data: brands})
	})

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
				Href:     strings.ToLower(fmt.Sprintf("/item/%s/%s", item.BrandName, item.ItemName)),
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

	mux.HandleFunc("GET /item/{brand_name}/{base_item_name}", func(w http.ResponseWriter, r *http.Request) {
		brandName := r.PathValue("brand_name")
		baseItemName := r.PathValue("base_item_name")

		detail, err := models.ApiQuery[models.Detail](r.Context(), "detail", baseItemName, brandName)
		if err != nil {
			http.Error(w, "Error querying database", http.StatusInternalServerError)
			return
		}

		data := struct {
			Brand       string
			ItemName    string
			Rating      widgets.Rating
			ImageViewer widgets.ImageViewer
		}{
			Brand:    detail.BrandName,
			ItemName: detail.ItemName,
			Rating: widgets.Rating{
				Rating: 3.5,
				Max:    5,
			},
			ImageViewer: widgets.ImageViewer{
				ImageUrls: []string{},
			},
		}

		imageUrls := []string{}
		for _, img := range detail.ImageUrls {
			imageUrls = append(imageUrls, fmt.Sprintf("/static/images/%s", img))
		}
		data.ImageViewer.ImageUrls = imageUrls

		views.RenderPage("detail", w, views.PageData{Title: "Item Detail", Data: data})
	})
	// 	}
	// 	views.RenderPage("detail", w, views.PageData{Title: "Item Detail", Data: map[string]any{
	// 		"ImageUrl":  fmt.Sprintf("/static/images/%s", detail.ThumbnailUrl),
	// 		"Brand":     detail.BrandName,
	// 		"ItemName":  detail.ItemName,
	// 		"Color":     "Color",
	// 		"Rating":    widgets.Rating{Rating: 3.25, Max: 5},
	// 		"ImageUrls": imageUrls,
	// 	}})
	// })

	return gzipMiddleware(loggingMiddleware(mux))
}
