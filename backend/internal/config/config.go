package config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoURI            string
	JWTSecret           string
	JWTExpiry           time.Duration
	Port                string
	Environment         string
	CORSOrigin          string
	CloudflareAppID     string
	CloudflareAppSecret string
}

func Load() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	jwtExpiry, err := time.ParseDuration(getEnv("JWT_EXPIRY", "24h"))
	if err != nil {
		jwtExpiry = 24 * time.Hour
	}

	return &Config{
		MongoURI:            getEnv("MONGODB_URI", "mongodb://localhost:27017/meet-clone"),
		JWTSecret:           getEnv("JWT_SECRET", "default-secret-key-change-in-production"),
		JWTExpiry:           jwtExpiry,
		Port:                getEnv("PORT", "8080"),
		Environment:         getEnv("ENV", "development"),
		CORSOrigin:          getEnv("CORS_ORIGIN", "http://localhost:3000"),
		CloudflareAppID:     getEnv("CLOUDFLARE_APP_ID", ""),
		CloudflareAppSecret: getEnv("CLOUDFLARE_APP_SECRET", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
