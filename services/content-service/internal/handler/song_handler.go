package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"content-service/internal/cache"
	"content-service/internal/dto"
	"content-service/internal/events"
	"content-service/internal/logger"
	"content-service/internal/middleware"
	"content-service/internal/model"
	"content-service/internal/storage"
	"content-service/internal/store"
	"shared/analytics"
)

func extractSongID(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 2 && parts[0] == "songs" {
		return parts[1]
	}
	return ""
}

type SongHandler struct {
	Repo                     *store.SongRepository
	AlbumRepo                *store.AlbumRepository
	ArtistRepo               *store.ArtistRepository
	SubscriptionsServiceURL  string
	RecommendationServiceURL  string
	RatingsServiceURL        string
	AnalyticsServiceURL      string
	SagaServiceURL           string // (2.13)
	Logger                   *logger.Logger
	HDFSClient               *storage.HDFSClient
	RedisCache               *cache.RedisCache // (2.12)
}

func NewSongHandler(repo *store.SongRepository, albumRepo *store.AlbumRepository, artistRepo *store.ArtistRepository, subscriptionsServiceURL, recommendationServiceURL, ratingsServiceURL, analyticsServiceURL, sagaServiceURL string, log *logger.Logger, hdfsClient *storage.HDFSClient, redisCache *cache.RedisCache) *SongHandler {
	return &SongHandler{
		Repo:                     repo,
		AlbumRepo:                albumRepo,
		ArtistRepo:               artistRepo,
		SubscriptionsServiceURL:  subscriptionsServiceURL,
		RecommendationServiceURL: recommendationServiceURL,
		RatingsServiceURL:        ratingsServiceURL,
		AnalyticsServiceURL:      analyticsServiceURL,
		SagaServiceURL:           sagaServiceURL,
		Logger:                   log,
		HDFSClient:               hdfsClient,
		RedisCache:               redisCache,
	}
}

// getAdminIDFromContext extracts admin user ID from request context
func getAdminIDFromSongContext(ctx context.Context) string {
	claims, ok := ctx.Value(middleware.UserContextKey).(*middleware.UserClaims)
	if !ok || claims == nil {
		return ""
	}
	return claims.UserID
}

func (h *SongHandler) CreateSong(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req dto.CreateSongRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	// Validation
	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	if req.Duration <= 0 {
		http.Error(w, "duration must be greater than 0", http.StatusBadRequest)
		return
	}
	if req.Genre == "" {
		http.Error(w, "genre is required", http.StatusBadRequest)
		return
	}
	if req.AlbumID == "" {
		http.Error(w, "albumId is required", http.StatusBadRequest)
		return
	}
	if len(req.ArtistIDs) == 0 {
		http.Error(w, "at least one artist ID is required", http.StatusBadRequest)
		return
	}

	// Check if album exists
	_, err := h.AlbumRepo.GetByID(r.Context(), req.AlbumID)
	if err != nil {
		http.Error(w, "album not found", http.StatusBadRequest)
		return
	}

	song := &model.Song{
		Name:         req.Name,
		Duration:     req.Duration,
		Genre:        req.Genre,
		AlbumID:      req.AlbumID,
		ArtistIDs:    req.ArtistIDs,
		AudioFileURL: req.AudioFileURL,
	}

	if err := h.Repo.Create(r.Context(), song); err != nil {
		http.Error(w, "failed to create song: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Log admin activity
	if h.Logger != nil {
		adminID := getAdminIDFromSongContext(r.Context())
		h.Logger.LogAdminActivity(adminID, "CREATE_SONG", "songs", map[string]interface{}{
			"songId":    song.ID,
			"name":     song.Name,
			"genre":    song.Genre,
			"albumId":  song.AlbumID,
			"artistIDs": song.ArtistIDs,
		})
	}

	// Get artist names for the event
	artistNames := make([]string, 0, len(song.ArtistIDs))
	for _, artistID := range song.ArtistIDs {
		artist, err := h.ArtistRepo.GetByID(r.Context(), artistID)
		if err == nil && artist != nil {
			artistNames = append(artistNames, artist.Name)
		}
	}

	// Emit event for new song (asynchronous)
	event := events.NewSongEvent{
		Type:        events.EventTypeNewSong,
		SongID:      song.ID,
		Name:        song.Name,
		Genre:       song.Genre,
		ArtistIDs:   song.ArtistIDs,
		ArtistNames: artistNames,
		AlbumID:     song.AlbumID,
	}
	events.EmitEvent(r.Context(), h.SubscriptionsServiceURL, event)
	// Also emit to recommendation-service
	events.EmitEvent(r.Context(), h.RecommendationServiceURL, map[string]interface{}{
		"type":      "song_created",
		"songId":    song.ID,
		"name":      song.Name,
		"genre":     song.Genre,
		"artistIds": song.ArtistIDs,
		"albumId":   song.AlbumID,
		"duration":  song.Duration,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(toSongResponse(song))
}

func (h *SongHandler) GetSong(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := extractSongID(r.URL.Path)
	if id == "" {
		http.Error(w, "song ID is required", http.StatusBadRequest)
		return
	}

	song, err := h.Repo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "song not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(toSongResponse(song))
}

func (h *SongHandler) GetAllSongs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	songs, err := h.Repo.GetAll(r.Context())
	if err != nil {
		http.Error(w, "failed to get songs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	responses := make([]*dto.SongResponse, len(songs))
	for i, song := range songs {
		responses[i] = toSongResponse(song)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responses)
}

func (h *SongHandler) GetSongsByAlbum(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	albumID := r.URL.Query().Get("albumId")
	if albumID == "" {
		http.Error(w, "albumId query parameter is required", http.StatusBadRequest)
		return
	}

	songs, err := h.Repo.GetByAlbumID(r.Context(), albumID)
	if err != nil {
		http.Error(w, "failed to get songs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	responses := make([]*dto.SongResponse, len(songs))
	for i, song := range songs {
		responses[i] = toSongResponse(song)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responses)
}

func (h *SongHandler) UpdateSong(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := extractSongID(r.URL.Path)
	if id == "" {
		http.Error(w, "song ID is required", http.StatusBadRequest)
		return
	}

	var req dto.UpdateSongRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	// Validation
	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	if req.Duration <= 0 {
		http.Error(w, "duration must be greater than 0", http.StatusBadRequest)
		return
	}
	if req.Genre == "" {
		http.Error(w, "genre is required", http.StatusBadRequest)
		return
	}
	if req.AlbumID == "" {
		http.Error(w, "albumId is required", http.StatusBadRequest)
		return
	}
	if len(req.ArtistIDs) == 0 {
		http.Error(w, "at least one artist ID is required", http.StatusBadRequest)
		return
	}

	// Check if album exists
	_, err := h.AlbumRepo.GetByID(r.Context(), req.AlbumID)
	if err != nil {
		http.Error(w, "album not found", http.StatusBadRequest)
		return
	}

	// Get existing song to preserve ID and timestamps
	existingSong, err := h.Repo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "song not found", http.StatusNotFound)
		return
	}

	// Store old state for logging
	oldState := map[string]interface{}{
		"name":        existingSong.Name,
		"duration":    existingSong.Duration,
		"genre":       existingSong.Genre,
		"albumID":     existingSong.AlbumID,
		"artistIDs":   existingSong.ArtistIDs,
		"audioFileURL": existingSong.AudioFileURL,
	}

	// Update fields
	existingSong.Name = req.Name
	existingSong.Duration = req.Duration
	existingSong.Genre = req.Genre
	existingSong.AlbumID = req.AlbumID
	existingSong.ArtistIDs = req.ArtistIDs
	existingSong.AudioFileURL = req.AudioFileURL

	newState := map[string]interface{}{
		"name":        existingSong.Name,
		"duration":    existingSong.Duration,
		"genre":       existingSong.Genre,
		"albumID":     existingSong.AlbumID,
		"artistIDs":   existingSong.ArtistIDs,
		"audioFileURL": existingSong.AudioFileURL,
	}

	if err := h.Repo.Update(r.Context(), id, existingSong); err != nil {
		http.Error(w, "failed to update song: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Log admin activity and state change
	if h.Logger != nil {
		adminID := getAdminIDFromSongContext(r.Context())
		h.Logger.LogAdminActivity(adminID, "UPDATE_SONG", "songs", map[string]interface{}{
			"songId": id,
		})
		h.Logger.LogStateChange("song", oldState, newState, adminID)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(toSongResponse(existingSong))
}

func (h *SongHandler) DeleteSong(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := extractSongID(r.URL.Path)
	if id == "" {
		http.Error(w, "song ID is required", http.StatusBadRequest)
		return
	}

	// Get song before deletion for logging
	song, _ := h.Repo.GetByID(r.Context(), id)

	// Use Saga pattern for distributed transaction (2.13)
	if h.SagaServiceURL != "" {
		// Call saga-service to orchestrate the deletion
		client := &http.Client{Timeout: 30 * time.Second}
		sagaURL := fmt.Sprintf("%s/sagas/delete-song", h.SagaServiceURL)
		
		reqBody := map[string]interface{}{
			"songId": id,
		}
		reqJSON, err := json.Marshal(reqBody)
		if err != nil {
			log.Printf("Error marshaling saga request: %v", err)
			http.Error(w, "failed to create saga request", http.StatusInternalServerError)
			return
		}

		req, err := http.NewRequestWithContext(r.Context(), "POST", sagaURL, bytes.NewBuffer(reqJSON))
		if err != nil {
			log.Printf("Error creating saga request: %v", err)
			http.Error(w, "failed to create saga request", http.StatusInternalServerError)
			return
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Error calling saga-service: %v", err)
			http.Error(w, "failed to execute saga transaction", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			var sagaResp map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&sagaResp); err == nil {
				if errorMsg, ok := sagaResp["error"].(string); ok {
					log.Printf("Saga transaction failed: %s", errorMsg)
					http.Error(w, fmt.Sprintf("Saga transaction failed: %s", errorMsg), http.StatusInternalServerError)
					return
				}
			}
			http.Error(w, "saga transaction failed", http.StatusInternalServerError)
			return
		}

		// Log admin activity
		if h.Logger != nil {
			adminID := getAdminIDFromSongContext(r.Context())
			h.Logger.LogAdminActivity(adminID, "DELETE_SONG", "songs", map[string]interface{}{
				"songId": id,
				"name":   song.Name,
				"method": "saga",
			})
		}

		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Fallback to old implementation if saga-service is not available
	log.Printf("Saga service not configured, using direct deletion")
	
	// Delete all ratings for this song from ratings-service
	if h.RatingsServiceURL != "" {
		go func() {
			client := &http.Client{Timeout: 5 * time.Second}
			deleteURL := fmt.Sprintf("%s/delete-ratings-by-song?songId=%s", h.RatingsServiceURL, id)
			req, err := http.NewRequest("DELETE", deleteURL, nil)
			if err == nil {
				resp, err := client.Do(req)
				if err != nil {
					log.Printf("Error calling ratings-service to delete ratings for song %s: %v", id, err)
				} else {
					resp.Body.Close()
					if resp.StatusCode == http.StatusOK {
						log.Printf("Successfully deleted all ratings for song %s", id)
					} else {
						log.Printf("Failed to delete ratings for song %s: status %d", id, resp.StatusCode)
					}
				}
			}
		}()
	}

	if err := h.Repo.Delete(r.Context(), id); err != nil {
		if err.Error() == "song not found" {
			http.Error(w, "song not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to delete song: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Emit deletion event to recommendation-service (asynchronous)
	if song != nil {
		events.EmitEvent(context.Background(), h.RecommendationServiceURL, events.DeletedSongEvent{
			Type:   events.EventTypeDeletedSong,
			SongID: id,
		})
	}

	// Log admin activity
	if h.Logger != nil {
		adminID := getAdminIDFromSongContext(r.Context())
		h.Logger.LogAdminActivity(adminID, "DELETE_SONG", "songs", map[string]interface{}{
			"songId": id,
			"name":   song.Name,
		})
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *SongHandler) StreamSong(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := extractSongID(r.URL.Path)
	if id == "" {
		http.Error(w, "song ID is required", http.StatusBadRequest)
		return
	}

	// Check if song exists
	song, err := h.Repo.GetByID(r.Context(), id)
	if err != nil {
		// If song not found, try to serve from HDFS directly by ID (fallback)
		hdfsPath := fmt.Sprintf("/audio/songs/%s.mp3", id)
		exists, err := h.HDFSClient.FileExists(hdfsPath)
		if err == nil && exists {
			// File exists on HDFS, serve it directly
			if h.Logger != nil {
				h.Logger.Log(logger.LevelInfo, logger.EventStateChange, "Serving audio from HDFS (song not in DB)", map[string]interface{}{
					"songId":   id,
					"hdfsPath": hdfsPath,
				})
			}
			audioData, err := h.HDFSClient.DownloadFile(hdfsPath)
			if err == nil {
				w.Header().Set("Content-Type", "audio/mpeg")
				w.Header().Set("Content-Length", fmt.Sprintf("%d", len(audioData)))
				w.Header().Set("Accept-Ranges", "bytes")
				w.Header().Set("Cache-Control", "no-cache")
				w.WriteHeader(http.StatusOK)
				w.Write(audioData)
				return
			}
			if h.Logger != nil {
				h.Logger.Log(logger.LevelError, logger.EventStateChange, "Failed to download audio from HDFS (fallback)", map[string]interface{}{
					"error":    err.Error(),
					"songId":   id,
					"hdfsPath": hdfsPath,
				})
			}
		} else {
			if h.Logger != nil {
				h.Logger.Log(logger.LevelError, logger.EventStateChange, "Song not found and HDFS file missing", map[string]interface{}{
					"songId":   id,
					"hdfsPath": hdfsPath,
					"exists":   exists,
					"error":    err,
				})
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Song not found and audio file not available on HDFS",
			"songId": id,
		})
		return
	}

	// Log activity if user is authenticated (1.15)
	// Try to get userID from JWT token (optional auth)
	claims, ok := r.Context().Value(middleware.UserContextKey).(*middleware.UserClaims)
	if ok && claims != nil && claims.UserID != "" {
		// Deduplication: Check if we already logged this activity recently (within 5 minutes)
		// This prevents multiple log entries for the same song play (e.g., from preload, range requests, etc.)
		activityKey := fmt.Sprintf("activity:%s:%s", claims.UserID, id)
		shouldLog := true
		
		if h.RedisCache != nil {
			ctx := r.Context()
			exists, err := h.RedisCache.Client().Exists(ctx, activityKey).Result()
			if err == nil && exists > 0 {
				// Activity already logged recently, skip
				shouldLog = false
				log.Printf("Skipping duplicate activity log for user %s, song %s (deduplication)", claims.UserID, id)
			}
		}
		
		if shouldLog {
			// Get artist names for the activity log
			artistNames := make(map[string]string)
			if len(song.ArtistIDs) > 0 && h.ArtistRepo != nil {
				for _, artistID := range song.ArtistIDs {
					if artistID != "" {
						artist, err := h.ArtistRepo.GetByID(r.Context(), artistID)
						if err == nil && artist != nil {
							artistNames[artistID] = artist.Name
						}
					}
				}
			}
			
			// Log activity with all available data
			activity := analytics.Activity{
				UserID:   claims.UserID,
				Type:     analytics.ActivityTypeSongPlayed,
				SongID:   id,
				SongName: song.Name,
				Genre:    song.Genre,
			}
			
			// Add first artist ID and name if available (for backward compatibility)
			if len(song.ArtistIDs) > 0 {
				activity.ArtistID = song.ArtistIDs[0]
				if name, ok := artistNames[song.ArtistIDs[0]]; ok {
					activity.ArtistName = name
				}
			}
			
			analytics.LogActivity(h.AnalyticsServiceURL, activity)
			
			// Set deduplication key in Redis (expires after 5 minutes to prevent duplicate processing)
			if h.RedisCache != nil {
				ctx := r.Context()
				h.RedisCache.Client().Set(ctx, activityKey, "1", 5*time.Minute)
			}
		}
	}

	// Increment play count in Redis cache (2.12)
	if h.RedisCache != nil {
		ctx := r.Context()
		if err := h.RedisCache.IncrementPlayCount(ctx, id); err != nil {
			log.Printf("Failed to increment play count for song %s: %v", id, err)
			// Don't fail the request if cache update fails
		}
	}

	// Set headers for audio streaming
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Cache-Control", "no-cache")

	if song.AudioFileURL != "" {
		// If song has an external URL, redirect to it
		if strings.HasPrefix(song.AudioFileURL, "http://") || strings.HasPrefix(song.AudioFileURL, "https://") {
			http.Redirect(w, r, song.AudioFileURL, http.StatusTemporaryRedirect)
			return
		}

		// Check if it's an HDFS path (starts with /audio/ or hdfs://)
		if strings.HasPrefix(song.AudioFileURL, "/audio/") || strings.HasPrefix(song.AudioFileURL, "hdfs://") {
			// Download from HDFS (2.11)
			hdfsPath := song.AudioFileURL
			if strings.HasPrefix(hdfsPath, "hdfs://") {
				// Extract path from hdfs://namenode:port/path
				parts := strings.SplitN(hdfsPath, "/", 4)
				if len(parts) >= 4 {
					hdfsPath = "/" + parts[3]
				}
			}

			// Check if file exists in HDFS before trying to download
			exists, err := h.HDFSClient.FileExists(hdfsPath)
			if err != nil || !exists {
				if h.Logger != nil {
					h.Logger.Log(logger.LevelError, logger.EventStateChange, "Audio file not found in HDFS", map[string]interface{}{
						"error":    err,
						"songId":   id,
						"hdfsPath": hdfsPath,
						"exists":   exists,
					})
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "Audio file not found in HDFS",
					"song":  song.Name,
					"path":  hdfsPath,
				})
				return
			}

			audioData, err := h.HDFSClient.DownloadFile(hdfsPath)
			if err != nil {
				if h.Logger != nil {
					h.Logger.Log(logger.LevelError, logger.EventStateChange, "Failed to download audio from HDFS", map[string]interface{}{
						"error":    err.Error(),
						"songId":   id,
						"hdfsPath": hdfsPath,
					})
				}
				http.Error(w, "failed to retrieve audio file: "+err.Error(), http.StatusInternalServerError)
				return
			}

			// Set content type based on file extension
			ext := strings.ToLower(filepath.Ext(hdfsPath))
			contentType := "audio/mpeg" // default
			switch ext {
			case ".mp3":
				contentType = "audio/mpeg"
			case ".wav":
				contentType = "audio/wav"
			case ".ogg":
				contentType = "audio/ogg"
			case ".m4a":
				contentType = "audio/mp4"
			case ".flac":
				contentType = "audio/flac"
			}
			w.Header().Set("Content-Type", contentType)
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(audioData)))

			// Stream the audio data
			w.WriteHeader(http.StatusOK)
			w.Write(audioData)
			return
		}

		// Check if it's a path to frontend/public/music files
		if strings.HasPrefix(song.AudioFileURL, "/music/") || strings.HasPrefix(song.AudioFileURL, "music/") {
			// Try to serve from /app/music directory (mounted from frontend/public/music)
			fileName := strings.TrimPrefix(song.AudioFileURL, "/music/")
			fileName = strings.TrimPrefix(fileName, "music/")
			musicPath := filepath.Join("/app/music", fileName)
			
			// Check if file exists
			if _, err := os.Stat(musicPath); err == nil {
				http.ServeFile(w, r, musicPath)
				return
			}
			
			// Fallback: redirect to frontend public folder
			frontendURL := fmt.Sprintf("http://localhost:3000/music/%s", fileName)
			http.Redirect(w, r, frontendURL, http.StatusTemporaryRedirect)
			return
		}

		// Fallback to local file (for backward compatibility)
		http.ServeFile(w, r, song.AudioFileURL)
		return
	}

	// If no audio file is specified, return a placeholder or error
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]string{
		"error": "No audio file available for this song",
		"song":  song.Name,
	})
}

// UploadAudio uploads an audio file to HDFS (2.11)
func (h *SongHandler) UploadAudio(w http.ResponseWriter, r *http.Request) {
	log.Printf("UploadAudio called: Method=%s, Path=%s, ContentType=%s", r.Method, r.URL.Path, r.Header.Get("Content-Type"))
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form (max 100MB)
	log.Printf("Parsing multipart form...")
	err := r.ParseMultipartForm(100 << 20) // 100 MB
	if err != nil {
		log.Printf("Error parsing multipart form: %v", err)
		http.Error(w, "failed to parse multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("Multipart form parsed successfully")

	// Get file from form
	log.Printf("Getting file from form...")
	file, header, err := r.FormFile("audio")
	if err != nil {
		log.Printf("Error getting file from form: %v", err)
		http.Error(w, "audio file is required: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()
	log.Printf("File retrieved: %s, size: %d", header.Filename, header.Size)

	// Get song ID from form or path
	songID := r.FormValue("songId")
	if songID == "" {
		// Try to extract from path
		path := strings.TrimPrefix(r.URL.Path, "/songs/")
		path = strings.TrimSuffix(path, "/upload")
		if path != "" {
			songID = path
		}
	}
	log.Printf("Song ID: %s", songID)

	if songID == "" {
		log.Printf("Song ID is empty")
		http.Error(w, "songId is required", http.StatusBadRequest)
		return
	}

	// Check if song exists
	log.Printf("Checking if song exists: %s", songID)
	song, err := h.Repo.GetByID(r.Context(), songID)
	if err != nil {
		log.Printf("Song not found: %v", err)
		http.Error(w, "song not found", http.StatusNotFound)
		return
	}
	log.Printf("Song found: %s", song.Name)

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	log.Printf("File extension: %s", ext)
	allowedExts := map[string]bool{
		".mp3":  true,
		".wav":  true,
		".ogg":  true,
		".m4a":  true,
		".flac": true,
	}
	if !allowedExts[ext] {
		log.Printf("File type not allowed: %s", ext)
		http.Error(w, "file type not allowed. Allowed: mp3, wav, ogg, m4a, flac", http.StatusBadRequest)
		return
	}

	// Read file data
	log.Printf("Reading file data...")
	fileData, err := io.ReadAll(file)
	if err != nil {
		log.Printf("Error reading file: %v", err)
		http.Error(w, "failed to read file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("File data read: %d bytes", len(fileData))

	// Upload to HDFS (2.11)
	hdfsPath := fmt.Sprintf("/audio/songs/%s%s", songID, ext)
	log.Printf("Uploading to HDFS: %s", hdfsPath)
	
	// Delay to ensure HDFS is ready (especially for newly created songs)
	// Check if this is a newly created song (no existing audioFileUrl or not HDFS path)
	isNewSong := song.AudioFileURL == "" || (!strings.HasPrefix(song.AudioFileURL, "/audio/") && !strings.HasPrefix(song.AudioFileURL, "hdfs://"))
	if isNewSong {
		log.Printf("New song or non-HDFS path detected, waiting longer for HDFS to be ready...")
		time.Sleep(2 * time.Second) // Wait 2 seconds for new songs
	} else {
		time.Sleep(1 * time.Second) // Wait 1 second for existing HDFS songs (to avoid connection issues)
	}
	
	err = h.HDFSClient.UploadData(fileData, hdfsPath)
	if err != nil {
		log.Printf("HDFS upload failed: %v", err)
		if h.Logger != nil {
			h.Logger.Log(logger.LevelError, logger.EventStateChange, "Failed to upload audio to HDFS", map[string]interface{}{
				"error":    err.Error(),
				"songId":   songID,
				"hdfsPath": hdfsPath,
			})
		}
		http.Error(w, "failed to upload file to HDFS: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("HDFS upload successful")

	// Update song with HDFS path
	song.AudioFileURL = hdfsPath
	log.Printf("Updating song with HDFS path...")
	err = h.Repo.Update(r.Context(), songID, song)
	if err != nil {
		log.Printf("Error updating song: %v", err)
		if h.Logger != nil {
			h.Logger.Log(logger.LevelError, logger.EventStateChange, "Failed to update song with HDFS path", map[string]interface{}{
				"error":  err.Error(),
				"songId": songID,
			})
		}
		// Don't fail the request, file is already uploaded
	} else {
		log.Printf("Song updated successfully")
	}

	// Log admin activity
	if h.Logger != nil {
		adminID := getAdminIDFromSongContext(r.Context())
		h.Logger.LogAdminActivity(adminID, "UPLOAD_AUDIO", "songs", map[string]interface{}{
			"songId":   songID,
			"hdfsPath": hdfsPath,
			"fileName": header.Filename,
			"fileSize": len(fileData),
		})
	}

	// Return success response
	log.Printf("Returning success response")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "Audio file uploaded successfully",
		"songId":   songID,
		"hdfsPath": hdfsPath,
		"fileName": header.Filename,
		"fileSize": len(fileData),
	})
	log.Printf("Upload completed successfully")
}

func toSongResponse(song *model.Song) *dto.SongResponse {
	return &dto.SongResponse{
		ID:           song.ID,
		Name:         song.Name,
		Duration:     song.Duration,
		Genre:        song.Genre,
		AlbumID:      song.AlbumID,
		ArtistIDs:    song.ArtistIDs,
		AudioFileURL: song.AudioFileURL,
		CreatedAt:    song.CreatedAt,
		UpdatedAt:    song.UpdatedAt,
	}
}
