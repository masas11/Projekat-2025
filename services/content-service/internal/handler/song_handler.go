package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"content-service/internal/dto"
	"content-service/internal/model"
	"content-service/internal/store"
)

func extractSongID(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 2 && parts[0] == "songs" {
		return parts[1]
	}
	return ""
}

type SongHandler struct {
	Repo        *store.SongRepository
	AlbumRepo   *store.AlbumRepository
}

func NewSongHandler(repo *store.SongRepository, albumRepo *store.AlbumRepository) *SongHandler {
	return &SongHandler{
		Repo:      repo,
		AlbumRepo: albumRepo,
	}
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
		Name:      req.Name,
		Duration:  req.Duration,
		Genre:     req.Genre,
		AlbumID:   req.AlbumID,
		ArtistIDs: req.ArtistIDs,
	}

	if err := h.Repo.Create(r.Context(), song); err != nil {
		http.Error(w, "failed to create song: "+err.Error(), http.StatusInternalServerError)
		return
	}

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

	// Update fields
	existingSong.Name = req.Name
	existingSong.Duration = req.Duration
	existingSong.Genre = req.Genre
	existingSong.AlbumID = req.AlbumID
	existingSong.ArtistIDs = req.ArtistIDs

	if err := h.Repo.Update(r.Context(), id, existingSong); err != nil {
		http.Error(w, "failed to update song: "+err.Error(), http.StatusInternalServerError)
		return
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

	if err := h.Repo.Delete(r.Context(), id); err != nil {
		if err.Error() == "song not found" {
			http.Error(w, "song not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to delete song: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func toSongResponse(song *model.Song) *dto.SongResponse {
	return &dto.SongResponse{
		ID:        song.ID,
		Name:      song.Name,
		Duration:  song.Duration,
		Genre:     song.Genre,
		AlbumID:   song.AlbumID,
		ArtistIDs: song.ArtistIDs,
		CreatedAt: song.CreatedAt,
		UpdatedAt: song.UpdatedAt,
	}
}
