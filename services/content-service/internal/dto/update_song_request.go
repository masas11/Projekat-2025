package dto

type UpdateSongRequest struct {
	Name      string   `json:"name"`
	Duration  int      `json:"duration"` // duration in seconds
	Genre     string   `json:"genre"`
	AlbumID   string   `json:"albumId"`
	ArtistIDs []string `json:"artistIds"`
}
