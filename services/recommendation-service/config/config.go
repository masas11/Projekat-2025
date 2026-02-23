package config

import "os"

type Config struct {
	Port                    string
	Neo4jURI                string
	Neo4jUser               string
	Neo4jPassword            string
	RatingsServiceURL       string
	SubscriptionsServiceURL string
	ContentServiceURL       string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8006"
	}

	return &Config{
		Port:                    port,
		Neo4jURI:                getEnv("NEO4J_URI", "bolt://localhost:7687"),
		Neo4jUser:               getEnv("NEO4J_USER", "neo4j"),
		Neo4jPassword:            getEnv("NEO4J_PASSWORD", "password"),
		RatingsServiceURL:       getEnv("RATINGS_SERVICE_URL", "http://ratings-service:8003"),
		SubscriptionsServiceURL: getEnv("SUBSCRIPTIONS_SERVICE_URL", "http://subscriptions-service:8004"),
		ContentServiceURL:       getEnv("CONTENT_SERVICE_URL", "http://content-service:8002"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
