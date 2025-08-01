package config

import (
	"log"
	"os"
	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL     string
	CloudinaryURL   string
	JWTSecret       string
	Port           string
	GinMode        string
}

func Load() *Config {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	config := &Config{
		DatabaseURL:   getEnv("DATABASE_URL", ""),
		CloudinaryURL: getEnv("CLOUDINARY_URL", ""),
		JWTSecret:     getEnv("JWT_SECRET", "default-secret-key"),
		Port:         getEnv("PORT", "8080"),
		GinMode:      getEnv("GIN_MODE", "debug"),
	}

	if config.DatabaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	return config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}