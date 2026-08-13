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
	c := Config{Port: os.Getenv("HTTP_PORT"), DatabaseURL: os.Getenv("DATABASE_URL"), JWTSecret: os.Getenv("JWT_SECRET"), KafkaTopic: os.Getenv("KAFKA_ORDER_TOPIC"), KafkaBrokers: strings.Split(os.Getenv("KAFKA_BROKERS"), ",")}
	if c.Port == "" || c.DatabaseURL == "" || len(c.JWTSecret) < 32 || c.KafkaTopic == "" || len(c.KafkaBrokers) == 0 || c.KafkaBrokers[0] == "" {
		return Config{}, errors.New("HTTP_PORT, DATABASE_URL, JWT_SECRET, KAFKA_ORDER_TOPIC, and KAFKA_BROKERS are required")
	}
	return c, nil
}
