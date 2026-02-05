package model

import "time"

type Notification struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	Type      string    `json:"type"` // "new_album", "new_song", "new_artist"
	Message   string    `json:"message"`
	ContentID string    `json:"contentId"` // ID of album, song, or artist
	Read      bool      `json:"read"`
	CreatedAt time.Time `json:"createdAt"`
}
