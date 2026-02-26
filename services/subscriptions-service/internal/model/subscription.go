package model

import "time"

type Subscription struct {
	ID         string    `json:"id" bson:"_id"`
	UserID     string    `json:"userId" bson:"userId"`
	Type       string    `json:"type" bson:"type"` // "artist" or "genre"
	ArtistID   string    `json:"artistId,omitempty" bson:"artistId,omitempty"`
	ArtistName string    `json:"artistName" bson:"artistName,omitempty"` // CQRS: denormalized data
	Genre      string    `json:"genre,omitempty" bson:"genre,omitempty"`
	CreatedAt  time.Time `json:"createdAt" bson:"createdAt"`
}
