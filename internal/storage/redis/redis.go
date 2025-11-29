package redisstorage

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/encoding/protojson"

	customer "github.com/shvdev1/HackNeChange/api-gateway/internal/gen"
)

const (
	cacheTTL = time.Minute * 10
)

// Storage provides an interface to Redis for get/set/invalidate
type Storage interface {
	Get(ctx context.Context, userID string) (*customer.GetUserSettingsResponse, error)
	Set(ctx context.Context, userID string, data *customer.GetUserSettingsResponse) error
	Invalidate(ctx context.Context, userID string) error
}

// redisStorage implements Storage
type redisStorage struct {
	client *redis.Client
}

// New creates a new redis storage client
func New(addr string) (Storage, error) {
	client := redis.NewClient(&redis.Options{Addr: addr})
	// verify connection
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return &redisStorage{client: client}, nil
}

// Get retrieves cached settings from Redis
func (r *redisStorage) Get(ctx context.Context, userID string) (*customer.GetUserSettingsResponse, error) {
	key := r.key(userID)
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var res customer.GetUserSettingsResponse
	if err := protojson.Unmarshal([]byte(val), &res); err != nil {
		log.Printf("failed to unmarshal cached value for %s: %v", userID, err)
		return nil, err
	}
	return &res, nil
}

// Set stores settings in Redis as JSON
func (r *redisStorage) Set(ctx context.Context, userID string, data *customer.GetUserSettingsResponse) error {
	key := r.key(userID)
	b, err := protojson.Marshal(data)
	if err != nil {
		return err
	}
	// set with TTL
	return r.client.Set(ctx, key, string(b), cacheTTL).Err()
}

// Invalidate removes cache entry for user
func (r *redisStorage) Invalidate(ctx context.Context, userID string) error {
	key := r.key(userID)
	return r.client.Del(ctx, key).Err()
}

func (r *redisStorage) key(userID string) string {
	return "user:settings:" + userID
}
