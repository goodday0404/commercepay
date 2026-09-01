package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	AppEnv    string
	HTTPPort  string
	LogConfig LogConfig
	DBConfig  DBConfig
}

func Load() (Config, error) {
	env := getEnv("APP_ENV", "local")
	logConfig, err := newLogConfig(env)
	if err != nil {
		return Config{}, nil
	}

	dbConfig, err := newDBConfig(env)
	if err != nil {
		return Config{}, nil
	}

	return Config{
		AppEnv:    env,
		HTTPPort:  getEnv("HTTP_PORT", "8080"),
		LogConfig: logConfig,
		DBConfig:  dbConfig,
	}, nil
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)

	if value == "" {
		return fallback
	}

	return value
}

func getEnvInt(key string, fallback int) (int, error) {
	value, exist := os.LookupEnv(key)
	if !exist || value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid %s, %q: %w ", key, value, err)
	}

	return parsed, nil
}

func getEnvBool(key string, fallback bool) (bool, error) {
	value, exist := os.LookupEnv(key)

	if !exist || value == "" {
		return fallback, nil
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("invalid %s, %q: %w", key, value, err)
	}

	return parsed, nil
}

func requireEnv(key string) (string, error) {
	value := os.Getenv(key)

	if value == "" {
		return "", fmt.Errorf("required environment variable %s is not set", key)
	}

	return value, nil
}
