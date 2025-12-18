package database

import (
	"context"

	"ebidsystem_csm/internal/config"

	"github.com/redis/go-redis/v9"
)

var Redis *redis.Client

func InitRedis(cfg config.RedisConfig) error {
	Redis = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	return Redis.Ping(context.Background()).Err()
}
