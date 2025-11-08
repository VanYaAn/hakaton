package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	BotToken       string
	DatabaseURL    string
	ServerPort     string
	MaxAPIBaseURL  string
	VotingDuration int // Duration in minutes before voting closes
}

// Load loads configuration from .env file and environment variables
// .env file values are overridden by actual environment variables
func Load() *Config {
	// Загружаем .env файл (если существует)
	// Игнорируем ошибку, если файл не найден - используем переменные окружения
	if err := godotenv.Load(); err != nil {
		log.Printf("Note: .env file not found, using environment variables only")
	}

	return &Config{
		BotToken:       mustGetEnv("BOT_TOKEN"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://localhost/meetingbot?sslmode=disable"),
		ServerPort:     getEnv("SERVER_PORT", "8080"),
		MaxAPIBaseURL:  getEnv("MAX_API_BASE_URL", "https://api.max.ru"),
		VotingDuration: getEnvInt("VOTING_DURATION", 120), // 2 hours default
	}
}

// mustGetEnv returns environment variable or panics if not set
func mustGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Required environment variable %s is not set", key)
	}
	return value
}

// getEnv returns environment variable or default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt returns environment variable as integer or default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		log.Printf("Warning: Invalid integer value for %s, using default: %d", key, defaultValue)
	}
	return defaultValue
}

// LoadFromPath loads configuration from specific .env file path
func LoadFromPath(path string) *Config {
	if err := godotenv.Load(path); err != nil {
		log.Printf("Warning: .env file not found at %s, using environment variables", path)
	}
	return Load()
}
