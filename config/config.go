package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	ServerPort  string
	JWTSecretKey string
}

func LoadConfig() *Config {
	// Try to load .env file if it exists (for local development)
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, assuming environment variables are set.")
	}

	cfg := &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		ServerPort:  os.Getenv("SERVER_PORT"),
		JWTSecretKey: os.Getenv("JWT_SECRET_KEY"), // <-- Muat ini
	}

	// Basic validation
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set.")
	}
	if cfg.ServerPort == "" {
		log.Fatal("SERVER_PORT environment variable is not set.")
	}
    if cfg.JWTSecretKey == "" { // <-- Validasi ini juga
        log.Fatal("JWT_SECRET_KEY environment variable is not set.")
    }

	return cfg
}
