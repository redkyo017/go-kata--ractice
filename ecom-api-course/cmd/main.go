package main

import (
	"context"
	"ecom-api/internal/env"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5"
)

func main() {
	ctx := context.Background()
	cfg := config{
		addr: ":8080",
		db: dbConfig{
			dsn: env.GetString("GOOSE_DBSTRING", "host=localhost user=postgres password=postgres dbname=ecom sslmode=disable"),
		},
	}
	// Logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Databatse
	conn, err := pgx.Connect(ctx, cfg.db.dsn)
	if err != nil {
		panic(err)
	}
	defer conn.Close(ctx)

	logger.Info("connected to database", "dsn", cfg.db.dsn)

	api := &application{
		config: cfg,
		db:     conn,
	}

	if err := api.run(api.mount()); err != nil {
		// log.Println("server has failed to start, err: %s", err)
		slog.Error("server failed to start", "error", err)
		os.Exit(1)
	}
}
