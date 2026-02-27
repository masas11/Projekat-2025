package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"content-service/internal/cache"
	"content-service/internal/store"
)

type MostPlayedHandler struct {
	SongRepo  *store.SongRepository
	Cache     *cache.RedisCache
}

func NewMostPlayedHandler(songRepo *store.SongRepository, redisCache *cache.RedisCache) *MostPlayedHandler {
	return &MostPlayedHandler{
		SongRepo: songRepo,
		Cache:    redisCache,
	}
}

// GetMostPlayedSongs returns the most played songs with full song details
func (h *MostPlayedHandler) GetMostPlayedSongs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get limit from query params (default 10)
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get most played song IDs from cache
	mostPlayed, err := h.Cache.GetMostPlayedSongs(ctx, limit)
	if err != nil {
		log.Printf("Error getting most played songs from cache: %v", err)
		http.Error(w, "failed to get most played songs", http.StatusInternalServerError)
		return
	}

	if len(mostPlayed) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}

	// Fetch full song details from MongoDB
	type SongWithPlayCount struct {
		ID           string   `json:"id"`
		Name         string   `json:"name"`
		Duration     int      `json:"duration"`
		Genre        string   `json:"genre"`
		AlbumID      string   `json:"albumId"`
		ArtistIDs    []string `json:"artistIds"`
		AudioFileURL string   `json:"audioFileUrl,omitempty"`
		PlayCount    int      `json:"playCount"`
	}

	result := make([]SongWithPlayCount, 0, len(mostPlayed))
	for _, mp := range mostPlayed {
		song, err := h.SongRepo.GetByID(ctx, mp.SongID)
		if err != nil {
			log.Printf("Error fetching song %s: %v", mp.SongID, err)
			continue // Skip songs that don't exist anymore
		}

		result = append(result, SongWithPlayCount{
			ID:           song.ID,
			Name:         song.Name,
			Duration:     song.Duration,
			Genre:        song.Genre,
			AlbumID:      song.AlbumID,
			ArtistIDs:    song.ArtistIDs,
			AudioFileURL: song.AudioFileURL,
			PlayCount:    mp.Count,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}
