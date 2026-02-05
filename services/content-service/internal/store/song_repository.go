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

type SongRepository struct {
	collection *mongo.Collection
}

func NewSongRepository(db *mongo.Database) *SongRepository {
	return &SongRepository{
		collection: db.Collection("songs"),
	}
}

func (r *SongRepository) Create(ctx context.Context, song *model.Song) error {
	song.ID = uuid.NewString()
	now := time.Now()
	song.CreatedAt = now
	song.UpdatedAt = now

	_, err := r.collection.InsertOne(ctx, song)
	return err
}

func (r *SongRepository) GetByID(ctx context.Context, id string) (*model.Song, error) {
	var song model.Song
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&song)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("song not found")
		}
		return nil, err
	}
	return &song, nil
}

func (r *SongRepository) GetByAlbumID(ctx context.Context, albumID string) ([]*model.Song, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"albumId": albumID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var songs []*model.Song
	if err = cursor.All(ctx, &songs); err != nil {
		return nil, err
	}

	return songs, nil
}

func (r *SongRepository) GetAll(ctx context.Context) ([]*model.Song, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var songs []*model.Song
	if err = cursor.All(ctx, &songs); err != nil {
		return nil, err
	}

	return songs, nil
}

func (r *SongRepository) Exists(ctx context.Context, id string) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"_id": id})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *SongRepository) Update(ctx context.Context, id string, song *model.Song) error {
	song.UpdatedAt = time.Now()
	
	update := bson.M{
		"$set": bson.M{
			"name":      song.Name,
			"duration":  song.Duration,
			"genre":     song.Genre,
			"albumId":   song.AlbumID,
			"artistIds": song.ArtistIDs,
			"updatedAt": song.UpdatedAt,
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("song not found")
	}
	return nil
}

func (r *SongRepository) Delete(ctx context.Context, id string) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.New("song not found")
	}
	return nil
}