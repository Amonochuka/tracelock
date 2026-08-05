package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	Port string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	JWTSecret     string
	DeviceAPIKey        string
	AllowedOrigin      string
	SessionTimeoutHours int
}

func Load() *Config {
	cfg := &Config{
		Port:          getEnv("PORT", "8080"),
		DBHost:        mustEnv("DB_HOST"),
		DBPort:        mustEnv("DB_PORT"),
		DBUser:        mustEnv("DB_USER"),
		DBPassword:    mustEnv("DB_PASSWORD"),
		DBName:        mustEnv("DB_NAME"),
		DBSSLMode:     getEnv("DB_SSLMODE", "disable"),
		JWTSecret:     mustEnv("JWT_SECRET"),
		DeviceAPIKey:        mustEnv("DEVICE_API_KEY"),
		AllowedOrigin:      getEnv("ALLOWED_ORIGIN", "*"),
		SessionTimeoutHours: getEnvInt("SESSION_TIMEOUT_HOURS", 12),
	}
	return cfg
}

func mustEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("%s environment variable required", key)
	}
	return val
}

func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}

func getEnvInt(key string, fallback int) int {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		log.Printf("invalid %s value %q, using default %d", key, val, fallback)
		return fallback
	}
	return n
}
