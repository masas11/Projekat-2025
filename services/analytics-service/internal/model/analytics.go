package model

// UserAnalytics represents analytics data for a user
type UserAnalytics struct {
	UserID                    string             `json:"userId"`
	TotalSongsPlayed          int                `json:"totalSongsPlayed"`
	AverageRating             float64            `json:"averageRating"`
	SongsPlayedByGenre        map[string]int     `json:"songsPlayedByGenre"`
	Top5Artists               []ArtistPlayCount  `json:"top5Artists"`
	SubscribedArtistsCount    int                `json:"subscribedArtistsCount"`
}

// ArtistPlayCount represents an artist with play count
type ArtistPlayCount struct {
	ArtistID   string `json:"artistId"`
	ArtistName string `json:"artistName"`
	PlayCount  int    `json:"playCount"`
}
