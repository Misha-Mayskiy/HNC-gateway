package redisstore

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"

	"github.com/shvdev1/HackNeChange/api-gateway/internal/entity"
)

// SettingsStorage provides methods to store and retrieve user settings in Redis
type SettingsStorage struct {
	rdb *redis.Client
}

// NewSettingsStorage creates a new SettingsStorage
func NewSettingsStorage(rdb *redis.Client) *SettingsStorage {
	return &SettingsStorage{rdb: rdb}
}

func keyFor(userID string) string {
	return fmt.Sprintf("user:settings:%s", userID)
}

// SaveSettings stores the provided settings into a Redis hash using HSet.
func (s *SettingsStorage) SaveSettings(ctx context.Context, userID string, settings entity.UserSettings) error {
	key := keyFor(userID)
	fields := map[string]interface{}{
		"theme":        settings.Theme,
		"picked_model": settings.PickedModel,
		"font":         settings.Font,
	}
	return s.rdb.HSet(ctx, key, fields).Err()
}

// GetSettings retrieves the settings for the given user from Redis.
// If no settings exist, returns an empty UserSettings and nil error.
func (s *SettingsStorage) GetSettings(ctx context.Context, userID string) (entity.UserSettings, error) {
	key := keyFor(userID)
	m, err := s.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return entity.UserSettings{}, err
	}

	// If not found, return defaults (empty fields) and no error.
	if len(m) == 0 {
		return entity.UserSettings{}, nil
	}

	us := entity.UserSettings{
		Theme:       m["theme"],
		PickedModel: m["picked_model"],
		Font:        m["font"],
	}
	return us, nil
}
