package config

import "os"

type Config struct {
	Port              string
	ContentServiceURL string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8003"
	}

	contentURL := os.Getenv("CONTENT_SERVICE_URL")
	if contentURL == "" {
		contentURL = "http://localhost:8081"
	}

	return &Config{
		Port:              port,
		ContentServiceURL: contentURL,
	}
}
