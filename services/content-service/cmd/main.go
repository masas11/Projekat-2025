package main

import (
	"log"
	"net/http"
	"strings"

	"content-service/config"
	"content-service/internal/handler"
	"content-service/internal/middleware"
	"content-service/internal/store"
)

func main() {
	cfg := config.Load()

	// Initialize MongoDB connection
	dbStore, err := store.NewMongoDBStore(cfg.MongoDBURI, cfg.MongoDBDatabase)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer dbStore.Close()
	log.Println("Connected to MongoDB")

	// Initialize repositories
	artistRepo := store.NewArtistRepository(dbStore.Database)

	// Initialize handlers
	artistHandler := handler.NewArtistHandler(artistRepo)

	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("content-service is running"))
	})

	// Dummy endpoint for 3.3 (song existence check)
	mux.HandleFunc("/songs/exists", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("true"))
	})

	// Artist routes
	// GET /artists - get all artists (public)
	mux.HandleFunc("/artists", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			artistHandler.GetAllArtists(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// POST /artists - create artist (admin only, requires JWT)
	mux.HandleFunc("/artists", middleware.JWTAuth(cfg)(middleware.AdminOnly(artistHandler.CreateArtist)))

	// GET /artists/{id} - get artist by ID (public)
	mux.HandleFunc("/artists/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/artists/")
		if path == "" {
			http.Error(w, "artist ID is required", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			artistHandler.GetArtist(w, r)
		case http.MethodPut:
			// PUT /artists/{id} - update artist (admin only, requires JWT)
			middleware.JWTAuth(cfg)(middleware.AdminOnly(artistHandler.UpdateArtist))(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Println("Content service running on port", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}
