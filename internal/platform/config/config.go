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
	Catalog   CatalogConfig
}

type CatalogConfig struct {
	DefaultPageSize int
	MaxPageSize     int
}

func Load() (Config, error) {
	env := getEnv("APP_ENV", "local")
	logConfig, err := newLogConfig(env)
	if err != nil {
		return Config{}, err
	}

	dbConfig, err := newDBConfig(env)
	if err != nil {
		return Config{}, err
	}

	catalogConfig, err := newCatalogConfig()
	if err != nil {
		return Config{}, err
	}

	return Config{
		AppEnv:    env,
		HTTPPort:  getEnv("HTTP_PORT", "8080"),
		LogConfig: logConfig,
		DBConfig:  dbConfig,
		Catalog:   catalogConfig,
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

const (
	catalogPaginationDefaultValue = 20
	catalogPaginationMaxValue     = 100
)

func newCatalogConfig() (CatalogConfig, error) {
	var cfg CatalogConfig

	size, err := getEnvInt("CATALOG_DEFAULT_PAGE_SIZE", catalogPaginationDefaultValue)
	if err != nil {
		return CatalogConfig{}, fmt.Errorf("convert catalog pagination default value: %w", err)
	}

	cfg.DefaultPageSize = size

	size, err = getEnvInt("CATALOG_MAX_PAGE_SIZE", catalogPaginationMaxValue)
	if err != nil {
		return CatalogConfig{}, fmt.Errorf("convert catalog pagination max value: %w", err)
	}

	cfg.MaxPageSize = size

	if cfg.DefaultPageSize <= 0 {
		return CatalogConfig{}, fmt.Errorf("pagination default page size must be greater than zero")
	}

	if cfg.MaxPageSize <= 0 {
		return CatalogConfig{}, fmt.Errorf("pagination max page size must be greater than zero")
	}

	if cfg.DefaultPageSize > cfg.MaxPageSize {
		return CatalogConfig{}, fmt.Errorf(
			"pagination default page size %d exceeds max page size %d",
			cfg.DefaultPageSize,
			cfg.MaxPageSize,
		)
	}

	return cfg, nil
}
