package config

import "os"

type Config struct {
	Port            string
	MongoDBURI      string
	MongoDBDatabase string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8005"
	}

	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	mongoDB := os.Getenv("MONGODB_DATABASE")
	if mongoDB == "" {
		mongoDB = "notifications_db"
	}

	return &Config{
		Port:            port,
		MongoDBURI:      mongoURI,
		MongoDBDatabase: mongoDB,
	}
}
