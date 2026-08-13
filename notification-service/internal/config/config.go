package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	HTTPPort, RedisAddr, KafkaTopic, KafkaGroup string
	KafkaBrokers                                []string
	SMTPHost, SMTPUser, SMTPPassword, EmailFrom string
	SMTPPort                                    int
	SMTPSecure                                  bool
}

func Load() (Config, error) {
	required := func(name string) (string, error) {
		value := strings.TrimSpace(os.Getenv(name))
		if value == "" {
			return "", fmt.Errorf("%s is required", name)
		}
		return value, nil
	}
	var c Config
	var err error
	if c.HTTPPort, err = required("HTTP_PORT"); err != nil {
		return c, err
	}
	if c.RedisAddr, err = required("REDIS_ADDR"); err != nil {
		return c, err
	}
	brokers, err := required("KAFKA_BROKERS")
	if err != nil {
		return c, err
	}
	for _, broker := range strings.Split(brokers, ",") {
		if broker = strings.TrimSpace(broker); broker != "" {
			c.KafkaBrokers = append(c.KafkaBrokers, broker)
		}
	}
	if len(c.KafkaBrokers) == 0 {
		return c, fmt.Errorf("KAFKA_BROKERS must contain at least one broker")
	}
	if c.KafkaTopic, err = required("KAFKA_ORDER_TOPIC"); err != nil {
		return c, err
	}
	if c.KafkaGroup, err = required("KAFKA_NOTIFICATION_GROUP"); err != nil {
		return c, err
	}
	if c.SMTPHost, err = required("SMTP_HOST"); err != nil {
		return c, err
	}
	if c.SMTPUser, err = required("SMTP_USER"); err != nil {
		return c, err
	}
	if c.SMTPPassword, err = required("SMTP_PASSWORD"); err != nil {
		return c, err
	}
	if c.EmailFrom, err = required("EMAIL_FROM"); err != nil {
		return c, err
	}
	port := os.Getenv("SMTP_PORT")
	if port == "" {
		port = "587"
	}
	if c.SMTPPort, err = strconv.Atoi(port); err != nil || c.SMTPPort < 1 || c.SMTPPort > 65535 {
		return c, fmt.Errorf("SMTP_PORT must be a valid port")
	}
	if secure := os.Getenv("SMTP_SECURE"); secure != "" {
		c.SMTPSecure, err = strconv.ParseBool(secure)
		if err != nil {
			return c, fmt.Errorf("SMTP_SECURE must be true or false")
		}
	}
	return c, nil
}
