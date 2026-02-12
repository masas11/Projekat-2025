package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"content-service/config"
	"content-service/internal/handler"
	"content-service/internal/logger"
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

	// Initialize logger
	logDir := os.Getenv("LOG_DIR")
	if logDir == "" {
		logDir = filepath.Join(".", "logs")
	}
	appLogger, err := logger.NewLogger(logDir)
	if err != nil {
		log.Printf("Warning: Failed to initialize logger: %v, using stdout", err)
		appLogger = logger.NewStdoutLogger()
	}
	defer appLogger.Close()

	// Initialize repositories
	artistRepo := store.NewArtistRepository(dbStore.Database)
	albumRepo := store.NewAlbumRepository(dbStore.Database)
	songRepo := store.NewSongRepository(dbStore.Database)

	// Initialize handlers
	artistHandler := handler.NewArtistHandler(artistRepo, cfg.SubscriptionsServiceURL, appLogger)
	albumHandler := handler.NewAlbumHandler(albumRepo, cfg.SubscriptionsServiceURL, appLogger)
	songHandler := handler.NewSongHandler(songRepo, albumRepo, cfg.SubscriptionsServiceURL, appLogger)

	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("content-service is running"))
	})

	// Static file server for music files
	mux.HandleFunc("/music/", func(w http.ResponseWriter, r *http.Request) {
		filePath := strings.TrimPrefix(r.URL.Path, "/music/")
		if filePath == "" {
			http.Error(w, "file path is required", http.StatusBadRequest)
			return
		}

		// For security, only allow specific file extensions
		ext := strings.ToLower(filepath.Ext(filePath))
		allowedExts := map[string]bool{
			".mp3":  true,
			".wav":  true,
			".ogg":  true,
			".m4a":  true,
			".flac": true,
		}

		if !allowedExts[ext] {
			http.Error(w, "file type not allowed", http.StatusForbidden)
			return
		}

		// Serve files from a music directory (adjust path as needed)
		musicDir := "./music" // Create this folder in your service directory
		fullPath := filepath.Join(musicDir, filePath)

		// Check if file exists
		http.ServeFile(w, r, fullPath)
	})

	// Album routes
	// GET /albums - get all albums (public)
	// POST /albums - create album (admin only, requires JWT)
	mux.HandleFunc("/albums", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			albumHandler.GetAllAlbums(w, r)
		case http.MethodPost:
			middleware.JWTAuth(cfg)(middleware.AdminOnly(albumHandler.CreateAlbum))(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// GET /albums?artistId={id} - get albums by artist
	mux.HandleFunc("/albums/by-artist", albumHandler.GetAlbumsByArtist)

	// GET /albums/{id} - get album by ID (public)
	// PUT /albums/{id} - update album (admin only, requires JWT)
	// DELETE /albums/{id} - delete album (admin only, requires JWT)
	mux.HandleFunc("/albums/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/albums/")
		if path == "" {
			http.Error(w, "album ID is required", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			albumHandler.GetAlbum(w, r)
		case http.MethodPut:
			middleware.JWTAuth(cfg)(middleware.AdminOnly(albumHandler.UpdateAlbum))(w, r)
		case http.MethodDelete:
			middleware.JWTAuth(cfg)(middleware.AdminOnly(albumHandler.DeleteAlbum))(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Song routes
	// GET /songs - get all songs (public)
	// POST /songs - create song (admin only, requires JWT)
	mux.HandleFunc("/songs", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			songHandler.GetAllSongs(w, r)
		case http.MethodPost:
			middleware.JWTAuth(cfg)(middleware.AdminOnly(songHandler.CreateSong))(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// GET /songs?albumId={id} - get songs by album
	mux.HandleFunc("/songs/by-album", songHandler.GetSongsByAlbum)

	// GET /songs/{id} - get song by ID (public)
	// PUT /songs/{id} - update song (admin only, requires JWT)
	// DELETE /songs/{id} - delete song (admin only, requires JWT)
	// GET /songs/{id}/stream - stream song audio (public)
	mux.HandleFunc("/songs/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/songs/")
		if path == "" {
			http.Error(w, "song ID is required", http.StatusBadRequest)
			return
		}

		// Check if this is a streaming request
		if strings.HasSuffix(path, "/stream") {
			songID := strings.TrimSuffix(path, "/stream")
			r.URL.Path = "/songs/" + songID + "/stream"
			songHandler.StreamSong(w, r)
			return
		}

		switch r.Method {
		case http.MethodGet:
			songHandler.GetSong(w, r)
		case http.MethodPut:
			middleware.JWTAuth(cfg)(middleware.AdminOnly(songHandler.UpdateSong))(w, r)
		case http.MethodDelete:
			middleware.JWTAuth(cfg)(middleware.AdminOnly(songHandler.DeleteSong))(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Song existence check endpoint
	mux.HandleFunc("/songs/exists", func(w http.ResponseWriter, r *http.Request) {
		songID := r.URL.Query().Get("id")
		if songID == "" {
			http.Error(w, "id query parameter is required", http.StatusBadRequest)
			return
		}

		exists, err := songRepo.Exists(r.Context(), songID)
		if err != nil {
			http.Error(w, "failed to check song existence", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]bool{"exists": exists})
	})

	// Artist routes
	// GET /artists - get all artists (public)
	// POST /artists - create artist (admin only, requires JWT)
	mux.HandleFunc("/artists", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			artistHandler.GetAllArtists(w, r)
		case http.MethodPost:
			middleware.JWTAuth(cfg)(middleware.AdminOnly(artistHandler.CreateArtist))(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// GET /artists/{id} - get artist by ID (public)
	// PUT /artists/{id} - update artist (admin only, requires JWT)
	// DELETE /artists/{id} - delete artist (admin only, requires JWT)
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
		case http.MethodDelete:
			// DELETE /artists/{id} - delete artist (admin only, requires JWT)
			middleware.JWTAuth(cfg)(middleware.AdminOnly(artistHandler.DeleteArtist))(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Println("Content service running on port", cfg.Port)
	
	// Support HTTPS if certificates are provided
	certFile := os.Getenv("TLS_CERT_FILE")
	keyFile := os.Getenv("TLS_KEY_FILE")
	if certFile != "" && keyFile != "" {
		log.Println("Starting HTTPS server on port", cfg.Port)
		server := &http.Server{
			Addr:    ":" + cfg.Port,
			Handler: mux,
		}
		if err := server.ListenAndServeTLS(certFile, keyFile); err != nil {
			if appLogger != nil {
				appLogger.LogTLSFailure("content-service", err.Error(), "")
			}
			log.Fatal("HTTPS server failed:", err)
		}
	} else {
		log.Println("Starting HTTP server on port", cfg.Port)
		log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
	}
}
