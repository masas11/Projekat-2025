package store

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"ratings-service/internal/model"
)

type RecommendationStore struct {
	ratingCollection       *mongo.Collection
	songCollection         *mongo.Collection
	subscriptionCollection *mongo.Collection
}

func NewRecommendationStore(ratingsDB, contentDB, subscriptionsDB *mongo.Database) *RecommendationStore {
	return &RecommendationStore{
		ratingCollection:       ratingsDB.Collection("ratings"),
		songCollection:         contentDB.Collection("songs"),
		subscriptionCollection: subscriptionsDB.Collection("subscriptions"),
	}
}

// GetSongsByUserSubscribedGenres returns songs that belong to genres the user is subscribed to
// and the user hasn't rated below 4
func (s *RecommendationStore) GetSongsByUserSubscribedGenres(ctx context.Context, userID string) ([]*model.SongRecommendation, error) {
	log.Printf("Getting recommendations for user: %s", userID)

	// Get user's genre subscriptions
	pipeline := bson.A{
		bson.M{"$match": bson.M{"userId": userID, "type": "genre"}},
		bson.M{"$group": bson.M{"_id": nil, "genres": bson.M{"$addToSet": "$genre"}}},
	}

	cursor, err := s.subscriptionCollection.Aggregate(ctx, pipeline)
	if err != nil {
		log.Printf("Error in subscription aggregation: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var result struct {
		Genres []string `bson:"genres"`
	}
	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			log.Printf("Error decoding subscription result: %v", err)
			return nil, err
		}
	}

	log.Printf("User subscribed genres: %v", result.Genres)

	if len(result.Genres) == 0 {
		log.Printf("No subscribed genres found for user: %s", userID)
		return []*model.SongRecommendation{}, nil
	}

	// Get user's ratings to filter out songs rated below 4
	ratingCursor, err := s.ratingCollection.Find(ctx, bson.M{
		"userId": userID,
		"rating": bson.M{"$lt": 4},
	})
	if err != nil {
		return nil, err
	}
	defer ratingCursor.Close(ctx)

	lowRatedSongIDs := make(map[string]bool)
	for ratingCursor.Next(ctx) {
		var rating model.Rating
		if err := ratingCursor.Decode(&rating); err != nil {
			continue
		}
		lowRatedSongIDs[rating.SongID] = true
	}

	// Get songs in subscribed genres that user hasn't rated low
	songFilter := bson.M{
		"genre": bson.M{"$in": result.Genres},
	}
	if len(lowRatedSongIDs) > 0 {
		songFilter["_id"] = bson.M{"$nin": getKeys(lowRatedSongIDs)}
	}

	songCursor, err := s.songCollection.Find(ctx, songFilter)
	if err != nil {
		return nil, err
	}
	defer songCursor.Close(ctx)

	var recommendations []*model.SongRecommendation
	for songCursor.Next(ctx) {
		var song model.Song
		if err := songCursor.Decode(&song); err != nil {
			continue
		}

		recommendations = append(recommendations, &model.SongRecommendation{
			SongID:    song.ID,
			Name:      song.Name,
			Genre:     song.Genre,
			ArtistIDs: song.ArtistIDs,
			AlbumID:   song.AlbumID,
			Duration:  song.Duration,
			Reason:    "Based on your genre subscriptions",
		})
	}

	return recommendations, nil
}

// GetTopRatedSongFromUnsubscribedGenre returns the song with most 5-star ratings
// from a genre the user is not subscribed to
func (s *RecommendationStore) GetTopRatedSongFromUnsubscribedGenre(ctx context.Context, userID string) (*model.SongRecommendation, error) {
	// Get user's subscribed genres
	pipeline := bson.A{
		bson.M{"$match": bson.M{"userId": userID, "type": "genre"}},
		bson.M{"$group": bson.M{"_id": nil, "genres": bson.M{"$addToSet": "$genre"}}},
	}

	cursor, err := s.subscriptionCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var result struct {
		Genres []string `bson:"genres"`
	}
	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
	}

	// Build filter for genres user is NOT subscribed to
	genreFilter := bson.M{}
	if len(result.Genres) > 0 {
		genreFilter["genre"] = bson.M{"$nin": result.Genres}
	}

	// Aggregate to find song with most 5-star ratings from unsubscribed genres
	pipeline = bson.A{
		bson.M{"$match": genreFilter},
		bson.M{"$lookup": bson.M{
			"from":         "ratings",
			"localField":   "_id",
			"foreignField": "songId",
			"as":           "ratings",
		}},
		bson.M{"$project": bson.M{
			"name":      1,
			"genre":     1,
			"albumId":   1,
			"artistIds": 1,
			"duration":  1,
			"fiveStarCount": bson.M{
				"$size": bson.M{
					"$filter": bson.M{
						"input": "$ratings",
						"cond":  bson.M{"$eq": bson.A{"$$this.rating", 5}},
					},
				},
			},
		}},
		bson.M{"$sort": bson.M{"fiveStarCount": -1}},
		bson.M{"$limit": 1},
	}

	cursor, err = s.songCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if cursor.Next(ctx) {
		var song struct {
			ID            string   `bson:"_id"`
			Name          string   `bson:"name"`
			Genre         string   `bson:"genre"`
			AlbumID       string   `bson:"albumId"`
			ArtistIDs     []string `bson:"artistIds"`
			Duration      int      `bson:"duration"`
			FiveStarCount int      `bson:"fiveStarCount"`
		}
		if err := cursor.Decode(&song); err != nil {
			return nil, err
		}

		return &model.SongRecommendation{
			SongID:    song.ID,
			Name:      song.Name,
			Genre:     song.Genre,
			ArtistIDs: song.ArtistIDs,
			AlbumID:   song.AlbumID,
			Duration:  song.Duration,
			Reason:    "Popular in genre you might like",
		}, nil
	}

	return nil, nil // No songs found
}

func getKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
