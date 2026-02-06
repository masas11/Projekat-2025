package config

import "os"

type Config struct {
	Port                    string
	JWTSecret               string
	UsersServiceURL         string
	ContentServiceURL       string
	NotificationsServiceURL string
	SubscriptionsServiceURL string
	RatingsServiceURL       string
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

	subscriptionsURL := os.Getenv("SUBSCRIPTIONS_SERVICE_URL")
	if subscriptionsURL == "" {
		subscriptionsURL = "http://localhost:8004"
	}

	ratingsURL := os.Getenv("RATINGS_SERVICE_URL")
	if ratingsURL == "" {
		ratingsURL = "http://localhost:8003"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-secret-key-change-in-production" // Default, should match users-service
	}

	return &Config{
		Port:                    port,
		JWTSecret:               jwtSecret,
		UsersServiceURL:         usersURL,
		ContentServiceURL:       contentURL,
		NotificationsServiceURL: notificationsURL,
		SubscriptionsServiceURL: subscriptionsURL,
		RatingsServiceURL:       ratingsURL,
	}
}
