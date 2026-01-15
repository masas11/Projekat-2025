package model

type Song struct {
	ID        string   `json:"id" bson:"_id"`
	Name      string   `json:"name" bson:"name"`
	Duration  int      `json:"duration" bson:"duration"` // duration in seconds
	Genre     string   `json:"genre" bson:"genre"`
	AlbumID   string   `json:"albumId" bson:"albumId"`
	ArtistIDs []string `json:"artistIds" bson:"artistIds"`
	CreatedAt string   `json:"createdAt" bson:"createdAt"`
	UpdatedAt string   `json:"updatedAt" bson:"updatedAt"`
}
