package model

import "time"

type Notification struct {
	ID        string    `json:"id" bson:"_id"`
	UserID    string    `json:"userId" bson:"userId"`
	Type      string    `json:"type" bson:"type"` // "new_album", "new_song", "new_artist"
	Message   string    `json:"message" bson:"message"`
	ContentID string    `json:"contentId" bson:"contentId"` // ID of album, song, or artist
	Read      bool      `json:"read" bson:"read"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
}
