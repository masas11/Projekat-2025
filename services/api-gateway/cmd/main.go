package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"api-gateway/config"
	"api-gateway/internal/logger"
	"api-gateway/internal/middleware"
	"shared/tracing"

	"go.opentelemetry.io/otel/propagation"
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
func proxyRequest(w http.ResponseWriter, r *http.Request, targetURL string, appLogger *logger.Logger) {
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

	// Eksplicitno postavljen timeout za vraćanje odgovora korisniku (2.7.6)
	// Koristimo request context tako da se može otkazati
	// Povećan timeout za notifications-service zbog Cassandra inicijalizacije
	timeout := 5 * time.Second
	if strings.Contains(targetURL, "notifications-service") {
		timeout = 15 * time.Second
	}
	// Povećan timeout za upload operacije (HDFS može da traje)
	if strings.Contains(targetURL, "/upload") {
		timeout = 30 * time.Second
	}
	
	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	// Za multipart/form-data, prosleđujemo body direktno bez čitanja
	// Inače čitamo body i prosleđujemo ga
	var reqBody io.Reader
	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		// Za multipart, prosleđujemo originalni body direktno
		reqBody = r.Body
	} else {
		// Za ostale tipove, čitamo body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()
		reqBody = bytes.NewBuffer(body)
	}

	// Kreiranje novog zahteva ka backend servisu sa context-om
	req, err := http.NewRequestWithContext(ctx, r.Method, targetURL, reqBody)
	if err != nil {
		// CORS headers već postavljeni na početku funkcije
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	// Kopiranje headers-a (ali ne kopiraj Origin i CORS headers ka backend-u)
	// Tracing (2.10): Propagate trace context to downstream services
	propagator := tracing.GetPropagator()
	if propagator != nil {
		propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))
	}
	
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
	// Konfiguriši HTTP klijent da ignoriše sertifikate za inter-service komunikaciju
	// (jer koristimo samopotpisane sertifikate)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	
	// Client timeout mora biti veći od context timeout-a da bi context imao prioritet
	clientTimeout := timeout + 1*time.Second
	client := &http.Client{
		Timeout:   clientTimeout,
		Transport: tr,
	}
	// Wrap client with tracing (2.10)
	client = tracing.HTTPClient(client)
	
	// Kanal za rezultat zahteva
	type result struct {
		resp *http.Response
		err  error
	}
	resultChan := make(chan result, 1)
	
	// Pokreni zahtev u gorutini
	go func() {
		resp, err := client.Do(req)
		resultChan <- result{resp: resp, err: err}
	}()
	
	// Čekaj na rezultat ili timeout
	var resp *http.Response
	select {
	case <-ctx.Done():
		// Timeout istekao - vraćamo odgovor korisniku (2.7.6)
		log.Printf("Request timeout for %s: %v", targetURL, ctx.Err())
		enableCORS(w, r)
		w.WriteHeader(http.StatusRequestTimeout)
		w.Write([]byte("Request timeout - service did not respond in time"))
		return
	case res := <-resultChan:
		resp = res.resp
		err = res.err
	}
	
	if err != nil {
		// Log TLS/connection errors
		if appLogger != nil {
			errorMsg := err.Error()
			if strings.Contains(errorMsg, "tls") || strings.Contains(errorMsg, "TLS") ||
				strings.Contains(errorMsg, "certificate") || strings.Contains(errorMsg, "handshake") {
				serviceName := extractServiceName(targetURL)
				appLogger.LogTLSFailure(serviceName, errorMsg, r.RemoteAddr)
			}
		}
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

// extractServiceName extracts service name from URL
func extractServiceName(url string) string {
	if strings.Contains(url, "users-service") {
		return "users-service"
	} else if strings.Contains(url, "content-service") {
		return "content-service"
	} else if strings.Contains(url, "notifications-service") {
		return "notifications-service"
	} else if strings.Contains(url, "subscriptions-service") {
		return "subscriptions-service"
	} else if strings.Contains(url, "ratings-service") {
		return "ratings-service"
	}
	return "unknown-service"
}

// composeSongsWithRatings implements API Composition pattern
// Combines songs from content-service with ratings from ratings-service
func composeSongsWithRatings(w http.ResponseWriter, r *http.Request, cfg *config.Config, appLogger *logger.Logger) {
	enableCORS(w, r)

	// Step 1: Get songs from content-service
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	contentReq, err := http.NewRequestWithContext(ctx, "GET", cfg.ContentServiceURL+"/songs", nil)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Timeout:   5 * time.Second,
		Transport: tr,
	}
	// Wrap client with tracing (2.10)
	client = tracing.HTTPClient(client)

	contentResp, err := client.Do(contentReq)
	if err != nil {
		log.Printf("Error calling content-service: %v", err)
		http.Error(w, "Content service unavailable", http.StatusServiceUnavailable)
		return
	}
	defer contentResp.Body.Close()

	if contentResp.StatusCode != http.StatusOK {
		http.Error(w, "Failed to get songs", contentResp.StatusCode)
		return
	}

	var songs []map[string]interface{}
	if err := json.NewDecoder(contentResp.Body).Decode(&songs); err != nil {
		log.Printf("Error decoding songs: %v", err)
		http.Error(w, "Failed to decode songs", http.StatusInternalServerError)
		return
	}

	// Step 2: For each song, get average rating and count from ratings-service
	// Use goroutines for parallel requests
	type ratingResult struct {
		index        int
		averageRating float64
		ratingCount   int
		err          error
	}

	ratingChan := make(chan ratingResult, len(songs))
	
	for i, song := range songs {
		go func(idx int, s map[string]interface{}) {
			songID, ok := s["id"].(string)
			if !ok {
				ratingChan <- ratingResult{index: idx, averageRating: 0, ratingCount: 0, err: nil}
				return
			}

			ratingCtx, ratingCancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer ratingCancel()

			ratingURL := cfg.RatingsServiceURL + "/average-rating?songId=" + songID
			ratingReq, err := http.NewRequestWithContext(ratingCtx, "GET", ratingURL, nil)
			if err != nil {
				ratingChan <- ratingResult{index: idx, averageRating: 0, ratingCount: 0, err: err}
				return
			}

			ratingClient := &http.Client{
				Timeout:   3 * time.Second,
				Transport: tr,
			}
			// Wrap client with tracing (2.10)
			ratingClient = tracing.HTTPClient(ratingClient)

			ratingResp, err := ratingClient.Do(ratingReq)
			if err != nil {
				// If ratings-service is unavailable, use default values
				ratingChan <- ratingResult{index: idx, averageRating: 0, ratingCount: 0, err: nil}
				return
			}
			defer ratingResp.Body.Close()

			if ratingResp.StatusCode == http.StatusOK {
				var ratingData map[string]interface{}
				if err := json.NewDecoder(ratingResp.Body).Decode(&ratingData); err == nil {
					avg, _ := ratingData["averageRating"].(float64)
					count, _ := ratingData["ratingCount"].(float64)
					ratingChan <- ratingResult{
						index:        idx,
						averageRating: avg,
						ratingCount:   int(count),
						err:          nil,
					}
					return
				}
			}

			// Default values if rating not found
			ratingChan <- ratingResult{index: idx, averageRating: 0, ratingCount: 0, err: nil}
		}(i, song)
	}

	// Collect all rating results
	ratings := make([]ratingResult, len(songs))
	for i := 0; i < len(songs); i++ {
		ratings[i] = <-ratingChan
	}

	// Step 3: Combine songs with ratings
	for _, rating := range ratings {
		if rating.index < len(songs) {
			songs[rating.index]["averageRating"] = rating.averageRating
			songs[rating.index]["ratingCount"] = rating.ratingCount
		}
	}

	// Step 4: Return composed response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(songs)
}

// composeSongsByAlbumWithRatings implements API Composition pattern for songs by album
func composeSongsByAlbumWithRatings(w http.ResponseWriter, r *http.Request, cfg *config.Config, appLogger *logger.Logger) {
	enableCORS(w, r)

	albumID := r.URL.Query().Get("albumId")
	if albumID == "" {
		http.Error(w, "albumId parameter is required", http.StatusBadRequest)
		return
	}

	// Step 1: Get songs by album from content-service
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	contentURL := cfg.ContentServiceURL + "/songs/by-album?albumId=" + albumID
	contentReq, err := http.NewRequestWithContext(ctx, "GET", contentURL, nil)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Timeout:   5 * time.Second,
		Transport: tr,
	}
	// Wrap client with tracing (2.10)
	client = tracing.HTTPClient(client)

	contentResp, err := client.Do(contentReq)
	if err != nil {
		log.Printf("Error calling content-service: %v", err)
		http.Error(w, "Content service unavailable", http.StatusServiceUnavailable)
		return
	}
	defer contentResp.Body.Close()

	if contentResp.StatusCode != http.StatusOK {
		http.Error(w, "Failed to get songs", contentResp.StatusCode)
		return
	}

	var songs []map[string]interface{}
	if err := json.NewDecoder(contentResp.Body).Decode(&songs); err != nil {
		log.Printf("Error decoding songs: %v", err)
		http.Error(w, "Failed to decode songs", http.StatusInternalServerError)
		return
	}

	// Step 2: For each song, get average rating and count from ratings-service
	type ratingResult struct {
		index        int
		averageRating float64
		ratingCount   int
		err          error
	}

	ratingChan := make(chan ratingResult, len(songs))
	
	for i, song := range songs {
		go func(idx int, s map[string]interface{}) {
			songID, ok := s["id"].(string)
			if !ok {
				ratingChan <- ratingResult{index: idx, averageRating: 0, ratingCount: 0, err: nil}
				return
			}

			ratingCtx, ratingCancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer ratingCancel()

			ratingURL := cfg.RatingsServiceURL + "/average-rating?songId=" + songID
			ratingReq, err := http.NewRequestWithContext(ratingCtx, "GET", ratingURL, nil)
			if err != nil {
				ratingChan <- ratingResult{index: idx, averageRating: 0, ratingCount: 0, err: err}
				return
			}

			ratingClient := &http.Client{
				Timeout:   3 * time.Second,
				Transport: tr,
			}
			// Wrap client with tracing (2.10)
			ratingClient = tracing.HTTPClient(ratingClient)

			ratingResp, err := ratingClient.Do(ratingReq)
			if err != nil {
				ratingChan <- ratingResult{index: idx, averageRating: 0, ratingCount: 0, err: nil}
				return
			}
			defer ratingResp.Body.Close()

			if ratingResp.StatusCode == http.StatusOK {
				var ratingData map[string]interface{}
				if err := json.NewDecoder(ratingResp.Body).Decode(&ratingData); err == nil {
					avg, _ := ratingData["averageRating"].(float64)
					count, _ := ratingData["ratingCount"].(float64)
					ratingChan <- ratingResult{
						index:        idx,
						averageRating: avg,
						ratingCount:   int(count),
						err:          nil,
					}
					return
				}
			}

			ratingChan <- ratingResult{index: idx, averageRating: 0, ratingCount: 0, err: nil}
		}(i, song)
	}

	// Collect all rating results
	ratings := make([]ratingResult, len(songs))
	for i := 0; i < len(songs); i++ {
		ratings[i] = <-ratingChan
	}

	// Step 3: Combine songs with ratings
	for _, rating := range ratings {
		if rating.index < len(songs) {
			songs[rating.index]["averageRating"] = rating.averageRating
			songs[rating.index]["ratingCount"] = rating.ratingCount
		}
	}

	// Step 4: Return composed response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(songs)
}

func main() {
	cfg := config.Load()

	// Initialize tracing (2.10)
	cleanup, err := tracing.InitTracing("api-gateway")
	if err != nil {
		log.Printf("Warning: Failed to initialize tracing: %v", err)
	} else {
		defer cleanup()
		log.Println("Tracing initialized for api-gateway")
	}

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

	mux := http.NewServeMux()

	// Health endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "healthy",
			"service": "api-gateway",
		})
	})

	// Global rate limiting: 100 requests per minute per IP (DoS protection)
	globalRateLimit := middleware.RateLimit(100, 1*time.Minute)

	// Helper function to check if user is not admin
	requireNonAdmin := func(next http.HandlerFunc) http.HandlerFunc {
		return middleware.RequireAuth(cfg, appLogger)(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(middleware.UserContextKey).(*middleware.UserClaims)
			if !ok || claims == nil {
				enableCORS(w, r)
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			// Check if user is NOT admin
			if claims.Role == "ADMIN" {
				enableCORS(w, r)
				http.Error(w, "admin users cannot perform this action", http.StatusForbidden)
				return
			}

			next(w, r)
		})
	}

	// USERS SERVICE ROUTES
	mux.HandleFunc("/api/users/health", globalRateLimit(func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.UsersServiceURL+"/health", appLogger)
	}))

	mux.HandleFunc("/api/users/register", globalRateLimit(func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.UsersServiceURL+"/register", appLogger)
	}))

	mux.HandleFunc("/api/users/verify-email", globalRateLimit(func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.UsersServiceURL+"/verify-email", appLogger)
	}))

	mux.HandleFunc("/api/users/login/request-otp", globalRateLimit(func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.UsersServiceURL+"/login/request-otp", appLogger)
	}))

	mux.HandleFunc("/api/users/login/verify-otp", globalRateLimit(func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.UsersServiceURL+"/login/verify-otp", appLogger)
	}))

	mux.HandleFunc("/api/users/logout", globalRateLimit(middleware.RequireAuth(cfg, appLogger)(func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.UsersServiceURL+"/logout", appLogger)
	})))

	mux.HandleFunc("/api/users/password/change", globalRateLimit(middleware.RequireAuth(cfg, appLogger)(func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.UsersServiceURL+"/password/change", appLogger)
	})))

	mux.HandleFunc("/api/users/password/reset/request", globalRateLimit(func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.UsersServiceURL+"/password/reset/request", appLogger)
	}))

	mux.HandleFunc("/api/users/password/reset", globalRateLimit(func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.UsersServiceURL+"/password/reset", appLogger)
	}))

	// Magic link endpoints (account recovery)
	mux.HandleFunc("/api/users/recover/request", globalRateLimit(func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.UsersServiceURL+"/recover/request", appLogger)
	}))
	mux.HandleFunc("/api/users/recover/verify", globalRateLimit(func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.UsersServiceURL+"/recover/verify", appLogger)
	}))

	// CONTENT SERVICE ROUTES
	mux.HandleFunc("/api/content/health", func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.ContentServiceURL+"/health", appLogger)
	})

	// Artists routes
	// GET /api/content/artists - get all artists (public)
	// POST /api/content/artists - create artist (admin only)
	mux.HandleFunc("/api/content/artists", globalRateLimit(middleware.OptionalAuth(cfg, appLogger)(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			middleware.RequireRole("ADMIN", cfg, appLogger)(func(w http.ResponseWriter, r *http.Request) {
				proxyRequest(w, r, cfg.ContentServiceURL+"/artists", appLogger)
			})(w, r)
		} else {
			proxyRequest(w, r, cfg.ContentServiceURL+"/artists", appLogger)
		}
	})))

	// GET /api/content/artists/{id} - get artist by ID (public)
	// PUT /api/content/artists/{id} - update artist (admin only)
	mux.HandleFunc("/api/content/artists/", globalRateLimit(middleware.OptionalAuth(cfg, appLogger)(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path[len("/api/content/artists/"):]
		if r.Method == http.MethodPut || r.Method == http.MethodDelete {
			middleware.RequireRole("ADMIN", cfg, appLogger)(func(w http.ResponseWriter, r *http.Request) {
				proxyRequest(w, r, cfg.ContentServiceURL+"/artists/"+path, appLogger)
			})(w, r)
		} else {
			proxyRequest(w, r, cfg.ContentServiceURL+"/artists/"+path, appLogger)
		}
	})))

	// Album routes
	// GET /api/content/albums - get all albums (public)
	// POST /api/content/albums - create album (admin only)
	mux.HandleFunc("/api/content/albums", globalRateLimit(middleware.OptionalAuth(cfg, appLogger)(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			middleware.RequireRole("ADMIN", cfg, appLogger)(func(w http.ResponseWriter, r *http.Request) {
				proxyRequest(w, r, cfg.ContentServiceURL+"/albums", appLogger)
			})(w, r)
		} else {
			proxyRequest(w, r, cfg.ContentServiceURL+"/albums", appLogger)
		}
	})))

	// GET /api/content/albums/by-artist?artistId={id} - get albums by artist
	mux.HandleFunc("/api/content/albums/by-artist", globalRateLimit(func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.ContentServiceURL+"/albums/by-artist", appLogger)
	}))

	// GET /api/content/albums/{id} - get album by ID
	mux.HandleFunc("/api/content/albums/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path[len("/api/content/albums/"):]
		proxyRequest(w, r, cfg.ContentServiceURL+"/albums/"+path, appLogger)
	})

	// GET /api/content/songs/by-album?albumId={id} - get songs by album with ratings (API Composition)
	mux.HandleFunc("/api/content/songs/by-album", globalRateLimit(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			// API Composition: Combine songs from content-service with ratings from ratings-service
			composeSongsByAlbumWithRatings(w, r, cfg, appLogger)
		} else {
			enableCORS(w, r)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// GET /api/content/songs/most-played?limit={n} - get most played songs (2.12)
	mux.HandleFunc("/api/content/songs/most-played", globalRateLimit(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			enableCORS(w, r)
			proxyRequest(w, r, cfg.ContentServiceURL+"/songs/most-played", appLogger)
		} else if r.Method == "OPTIONS" {
			enableCORS(w, r)
			w.WriteHeader(http.StatusOK)
		} else {
			enableCORS(w, r)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// GET /api/content/songs - get all songs with ratings (API Composition)
	// POST /api/content/songs - create song (admin only)
	mux.HandleFunc("/api/content/songs", globalRateLimit(middleware.OptionalAuth(cfg, appLogger)(func(w http.ResponseWriter, r *http.Request) {
		// Handle preflight OPTIONS request
		if r.Method == "OPTIONS" {
			enableCORS(w, r)
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method == http.MethodPost {
			middleware.RequireRole("ADMIN", cfg, appLogger)(func(w http.ResponseWriter, r *http.Request) {
				proxyRequest(w, r, cfg.ContentServiceURL+"/songs", appLogger)
			})(w, r)
		} else if r.Method == http.MethodGet {
			// API Composition: Combine songs from content-service with ratings from ratings-service
			composeSongsWithRatings(w, r, cfg, appLogger)
		} else {
			enableCORS(w, r)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// GET /api/content/songs/{id} - get song by ID
	// PUT /api/content/songs/{id} - update song (admin only)
	// DELETE /api/content/songs/{id} - delete song via saga (admin only) (2.13)
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
			// Use OptionalAuth to extract userID if token is present (for activity logging 1.15)
			middleware.OptionalAuth(cfg, appLogger)(func(w http.ResponseWriter, r *http.Request) {
				proxyRequest(w, r, cfg.ContentServiceURL+"/songs/"+path, appLogger)
			})(w, r)
			return
		}

		// Check if this is an upload request (2.11)
		if strings.HasSuffix(path, "/upload") && r.Method == http.MethodPost {
			log.Printf("Upload request detected: %s", path)
			middleware.RequireRole("ADMIN", cfg, appLogger)(func(w http.ResponseWriter, r *http.Request) {
				targetURL := cfg.ContentServiceURL + "/songs/" + path
				log.Printf("Proxying upload to: %s", targetURL)
				proxyRequest(w, r, targetURL, appLogger)
			})(w, r)
			return
		}

		// Handle specific song ID routes (GET, PUT, DELETE)
		if path != "" && !strings.Contains(path, "/") {
			if r.Method == http.MethodGet {
				proxyRequest(w, r, cfg.ContentServiceURL+"/songs/"+path, appLogger)
			} else if r.Method == http.MethodPut {
				middleware.RequireRole("ADMIN", cfg, appLogger)(func(w http.ResponseWriter, r *http.Request) {
					proxyRequest(w, r, cfg.ContentServiceURL+"/songs/"+path, appLogger)
				})(w, r)
			} else if r.Method == http.MethodDelete {
				middleware.RequireRole("ADMIN", cfg, appLogger)(func(w http.ResponseWriter, r *http.Request) {
					proxyRequest(w, r, cfg.ContentServiceURL+"/songs/"+path, appLogger)
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
		proxyRequest(w, r, cfg.NotificationsServiceURL+"/health", appLogger)
	})

	// GET /api/notifications - get notifications for authenticated user (requires auth)
	// userId is extracted from JWT token, not from query parameter for security
	mux.HandleFunc("/api/notifications", globalRateLimit(middleware.RequireAuth(cfg, appLogger)(func(w http.ResponseWriter, r *http.Request) {
		// Get userId from JWT token (set by RequireAuth middleware)
		claims, ok := r.Context().Value(middleware.UserContextKey).(*middleware.UserClaims)
		if !ok || claims == nil {
			enableCORS(w, r)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Use userId from JWT token, ignore any userId in query parameters for security
		query := "?userId=" + claims.UserID
		proxyRequest(w, r, cfg.NotificationsServiceURL+"/notifications"+query, appLogger)
	})))

	// SUBSCRIPTIONS SERVICE ROUTES
	mux.HandleFunc("/api/subscriptions/health", globalRateLimit(func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.SubscriptionsServiceURL+"/health", appLogger)
	}))

	// GET /api/subscriptions - get user subscriptions (requires auth)
	mux.HandleFunc("/api/subscriptions", globalRateLimit(middleware.RequireAuth(cfg, appLogger)(func(w http.ResponseWriter, r *http.Request) {
		// Get userId from JWT token
		claims, ok := r.Context().Value(middleware.UserContextKey).(*middleware.UserClaims)
		if !ok || claims == nil {
			enableCORS(w, r)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		// Add userId to query params
		query := "?userId=" + claims.UserID
		proxyRequest(w, r, cfg.SubscriptionsServiceURL+"/subscriptions"+query, appLogger)
	})))

	// POST /api/subscriptions/subscribe-artist - subscribe to artist (requires auth, non-admin only)
	// DELETE /api/subscriptions/subscribe-artist - unsubscribe from artist (requires auth, non-admin only)
	mux.HandleFunc("/api/subscriptions/subscribe-artist", globalRateLimit(requireNonAdmin(func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.SubscriptionsServiceURL+"/subscribe-artist", appLogger)
	})))

	// POST /api/subscriptions/subscribe-genre - subscribe to genre (requires auth, non-admin only)
	// DELETE /api/subscriptions/subscribe-genre - unsubscribe from genre (requires auth, non-admin only)
	mux.HandleFunc("/api/subscriptions/subscribe-genre", globalRateLimit(requireNonAdmin(func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.SubscriptionsServiceURL+"/subscribe-genre", appLogger)
	})))

	// RATINGS SERVICE ROUTES
	mux.HandleFunc("/api/ratings/health", globalRateLimit(func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.RatingsServiceURL+"/health", appLogger)
	}))

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
		proxyRequest(w, r, targetURL, appLogger)
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
		proxyRequest(w, r, targetURL, appLogger)
	})))

	// GET /api/ratings/average-rating - get average rating and count for a song (public)
	mux.HandleFunc("/api/ratings/average-rating", globalRateLimit(func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.RatingsServiceURL+"/average-rating", appLogger)
	}))

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
		proxyRequest(w, r, targetURL, appLogger)
	})))

	// GET /api/ratings/recommendations - get personalized recommendations (requires auth, non-admin only)
	mux.HandleFunc("/api/ratings/recommendations", globalRateLimit(requireNonAdmin(func(w http.ResponseWriter, r *http.Request) {
		// Get userId from JWT token
		claims, ok := r.Context().Value(middleware.UserContextKey).(*middleware.UserClaims)
		if !ok || claims == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Get query parameters
		query := r.URL.Query()
		userId := query.Get("userId")
		if userId == "" {
			userId = claims.UserID
		}

		// Use recommendation-service instead of ratings-service
		// Clean userId to avoid double encoding
		cleanUserId := userId
		if idx := strings.Index(cleanUserId, "?"); idx != -1 {
			cleanUserId = cleanUserId[:idx]
		}
		if idx := strings.Index(cleanUserId, "&"); idx != -1 {
			cleanUserId = cleanUserId[:idx]
		}
		targetURL := cfg.RecommendationServiceURL + "/recommendations?userId=" + cleanUserId
		proxyRequest(w, r, targetURL, appLogger)
	})))

	// ANALYTICS SERVICE ROUTES (1.15)
	// GET /api/analytics/activities - get user activities (requires auth, non-admin only)
	mux.HandleFunc("/api/analytics/activities", globalRateLimit(requireNonAdmin(func(w http.ResponseWriter, r *http.Request) {
		// Get userId from JWT token
		claims, ok := r.Context().Value(middleware.UserContextKey).(*middleware.UserClaims)
		if !ok || claims == nil {
			log.Printf("Failed to get user claims from context")
			enableCORS(w, r)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		log.Printf("Getting activities for user: %s", claims.UserID)

		// Add userId from JWT token to query params
		query := r.URL.RawQuery
		if query != "" {
			query += "&userId=" + claims.UserID
		} else {
			query = "userId=" + claims.UserID
		}

		// Create new request with updated query
		targetURL := cfg.AnalyticsServiceURL + "/activities?" + query
		log.Printf("Proxying analytics activities request to: %s", targetURL)
		proxyRequest(w, r, targetURL, appLogger)
	})))

	// EVENT SOURCING ROUTES (2.14)
	// GET /api/analytics/events/stream - get event stream for a user
	mux.HandleFunc("/api/analytics/events/stream", globalRateLimit(requireNonAdmin(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(middleware.UserContextKey).(*middleware.UserClaims)
		if !ok || claims == nil {
			log.Printf("Failed to get user claims from context")
			enableCORS(w, r)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		query := r.URL.RawQuery
		if query != "" {
			query += "&userId=" + claims.UserID
		} else {
			query = "userId=" + claims.UserID
		}

		targetURL := cfg.AnalyticsServiceURL + "/events/stream?" + query
		log.Printf("Proxying event stream request to: %s", targetURL)
		proxyRequest(w, r, targetURL, appLogger)
	})))

	// GET /api/analytics/events/replay - replay events to reconstruct state
	mux.HandleFunc("/api/analytics/events/replay", globalRateLimit(requireNonAdmin(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(middleware.UserContextKey).(*middleware.UserClaims)
		if !ok || claims == nil {
			log.Printf("Failed to get user claims from context")
			enableCORS(w, r)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		query := r.URL.RawQuery
		if query != "" {
			query += "&userId=" + claims.UserID
		} else {
			query = "userId=" + claims.UserID
		}

		targetURL := cfg.AnalyticsServiceURL + "/events/replay?" + query
		log.Printf("Proxying event replay request to: %s", targetURL)
		proxyRequest(w, r, targetURL, appLogger)
	})))

	// SAGA SERVICE ROUTES (2.13)
	// POST /api/sagas/delete-song - start saga transaction for song deletion (admin only)
	mux.HandleFunc("/api/sagas/delete-song", globalRateLimit(middleware.RequireRole("ADMIN", cfg, appLogger)(func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method == http.MethodPost {
			proxyRequest(w, r, cfg.SagaServiceURL+"/sagas/delete-song", appLogger)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// GET /api/sagas/{id} - get saga transaction status
	mux.HandleFunc("/api/sagas/", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method == http.MethodGet {
			sagaID := r.URL.Path[len("/api/sagas/"):]
			if sagaID == "" {
				http.Error(w, "saga ID is required", http.StatusBadRequest)
				return
			}
			proxyRequest(w, r, cfg.SagaServiceURL+"/sagas/"+sagaID, appLogger)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Note: Root endpoint "/" is intentionally not registered
	// In Go ServeMux, "/" is a catch-all that would interfere with other routes
	// Use /health endpoint instead for API Gateway status
	// All API endpoints are under /api/*

	log.Println("API Gateway running on port", cfg.Port)

	// Support HTTPS if certificates are provided
	certFile := os.Getenv("TLS_CERT_FILE")
	keyFile := os.Getenv("TLS_KEY_FILE")
	if certFile != "" && keyFile != "" {
		log.Println("Starting HTTPS server on port", cfg.Port)
		server := &http.Server{
			Addr:    ":" + cfg.Port,
			Handler: mux,
		}
		// Wrap server handler with tracing middleware
		server.Handler = tracing.HTTPMiddleware(mux)
		if err := server.ListenAndServeTLS(certFile, keyFile); err != nil {
			if appLogger != nil {
				appLogger.LogTLSFailure("api-gateway", err.Error(), "")
			}
			log.Fatal("HTTPS server failed:", err)
		}
	} else {
		log.Println("Starting HTTP server on port", cfg.Port)
		// Wrap mux with tracing middleware
		handler := tracing.HTTPMiddleware(mux)
		log.Fatal(http.ListenAndServe(":"+cfg.Port, handler))
	}
}
