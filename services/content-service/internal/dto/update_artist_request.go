package dto

type UpdateArtistRequest struct {
	Name      string   `json:"name"`
	Biography string   `json:"biography"`
	Genres    []string `json:"genres"`
}
