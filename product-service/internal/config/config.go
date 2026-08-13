package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port          string
	DatabaseURL   string
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	RedisCacheTTL time.Duration
	JWTSecret     string
	KafkaBrokers  []string
	KafkaTopic    string
	Environment   string
}

func Load() (Config, error) {
	port := getEnv("PORT", "8080")

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	redisAddr := getEnv("REDIS_URL", "localhost:6379")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDB := getEnvInt("REDIS_DB", 0)

	redisCacheTTL := time.Minute * 5
	if ttl := os.Getenv("REDIS_CACHE_TTL"); ttl != "" {
		if parsed, err := time.ParseDuration(ttl); err == nil {
			redisCacheTTL = parsed
		}
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET is required")
	}

	kafkaBrokers := []string{"localhost:9092"}
	if brokers := os.Getenv("KAFKA_BROKERS"); brokers != "" {
		kafkaBrokers = splitCommaSeparated(brokers)
	}

	kafkaTopic := getEnv("KAFKA_PRODUCT_TOPIC", "product.events")

	environment := getEnv("ENVIRONMENT", "development")

	return Config{
		Port:          port,
		DatabaseURL:   databaseURL,
		RedisAddr:     redisAddr,
		RedisPassword: redisPassword,
		RedisDB:       redisDB,
		RedisCacheTTL: redisCacheTTL,
		JWTSecret:     jwtSecret,
		KafkaBrokers:  kafkaBrokers,
		KafkaTopic:    kafkaTopic,
		Environment:   environment,
	}, nil
}

func getEnv(key string, defaultValue string) string {
	value := os.Getenv(key)

	if value == "" {
		return defaultValue
	}

	return value
}

func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)

	if value == "" {
		return defaultValue
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return parsed
}

func splitCommaSeparated(value string) []string {
	parts := make([]string, 0)

	current := ""

	for _, char := range value {
		if char == ',' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}

	if current != "" {
		parts = append(parts, current)
	}

	return parts
}
