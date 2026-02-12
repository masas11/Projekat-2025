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

func extractAlbumID(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 2 && parts[0] == "albums" {
		return parts[1]
	}
	return ""
}

type AlbumHandler struct {
	Repo                    *store.AlbumRepository
	SubscriptionsServiceURL string
	Logger                  *logger.Logger
}

func NewAlbumHandler(repo *store.AlbumRepository, subscriptionsServiceURL string, log *logger.Logger) *AlbumHandler {
	return &AlbumHandler{
		Repo:                    repo,
		SubscriptionsServiceURL: subscriptionsServiceURL,
		Logger:                  log,
	}
}

// getAdminID extracts admin user ID from request context
func getAdminIDFromContext(ctx context.Context) string {
	claims, ok := ctx.Value(middleware.UserContextKey).(*middleware.UserClaims)
	if !ok || claims == nil {
		return ""
	}
	return claims.UserID
}

func (h *AlbumHandler) CreateAlbum(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req dto.CreateAlbumRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	// Validation
	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	if req.Genre == "" {
		http.Error(w, "genre is required", http.StatusBadRequest)
		return
	}
	if len(req.ArtistIDs) == 0 {
		http.Error(w, "at least one artist ID is required", http.StatusBadRequest)
		return
	}

	album := &model.Album{
		Name:        req.Name,
		ReleaseDate: req.ReleaseDate,
		Genre:       req.Genre,
		ArtistIDs:   req.ArtistIDs,
	}

	if err := h.Repo.Create(r.Context(), album); err != nil {
		http.Error(w, "failed to create album: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Log admin activity
	if h.Logger != nil {
		adminID := getAdminIDFromContext(r.Context())
		h.Logger.LogAdminActivity(adminID, "CREATE_ALBUM", "albums", map[string]interface{}{
			"albumId":   album.ID,
			"name":      album.Name,
			"genre":     album.Genre,
			"artistIDs": album.ArtistIDs,
		})
	}

	// Emit event for new album (asynchronous)
	event := events.NewAlbumEvent{
		Type:      events.EventTypeNewAlbum,
		AlbumID:   album.ID,
		Name:      album.Name,
		Genre:     album.Genre,
		ArtistIDs: album.ArtistIDs,
	}
	events.EmitEvent(h.SubscriptionsServiceURL, event)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(toAlbumResponse(album))
}

func (h *AlbumHandler) GetAlbum(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := extractAlbumID(r.URL.Path)
	if id == "" {
		http.Error(w, "album ID is required", http.StatusBadRequest)
		return
	}

	album, err := h.Repo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "album not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(toAlbumResponse(album))
}

func (h *AlbumHandler) GetAllAlbums(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	albums, err := h.Repo.GetAll(r.Context())
	if err != nil {
		http.Error(w, "failed to get albums: "+err.Error(), http.StatusInternalServerError)
		return
	}

	responses := make([]*dto.AlbumResponse, len(albums))
	for i, album := range albums {
		responses[i] = toAlbumResponse(album)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responses)
}

func (h *AlbumHandler) GetAlbumsByArtist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	artistID := r.URL.Query().Get("artistId")
	if artistID == "" {
		http.Error(w, "artistId query parameter is required", http.StatusBadRequest)
		return
	}

	albums, err := h.Repo.GetByArtistID(r.Context(), artistID)
	if err != nil {
		http.Error(w, "failed to get albums: "+err.Error(), http.StatusInternalServerError)
		return
	}

	responses := make([]*dto.AlbumResponse, len(albums))
	for i, album := range albums {
		responses[i] = toAlbumResponse(album)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responses)
}

func (h *AlbumHandler) UpdateAlbum(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := extractAlbumID(r.URL.Path)
	if id == "" {
		http.Error(w, "album ID is required", http.StatusBadRequest)
		return
	}

	var req dto.UpdateAlbumRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	// Validation
	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	if req.Genre == "" {
		http.Error(w, "genre is required", http.StatusBadRequest)
		return
	}
	if len(req.ArtistIDs) == 0 {
		http.Error(w, "at least one artist ID is required", http.StatusBadRequest)
		return
	}

	// Get existing album to preserve ID and timestamps
	existingAlbum, err := h.Repo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "album not found", http.StatusNotFound)
		return
	}

	// Store old state for logging
	oldState := map[string]interface{}{
		"name":        existingAlbum.Name,
		"releaseDate": existingAlbum.ReleaseDate,
		"genre":       existingAlbum.Genre,
		"artistIDs":   existingAlbum.ArtistIDs,
	}

	// Update fields
	existingAlbum.Name = req.Name
	existingAlbum.ReleaseDate = req.ReleaseDate
	existingAlbum.Genre = req.Genre
	existingAlbum.ArtistIDs = req.ArtistIDs

	newState := map[string]interface{}{
		"name":        existingAlbum.Name,
		"releaseDate": existingAlbum.ReleaseDate,
		"genre":       existingAlbum.Genre,
		"artistIDs":   existingAlbum.ArtistIDs,
	}

	if err := h.Repo.Update(r.Context(), id, existingAlbum); err != nil {
		http.Error(w, "failed to update album: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Log admin activity and state change
	if h.Logger != nil {
		adminID := getAdminIDFromContext(r.Context())
		h.Logger.LogAdminActivity(adminID, "UPDATE_ALBUM", "albums", map[string]interface{}{
			"albumId": id,
		})
		h.Logger.LogStateChange("album", oldState, newState, adminID)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(toAlbumResponse(existingAlbum))
}

func (h *AlbumHandler) DeleteAlbum(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := extractAlbumID(r.URL.Path)
	if id == "" {
		http.Error(w, "album ID is required", http.StatusBadRequest)
		return
	}

	// Get album before deletion for logging
	album, _ := h.Repo.GetByID(r.Context(), id)

	if err := h.Repo.Delete(r.Context(), id); err != nil {
		if err.Error() == "album not found" {
			http.Error(w, "album not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to delete album: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Log admin activity
	if h.Logger != nil {
		adminID := getAdminIDFromContext(r.Context())
		h.Logger.LogAdminActivity(adminID, "DELETE_ALBUM", "albums", map[string]interface{}{
			"albumId": id,
			"name":    album.Name,
		})
	}

	w.WriteHeader(http.StatusNoContent)
}

func toAlbumResponse(album *model.Album) *dto.AlbumResponse {
	return &dto.AlbumResponse{
		ID:          album.ID,
		Name:        album.Name,
		ReleaseDate: album.ReleaseDate,
		Genre:       album.Genre,
		ArtistIDs:   album.ArtistIDs,
		CreatedAt:   album.CreatedAt,
		UpdatedAt:   album.UpdatedAt,
	}
}
