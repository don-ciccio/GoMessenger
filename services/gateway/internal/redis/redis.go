package redisutil

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"
)

func NewRedisClient() (*redis.Client, error) {
	addr := os.Getenv("REDIS_URL")
	if addr == "" {
		addr = "localhost:6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr: addr,
		MaintNotificationsConfig: &maintnotifications.Config{
			Mode: maintnotifications.ModeDisabled,
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("falha ao conectar ao Redis: %w", err)
	}

	return client, nil
}

type RedisConfig struct {
	StreamChat  string
	ChannelChat string
}

func LoadRedisConfig() *RedisConfig {
	return &RedisConfig{
		StreamChat:  os.Getenv("REDIS_STREAM_CHAT"),
		ChannelChat: os.Getenv("REDIS_CHANNEL_CHAT"),
	}
}
