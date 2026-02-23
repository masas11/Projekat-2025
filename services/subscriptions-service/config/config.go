package config

import "os"

type Config struct {
	Port                    string
	ContentServiceURL       string
	MongoDBURI              string
	MongoDBDatabase         string
	NotificationsServiceURL string
	RecommendationServiceURL string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8004"
	}

	contentServiceURL := os.Getenv("CONTENT_SERVICE_URL")
	if contentServiceURL == "" {
		contentServiceURL = "http://localhost:8081"
	}

	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	mongoDB := os.Getenv("MONGODB_DATABASE")
	if mongoDB == "" {
		mongoDB = "subscriptions_db"
	}

	notificationsServiceURL := os.Getenv("NOTIFICATIONS_SERVICE_URL")
	if notificationsServiceURL == "" {
		notificationsServiceURL = "http://notifications-service:8005"
	}

	recommendationServiceURL := os.Getenv("RECOMMENDATION_SERVICE_URL")
	if recommendationServiceURL == "" {
		recommendationServiceURL = "http://recommendation-service:8006"
	}

	return &Config{
		Port:                     port,
		ContentServiceURL:        contentServiceURL,
		MongoDBURI:               mongoURI,
		MongoDBDatabase:          mongoDB,
		NotificationsServiceURL:  notificationsServiceURL,
		RecommendationServiceURL: recommendationServiceURL,
	}
}
