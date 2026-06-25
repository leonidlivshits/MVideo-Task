package config

import (
	"net"
	"net/url"
	"os"
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
	return Config{
		HTTPAddr: getEnv("HTTP_ADDR", ":8080"),
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			Name:     getEnv("DB_NAME", "price_service"),
			User:     getEnv("DB_USER", "price_user"),
			Password: os.Getenv("DB_PASSWORD"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
	}, nil
}

func (c DatabaseConfig) ConnString() string {
	host := c.Host
	if c.Port != "" {
		host = net.JoinHostPort(c.Host, c.Port)
	}

	connectionURL := url.URL{
		Scheme: "postgres",
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
