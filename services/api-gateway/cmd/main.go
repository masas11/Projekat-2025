package main

import (
	"bytes"
	"io"
	"log"
	"net/http"

	"api-gateway/config"
)

// enableCORS dodaje CORS headers u odgovor
func enableCORS(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin == "" {
		origin = "*"
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
}

// proxyRequest prosleđuje zahtev ka backend servisu
func proxyRequest(w http.ResponseWriter, r *http.Request, targetURL string) {
	// Dodaj CORS headers
	enableCORS(w, r)
	
	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	// Čitanje body-ja zahteva
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Kreiranje novog zahteva ka backend servisu
	req, err := http.NewRequest(r.Method, targetURL, bytes.NewBuffer(body))
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	// Kopiranje headers-a (ali ne kopiraj Origin i CORS headers ka backend-u)
	for key, values := range r.Header {
		// Preskoči CORS headers i Origin pri prosleđivanju ka backend-u
		if key == "Origin" || key == "Access-Control-Request-Method" || key == "Access-Control-Request-Headers" {
			continue
		}
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Slanje zahteva
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	// Kopiranje status koda i headers-a
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(resp.StatusCode)

	// Kopiranje body-ja odgovora
	responseBody, _ := io.ReadAll(resp.Body)
	w.Write(responseBody)
}

func main() {
	cfg := config.Load()

	mux := http.NewServeMux()

	// USERS SERVICE ROUTES
	mux.HandleFunc("/api/users/health", func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.UsersServiceURL+"/health")
	})

	mux.HandleFunc("/api/users/register", func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.UsersServiceURL+"/register")
	})

	mux.HandleFunc("/api/users/login/request-otp", func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.UsersServiceURL+"/login/request-otp")
	})

	mux.HandleFunc("/api/users/login/verify-otp", func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.UsersServiceURL+"/login/verify-otp")
	})

	mux.HandleFunc("/api/users/password/change", func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.UsersServiceURL+"/password/change")
	})

	mux.HandleFunc("/api/users/password/reset", func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.UsersServiceURL+"/password/reset")
	})

	// CONTENT SERVICE ROUTES
	mux.HandleFunc("/api/content/health", func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.ContentServiceURL+"/health")
	})

	// Artists routes
	// GET /api/content/artists - get all artists
	// POST /api/content/artists - create artist (admin only)
	mux.HandleFunc("/api/content/artists", func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.ContentServiceURL+"/artists")
	})

	// GET /api/content/artists/{id} - get artist by ID
	// PUT /api/content/artists/{id} - update artist (admin only)
	mux.HandleFunc("/api/content/artists/", func(w http.ResponseWriter, r *http.Request) {
		// Extract the path after /api/content/artists/
		path := r.URL.Path[len("/api/content/artists/"):]
		proxyRequest(w, r, cfg.ContentServiceURL+"/artists/"+path)
	})

	// Album routes
	// GET /api/content/albums - get all albums
	// POST /api/content/albums - create album (admin only)
	mux.HandleFunc("/api/content/albums", func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.ContentServiceURL+"/albums")
	})

	// GET /api/content/albums/by-artist?artistId={id} - get albums by artist
	mux.HandleFunc("/api/content/albums/by-artist", func(w http.ResponseWriter, r *http.Request) {
		query := ""
		if r.URL.RawQuery != "" {
			query = "?" + r.URL.RawQuery
		}
		proxyRequest(w, r, cfg.ContentServiceURL+"/albums/by-artist"+query)
	})

	// GET /api/content/albums/{id} - get album by ID
	mux.HandleFunc("/api/content/albums/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path[len("/api/content/albums/"):]
		proxyRequest(w, r, cfg.ContentServiceURL+"/albums/"+path)
	})

	// Song routes
	// GET /api/content/songs - get all songs
	// POST /api/content/songs - create song (admin only)
	mux.HandleFunc("/api/content/songs", func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.ContentServiceURL+"/songs")
	})

	// GET /api/content/songs/by-album?albumId={id} - get songs by album
	mux.HandleFunc("/api/content/songs/by-album", func(w http.ResponseWriter, r *http.Request) {
		query := ""
		if r.URL.RawQuery != "" {
			query = "?" + r.URL.RawQuery
		}
		proxyRequest(w, r, cfg.ContentServiceURL+"/songs/by-album"+query)
	})

	// GET /api/content/songs/{id} - get song by ID
	mux.HandleFunc("/api/content/songs/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path[len("/api/content/songs/"):]
		proxyRequest(w, r, cfg.ContentServiceURL+"/songs/"+path)
	})

	// NOTIFICATIONS SERVICE ROUTES
	mux.HandleFunc("/api/notifications/health", func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.NotificationsServiceURL+"/health")
	})

	// GET /api/notifications?userId={id} - get notifications for user
	mux.HandleFunc("/api/notifications", func(w http.ResponseWriter, r *http.Request) {
		query := ""
		if r.URL.RawQuery != "" {
			query = "?" + r.URL.RawQuery
		}
		proxyRequest(w, r, cfg.NotificationsServiceURL+"/notifications"+query)
	})

	log.Println("API Gateway running on port", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}
