package models

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"slices"

	"github.com/jackc/pgx/v5/pgxpool"
)

const sqlDir = "models/sql"

var pool *pgxpool.Pool

func GetDb() *pgxpool.Pool {
	return pool
}

func init() {
	connectionString := fmt.Sprintf("postgresql:///postgres?user=%s", os.Getenv("USER"))
	slog.Info("Connecting to database", "connectionString", connectionString)
	p, err := pgxpool.New(context.Background(), connectionString)
	if err != nil {
		slog.Error("Failed to create connection pool", "error", err)
		os.Exit(1)
	}

	files, err := os.ReadDir(sqlDir)
	if err != nil {
		slog.Error("Failed to read sql directory", "error", err)
		os.Exit(1)
	}

	sqlFilePaths := []string{}
	for _, f := range files {
		if f.IsDir() {
			slog.Warn("Skipping directory in sqlDir", "name", f.Name())
			continue
		}
		sqlFilePaths = append(sqlFilePaths, fmt.Sprintf("%s/%s", sqlDir, f.Name()))
	}

	slices.Sort(sqlFilePaths)

	// get a single connection from the pool
	for _, f := range sqlFilePaths {
		slog.Info("Executing sql file", "name", f)
		content, err := os.ReadFile(f)
		if err != nil {
			slog.Error("Failed to read SQL file", "file", f, "error", err)
			os.Exit(1)
		}
		_, err = p.Exec(context.Background(), string(content))
		if err != nil {
			slog.Error("Failed to execute SQL file", "file", f, "error", err)
			os.Exit(1)
		}
	}

	slog.Info("Database connection pool established and schema initialized")
	pool = p
}
