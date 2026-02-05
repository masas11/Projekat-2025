package model

import "time"

type Song struct {
	ID           string    `json:"id" bson:"_id"`
	Name         string    `json:"name" bson:"name"`
	Duration     int       `json:"duration" bson:"duration"` // duration in seconds
	Genre        string    `json:"genre" bson:"genre"`
	AlbumID      string    `json:"albumId" bson:"albumId"`
	ArtistIDs    []string  `json:"artistIds" bson:"artistIds"`
	AudioFileURL string    `json:"audioFileUrl" bson:"audioFileUrl,omitempty"` // path or URL to audio file
	CreatedAt    time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt" bson:"updatedAt"`
}
