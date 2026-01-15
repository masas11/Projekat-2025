package model

import "time"

type Artist struct {
	ID        string    `json:"id" bson:"_id"`
	Name      string    `json:"name" bson:"name"`
	Biography string    `json:"biography" bson:"biography"`
	Genres    []string  `json:"genres" bson:"genres"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
}
