package config

import "os"

type Config struct {
	Port            string
	JWTSecret       string
	MongoDBURI      string
	MongoDBDatabase string
	BaseURL         string
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

	return &Config{
		Port:            port,
		JWTSecret:       jwtSecret,
		MongoDBURI:      mongoURI,
		MongoDBDatabase: mongoDB,
		BaseURL:         baseURL,
	}
}
