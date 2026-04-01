package config

import (
	"os"
	"strings"
)

const (
	defaultPort           = "8080"
	defaultWelcomeMessage = "Welcome to the custom app"
	defaultLogLevel       = "info"
)

type Config struct {
	Port           string
	WelcomeMessage string
	LogLevel       string
}

func Load() Config {
	cfg := Config{
		Port:           defaultPort,
		WelcomeMessage: defaultWelcomeMessage,
		LogLevel:       defaultLogLevel,
	}

	if value := strings.TrimSpace(os.Getenv("APP_PORT")); value != "" {
		cfg.Port = value
	}

	if value := strings.TrimSpace(os.Getenv("APP_WELCOME_MESSAGE")); value != "" {
		cfg.WelcomeMessage = value
	}

	if value := strings.TrimSpace(os.Getenv("APP_LOG_LEVEL")); value != "" {
		cfg.LogLevel = value
	}

	return cfg
}
