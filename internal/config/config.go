package config

import "os"

type Config struct {
	Port       string
	APIKey     string
	MaxPDFSize int64
}

func Load() *Config {
	return &Config{
		Port:       getEnv("PORT", "8080"),
		APIKey:     getEnv("GEMINI_API_KEY", ""),
		MaxPDFSize: 10 * 1024 * 1024, // 10MB
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
