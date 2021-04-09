package cache

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
)

type Client struct {
	Host string
	Port string
}

func (cacheClient *Client) DeleteKeysFromCache(ctx context.Context, keys []string) error {
	rdb := redis.NewClient(&redis.Options{
		Addr: cacheClient.Host + ":" + cacheClient.Port,
		Password: "",
		DB: 0,
	})
	if rdb != nil {
		defer rdb.Close()

		rdb.Del(ctx, keys...)
		return nil
	}

	return errors.New("redis client cannot be initialized")
}

func (cacheClient *Client) DeleteKeyFromMapCache(ctx context.Context, mapName string, key string) error {
	rdb := redis.NewClient(&redis.Options{
		Addr: cacheClient.Host + ":" + cacheClient.Port,
		Password: "",
		DB: 0,
	})
	if rdb != nil {
		defer rdb.Close()

		rdb.HDel(ctx, mapName, key)
		return nil
	}

	return errors.New("redis client cannot be initialized")
}

func (cacheClient *Client) GetAllKeysFromMapCache(ctx context.Context, mapName string) (map[string]string, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: cacheClient.Host + ":" + cacheClient.Port,
		Password: "",
		DB: 0,
	})
	if rdb != nil {
		defer rdb.Close()

		res, err := rdb.HGetAll(ctx, mapName).Result()
		if err != nil {
			return nil, err
		}
		return res, nil
	}

	return nil, errors.New("redis client cannot be initialized")
}