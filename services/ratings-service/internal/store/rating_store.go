package store

import (
	"context"
	"log"
	"time"

	"ratings-service/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type RatingStore struct {
	collection *mongo.Collection
}

func NewRatingStore(db *mongo.Database) *RatingStore {
	return &RatingStore{
		collection: db.Collection("ratings"),
	}
}

func (rs *RatingStore) Create(ctx context.Context, rating *model.Rating) error {
	rating.ID = primitive.NewObjectID().Hex()
	rating.CreatedAt = time.Now()
	rating.UpdatedAt = time.Now()

	_, err := rs.collection.InsertOne(ctx, rating)
	if err != nil {
		log.Printf("Error creating rating: %v", err)
		return err
	}

	log.Printf("Created rating: %+v", rating)
	return nil
}

func (rs *RatingStore) GetBySongAndUser(ctx context.Context, songID, userID string) (*model.Rating, error) {
	var rating model.Rating
	err := rs.collection.FindOne(ctx, bson.M{
		"songId": songID,
		"userId": userID,
	}).Decode(&rating)

	if err == mongo.ErrNoDocuments {
		return nil, nil
	}

	if err != nil {
		log.Printf("Error getting rating: %v", err)
		return nil, err
	}

	return &rating, nil
}

func (rs *RatingStore) Update(ctx context.Context, rating *model.Rating) error {
	rating.UpdatedAt = time.Now()

	filter := bson.M{
		"songId": rating.SongID,
		"userId": rating.UserID,
	}

	update := bson.M{
		"$set": bson.M{
			"rating":    rating.Rating,
			"updatedAt": rating.UpdatedAt,
		},
	}

	_, err := rs.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Printf("Error updating rating: %v", err)
		return err
	}

	log.Printf("Updated rating: %+v", rating)
	return nil
}

func (rs *RatingStore) GetAverageRating(ctx context.Context, songID string) (float64, int, error) {
	pipeline := []bson.M{
		{"$match": bson.M{"songId": songID}},
		{"$group": bson.M{
			"_id":   nil,
			"avg":   bson.M{"$avg": "$rating"},
			"count": bson.M{"$sum": 1},
		}},
	}

	cursor, err := rs.collection.Aggregate(ctx, pipeline)
	if err != nil {
		log.Printf("Error getting average rating: %v", err)
		return 0, 0, err
	}
	defer cursor.Close(ctx)

	var result struct {
		Avg   float64 `bson:"avg"`
		Count int     `bson:"count"`
	}

	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return 0, 0, err
		}
		return result.Avg, result.Count, nil
	}

	return 0, 0, nil
}
