package model

import "time"

type RecommendationRequest struct {
	UserID string `json:"userId"`
}

type RecommendationResponse struct {
	SubscribedGenreSongs []*SongRecommendation `json:"subscribedGenreSongs"`
	TopRatedSong         *SongRecommendation   `json:"topRatedSong"`
}

type SongRecommendation struct {
	SongID    string   `json:"songId"`
	Name      string   `json:"name"`
	Genre     string   `json:"genre"`
	ArtistIDs []string `json:"artistIds"`
	AlbumID   string   `json:"albumId"`
	Duration  int      `json:"duration"`
	Reason    string   `json:"reason"` // Why this song was recommended
}

type Song struct {
	ID        string    `json:"id" bson:"_id"`
	Name      string    `json:"name" bson:"name"`
	Duration  int       `json:"duration" bson:"duration"`
	Genre     string    `json:"genre" bson:"genre"`
	AlbumID   string    `json:"albumId" bson:"albumId"`
	ArtistIDs []string  `json:"artistIds" bson:"artistIds"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
}
