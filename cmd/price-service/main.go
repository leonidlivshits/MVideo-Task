package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	nethttp "net/http"
	"os"
	"time"

	"mvideo-task/internal/config"
	httpapi "mvideo-task/internal/http"
	"mvideo-task/internal/postgres"
	"mvideo-task/internal/service"
)

const (
	databaseConnectTimeout = 5 * time.Second
	exitFailure = 1
)

func main() {
	if err := run(); err != nil {
		slog.Error("application stopped", "error", err)
		os.Exit(exitFailure) // надо подумать
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), databaseConnectTimeout)
	defer cancel()

	pool, err := postgres.NewPool(ctx, cfg.Database.ConnString())
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer pool.Close()

	priceRepository := postgres.NewPriceRepository(pool)
	priceService := service.NewPriceService(priceRepository)
	server := httpapi.NewServer(cfg.HTTPAddr, priceService)

	slog.Info("starting HTTP server", "addr", cfg.HTTPAddr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, nethttp.ErrServerClosed) {
		return fmt.Errorf("listen and serve: %w", err)
	}

	return nil
}
