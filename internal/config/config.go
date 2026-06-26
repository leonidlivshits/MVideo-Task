package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPAddr string
	Database DatabaseConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
	SSLMode  string
}

func Load() (Config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("load .env: %w", err)
	}

	cfg := Config{
		HTTPAddr: getEnv("HTTP_ADDR", ":8080"),
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			Name:     getEnv("DB_NAME", "price_service"),
			User:     getEnv("DB_USER", "price_user"),
			Password: os.Getenv("DB_PASSWORD"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
	}

	if cfg.Database.Password == "" {
		return Config{}, fmt.Errorf("DB_PASSWORD is required")
	}

	return cfg, nil
}

func (c DatabaseConfig) ConnString() string {
	return c.connString("postgres")
}

func (c DatabaseConfig) MigrationConnString() string {
	return c.connString("pgx5")
}

func (c DatabaseConfig) connString(scheme string) string {
	host := c.Host
	if c.Port != "" {
		host = net.JoinHostPort(c.Host, c.Port)
	}

	connectionURL := url.URL{
		Scheme: scheme,
		User:   url.UserPassword(c.User, c.Password),
		Host:   host,
		Path:   c.Name,
	}

	query := connectionURL.Query()
	query.Set("sslmode", c.SSLMode)
	connectionURL.RawQuery = query.Encode()

	return connectionURL.String()
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
