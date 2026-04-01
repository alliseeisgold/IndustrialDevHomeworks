package config

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultPort           = "8080"
	defaultWelcomeMessage = "Welcome to the custom app"
	defaultLogLevel       = "info"
	configDirPath         = "/app/config"
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

	if value := readConfigFile("APP_PORT"); value != "" {
		cfg.Port = value
	}

	if value := readConfigFile("APP_WELCOME_MESSAGE"); value != "" {
		cfg.WelcomeMessage = value
	}

	if value := readConfigFile("APP_LOG_LEVEL"); value != "" {
		cfg.LogLevel = value
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

func readConfigFile(name string) string {
	content, err := os.ReadFile(filepath.Join(configDirPath, name))
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(content))
}
