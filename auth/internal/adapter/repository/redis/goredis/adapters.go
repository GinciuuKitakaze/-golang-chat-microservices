package goredis_pool

import (
	"errors"

	redis_pool "github.com/GinciuuKitakaze/auth/internal/adapter/repository/redis"
	"github.com/redis/go-redis/v9"
)

// goredisStringCmd оборачивает *redis.StringCmd
// и реализует redis_pool.StringCmd.
type goredisStringCmd struct {
	*redis.StringCmd
}

func (c goredisStringCmd) Bytes() ([]byte, error) {
	data, err := c.StringCmd.Bytes()
	if err != nil {
		return nil, mapError(err)
	}

	return data, nil
}

// goredisStatusCmd оборачивает *redis.StatusCmd
// и реализует redis_pool.StatusCmd.
type goredisStatusCmd struct {
	*redis.StatusCmd
}

func (c goredisStatusCmd) Err() error {
	return mapError(c.StatusCmd.Err())
}

// goredisIntCmd оборачивает *redis.IntCmd
// и реализует redis_pool.IntCmd.
type goredisIntCmd struct {
	*redis.IntCmd
}

func (c goredisIntCmd) Err() error {
	return mapError(c.IntCmd.Err())
}

// goredisBoolCmd оборачивает *redis.BoolCmd
// и реализует redis_pool.BoolCmd.
type goredisBoolCmd struct {
	*redis.BoolCmd
}

func (c goredisBoolCmd) Err() error {
	return mapError(c.BoolCmd.Err())
}

// mapError переводит go-redis-специфичные ошибки
// в ошибки, используемые абстракцией Redis.
func mapError(err error) error {
	if errors.Is(err, redis.Nil) {
		return redis_pool.NotFound
	}

	return err
}
