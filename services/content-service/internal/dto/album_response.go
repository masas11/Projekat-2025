package dto

import "time"

type AlbumResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	ReleaseDate time.Time `json:"releaseDate"`
	Genre       string    `json:"genre"`
	ArtistIDs   []string  `json:"artistIds"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
