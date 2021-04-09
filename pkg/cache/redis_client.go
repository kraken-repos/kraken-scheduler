package cache

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"time"
	"github.com/gomodule/redigo/redis"
)

type RedisCli struct {
	Host string
	Port string
}

func (redisCli *RedisCli) New() (redis.Conn, error) {
//	redisCli.Host = "10.224.179.243"
//	redisCli.Port = "6379"

	conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%s", redisCli.Host, redisCli.Port))
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (redisCli *RedisCli) SetMapEntry(mapName string, key string, value string) error {
	conn, err := redisCli.New()
	if err != nil {
		return err
	}

	defer conn.Close()

	mapVal := []interface{}{mapName, key, value}

	_, err = conn.Do("HSET", mapVal...)
	if err != nil {
	return err
	}

	return nil
}

func (redisCli *RedisCli) GetMapEntry(mapName string, key string) (v1.PullPolicy, bool) {
	conn, err := redisCli.New()
	if err != nil {
		return v1.PullIfNotPresent, false
	}

	defer conn.Close()

	value, err := redis.Bytes(conn.Do("HGET", mapName, key))
	if err != nil {
		return v1.PullIfNotPresent, false
	}

	switch pullPolicy := string(value); pullPolicy {
	case "Always":
		return v1.PullAlways, true
	case "Never":
		return v1.PullNever, true
	case "IfNotPresent":
		return v1.PullIfNotPresent, true
	}

	return v1.PullIfNotPresent, false
}

func (redisCli *RedisCli) DeleteMap(mapNames []string) error {
	conn, err := redisCli.New()
	if err != nil {
		return err
	}

	defer conn.Close()

	for _, mapName := range mapNames {
		_, err = conn.Do("DEL", mapName)
		if err != nil {
			return err
		}
	}

	return nil
}

func (redisCli *RedisCli) SetEntry(mapName string, key string, value time.Time) error {
	conn, err := redisCli.New()
	if err != nil {
		return err
	}

	defer conn.Close()

	timeVal, err := value.MarshalBinary()
	if err != nil {
		return err
	}

	mapVal := []interface{}{mapName, key, timeVal}

	_, err = conn.Do("HSET", mapVal...)
	if err != nil {
		return err
	}

	return nil
}

func (redisCli *RedisCli) GetEntry(mapName string, key string) (time.Time, bool) {
	conn, err := redisCli.New()
	if err != nil {
		return time.Time{}, false
	}

	defer conn.Close()

	value, err := redis.Bytes(conn.Do("HGET", mapName, key))
	if err != nil {
		return time.Time{}, false
	}

	var timeVal time.Time
	err = timeVal.UnmarshalBinary(value)
	if err != nil {
		return time.Time{}, false
	}

	return timeVal, true
}

func (redisCli *RedisCli) GetAllEntry(mapName string) (map[string]time.Time, error) {
	conn, err := redisCli.New()
	if err != nil {
		return nil, err
	}

	defer conn.Close()

	mapVal := map[string]time.Time{}

	values, err := redis.Values(conn.Do("HGETALL", mapName))
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(values); i +=2 {
		var key, okk = values[i].([]byte)
		var val, okv = values[i+1].([]byte)

		if okk && okv {
			var timeVal time.Time
			err = timeVal.UnmarshalBinary(val)
			if err != nil {
				return nil, err
			}

			mapVal[string(key)] = timeVal
		}
	}

	return mapVal, nil
}

func (redisCli *RedisCli) DeleteEntry(mapName string, key string) error {
	conn, err := redisCli.New()
	if err != nil {
		return err
	}

	defer conn.Close()

	mapVal := []interface{}{mapName, key}

	_, err = conn.Do("HDEL", mapVal...)
	if err != nil {
		return err
	}

	return nil
}