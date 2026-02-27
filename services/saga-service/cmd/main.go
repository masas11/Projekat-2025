package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"saga-service/config"
	"saga-service/internal/orchestrator"
	"saga-service/internal/store"
)

func main() {
	cfg := config.Load()

	// Initialize MongoDB connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoDBURI))
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer client.Disconnect(ctx)

	db := client.Database(cfg.MongoDBDatabase)
	log.Println("Connected to MongoDB")

	// Initialize store and orchestrator
	sagaStore := store.NewSagaStore(db)
	songDeletionSaga := orchestrator.NewSongDeletionSaga(sagaStore, cfg)

	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("saga-service is running"))
	})

	// Start saga transaction for song deletion
	mux.HandleFunc("/sagas/delete-song", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			SongID string `json:"songId"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}

		if req.SongID == "" {
			http.Error(w, "songId is required", http.StatusBadRequest)
			return
		}

		// Execute saga
		sagaCtx, sagaCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer sagaCancel()

		saga, err := songDeletionSaga.Execute(sagaCtx, req.SongID)
		if err != nil {
			log.Printf("Saga execution failed: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": err.Error(),
				"saga":  saga,
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(saga)
	})

	// Get saga transaction status
	mux.HandleFunc("/sagas/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		sagaID := r.URL.Path[len("/sagas/"):]
		if sagaID == "" {
			http.Error(w, "saga ID is required", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		saga, err := sagaStore.GetTransaction(ctx, sagaID)
		if err != nil {
			http.Error(w, "saga not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(saga)
	})

	port := cfg.Port
	if port == "" {
		port = "8008"
	}

	log.Printf("Saga service running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
