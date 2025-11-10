package controllers

import (
	"clothes/models"
	"clothes/views"
	"clothes/views/widgets"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
)

func GetServerMux() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		views.RenderPage("home", w, views.PageData{Title: "Home Page"})
	})

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	mux.HandleFunc("GET /api/search_bar", func(w http.ResponseWriter, r *http.Request) {
		// searchString := r.URL.Query().Get("input")
		// searchBarResult, err := models.ApiQuery[models.SearchBar](r.Context(), "search_bar", searchString)
		// if err != nil {
		// 	http.Error(w, "Error querying database", http.StatusInternalServerError)
		// 	return
		// }

		views.RenderWidget("searchResults", w, nil)
		// w.Header().Set("Content-Type", "application/json")
		// if err := json.NewEncoder(w).Encode(searchBarResult); err != nil {
		// 	http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		// 	return
		// }
	})

	mux.HandleFunc("GET /brands", func(w http.ResponseWriter, r *http.Request) {
		brands, err := models.ApiQuery[models.Brands](r.Context(), "brands")
		if err != nil {
			http.Error(w, "Error querying database", http.StatusInternalServerError)
			return
		}

		views.RenderPage("brands", w, views.PageData{Title: "Brands", Data: brands})
	})

	mux.HandleFunc("GET /browse/{top_level_tag}", func(w http.ResponseWriter, r *http.Request) {
		topLevelTag := r.PathValue("top_level_tag")
		// these are other filter tags
		tags := r.URL.Query()["tag"]
		tags = append(tags, topLevelTag)

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

		items, err := models.ApiQuery[models.Browse](r.Context(), "browse", page, pageSize, tags)
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

		views.RenderPage("browse", w, views.PageData{Title: topLevelTag, Data: data})
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
			SizeInfo    []struct {
				Size    string
				InStock bool
			}
			Details struct {
				Description string
			}
		}{
			Brand:    detail.BrandName,
			ItemName: detail.ItemName,
			Rating: widgets.Rating{
				Rating: detail.Rating,
				Max:    5,
			},
			ImageViewer: widgets.ImageViewer{
				ImageUrls: []string{},
			},
			SizeInfo: []struct {
				Size    string
				InStock bool
			}(detail.ItemSpecificDetails),
			Details: struct {
				Description string
			}{
				Description: detail.Description,
			},
		}

		imageUrls := []string{}
		for _, img := range detail.ImageUrls {
			imageUrls = append(imageUrls, fmt.Sprintf("/static/images/%s", img))
		}
		data.ImageViewer.ImageUrls = imageUrls

		views.RenderPage("detail", w, views.PageData{Title: "Item Detail", Data: data})
	})

	mux.HandleFunc("GET /sign-in", func(w http.ResponseWriter, r *http.Request) {
		views.RenderPage("sign-in", w, views.PageData{Title: "Sign In", Data: nil})
	})

	mux.HandleFunc("POST /sign-in", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}
		email := r.FormValue("email")
		password := r.FormValue("password")

		session, err := models.ApiQuery[string](r.Context(), "user_authenticate", email, password)
		if err != nil {
			slog.Error("Error signing in user", "error", err)
			http.Error(w, "Error signing in user", http.StatusInternalServerError)
			return
		}

		if session == nil {
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    *session,
			HttpOnly: true,
			Secure:   true,
		})

		http.Redirect(w, r, "/", http.StatusSeeOther)

	})

	mux.HandleFunc("GET /sign-up", func(w http.ResponseWriter, r *http.Request) {
		views.RenderPage("sign-up", w, views.PageData{Title: "Sign Up", Data: nil})
	})

	mux.HandleFunc("POST /sign-up", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}
		email := r.FormValue("email")
		password := r.FormValue("password")

		_, err := models.ApiQuery[string](r.Context(), "user_signup", email, password)
		if err != nil {
			slog.Error("Error signing up user", "error", err)
			http.Error(w, "Error signing up user", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)

	})

	return gzipMiddleware(loggingMiddleware(mux))
}
