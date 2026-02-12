package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"content-service/internal/dto"
	"content-service/internal/events"
	"content-service/internal/logger"
	"content-service/internal/middleware"
	"content-service/internal/model"
	"content-service/internal/store"
)

// extractArtistID extracts the artist ID from the URL path
// Expected format: /artists/{id}
func extractArtistID(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 2 && parts[0] == "artists" {
		return parts[1]
	}
	return ""
}

type ArtistHandler struct {
	Repo                    *store.ArtistRepository
	SubscriptionsServiceURL string
	Logger                  *logger.Logger
}

func NewArtistHandler(repo *store.ArtistRepository, subscriptionsServiceURL string, log *logger.Logger) *ArtistHandler {
	return &ArtistHandler{
		Repo:                    repo,
		SubscriptionsServiceURL: subscriptionsServiceURL,
		Logger:                  log,
	}
}

// getAdminID extracts admin user ID from request context
func getAdminID(ctx context.Context) string {
	claims, ok := ctx.Value(middleware.UserContextKey).(*middleware.UserClaims)
	if !ok || claims == nil {
		return ""
	}
	return claims.UserID
}

func (h *ArtistHandler) CreateArtist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req dto.CreateArtistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	// Validation
	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	if req.Biography == "" {
		http.Error(w, "biography is required", http.StatusBadRequest)
		return
	}
	if len(req.Genres) == 0 {
		http.Error(w, "at least one genre is required", http.StatusBadRequest)
		return
	}

	artist := &model.Artist{
		Name:      req.Name,
		Biography: req.Biography,
		Genres:    req.Genres,
	}

	if err := h.Repo.Create(r.Context(), artist); err != nil {
		http.Error(w, "failed to create artist: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Log admin activity
	if h.Logger != nil {
		adminID := getAdminID(r.Context())
		h.Logger.LogAdminActivity(adminID, "CREATE_ARTIST", "artists", map[string]interface{}{
			"artistId": artist.ID,
			"name":     artist.Name,
			"genres":   artist.Genres,
		})
	}

	// Emit event for new artist (asynchronous)
	event := events.NewArtistEvent{
		Type:     events.EventTypeNewArtist,
		ArtistID: artist.ID,
		Name:     artist.Name,
		Genres:   artist.Genres,
	}
	events.EmitEvent(h.SubscriptionsServiceURL, event)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto.ToArtistResponse(artist))
}

func (h *ArtistHandler) UpdateArtist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract artist ID from URL path
	id := extractArtistID(r.URL.Path)
	if id == "" {
		http.Error(w, "artist ID is required", http.StatusBadRequest)
		return
	}

	var req dto.UpdateArtistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	// Validation
	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	if req.Biography == "" {
		http.Error(w, "biography is required", http.StatusBadRequest)
		return
	}
	if len(req.Genres) == 0 {
		http.Error(w, "at least one genre is required", http.StatusBadRequest)
		return
	}

	// Get existing artist to preserve ID and timestamps
	existingArtist, err := h.Repo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "artist not found", http.StatusNotFound)
		return
	}

	// Store old state for logging
	oldState := map[string]interface{}{
		"name":      existingArtist.Name,
		"biography": existingArtist.Biography,
		"genres":    existingArtist.Genres,
	}

	// Update fields
	existingArtist.Name = req.Name
	existingArtist.Biography = req.Biography
	existingArtist.Genres = req.Genres

	newState := map[string]interface{}{
		"name":      existingArtist.Name,
		"biography": existingArtist.Biography,
		"genres":    existingArtist.Genres,
	}

	if err := h.Repo.Update(r.Context(), id, existingArtist); err != nil {
		http.Error(w, "failed to update artist: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Log admin activity and state change
	if h.Logger != nil {
		adminID := getAdminID(r.Context())
		h.Logger.LogAdminActivity(adminID, "UPDATE_ARTIST", "artists", map[string]interface{}{
			"artistId": id,
		})
		// Log state change if there are unexpected changes
		h.Logger.LogStateChange("artist", oldState, newState, adminID)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(dto.ToArtistResponse(existingArtist))
}

func (h *ArtistHandler) GetArtist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := extractArtistID(r.URL.Path)
	if id == "" {
		http.Error(w, "artist ID is required", http.StatusBadRequest)
		return
	}

	artist, err := h.Repo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "artist not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(dto.ToArtistResponse(artist))
}

func (h *ArtistHandler) GetAllArtists(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	artists, err := h.Repo.GetAll(r.Context())
	if err != nil {
		http.Error(w, "failed to get artists: "+err.Error(), http.StatusInternalServerError)
		return
	}

	responses := make([]*dto.ArtistResponse, len(artists))
	for i, artist := range artists {
		responses[i] = dto.ToArtistResponse(artist)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responses)
}

func (h *ArtistHandler) DeleteArtist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := extractArtistID(r.URL.Path)
	if id == "" {
		http.Error(w, "artist ID is required", http.StatusBadRequest)
		return
	}

	// Get artist before deletion for logging
	artist, _ := h.Repo.GetByID(r.Context(), id)

	if err := h.Repo.Delete(r.Context(), id); err != nil {
		if err.Error() == "artist not found" {
			http.Error(w, "artist not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to delete artist: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Log admin activity
	if h.Logger != nil {
		adminID := getAdminID(r.Context())
		h.Logger.LogAdminActivity(adminID, "DELETE_ARTIST", "artists", map[string]interface{}{
			"artistId": id,
			"name":     artist.Name,
		})
	}

	w.WriteHeader(http.StatusNoContent)
}