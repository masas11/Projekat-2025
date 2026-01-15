package model

import "time"

type Album struct {
	ID        string    `json:"id" bson:"_id"`
	Name      string    `json:"name" bson:"name"`
	ReleaseDate time.Time `json:"releaseDate" bson:"releaseDate"`
	Genre     string    `json:"genre" bson:"genre"`
	ArtistIDs []string  `json:"artistIds" bson:"artistIds"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
}
