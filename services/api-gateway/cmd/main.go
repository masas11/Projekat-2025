package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"api-gateway/config"
	"api-gateway/internal/middleware"
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

	// Preslikaj query parametre iz originalnog zahteva
	if r.URL.RawQuery != "" {
		targetURL = targetURL + "?" + r.URL.RawQuery
	}

	// Čitanje body-ja zahteva
	body, err := io.ReadAll(r.Body)
	if err != nil {
		// CORS headers već postavljeni na početku funkcije
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Kreiranje novog zahteva ka backend servisu
	req, err := http.NewRequest(r.Method, targetURL, bytes.NewBuffer(body))
	if err != nil {
		// CORS headers već postavljeni na početku funkcije
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
	client := &http.Client{
		Timeout: 5 * time.Second, // Timeout za pozive backend servisa
	}
	resp, err := client.Do(req)
	if err != nil {
		// CORS headers već postavljeni na početku funkcije
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	// Kopiranje status koda i headers-a (ali ne kopiraj CORS headers - API Gateway kontroliše CORS)
	for key, values := range resp.Header {
		// Preskoči CORS headers pri kopiranju odgovora - API Gateway kontroliše CORS
		if strings.HasPrefix(key, "Access-Control-") {
			continue
		}
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	// Postavi CORS headers nakon kopiranja backend headers-a (ovo prepisuje one koje smo možda propustili)
	enableCORS(w, r)
	w.WriteHeader(resp.StatusCode)

	// Kopiranje body-ja odgovora
	responseBody, _ := io.ReadAll(resp.Body)
	w.Write(responseBody)
}

func main() {
	cfg := config.Load()

	mux := http.NewServeMux()

	// Global rate limiting: 100 requests per minute per IP (DoS protection)
	globalRateLimit := middleware.RateLimit(100, 1*time.Minute)

	// USERS SERVICE ROUTES
	mux.HandleFunc("/api/users/health", globalRateLimit(func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.UsersServiceURL+"/health")
	}))

	mux.HandleFunc("/api/users/register", globalRateLimit(func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.UsersServiceURL+"/register")
	}))

	mux.HandleFunc("/api/users/verify-email", globalRateLimit(func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.UsersServiceURL+"/verify-email")
	}))

	mux.HandleFunc("/api/users/login/request-otp", globalRateLimit(func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.UsersServiceURL+"/login/request-otp")
	}))

	mux.HandleFunc("/api/users/login/verify-otp", globalRateLimit(func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.UsersServiceURL+"/login/verify-otp")
	}))

	mux.HandleFunc("/api/users/logout", globalRateLimit(middleware.RequireAuth(cfg)(func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.UsersServiceURL+"/logout")
	})))

	mux.HandleFunc("/api/users/password/change", globalRateLimit(middleware.RequireAuth(cfg)(func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.UsersServiceURL+"/password/change")
	})))

	mux.HandleFunc("/api/users/password/reset/request", globalRateLimit(func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.UsersServiceURL+"/password/reset/request")
	}))

	mux.HandleFunc("/api/users/password/reset", globalRateLimit(func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.UsersServiceURL+"/password/reset")
	}))

	// Magic link endpoints (account recovery)
	mux.HandleFunc("/api/users/recover/request", globalRateLimit(func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.UsersServiceURL+"/recover/request")
	}))
	mux.HandleFunc("/api/users/recover/verify", globalRateLimit(func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.UsersServiceURL+"/recover/verify")
	}))

	// CONTENT SERVICE ROUTES
	mux.HandleFunc("/api/content/health", func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.ContentServiceURL+"/health")
	})

	// Artists routes
	// GET /api/content/artists - get all artists (public)
	// POST /api/content/artists - create artist (admin only)
	mux.HandleFunc("/api/content/artists", globalRateLimit(middleware.OptionalAuth(cfg)(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			middleware.RequireRole("ADMIN", cfg)(func(w http.ResponseWriter, r *http.Request) {
				proxyRequest(w, r, cfg.ContentServiceURL+"/artists")
			})(w, r)
		} else {
			proxyRequest(w, r, cfg.ContentServiceURL+"/artists")
		}
	})))

	// GET /api/content/artists/{id} - get artist by ID (public)
	// PUT /api/content/artists/{id} - update artist (admin only)
	mux.HandleFunc("/api/content/artists/", globalRateLimit(middleware.OptionalAuth(cfg)(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path[len("/api/content/artists/"):]
		if r.Method == http.MethodPut || r.Method == http.MethodDelete {
			middleware.RequireRole("ADMIN", cfg)(func(w http.ResponseWriter, r *http.Request) {
				proxyRequest(w, r, cfg.ContentServiceURL+"/artists/"+path)
			})(w, r)
		} else {
			proxyRequest(w, r, cfg.ContentServiceURL+"/artists/"+path)
		}
	})))

	// Album routes
	// GET /api/content/albums - get all albums (public)
	// POST /api/content/albums - create album (admin only)
	mux.HandleFunc("/api/content/albums", globalRateLimit(middleware.OptionalAuth(cfg)(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			middleware.RequireRole("ADMIN", cfg)(func(w http.ResponseWriter, r *http.Request) {
				proxyRequest(w, r, cfg.ContentServiceURL+"/albums")
			})(w, r)
		} else {
			proxyRequest(w, r, cfg.ContentServiceURL+"/albums")
		}
	})))

	// GET /api/content/albums/by-artist?artistId={id} - get albums by artist
	mux.HandleFunc("/api/content/albums/by-artist", globalRateLimit(func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.ContentServiceURL+"/albums/by-artist")
	}))

	// GET /api/content/albums/{id} - get album by ID
	mux.HandleFunc("/api/content/albums/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path[len("/api/content/albums/"):]
		proxyRequest(w, r, cfg.ContentServiceURL+"/albums/"+path)
	})

	// GET /api/content/songs/by-album?albumId={id} - get songs by album
	mux.HandleFunc("/api/content/songs/by-album", globalRateLimit(func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.ContentServiceURL+"/songs/by-album")
	}))

	// GET /api/content/songs - get all songs (public)
	// POST /api/content/songs - create song (admin only)
	mux.HandleFunc("/api/content/songs", globalRateLimit(middleware.OptionalAuth(cfg)(func(w http.ResponseWriter, r *http.Request) {
		// Handle preflight OPTIONS request
		if r.Method == "OPTIONS" {
			enableCORS(w, r)
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method == http.MethodPost {
			middleware.RequireRole("ADMIN", cfg)(func(w http.ResponseWriter, r *http.Request) {
				proxyRequest(w, r, cfg.ContentServiceURL+"/songs")
			})(w, r)
		} else {
			proxyRequest(w, r, cfg.ContentServiceURL+"/songs")
		}
	})))

	// GET /api/content/songs/{id} - get song by ID
	// PUT /api/content/songs/{id} - update song (admin only)
	// DELETE /api/content/songs/{id} - delete song (admin only)
	// GET /api/content/songs/{id}/stream - stream song audio (public)
	mux.HandleFunc("/api/content/songs/", func(w http.ResponseWriter, r *http.Request) {
		// Handle preflight OPTIONS request
		if r.Method == "OPTIONS" {
			enableCORS(w, r)
			w.WriteHeader(http.StatusOK)
			return
		}

		path := r.URL.Path[len("/api/content/songs/"):]

		// Check if this is a streaming request
		if strings.HasSuffix(path, "/stream") {
			proxyRequest(w, r, cfg.ContentServiceURL+"/songs/"+path)
			return
		}

		// Handle specific song ID routes (GET, PUT, DELETE)
		if path != "" && !strings.Contains(path, "/") {
			if r.Method == http.MethodGet {
				proxyRequest(w, r, cfg.ContentServiceURL+"/songs/"+path)
			} else if r.Method == http.MethodPut {
				middleware.RequireRole("ADMIN", cfg)(func(w http.ResponseWriter, r *http.Request) {
					proxyRequest(w, r, cfg.ContentServiceURL+"/songs/"+path)
				})(w, r)
			} else if r.Method == http.MethodDelete {
				middleware.RequireRole("ADMIN", cfg)(func(w http.ResponseWriter, r *http.Request) {
					proxyRequest(w, r, cfg.ContentServiceURL+"/songs/"+path)
				})(w, r)
			} else {
				enableCORS(w, r)
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		// This should not be reached due to the above logic
		enableCORS(w, r)
		http.Error(w, "invalid request", http.StatusBadRequest)
	})

	// NOTIFICATIONS SERVICE ROUTES
	mux.HandleFunc("/api/notifications/health", func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.NotificationsServiceURL+"/health")
	})

	// GET /api/notifications - get notifications for authenticated user (requires auth)
	// userId is extracted from JWT token, not from query parameter for security
	mux.HandleFunc("/api/notifications", globalRateLimit(middleware.RequireAuth(cfg)(func(w http.ResponseWriter, r *http.Request) {
		// Get userId from JWT token (set by RequireAuth middleware)
		claims, ok := r.Context().Value(middleware.UserContextKey).(*middleware.UserClaims)
		if !ok || claims == nil {
			enableCORS(w, r)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Use userId from JWT token, ignore any userId in query parameters for security
		query := "?userId=" + claims.UserID
		proxyRequest(w, r, cfg.NotificationsServiceURL+"/notifications"+query)
	})))

	// SUBSCRIPTIONS SERVICE ROUTES
	mux.HandleFunc("/api/subscriptions/health", globalRateLimit(func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.SubscriptionsServiceURL+"/health")
	}))

	// GET /api/subscriptions - get user subscriptions (requires auth)
	mux.HandleFunc("/api/subscriptions", globalRateLimit(middleware.RequireAuth(cfg)(func(w http.ResponseWriter, r *http.Request) {
		// Get userId from JWT token
		claims, ok := r.Context().Value(middleware.UserContextKey).(*middleware.UserClaims)
		if !ok || claims == nil {
			enableCORS(w, r)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		// Add userId to query params
		query := "?userId=" + claims.UserID
		proxyRequest(w, r, cfg.SubscriptionsServiceURL+"/subscriptions"+query)
	})))

	// POST /api/subscriptions/subscribe-artist - subscribe to artist (requires auth)
	// DELETE /api/subscriptions/subscribe-artist - unsubscribe from artist (requires auth)
	mux.HandleFunc("/api/subscriptions/subscribe-artist", globalRateLimit(middleware.RequireAuth(cfg)(func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.SubscriptionsServiceURL+"/subscribe-artist")
	})))

	// POST /api/subscriptions/subscribe-genre - subscribe to genre (requires auth)
	// DELETE /api/subscriptions/subscribe-genre - unsubscribe from genre (requires auth)
	mux.HandleFunc("/api/subscriptions/subscribe-genre", globalRateLimit(middleware.RequireAuth(cfg)(func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.SubscriptionsServiceURL+"/subscribe-genre")
	})))

	// RATINGS SERVICE ROUTES
	mux.HandleFunc("/api/ratings/health", globalRateLimit(func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.RatingsServiceURL+"/health")
	}))

	// Helper function to check if user is not admin
	requireNonAdmin := func(next http.HandlerFunc) http.HandlerFunc {
		return middleware.RequireAuth(cfg)(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(middleware.UserContextKey).(*middleware.UserClaims)
			if !ok || claims == nil {
				enableCORS(w, r)
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			// Check if user is NOT admin
			if claims.Role == "ADMIN" {
				enableCORS(w, r)
				http.Error(w, "admin users cannot rate songs", http.StatusForbidden)
				return
			}

			next(w, r)
		})
	}

	// POST /api/ratings/rate-song - rate/update a song (requires auth, non-admin only)
	mux.HandleFunc("/api/ratings/rate-song", globalRateLimit(requireNonAdmin(func(w http.ResponseWriter, r *http.Request) {
		// Get userId from JWT token
		claims, ok := r.Context().Value(middleware.UserContextKey).(*middleware.UserClaims)
		if !ok || claims == nil {
			enableCORS(w, r)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Add userId from JWT token to query params
		query := r.URL.RawQuery
		if query != "" {
			query += "&userId=" + claims.UserID
		} else {
			query = "userId=" + claims.UserID
		}

		// Create new request with updated query
		targetURL := cfg.RatingsServiceURL + "/rate-song?" + query
		proxyRequest(w, r, targetURL)
	})))

	// DELETE /api/ratings/delete-rating - delete a rating (requires auth, non-admin only)
	mux.HandleFunc("/api/ratings/delete-rating", globalRateLimit(requireNonAdmin(func(w http.ResponseWriter, r *http.Request) {
		// Get userId from JWT token
		claims, ok := r.Context().Value(middleware.UserContextKey).(*middleware.UserClaims)
		if !ok || claims == nil {
			enableCORS(w, r)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Add userId from JWT token to query params
		query := r.URL.RawQuery
		if query != "" {
			query += "&userId=" + claims.UserID
		} else {
			query = "userId=" + claims.UserID
		}

		// Create new request with updated query
		targetURL := cfg.RatingsServiceURL + "/delete-rating?" + query
		proxyRequest(w, r, targetURL)
	})))

	// GET /api/ratings/get-rating - get user's rating for a song (requires auth, non-admin only)
	mux.HandleFunc("/api/ratings/get-rating", globalRateLimit(requireNonAdmin(func(w http.ResponseWriter, r *http.Request) {
		// Get userId from JWT token
		claims, ok := r.Context().Value(middleware.UserContextKey).(*middleware.UserClaims)
		if !ok || claims == nil {
			enableCORS(w, r)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Add userId from JWT token to query params
		query := r.URL.RawQuery
		if query != "" {
			query += "&userId=" + claims.UserID
		} else {
			query = "userId=" + claims.UserID
		}

		// Create new request with updated query
		targetURL := cfg.RatingsServiceURL + "/get-rating?" + query
		proxyRequest(w, r, targetURL)
	})))

	log.Println("API Gateway running on port", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}
