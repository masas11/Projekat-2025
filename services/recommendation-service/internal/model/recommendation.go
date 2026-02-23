package model

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
	Reason    string   `json:"reason"`
}
