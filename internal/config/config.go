package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                  string
	AnthropicAPIKey       string
	GoogleCredentialsJSON string
	GoogleCalendarID      string
	AllowedOrigins        string
	DashboardKey          string
	DataDir               string
	// StoreURL selects the persistence backend.
	// Leave empty to use the embedded bolt DB (default).
	// Set to "redis://<host>:<port>" or "rediss://..." to use Redis.
	StoreURL        string
	WaniKaniToken   string
}

func Load() (*Config, error) {
	// Load .env if present (ignored in production where env vars are set directly)
	_ = godotenv.Load()

	cfg := &Config{
		Port:                  getEnv("PORT", "8080"),
		AnthropicAPIKey:       os.Getenv("ANTHROPIC_API_KEY"),
		GoogleCredentialsJSON: os.Getenv("GOOGLE_CREDENTIALS_JSON"),
		GoogleCalendarID:      os.Getenv("GOOGLE_CALENDAR_ID"),
		AllowedOrigins:        getEnv("ALLOWED_ORIGINS", "http://localhost:5173"),
		DashboardKey:          os.Getenv("DASHBOARD_KEY"),
		DataDir:               getEnv("DATA_DIR", "./data"),
		StoreURL:              os.Getenv("STORE_URL"),
		WaniKaniToken:         os.Getenv("WANIKANI_API_TOKEN"),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.AnthropicAPIKey == "" {
		return fmt.Errorf("ANTHROPIC_API_KEY is required")
	}
	return nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

