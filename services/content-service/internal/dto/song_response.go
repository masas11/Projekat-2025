package dto

type SongResponse struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Duration  int      `json:"duration"`
	Genre     string   `json:"genre"`
	AlbumID   string   `json:"albumId"`
	ArtistIDs []string `json:"artistIds"`
	CreatedAt string   `json:"createdAt"`
	UpdatedAt string   `json:"updatedAt"`
}
