package main

import (
	"clothes/controllers"
	"clothes/models"
	"clothes/scraper"
	"flag"
	"log/slog"
	"net/http"

	"github.com/evanw/esbuild/pkg/api"
)

func BuildWebApps(entrypoints ...string) {
	result := api.Build(api.BuildOptions{
		EntryPoints:       entrypoints,
		Bundle:            true,
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
		Engines: []api.Engine{
			{Name: api.EngineChrome, Version: "60"},
			{Name: api.EngineFirefox, Version: "60"},
			{Name: api.EngineSafari, Version: "12"},
			{Name: api.EngineEdge, Version: "79"},
		},
		Write:    true,
		Tsconfig: "./webcomponents/tsconfig.json",
		Outfile:  "static/apps/bundle.js",
	})

	if len(result.Errors) > 0 {
		panic("Error building web apps: " + result.Errors[0].Text)
	}
}

func main() {
	slog.Info("Starting clothes app")

	scrapeBrand := flag.Bool("scrape", false, "Run scraper for given brand (nike, adidas, puma)")
	databaseMigrate := flag.Bool("migrate", false, "Run database migrations")
	flag.Parse()

	BuildWebApps("webcomponents/src/_bundle.ts")

	if *databaseMigrate {
		models.Migrate()
	}
	if *scrapeBrand {
		go scraper.ScrapeAll()
	}

	if err := http.ListenAndServe(":8080", controllers.GetServerMux()); err != nil {
		slog.Error("Failed to start server", "error", err)
	}
}
