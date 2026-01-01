package controllers

import (
	"clothes/models"
	"encoding/json"
	"fmt"
	"net/http"
)

func GetApiMux() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /user/closets", func(w http.ResponseWriter, r *http.Request) {
		siteUser, err := getSession(w, r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		closets, err := models.ApiQuery[[]models.SiteUserCloset](r.Context(), "site_user_get_closets", siteUser.Username)
		if err != nil {
			http.Error(w, "Error querying database", http.StatusInternalServerError)
			return
		}

		data, err := json.Marshal(closets)
		if err != nil {
			http.Error(w, "Error serializing response", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	})

	mux.HandleFunc("POST /user/closets/add_item", func(w http.ResponseWriter, r *http.Request) {
		var info struct {
			ClosetName string `json:"closet_name"`
			Brand      string `json:"brand"`
			Item       string `json:"item"`
		}

		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&info); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		siteUser, err := getSession(w, r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		_, err = models.ApiQuery[string](r.Context(), "site_user_add_item_to_closet", siteUser.Username, info.ClosetName, info.Item, info.Brand)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Error adding item to closet", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("GET /search_bar", func(w http.ResponseWriter, r *http.Request) {
		// TODO
		input := r.URL.Query().Get("input")
		results, err := models.ApiQuery[models.SearchBar](r.Context(), "search_bar", input)
		if err != nil {
			http.Error(w, "Error querying database", http.StatusInternalServerError)
			return
		}

		fmt.Println(results)

		data, err := json.Marshal(results)
		if err != nil {
			http.Error(w, "Error serializing response", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	})
	return loggingMiddleware(mux)
}
