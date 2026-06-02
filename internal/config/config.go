package config

import "os"

type Config struct {
	AppName string
	AppEnv  string
	AppPort string
}

func Load() Config {
	return Config{
		AppName: getEnv("APP_NAME", "DuwitTrackerApp"),
		AppEnv:  getEnv("APP_ENV", "development"),
		AppPort: getEnv("APP_PORT", "8080"),
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}