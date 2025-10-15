// file: internal/adapters/cache/redis/game_config_cache.go
package redis

import (
	"context"
	"sage-of-elements-backend/internal/domain"

	"github.com/redis/go-redis/v9"
)

const gameConfigsCacheKey = "game_configs:v1"

// GameConfigCacheRepository คือ "คนทำงาน" ที่คุยกับ Redis สำหรับ Game Configs
type GameConfigCacheRepository struct {
	client *redis.Client
}

func NewGameConfigCacheRepository(client *redis.Client) *GameConfigCacheRepository {
	return &GameConfigCacheRepository{client: client}
}

// SetAllConfigs บันทึก Configs ทั้งหมดลงใน Redis Hash
func (r *GameConfigCacheRepository) SetAllConfigs(configs []domain.GameConfig) error {
	ctx := context.Background()
	pipe := r.client.Pipeline()
	for _, config := range configs {
		pipe.HSet(ctx, gameConfigsCacheKey, config.Key, config.Value)
	}
	_, err := pipe.Exec(ctx)
	return err
}

// GetConfig พยายามดึง Config 1 ตัวจาก Redis Hash
func (r *GameConfigCacheRepository) GetConfig(key string) (string, error) {
	ctx := context.Background()
	val, err := r.client.HGet(ctx, gameConfigsCacheKey, key).Result()
	if err == redis.Nil {
		return "", nil // Cache miss
	}
	return val, err
}
