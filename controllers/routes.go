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

func NewPageData(w http.ResponseWriter, r *http.Request, title string, data any) views.PageData {
	slog.Info("Creating PageData for", "title", title)
	pd := views.PageData{
		Title: title,
		Data:  data,
	}

	c, err := r.Cookie("alert")
	if err == nil && c.Value != "" {
		fmt.Println("Alert cookie value:", c.Value)
		pd.Alert = &widgets.Alert{
			Message: c.Value,
			Level:   widgets.AlertLevelDanger,
		}
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "alert",
		Value:    "",
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		MaxAge:   -1,
		SameSite: http.SameSiteLaxMode,
	})

	c, err = r.Cookie("session_token")
	if err != nil || c.Value == "" {
		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    "",
			HttpOnly: true,
			Secure:   true,
			Path:     "/",
			MaxAge:   -1,
			SameSite: http.SameSiteLaxMode,
		})

		return pd
	}

	pd.SiteUser, err = models.ApiQuery[models.SiteUser](r.Context(), "user_validate_session", c.Value)
	if err != nil {
		slog.Error("Error validating session token", "error", err)
	}

	return pd
}

func GetAuthenticatedServerMux() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		slog.Info("AuthentHicated request to /account/")
		views.RenderPage("account", w, NewPageData(w, r, "Account", nil))
	})

	return authenticateMiddleware(mux)
}

func GetServerMux() http.Handler {
	mux := http.NewServeMux()

	// everything under account is authenticated
	mux.Handle("/account/", http.StripPrefix("/account", GetAuthenticatedServerMux()))

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		views.RenderPage("home", w, NewPageData(w, r, "Home", nil))
	})

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	mux.HandleFunc("GET /api/search_bar", func(w http.ResponseWriter, r *http.Request) {
		// TODO
	})

	mux.HandleFunc("GET /brands", func(w http.ResponseWriter, r *http.Request) {
		brands, err := models.ApiQuery[models.Brands](r.Context(), "brands")
		if err != nil {
			http.Error(w, "Error querying database", http.StatusInternalServerError)
			return
		}

		views.RenderPage("brands", w, NewPageData(w, r, "Brands", brands))
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

		slog.Info("Browsing with tags", "tags", tags)
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

		title := strings.Title(strings.ReplaceAll(topLevelTag, "_", " "))

		views.RenderPage("browse", w, NewPageData(w, r, title, data))
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

		views.RenderPage("browse", w, NewPageData(w, r, "Clothes", data))
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
			Brand     string
			ItemName  string
			Rating    widgets.Rating
			ImageUrls []string
			SizeInfo  []struct {
				Size    string
				InStock bool
			}
			Details struct {
				Description string
			}
			Tags         []string
			MoreOfBrand  widgets.MoreLike
			SimilarItems widgets.MoreLike
		}{
			Brand:    detail.BrandName,
			ItemName: detail.ItemName,
			Rating: widgets.Rating{
				Rating: detail.Rating,
				Max:    5,
			},
			ImageUrls: detail.ImageUrls,
			SizeInfo: []struct {
				Size    string
				InStock bool
			}(detail.ItemSpecificDetails),
			Details: struct {
				Description string
			}{
				Description: detail.Description,
			},
			Tags:         detail.Tags,
			MoreOfBrand:  widgets.MoreLike{Title: "More from " + detail.BrandName},
			SimilarItems: widgets.MoreLike{Title: "Similar to this"},
		}
		for _, item := range detail.MoreLike.SameBrand {
			card := widgets.Card{
				Title:    item.ItemName,
				Content:  item.BrandName,
				ImageURL: fmt.Sprintf("/static/images/%s", item.ThumbnailUrl),
				ImageAlt: fmt.Sprintf("%s %s", item.BrandName, item.ItemName),
				Href:     strings.ToLower(fmt.Sprintf("/item/%s/%s", item.BrandName, item.ItemName)),
			}
			data.MoreOfBrand.Items = append(data.MoreOfBrand.Items, card)
		}
		for _, item := range detail.MoreLike.SimilarItems {
			card := widgets.Card{
				Title:    item.ItemName,
				Content:  item.BrandName,
				ImageURL: fmt.Sprintf("/static/images/%s", item.ThumbnailUrl),
				ImageAlt: fmt.Sprintf("%s %s", item.BrandName, item.ItemName),
				Href:     strings.ToLower(fmt.Sprintf("/item/%s/%s", item.BrandName, item.ItemName)),
			}
			data.SimilarItems.Items = append(data.SimilarItems.Items, card)
		}

		imageUrls := []string{}
		for _, img := range detail.ImageUrls {
			imageUrls = append(imageUrls, fmt.Sprintf("/static/images/%s", img))
		}
		data.ImageUrls = imageUrls

		views.RenderPage("detail", w, NewPageData(w, r, detail.ItemName, data))
	})

	mux.HandleFunc("GET /sign-in", func(w http.ResponseWriter, r *http.Request) {
		views.RenderPage("sign-in", w, NewPageData(w, r, "Sign In", nil))
	})

	mux.HandleFunc("POST /sign-in", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}
		email := r.FormValue("email")
		password := r.FormValue("password")

		session, err := models.ApiQuery[string](r.Context(), "site_user_authenticate", email, password)
		if err != nil || session == nil {
			slog.Error("Error signing in user", "error", err)
			http.SetCookie(w, &http.Cookie{
				Name:     "alert",
				Value:    "Error signing in user",
				HttpOnly: true,
				Secure:   true,
				Path:     "/",
				SameSite: http.SameSiteLaxMode,
			})
			// redirect to sign-in
			http.Redirect(w, r, "/sign-in", http.StatusSeeOther)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    *session,
			HttpOnly: true,
			Secure:   true,
			Path:     "/",
			SameSite: http.SameSiteLaxMode,
		})

		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	mux.HandleFunc("GET /sign-up", func(w http.ResponseWriter, r *http.Request) {
		views.RenderPage("sign-up", w, NewPageData(w, r, "Sign Up", nil))
	})

	mux.HandleFunc("POST /sign-up", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}
		firstName := r.FormValue("first_name")
		lastName := r.FormValue("last_name")
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")

		session, err := models.ApiQuery[string](r.Context(), "site_user_signup", firstName, lastName, username, email, password)
		if err != nil {
			slog.Error("Error signing up user", "error", err)
			http.Error(w, "Error signing up user", http.StatusInternalServerError)
			return
		}

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
			Path:     "/",
			SameSite: http.SameSiteLaxMode,
		})

		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	mux.HandleFunc("GET /sign-out", func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("session_token")
		if err != nil || c.Value == "" {
			http.Redirect(w, r, "/sign-in", http.StatusSeeOther)
			return
		}

		_, err = models.ApiQuery[string](r.Context(), "user_signout", c.Value)
		if err != nil {
			slog.Error("Error signing out user", "error", err)
			http.Error(w, "Error signing out user", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    "",
			HttpOnly: true,
			Secure:   true,
		})
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		views.RenderPage("404", w, NewPageData(w, r, "Page Not Found", nil))
	}))

	return loggingMiddleware(gzipMiddleware(mux))
}
