package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"content-service/internal/model"
)

type ArtistRepository struct {
	collection *mongo.Collection
}

func NewArtistRepository(db *mongo.Database) *ArtistRepository {
	return &ArtistRepository{
		collection: db.Collection("artists"),
	}
}

func (r *ArtistRepository) Create(ctx context.Context, artist *model.Artist) error {
	artist.ID = uuid.NewString()
	artist.CreatedAt = time.Now()
	artist.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, artist)
	return err
}

func (r *ArtistRepository) GetByID(ctx context.Context, id string) (*model.Artist, error) {
	var artist model.Artist
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&artist)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("artist not found")
		}
		return nil, err
	}
	return &artist, nil
}

func (r *ArtistRepository) Update(ctx context.Context, id string, artist *model.Artist) error {
	artist.UpdatedAt = time.Now()
	
	update := bson.M{
		"$set": bson.M{
			"name":      artist.Name,
			"biography": artist.Biography,
			"genres":    artist.Genres,
			"updatedAt": artist.UpdatedAt,
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("artist not found")
	}
	return nil
}

func (r *ArtistRepository) GetAll(ctx context.Context) ([]*model.Artist, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var artists []*model.Artist
	if err = cursor.All(ctx, &artists); err != nil {
		return nil, err
	}

	return artists, nil
}
