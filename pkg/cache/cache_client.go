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