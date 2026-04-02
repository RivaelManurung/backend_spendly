package config

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// AppConfig is the main configuration structure for the application.
type AppConfig struct {
	AppEnv       string `envconfig:"APP_ENV" default:"development"`
	GeminiApiKey string `envconfig:"GEMINI_API_KEY" required:"true"`
	HTTPPort     string `envconfig:"HTTP_PORT" default:"8080"`
	
	// Auth
	JWTSecret         string `envconfig:"JWT_SECRET" required:"true"`
	GoogleClientID     string `envconfig:"GOOGLE_CLIENT_ID" required:"true"`
	GoogleClientSecret string `envconfig:"GOOGLE_CLIENT_SECRET" required:"true"`
	
	Database DBConfig
}

// DBConfig stores connection parameters for the database.
type DBConfig struct {
	DSN             string `envconfig:"DB_DSN" required:"true"`
	MaxOpenConns    int    `envconfig:"DB_MAX_OPEN_CONNS" default:"25"`
	MaxIdleConns    int    `envconfig:"DB_MAX_IDLE_CONNS" default:"25"`
	ConnMaxLifetime int    `envconfig:"DB_CONN_MAX_LIFETIME_MINUTES" default:"15"`
	ConnMaxIdleTime int    `envconfig:"DB_CONN_MAX_IDLE_TIME_MINUTES" default:"5"`
}

// LoadConfig loads the configuration from .env file (if exists) and environment variables.
func LoadConfig() (*AppConfig, error) {
	_ = godotenv.Load()

	var cfg AppConfig
	if err := envconfig.Process("", &cfg); err != nil {
		log.Printf("Failed to process env: %v\n", err)
		return nil, err
	}

	return &cfg, nil
}
