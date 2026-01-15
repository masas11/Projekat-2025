package config

import "os"

type Config struct {
	Port                string
	UsersServiceURL     string
	ContentServiceURL   string
	NotificationsServiceURL string
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

	notificationsURL := os.Getenv("NOTIFICATIONS_SERVICE_URL")
	if notificationsURL == "" {
		notificationsURL = "http://localhost:8005"
	}

	return &Config{
		Port:                port,
		UsersServiceURL:     usersURL,
		ContentServiceURL:   contentURL,
		NotificationsServiceURL: notificationsURL,
	}
}
