package store

import (
	"context"
	"log"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"recommendation-service/internal/model"
)

// AddOrUpdateUser creates or updates a user node
func (s *Neo4jStore) AddOrUpdateUser(ctx context.Context, userID string) error {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	query := `
		MERGE (u:User {id: $userID})
		SET u.updatedAt = datetime()
		RETURN u
	`

	_, err := session.Run(ctx, query, map[string]interface{}{
		"userID": userID,
	})
	return err
}

// AddOrUpdateArtist creates or updates an artist node
func (s *Neo4jStore) AddOrUpdateArtist(ctx context.Context, artistID, artistName string, genres []string) error {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	query := `
		MERGE (a:Artist {id: $artistID})
		SET a.name = $artistName, a.updatedAt = datetime()
		WITH a
		UNWIND $genres AS genreName
		MERGE (g:Genre {name: genreName})
		MERGE (a)-[:PERFORMS_IN]->(g)
		RETURN a
	`

	_, err := session.Run(ctx, query, map[string]interface{}{
		"artistID":   artistID,
		"artistName": artistName,
		"genres":     genres,
	})
	return err
}

// AddOrUpdateSong creates or updates a song node and connects it to genre
func (s *Neo4jStore) AddOrUpdateSong(ctx context.Context, songID, songName, genre string, artistIDs []string, albumID string, duration int) error {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	query := `
		MERGE (s:Song {id: $songID})
		SET s.name = $songName, s.albumId = $albumID, s.duration = $duration, s.updatedAt = datetime()
		WITH s
		MERGE (g:Genre {name: $genre})
		MERGE (s)-[:BELONGS_TO]->(g)
		WITH s
		UNWIND $artistIDs AS artistID
		MERGE (a:Artist {id: artistID})
		MERGE (s)-[:PERFORMED_BY]->(a)
		RETURN s
	`

	_, err := session.Run(ctx, query, map[string]interface{}{
		"songID":    songID,
		"songName":  songName,
		"genre":     genre,
		"artistIDs": artistIDs,
		"albumID":   albumID,
		"duration":  duration,
	})
	return err
}

// AddRating creates or updates a RATED relationship between user and song
func (s *Neo4jStore) AddRating(ctx context.Context, userID, songID string, rating int) error {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	query := `
		MATCH (u:User {id: $userID})
		MATCH (s:Song {id: $songID})
		MERGE (u)-[r:RATED]->(s)
		SET r.rating = $rating, r.updatedAt = datetime()
		RETURN r
	`

	_, err := session.Run(ctx, query, map[string]interface{}{
		"userID": userID,
		"songID": songID,
		"rating": rating,
	})
	return err
}

// AddSubscription creates a SUBSCRIBED_TO relationship between user and genre
func (s *Neo4jStore) AddSubscription(ctx context.Context, userID, genre string) error {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	query := `
		MATCH (u:User {id: $userID})
		MERGE (g:Genre {name: $genre})
		MERGE (u)-[r:SUBSCRIBED_TO]->(g)
		SET r.createdAt = datetime()
		RETURN r
	`

	_, err := session.Run(ctx, query, map[string]interface{}{
		"userID": userID,
		"genre":  genre,
	})
	return err
}

// RemoveSubscription removes a SUBSCRIBED_TO relationship
func (s *Neo4jStore) RemoveSubscription(ctx context.Context, userID, genre string) error {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	query := `
		MATCH (u:User {id: $userID})-[r:SUBSCRIBED_TO]->(g:Genre {name: $genre})
		DELETE r
	`

	_, err := session.Run(ctx, query, map[string]interface{}{
		"userID": userID,
		"genre":  genre,
	})
	return err
}

// GetSubscribedGenreSongs returns songs from genres user is subscribed to,
// excluding songs rated below 4
func (s *Neo4jStore) GetSubscribedGenreSongs(ctx context.Context, userID string) ([]*model.SongRecommendation, error) {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	// First, check if user has subscriptions
	checkQuery := `
		MATCH (u:User {id: $userID})-[:SUBSCRIBED_TO]->(g:Genre)
		RETURN g.name AS genre
	`
	checkResult, err := session.Run(ctx, checkQuery, map[string]interface{}{
		"userID": userID,
	})
	if err == nil {
		var subscribedGenres []string
		for checkResult.Next(ctx) {
			record := checkResult.Record()
			if genre, ok := record.Get("genre"); ok && genre != nil {
				subscribedGenres = append(subscribedGenres, genre.(string))
			}
		}
		if len(subscribedGenres) == 0 {
			log.Printf("User %s has no subscriptions in Neo4j", userID)
			return []*model.SongRecommendation{}, nil
		}
		log.Printf("User %s is subscribed to genres: %v", userID, subscribedGenres)
	}

	query := `
		MATCH (u:User {id: $userID})-[:SUBSCRIBED_TO]->(g:Genre)<-[:BELONGS_TO]-(s:Song)
		OPTIONAL MATCH (u)-[r:RATED]->(s)
		WHERE r IS NULL OR r.rating >= 4
		OPTIONAL MATCH (s)-[:PERFORMED_BY]->(a:Artist)
		WITH DISTINCT s, g, collect(a.id) AS artistIds
		RETURN s.id AS songId, s.name AS name, g.name AS genre, 
		       s.albumId AS albumId, s.duration AS duration, artistIds
		LIMIT 50
	`

	result, err := session.Run(ctx, query, map[string]interface{}{
		"userID": userID,
	})
	if err != nil {
		return nil, err
	}

	var recommendations []*model.SongRecommendation
	for result.Next(ctx) {
		record := result.Record()
		songID, _ := record.Get("songId")
		name, _ := record.Get("name")
		genre, _ := record.Get("genre")
		albumID, _ := record.Get("albumId")
		duration, _ := record.Get("duration")
		artistIDsInterface, _ := record.Get("artistIds")

		artistIDs := make([]string, 0)
		if artistIDsInterface != nil {
			if ids, ok := artistIDsInterface.([]interface{}); ok {
				for _, id := range ids {
					if idStr, ok := id.(string); ok {
						artistIDs = append(artistIDs, idStr)
					}
				}
			}
		}

		albumIDStr := ""
		if albumID != nil {
			albumIDStr = albumID.(string)
		}

		durationInt := 0
		if duration != nil {
			if d, ok := duration.(int64); ok {
				durationInt = int(d)
			}
		}

		recommendations = append(recommendations, &model.SongRecommendation{
			SongID:    songID.(string),
			Name:      name.(string),
			Genre:     genre.(string),
			ArtistIDs: artistIDs,
			AlbumID:   albumIDStr,
			Duration:  durationInt,
			Reason:    "Based on your genre subscriptions",
		})
	}

	if err := result.Err(); err != nil {
		return nil, err
	}

	return recommendations, nil
}

// GetTopRatedSongFromUnsubscribedGenre returns the song with most 5-star ratings
// from a genre the user is not subscribed to
func (s *Neo4jStore) GetTopRatedSongFromUnsubscribedGenre(ctx context.Context, userID string) (*model.SongRecommendation, error) {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	query := `
		OPTIONAL MATCH (u:User {id: $userID})-[:SUBSCRIBED_TO]->(subscribedGenre:Genre)
		WITH collect(DISTINCT subscribedGenre.name) AS subscribedGenres
		MATCH (s:Song)-[:BELONGS_TO]->(g:Genre)
		WHERE (size(subscribedGenres) = 0 OR NOT g.name IN subscribedGenres)
		OPTIONAL MATCH (otherUser:User)-[r:RATED {rating: 5}]->(s)
		OPTIONAL MATCH (s)-[:PERFORMED_BY]->(a:Artist)
		WITH s, g, count(r) AS fiveStarCount, collect(a.id) AS artistIds
		ORDER BY fiveStarCount DESC
		LIMIT 1
		RETURN s.id AS songId, s.name AS name, g.name AS genre, 
		       s.albumId AS albumId, s.duration AS duration, artistIds, fiveStarCount
	`

	result, err := session.Run(ctx, query, map[string]interface{}{
		"userID": userID,
	})
	if err != nil {
		return nil, err
	}

	if result.Next(ctx) {
		record := result.Record()
		songID, _ := record.Get("songId")
		name, _ := record.Get("name")
		genre, _ := record.Get("genre")
		albumID, _ := record.Get("albumId")
		duration, _ := record.Get("duration")
		artistIDsInterface, _ := record.Get("artistIds")

		if songID != nil {
			artistIDs := make([]string, 0)
			if artistIDsInterface != nil {
				if ids, ok := artistIDsInterface.([]interface{}); ok {
					for _, id := range ids {
						if idStr, ok := id.(string); ok {
							artistIDs = append(artistIDs, idStr)
						}
					}
				}
			}

			albumIDStr := ""
			if albumID != nil {
				albumIDStr = albumID.(string)
			}

			durationInt := 0
			if duration != nil {
				if d, ok := duration.(int64); ok {
					durationInt = int(d)
				}
			}

			return &model.SongRecommendation{
				SongID:    songID.(string),
				Name:      name.(string),
				Genre:     genre.(string),
				ArtistIDs: artistIDs,
				AlbumID:   albumIDStr,
				Duration:  durationInt,
				Reason:    "Popular in genre you might like",
			}, nil
		}
	}

	if err := result.Err(); err != nil {
		return nil, err
	}

	return nil, nil
}

// DeleteSong removes a song and all its relationships
func (s *Neo4jStore) DeleteSong(ctx context.Context, songID string) error {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	query := `
		MATCH (s:Song {id: $songID})
		DETACH DELETE s
	`

	_, err := session.Run(ctx, query, map[string]interface{}{
		"songID": songID,
	})
	return err
}

// DeleteRating removes a RATED relationship
func (s *Neo4jStore) DeleteRating(ctx context.Context, userID, songID string) error {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	query := `
		MATCH (u:User {id: $userID})-[r:RATED]->(s:Song {id: $songID})
		DELETE r
	`

	_, err := session.Run(ctx, query, map[string]interface{}{
		"userID": userID,
		"songID": songID,
	})
	return err
}
