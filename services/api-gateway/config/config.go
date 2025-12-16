package config

import "os"

type Config struct {
	Port           string
	UsersServiceURL string
	ContentServiceURL string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	usersURL := os.Getenv("USERS_SERVICE_URL")
	if usersURL == "" {
		usersURL = "http://localhost:8001"
	}

	contentURL := os.Getenv("CONTENT_SERVICE_URL")
	if contentURL == "" {
		contentURL = "http://localhost:8002"
	}

	return &Config{
		Port: port,
		UsersServiceURL: usersURL,
		ContentServiceURL: contentURL,
	}
}
