package config

import (
	"os"
	"strings"
)

type Config struct {
	Port, RedisAddr, KafkaTopic, KafkaGroup string
	KafkaBrokers                            []string
}

func Load() Config {
	return Config{Port: env("HTTP_PORT", "8084"), RedisAddr: env("REDIS_ADDR", "redis:6379"), KafkaTopic: env("KAFKA_ORDER_TOPIC", "order.created"), KafkaGroup: env("KAFKA_NOTIFICATION_GROUP", "notification-service"), KafkaBrokers: strings.Split(env("KAFKA_BROKERS", "localhost:9092"), ",")}
}
func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
