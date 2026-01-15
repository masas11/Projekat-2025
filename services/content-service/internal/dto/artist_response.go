package dto

import "content-service/internal/model"

type ArtistResponse struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Biography string   `json:"biography"`
	Genres    []string `json:"genres"`
}

func ToArtistResponse(artist *model.Artist) *ArtistResponse {
	return &ArtistResponse{
		ID:        artist.ID,
		Name:      artist.Name,
		Biography: artist.Biography,
		Genres:    artist.Genres,
	}
}
