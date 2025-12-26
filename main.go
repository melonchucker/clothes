package main

import (
	"clothes/controllers"
	"clothes/models"
	"clothes/scraper"
	"flag"
	"log/slog"
	"net/http"
)

func main() {
	slog.Info("Starting clothes app")

	scrapeBrand := flag.Bool("scrape", false, "Run scraper for given brand (nike, adidas, puma)")
	databaseMigrate := flag.Bool("migrate", false, "Run database migrations")
	flag.Parse()

	if *databaseMigrate {
		models.Migrate()
	}
	if *scrapeBrand {
		go scraper.ScrapeAll()
	}

	if err := http.ListenAndServe("192.168.1.36:8080", controllers.GetServerMux()); err != nil {
		slog.Error("Failed to start server", "error", err)
	}
}
