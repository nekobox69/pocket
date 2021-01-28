// Package pocket Create at 2020-12-01 11:35
package pocket

import (
	"context"
	"sync"
	"time"

	redis "github.com/go-redis/redis/v8"
)

type RedisUtils struct {
	Client     *redis.Client
	expiration time.Duration
	once       sync.Once
	ctx        context.Context
}

// RedisConfig redis config
type RedisConfig struct {
	Host       string
	Pwd        string
	ExpireTime int
}

// NewRedis get redis client
func NewRedis(config RedisConfig) (*RedisUtils, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.Host,
		Password: config.Pwd, // no password set
		DB:       0,          // use default DB
	})
	ctx := context.Background()
	_, err := redisClient.Ping(ctx).Result()
	if nil != err {
		return nil, err
	}
	return &RedisUtils{Client: redisClient, expiration: time.Duration(config.ExpireTime) * time.Second, ctx: ctx}, nil
}

// Close close redis connect
func (r *RedisUtils) Close() {
	r.once.Do(func() {
		r.Client.Close()
	})
}
