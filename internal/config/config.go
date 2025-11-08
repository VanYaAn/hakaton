package config

import (
	"os"
)

type Config struct {
	BotToken       string
	DatabaseURL    string
	ServerPort     string
	MaxAPIBaseURL  string
	VotingDuration int // Duration in minutes before voting closes
}

func Load() *Config {
	return &Config{
		BotToken:       getEnv("BOT_TOKEN", "stub_token"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://localhost/meetingbot?sslmode=disable"),
		ServerPort:     getEnv("SERVER_PORT", "8080"),
		MaxAPIBaseURL:  getEnv("MAX_API_BASE_URL", "https://api.max.ru"),
		VotingDuration: getEnvInt("VOTING_DURATION", 120), // 2 hours default
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	// Simple stub implementation - in production, parse os.Getenv(key)
	_ = key // stub: will be used in production
	return defaultValue
}
