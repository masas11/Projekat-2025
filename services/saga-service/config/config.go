package config

import "os"

type Config struct {
	Port                    string
	ContentServiceURL       string
	RatingsServiceURL       string
	RecommendationServiceURL string
	MongoDBURI              string
	MongoDBDatabase         string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8008"
	}

	contentServiceURL := os.Getenv("CONTENT_SERVICE_URL")
	if contentServiceURL == "" {
		contentServiceURL = "http://content-service:8002"
	}

	ratingsServiceURL := os.Getenv("RATINGS_SERVICE_URL")
	if ratingsServiceURL == "" {
		ratingsServiceURL = "http://ratings-service:8003"
	}

	recommendationServiceURL := os.Getenv("RECOMMENDATION_SERVICE_URL")
	if recommendationServiceURL == "" {
		recommendationServiceURL = "http://recommendation-service:8006"
	}

	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	mongoDB := os.Getenv("MONGODB_DATABASE")
	if mongoDB == "" {
		mongoDB = "saga_db"
	}

	return &Config{
		Port:                     port,
		ContentServiceURL:        contentServiceURL,
		RatingsServiceURL:        ratingsServiceURL,
		RecommendationServiceURL: recommendationServiceURL,
		MongoDBURI:               mongoURI,
		MongoDBDatabase:          mongoDB,
	}
}
