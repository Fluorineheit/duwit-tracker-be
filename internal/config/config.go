package config

import (
	"os"
	"strconv"
)

type Config struct {
	AppName string
	AppEnv  string
	AppPort string

	AppUserEmail string

	DatabaseURL string
	DBMaxConns  int32
	DBMinConns  int32
}

func Load() Config {
	return Config{
		AppName: getEnv("APP_NAME", "DuwitTrackerApp"),
		AppEnv:  getEnv("APP_ENV", "development"),
		AppPort: getEnv("APP_PORT", "8080"),

		AppUserEmail: getEnv("APP_USER_EMAIL", ""),

		DatabaseURL: getEnv("DATABASE_URL", ""),
		DBMaxConns:  int32(getEnvAsInt("DB_MAX_CONNS", 5)),
		DBMinConns:  int32(getEnvAsInt("DB_MIN_CONNS", 1)),
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func getEnvAsInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}
