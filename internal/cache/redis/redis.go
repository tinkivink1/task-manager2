package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type Redis struct {
	redis  *redis.Client
	config *Config
}

func New(config *Config) *Redis {
	return &Redis{
		config: config,
	}
}

func (r *Redis) Open() {
	r.redis = redis.NewClient(&redis.Options{
		Addr:     r.config.addr,
		Password: r.config.password,
		DB:       0,
	})
}

func (r *Redis) Close() {
	r.redis.Close()
}

func (r *Redis) Del(key, value string, expiration time.Duration) error {
	res := r.redis.Del(context.TODO(), key)
	return res.Err()
}

func (r *Redis) Set(key, value string, expiration time.Duration) error {
	res := r.redis.Set(context.TODO(), key, value, expiration)
	return res.Err()
}

func (r *Redis) Get(key string) (string, error) {
	res := r.redis.Get(context.TODO(), key)
	if res.Err() != nil {
		return "", res.Err()
	}

	return res.Val(), nil
}
