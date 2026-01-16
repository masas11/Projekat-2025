package dto

import "time"

type UpdateAlbumRequest struct {
	Name        string    `json:"name"`
	ReleaseDate time.Time `json:"releaseDate"`
	Genre       string    `json:"genre"`
	ArtistIDs   []string  `json:"artistIds"`
}
