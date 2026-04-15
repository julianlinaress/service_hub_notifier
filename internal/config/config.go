package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	defaultPort              = "8081"
	defaultDeliveryTimeoutMS = 5000
	defaultShutdownTimeoutMS = 10000
	defaultTelegramAPIBase   = "https://api.telegram.org"
)

// Config stores runtime configuration for the notifier service.
type Config struct {
	Port               string
	DeliveryTimeout    time.Duration
	ShutdownTimeout    time.Duration
	TelegramAPIBaseURL string
}

// EnvGetter reads an environment variable value by key.
type EnvGetter func(key string) string

// LoadFromEnv builds service configuration from environment values.
func LoadFromEnv(getEnv EnvGetter) Config {
	if getEnv == nil {
		getEnv = func(string) string { return "" }
	}

	return Config{
		Port:               stringEnvOrDefault(getEnv, "PORT", defaultPort),
		DeliveryTimeout:    durationFromEnvMS(getEnv, "DELIVERY_TIMEOUT_MS", defaultDeliveryTimeoutMS),
		ShutdownTimeout:    durationFromEnvMS(getEnv, "SHUTDOWN_TIMEOUT_MS", defaultShutdownTimeoutMS),
		TelegramAPIBaseURL: stringEnvOrDefault(getEnv, "TELEGRAM_API_BASE_URL", defaultTelegramAPIBase),
	}
}

func stringEnvOrDefault(getEnv EnvGetter, key string, fallback string) string {
	value := strings.TrimSpace(getEnv(key))
	if value == "" {
		return fallback
	}

	return value
}

func durationFromEnvMS(getEnv EnvGetter, key string, fallbackMS int) time.Duration {
	raw := stringEnvOrDefault(getEnv, key, fmt.Sprintf("%d", fallbackMS))
	parsed, err := strconv.Atoi(raw)
	if err != nil || parsed <= 0 {
		parsed = fallbackMS
	}

	return time.Duration(parsed) * time.Millisecond
}
