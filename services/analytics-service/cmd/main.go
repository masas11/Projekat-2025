package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"analytics-service/config"
	"analytics-service/internal/handler"
	"analytics-service/internal/store"
	"shared/tracing"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	cfg := config.Load()

	// Initialize tracing
	cleanup, err := tracing.InitTracing("analytics-service")
	if err != nil {
		log.Printf("Warning: Failed to initialize tracing: %v", err)
	} else {
		defer cleanup()
		log.Println("Tracing initialized for analytics-service")
	}

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoDBURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(ctx)

	// Test the connection
	err = mongoClient.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	db := mongoClient.Database(cfg.MongoDBDatabase)
	activityStore := store.NewActivityStore(db)
	activityHandler := handler.NewActivityHandler(activityStore)

	log.Printf("Connected to MongoDB at %s, database: %s", cfg.MongoDBURI, cfg.MongoDBDatabase)

	// Setup routes
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("analytics-service is running"))
	})

	// Activity endpoints
	mux.HandleFunc("/activities", func(w http.ResponseWriter, r *http.Request) {
		// Add CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight request
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		switch r.Method {
		case http.MethodPost:
			activityHandler.LogActivity(w, r)
		case http.MethodGet:
			activityHandler.GetUserActivities(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Wrap with tracing middleware
	handler := tracing.HTTPMiddleware(mux)

	log.Println("Analytics service running on port", cfg.Port)

	// Support HTTPS if certificates are provided
	certFile := os.Getenv("TLS_CERT_FILE")
	keyFile := os.Getenv("TLS_KEY_FILE")
	if certFile != "" && keyFile != "" {
		log.Println("Starting HTTPS server on port", cfg.Port)
		log.Fatal(http.ListenAndServeTLS(":"+cfg.Port, certFile, keyFile, handler))
	} else {
		log.Println("Starting HTTP server on port", cfg.Port)
		log.Fatal(http.ListenAndServe(":"+cfg.Port, handler))
	}
}
