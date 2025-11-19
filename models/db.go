package models

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const sqlDir = "models/sql"

var pool *pgxpool.Pool

func GetDb() *pgxpool.Pool {
	return pool
}

func ApiQuery[T any](ctx context.Context, apiFunction string, args ...any) (*T, error) {
	argsStrings := make([]string, len(args))
	for ndx := range args {
		argsStrings[ndx] = fmt.Sprintf("$%d", ndx+1)
	}
	argsString := strings.Join(argsStrings, ", ")

	rows, err := pool.Query(ctx, fmt.Sprintf("SELECT * FROM api.%s(%s) AS result", apiFunction, argsString), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type resultStruct struct {
		Result T `json:"result"`
	}
	res, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[resultStruct])

	if err != nil {
		return nil, err
	}
	return &res.Result, nil
}

func init() {
	connectionString := fmt.Sprintf("postgresql:///postgres?user=%s", os.Getenv("USER"))
	slog.Info("Connecting to database", "connectionString", connectionString)
	p, err := pgxpool.New(context.Background(), connectionString)
	if err != nil {
		slog.Error("Failed to create connection pool", "error", err)
		os.Exit(1)
	}

	slog.Info("Database connection pool established and schema initialized")
	pool = p
}

func Migrate() {
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

	for _, f := range sqlFilePaths {
		slog.Info("Executing sql file", "name", f)
		content, err := os.ReadFile(f)
		if err != nil {
			slog.Error("Failed to read SQL file", "file", f, "error", err)
			os.Exit(1)
		}
		_, err = pool.Exec(context.Background(), string(content))
		if err != nil {
			slog.Error("Failed to execute SQL file", "file", f, "error", err)
			os.Exit(1)
		}
	}
}
