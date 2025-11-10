package main

import (
	"clothes/controllers"
	"clothes/scraper"
	"flag"
	"log/slog"
	"net/http"
)

func main() {
	slog.Info("Starting clothes app")

	scrapeBrand := flag.Bool("scrape", false, "Run scraper for given brand (nike, adidas, puma)")
	flag.Parse()

	if *scrapeBrand {
		go scraper.ScrapeAll()
	}

	if err := http.ListenAndServe(":8080", controllers.GetServerMux()); err != nil {
		slog.Error("Failed to start server", "error", err)
	}
}
