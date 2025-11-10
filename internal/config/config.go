package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/hakaton/meeting-bot/internal/storage"
	"github.com/joho/godotenv"
)

type Config struct {
	BotToken       string
	DatabaseURL    string
	ServerPort     string
	MaxAPIBaseURL  string
	VotingDuration int
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Printf("Note: .env file not found, using environment variables only")
	}

	return &Config{
		BotToken:       mustGetEnv("BOT_TOKEN"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://localhost/meetingbot?sslmode=disable"),
		ServerPort:     getEnv("SERVER_PORT", "8080"),
		MaxAPIBaseURL:  getEnv("MAX_API_BASE_URL", "https://api.max.ru"),
		VotingDuration: getEnvInt("VOTING_DURATION", 120),
	}
}

func mustGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Required environment variable %s is not set", key)
	}
	return value
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		log.Printf("Warning: Invalid integer value for %s, using default: %d", key, defaultValue)
	}
	return defaultValue
}

func LoadFromPath(path string) *Config {
	if err := godotenv.Load(path); err != nil {
		log.Printf("Warning: .env file not found at %s, using environment variables", path)
	}
	return Load()
}

func LoadDatabaseConfig() storage.Config {
	host := "localhost"
	port := 5432
	user := "postgres"
	password := "postgres"
	dbname := "meetingbot"
	sslmode := "disable"

	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		if parsed := parseDatabaseURL(dbURL); parsed != nil {
			if parsed.Host != "" {
				host = parsed.Host
			}
			if parsed.Port != 0 {
				port = parsed.Port
			}
			if parsed.User != "" {
				user = parsed.User
			}
			if parsed.Password != "" {
				password = parsed.Password
			}
			if parsed.DBName != "" {
				dbname = parsed.DBName
			}
			if parsed.SSLMode != "" {
				sslmode = parsed.SSLMode
			}
		}
	}

	return storage.Config{
		Host:            getEnv("DB_HOST", host),
		Port:            getEnvInt("DB_PORT", port),
		User:            getEnv("DB_USER", user),
		Password:        getEnv("DB_PASSWORD", password),
		DBName:          getEnv("DB_NAME", dbname),
		SSLMode:         getEnv("DB_SSLMODE", sslmode),
		MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
		ConnMaxLifetime: getEnvDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		ConnMaxIdleTime: getEnvDuration("DB_CONN_MAX_IDLE_TIME", 10*time.Minute),
	}
}

func parseDatabaseURL(dbURL string) *storage.Config {
	if len(dbURL) < 11 || dbURL[:11] != "postgres://" {
		return nil
	}
	dbURL = dbURL[11:]

	cfg := &storage.Config{}

	atIndex := -1
	for i := 0; i < len(dbURL); i++ {
		if dbURL[i] == '@' {
			atIndex = i
			break
		}
	}

	if atIndex > 0 {
		userPass := dbURL[:atIndex]
		colonIndex := -1
		for i := 0; i < len(userPass); i++ {
			if userPass[i] == ':' {
				colonIndex = i
				break
			}
		}
		if colonIndex > 0 {
			cfg.User = userPass[:colonIndex]
			cfg.Password = userPass[colonIndex+1:]
		} else {
			cfg.User = userPass
		}
		dbURL = dbURL[atIndex+1:]
	}

	slashIndex := -1
	for i := 0; i < len(dbURL); i++ {
		if dbURL[i] == '/' {
			slashIndex = i
			break
		}
	}

	if slashIndex > 0 {
		hostPort := dbURL[:slashIndex]
		colonIndex := -1
		for i := 0; i < len(hostPort); i++ {
			if hostPort[i] == ':' {
				colonIndex = i
				break
			}
		}
		if colonIndex > 0 {
			cfg.Host = hostPort[:colonIndex]
			if p, err := strconv.Atoi(hostPort[colonIndex+1:]); err == nil {
				cfg.Port = p
			}
		} else {
			cfg.Host = hostPort
		}
		dbURL = dbURL[slashIndex+1:]
	}

	questionIndex := -1
	for i := 0; i < len(dbURL); i++ {
		if dbURL[i] == '?' {
			questionIndex = i
			break
		}
	}

	if questionIndex > 0 {
		cfg.DBName = dbURL[:questionIndex]
		params := dbURL[questionIndex+1:]
		if len(params) > 8 && params[:8] == "sslmode=" {
			cfg.SSLMode = params[8:]
		}
	} else {
		cfg.DBName = dbURL
	}

	return cfg
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
