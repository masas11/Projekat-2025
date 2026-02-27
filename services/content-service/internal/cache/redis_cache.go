package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	MostPlayedCacheKey = "most_played_songs"
	SongPlayCountKey   = "song_play_count:%s"
	CacheTTL           = 1 * time.Hour // Cache expires after 1 hour
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(redisURL string) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     redisURL,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Printf("Connected to Redis at %s", redisURL)
	return &RedisCache{client: client}, nil
}

func (c *RedisCache) Close() error {
	return c.client.Close()
}

// IncrementPlayCount increments the play count for a song
func (c *RedisCache) IncrementPlayCount(ctx context.Context, songID string) error {
	key := fmt.Sprintf(SongPlayCountKey, songID)
	_, err := c.client.Incr(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to increment play count: %w", err)
	}

	// Set expiration for the counter (24 hours)
	c.client.Expire(ctx, key, 24*time.Hour)

	// Invalidate the most played cache so it gets refreshed
	c.client.Del(ctx, MostPlayedCacheKey)

	return nil
}

// GetMostPlayedSongs retrieves the most played songs from cache or computes them
func (c *RedisCache) GetMostPlayedSongs(ctx context.Context, limit int) ([]*MostPlayedSong, error) {
	// Try to get from cache first
	cached, err := c.client.Get(ctx, MostPlayedCacheKey).Result()
	if err == nil && cached != "" {
		var songs []*MostPlayedSong
		if err := json.Unmarshal([]byte(cached), &songs); err == nil {
			log.Printf("Retrieved %d most played songs from Redis cache", len(songs))
			// Return up to limit
			if len(songs) > limit {
				return songs[:limit], nil
			}
			return songs, nil
		}
	}

	// Cache miss or invalid data - need to compute from play counts
	log.Printf("Cache miss for most played songs, computing from play counts...")
	return c.computeMostPlayedSongs(ctx, limit)
}

// computeMostPlayedSongs computes the most played songs from Redis counters
func (c *RedisCache) computeMostPlayedSongs(ctx context.Context, limit int) ([]*MostPlayedSong, error) {
	// Get all keys matching song_play_count:*
	keys, err := c.client.Keys(ctx, "song_play_count:*").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get play count keys: %w", err)
	}

	if len(keys) == 0 {
		log.Printf("No play counts found in Redis")
		return []*MostPlayedSong{}, nil
	}

	// Get play counts for all songs
	type songCount struct {
		SongID string
		Count  int64
	}

	var songsWithCounts []songCount
	for _, key := range keys {
		count, err := c.client.Get(ctx, key).Int64()
		if err != nil {
			continue // Skip if count can't be retrieved
		}

		// Extract song ID from key (song_play_count:songID)
		songID := key[len("song_play_count:"):]
		songsWithCounts = append(songsWithCounts, songCount{
			SongID: songID,
			Count:  count,
		})
	}

	// Sort by count (descending) - simple bubble sort for small datasets
	for i := 0; i < len(songsWithCounts)-1; i++ {
		for j := i + 1; j < len(songsWithCounts); j++ {
			if songsWithCounts[i].Count < songsWithCounts[j].Count {
				songsWithCounts[i], songsWithCounts[j] = songsWithCounts[j], songsWithCounts[i]
			}
		}
	}

	// Convert to MostPlayedSong format
	result := make([]*MostPlayedSong, 0, limit)
	for i, sc := range songsWithCounts {
		if i >= limit {
			break
		}
		result = append(result, &MostPlayedSong{
			SongID: sc.SongID,
			Count:  int(sc.Count),
		})
	}

	// Cache the result
	if len(result) > 0 {
		cacheData, err := json.Marshal(result)
		if err == nil {
			c.client.Set(ctx, MostPlayedCacheKey, cacheData, CacheTTL)
			log.Printf("Cached %d most played songs in Redis", len(result))
		}
	}

	return result, nil
}

// MostPlayedSong represents a song with its play count
type MostPlayedSong struct {
	SongID string `json:"songId"`
	Count  int    `json:"count"`
}
