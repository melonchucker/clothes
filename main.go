package main

import (
	"clothes/controllers"
	"clothes/models"
	"context"
	"log/slog"
	"net/http"
)

func main() {
	slog.Info("Starting clothes app")

	rows, err := models.GetDb().Query(context.Background(), "SELECT 1;")
	if err != nil {
		slog.Error("Failed to query database", "error", err)
		return
	}
	// defer rows.Close()
	slog.Info("Database query successful")

	for rows.Next() {
		var n int
		if err := rows.Scan(&n); err != nil {
			slog.Error("Failed to scan row", "error", err)
			return
		}
		slog.Info("Query result", "value", n)
	}

	if err := http.ListenAndServe(":8080", controllers.GetServerMux()); err != nil {
		slog.Error("Failed to start server", "error", err)
	}

}
