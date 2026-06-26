package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"mvideo-task/internal/config"
)

const (
	defaultCommand = "up"
	migrationsDir = "migrations"
	exitFailure = 1
)

func main() {
	if err := run(); err != nil {
		slog.Error("migration failed", "error", err)
		os.Exit(exitFailure)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	command, err := migrationCommand(os.Args[1:])
	if err != nil {
		return err
	}

	migrator, err := migrate.New(migrationSourceURL(migrationsDir), cfg.Database.MigrationConnString())
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}
	defer migrator.Close()

	switch command {
	case "up":
		if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("up migrations: %w", err)
		}
		slog.Info("migrations are up to date")
	case "down":
		if err := migrator.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("down migrations: %w", err)
		}
		slog.Info("migrations rolled back")
	}

	return nil
}

func migrationCommand(args []string) (string, error) {
	if len(args) == 0 {
		return defaultCommand, nil
	}

	if len(args) > 1 {
		return "", fmt.Errorf("expected at most one command: up or down")
	}

	switch args[0] {
	case "up", "down":
		return args[0], nil
	default:
		return "", fmt.Errorf("unknown migration command %q", args[0])
	}
}

func migrationSourceURL(dir string) string {
	return "file://" + filepath.ToSlash(dir)
}
