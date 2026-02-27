package config

import "os"

type Config struct {
	Port                   string
	ContentServiceURL      string
	RecommendationServiceURL string
	AnalyticsServiceURL    string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8003"
	}

	contentURL := os.Getenv("CONTENT_SERVICE_URL")
	if contentURL == "" {
		contentURL = "http://content-service:8002"
	}

	recommendationURL := os.Getenv("RECOMMENDATION_SERVICE_URL")
	if recommendationURL == "" {
		recommendationURL = "http://recommendation-service:8006"
	}

	analyticsURL := os.Getenv("ANALYTICS_SERVICE_URL")
	if analyticsURL == "" {
		analyticsURL = "http://analytics-service:8007"
	}

	return &Config{
		Port:                    port,
		ContentServiceURL:       contentURL,
		RecommendationServiceURL: recommendationURL,
		AnalyticsServiceURL:     analyticsURL,
	}
}
