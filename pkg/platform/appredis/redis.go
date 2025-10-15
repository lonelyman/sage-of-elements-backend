package appredis

import (
	"context"
	"fmt"
	"sage-of-elements-backend/pkg/appconfig"
	"sage-of-elements-backend/pkg/applogger"

	"github.com/redis/go-redis/v9"
)

// NewRedisConnection สร้างการเชื่อมต่อกับ Redis
func NewConnection(cfg appconfig.RedisConfig, logger applogger.Logger) (*redis.Client, error) {
	redisHost := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	client := redis.NewClient(&redis.Options{
		Addr:     redisHost,
		Password: cfg.Password,
		DB:       0, // ใช้ DB default
	})

	// ลอง Ping เพื่อตรวจสอบว่าเชื่อมต่อสำเร็จหรือไม่
	if _, err := client.Ping(context.Background()).Result(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Success("Successfully connected to Redis", "host", redisHost)
	return client, nil
}
