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
	c := Config{Port: os.Getenv("HTTP_PORT"), DatabaseURL: os.Getenv("DATABASE_URL"), JWTSecret: os.Getenv("JWT_SECRET"), TokenTTL: 24 * time.Hour}
	if c.Port == "" || c.DatabaseURL == "" || len(c.JWTSecret) < 32 {
		return Config{}, errors.New("HTTP_PORT, DATABASE_URL, and JWT_SECRET (at least 32 characters) are required")
	}
	return c, nil
}
