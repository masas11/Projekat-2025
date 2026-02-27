package config

import "os"

type Config struct {
	Port                string
	MongoDBURI          string
	MongoDBDatabase     string
	ContentServiceURL   string
	RatingsServiceURL   string
	SubscriptionsServiceURL string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8007"
	}

	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://mongodb-analytics:27017"
	}

	mongoDB := os.Getenv("MONGODB_DATABASE")
	if mongoDB == "" {
		mongoDB = "analytics_db"
	}

	contentURL := os.Getenv("CONTENT_SERVICE_URL")
	if contentURL == "" {
		contentURL = "http://content-service:8002"
	}

	ratingsURL := os.Getenv("RATINGS_SERVICE_URL")
	if ratingsURL == "" {
		ratingsURL = "http://ratings-service:8003"
	}

	subscriptionsURL := os.Getenv("SUBSCRIPTIONS_SERVICE_URL")
	if subscriptionsURL == "" {
		subscriptionsURL = "http://subscriptions-service:8004"
	}

	return &Config{
		Port:                 port,
		MongoDBURI:           mongoURI,
		MongoDBDatabase:      mongoDB,
		ContentServiceURL:    contentURL,
		RatingsServiceURL:    ratingsURL,
		SubscriptionsServiceURL: subscriptionsURL,
	}
}
