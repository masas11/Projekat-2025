package config

import "os"

type Config struct {
	Port                    string
	MongoDBURI              string
	MongoDBDatabase         string
	JWTSecret               string
	SubscriptionsServiceURL string
	RecommendationServiceURL string
	RatingsServiceURL       string
	AnalyticsServiceURL     string
	HDFSNamenodeURL         string
	RedisURL                string
	SagaServiceURL          string
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

	recommendationServiceURL := os.Getenv("RECOMMENDATION_SERVICE_URL")
	if recommendationServiceURL == "" {
		recommendationServiceURL = "http://recommendation-service:8006"
	}

	ratingsServiceURL := os.Getenv("RATINGS_SERVICE_URL")
	if ratingsServiceURL == "" {
		ratingsServiceURL = "http://ratings-service:8003"
	}

	analyticsServiceURL := os.Getenv("ANALYTICS_SERVICE_URL")
	if analyticsServiceURL == "" {
		analyticsServiceURL = "http://analytics-service:8007"
	}

	hdfsNamenodeURL := os.Getenv("HDFS_NAMENODE_URL")
	if hdfsNamenodeURL == "" {
		hdfsNamenodeURL = "http://hdfs-namenode:9870"
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis:6379"
	}

	sagaServiceURL := os.Getenv("SAGA_SERVICE_URL")
	if sagaServiceURL == "" {
		sagaServiceURL = "http://saga-service:8008"
	}

	return &Config{
		Port:                     port,
		MongoDBURI:               mongoURI,
		MongoDBDatabase:          mongoDB,
		JWTSecret:                jwtSecret,
		SubscriptionsServiceURL:  subscriptionsServiceURL,
		RecommendationServiceURL: recommendationServiceURL,
		RatingsServiceURL:        ratingsServiceURL,
		AnalyticsServiceURL:      analyticsServiceURL,
		HDFSNamenodeURL:          hdfsNamenodeURL,
		RedisURL:                 redisURL,
		SagaServiceURL:           sagaServiceURL,
	}
}
