package store

import (
	"errors"

	"github.com/go-redis/redis"
)

var (
	ErrKeyNotFound = errors.New("key not found")
)

type Store interface {
	Set(key string, value string) error
	Get(key string) (string, error)
}

type Redis struct {
	redisClient *redis.Client
}

func NewRedis(addr string) (*Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	_, err := client.Ping().Result()
	if err != nil {
		return nil, err
	}
	return &Redis{
		redisClient: client,
	}, nil
}

func (r *Redis) Get(key string) (string, error) {
	val, err := r.redisClient.Get(key).Result()
	if err == redis.Nil {
		return "", ErrKeyNotFound
	}
	return val, err
}

func (r *Redis) Set(key, val string) error {
	return r.redisClient.Set(key, val, 0).Err()
}
