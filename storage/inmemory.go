package storage

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type InMemoryStorageI interface {
	Set(key string, value []byte) error
	// SetAnother(key int64, value string, text string) error
	Get(key string) ([]byte, error)
	// GetAnother(key int64, text string) (string, error)
	// Delete(key int64, text string) error
	// DeleteWithoutTxt(key int64) error
}

type storageRedis struct {
	client *redis.Client
}

func NewInMemoryStorage(rdb *redis.Client) InMemoryStorageI {
	return &storageRedis{
		client: rdb,
	}
}

func (rd *storageRedis) Set(key string, value []byte) error {
	err := rd.client.Set(context.Background(), key, value, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

func (rd *storageRedis) Get(key string) ([]byte, error) {
	val, err := rd.client.Get(context.Background(), key).Result()
	if err != nil {
		return nil, err
	}

	return []byte(val), nil
}
