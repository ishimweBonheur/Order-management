package config

import (
	"errors"
	"os"
	"strings"
)

type Config struct {
	Port, DatabaseURL, JWTSecret, KafkaTopic string
	KafkaBrokers                             []string
}

func Load() (Config, error) {
	c := Config{Port: env("HTTP_PORT", "8083"), DatabaseURL: os.Getenv("DATABASE_URL"), JWTSecret: os.Getenv("JWT_SECRET"), KafkaTopic: env("KAFKA_ORDER_TOPIC", "order.created"), KafkaBrokers: strings.Split(env("KAFKA_BROKERS", "localhost:9092"), ",")}
	if c.DatabaseURL == "" || len(c.JWTSecret) < 32 {
		return Config{}, errors.New("DATABASE_URL and JWT_SECRET are required")
	}
	return c, nil
}
func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
