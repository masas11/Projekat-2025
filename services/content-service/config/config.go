package config

import "os"

type Config struct {
	Port                 string
	MongoDBURI           string
	MongoDBDatabase      string
	JWTSecret            string
	SubscriptionsServiceURL string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8002"
	}

	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	mongoDB := os.Getenv("MONGODB_DATABASE")
	if mongoDB == "" {
		mongoDB = "music_streaming"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-secret-key-change-in-production" // Should match users-service secret
	}

	subscriptionsServiceURL := os.Getenv("SUBSCRIPTIONS_SERVICE_URL")
	if subscriptionsServiceURL == "" {
		subscriptionsServiceURL = "http://subscriptions-service:8004"
	}

	return &Config{
		Port:                 port,
		MongoDBURI:           mongoURI,
		MongoDBDatabase:      mongoDB,
		JWTSecret:            jwtSecret,
		SubscriptionsServiceURL: subscriptionsServiceURL,
	}
}
