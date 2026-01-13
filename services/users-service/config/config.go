package config

import "os"

type Config struct {
	Port      string
	JWTSecret string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8001"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-secret-key-change-in-production" // Default secret, should be changed in production
	}

	return &Config{
		Port:      port,
		JWTSecret: jwtSecret,
	}
}
