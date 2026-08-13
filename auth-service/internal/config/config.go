package config

import (
	"errors"
	"os"
	"time"
)

type Config struct {
	Port, DatabaseURL, JWTSecret string
	TokenTTL                     time.Duration
}

func Load() (Config, error) {
	c := Config{Port: env("HTTP_PORT", "8081"), DatabaseURL: os.Getenv("DATABASE_URL"), JWTSecret: os.Getenv("JWT_SECRET"), TokenTTL: 24 * time.Hour}
	if c.DatabaseURL == "" || len(c.JWTSecret) < 32 {
		return Config{}, errors.New("DATABASE_URL and JWT_SECRET (at least 32 characters) are required")
	}
	return c, nil
}
func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
