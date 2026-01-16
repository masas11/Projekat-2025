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

type AlbumRepository struct {
	collection *mongo.Collection
}

func NewAlbumRepository(db *mongo.Database) *AlbumRepository {
	return &AlbumRepository{
		collection: db.Collection("albums"),
	}
}

func (r *AlbumRepository) Create(ctx context.Context, album *model.Album) error {
	album.ID = uuid.NewString()
	album.CreatedAt = time.Now()
	album.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, album)
	return err
}

func (r *AlbumRepository) GetByID(ctx context.Context, id string) (*model.Album, error) {
	var album model.Album
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&album)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("album not found")
		}
		return nil, err
	}
	return &album, nil
}

func (r *AlbumRepository) GetByArtistID(ctx context.Context, artistID string) ([]*model.Album, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"artistIds": artistID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var albums []*model.Album
	if err = cursor.All(ctx, &albums); err != nil {
		return nil, err
	}

	return albums, nil
}

func (r *AlbumRepository) GetAll(ctx context.Context) ([]*model.Album, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var albums []*model.Album
	if err = cursor.All(ctx, &albums); err != nil {
		return nil, err
	}

	return albums, nil
}

func (r *AlbumRepository) Update(ctx context.Context, id string, album *model.Album) error {
	album.UpdatedAt = time.Now()
	
	update := bson.M{
		"$set": bson.M{
			"name":        album.Name,
			"releaseDate": album.ReleaseDate,
			"genre":       album.Genre,
			"artistIds":   album.ArtistIDs,
			"updatedAt":   album.UpdatedAt,
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("album not found")
	}
	return nil
}

func (r *AlbumRepository) Delete(ctx context.Context, id string) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.New("album not found")
	}
	return nil
}