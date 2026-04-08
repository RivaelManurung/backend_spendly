package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Env       string
	Port      string
	JwtSecret string
	ApiKey    string // Gemini API Key
	DBType    string // "sqlite" or "postgres"
	DBDSN     string
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load() // Loads .env if it exists

	return &Config{
		Env:       getEnv("APP_ENV", "development"),
		Port:      getEnv("PORT", "8080"),
		JwtSecret: getEnv("JWT_SECRET", "supersecretkey"),
		ApiKey:    getEnv("GEMINI_API_KEY", ""),
		DBType:    getEnv("DB_TYPE", "sqlite"),
		DBDSN:     getEnv("DB_DSN", "spendly.db"),
	}, nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
