package model

import "time"

type Rating struct {
	ID        string    `json:"id" bson:"_id"`
	SongID    string    `json:"songId" bson:"songId"`
	UserID    string    `json:"userId" bson:"userId"`
	Rating    int       `json:"rating" bson:"rating"` // 1-5
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
}
