package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port                   string
	JWTSecret              string
	MongoDBURI             string
	MongoDBDatabase        string
	BaseURL                string
	PasswordExpirationDays int // Number of days until password expires (default 60, can be overridden for testing)
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8001"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-secret-key-change-in-production" // Default secret, should be changed in production
	}

	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	mongoDB := os.Getenv("MONGODB_DATABASE")
	if mongoDB == "" {
		mongoDB = "users_db"
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8081" // Default to API Gateway URL
	}

	// Password expiration period - for testing, can be set to shorter periods (e.g., 1 hour, 1 day)
	// Default: 60 days as per specification
	// For demo: set PASSWORD_EXPIRATION_DAYS=1 to simulate 1 day expiration
	passwordExpirationDays := 60
	if days := os.Getenv("PASSWORD_EXPIRATION_DAYS"); days != "" {
		if parsedDays, err := strconv.Atoi(days); err == nil && parsedDays > 0 {
			passwordExpirationDays = parsedDays
		}
	}

	return &Config{
		Port:                   port,
		JWTSecret:              jwtSecret,
		MongoDBURI:             mongoURI,
		MongoDBDatabase:        mongoDB,
		BaseURL:                baseURL,
		PasswordExpirationDays: passwordExpirationDays,
	}
}
