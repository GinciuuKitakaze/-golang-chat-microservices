package goredis_pool

import (
	"context"
	"fmt"
	"time"

	redis_pool "github.com/GinciuuKitakaze/auth/internal/adapter/repository/redis"
	core_config "github.com/GinciuuKitakaze/auth/internal/core/config"
	"github.com/redis/go-redis/v9"
)

// Pool — конкретная реализация core_redis_pool.Pool на базе библиотеки go-redis.
// Методы являются тонкими обёртками: делегируют вызов клиенту go-redis
// и адаптируют возвращаемые типы к интерфейсам core_redis_pool.
type Pool struct {
	client *redis.Client
	ttl    time.Duration
}

func NewPool(ctx context.Context, cfg core_config.RedisConfig) (*Pool, error) {
	options := &redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	}
	client := redis.NewClient(options)

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()

		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &Pool{
		client: client,
		ttl:    cfg.TTL,
	}, nil
}

func (p *Pool) Get(
	ctx context.Context,
	key string,
) redis_pool.StringCmd {
	cmd := p.client.Get(ctx, key)

	return goredisStringCmd{cmd}
}

func (p *Pool) Set(
	ctx context.Context,
	key string,
	value any,
	ttl time.Duration,
) redis_pool.StatusCmd {
	cmd := p.client.Set(ctx, key, value, ttl)

	return goredisStatusCmd{cmd}
}

func (p *Pool) Exists(
	ctx context.Context,
	keys ...string,
) redis_pool.IntCmd {
	cmd := p.client.Exists(ctx, keys...)

	return goredisIntCmd{cmd}
}

func (p *Pool) Incr(
	ctx context.Context,
	key string,
) redis_pool.IntCmd {
	cmd := p.client.Incr(ctx, key)

	return goredisIntCmd{cmd}
}

func (p *Pool) Expire(
	ctx context.Context,
	key string,
	ttl time.Duration,
) redis_pool.BoolCmd {
	cmd := p.client.Expire(ctx, key, ttl)

	return goredisBoolCmd{cmd}
}

func (p *Pool) Del(
	ctx context.Context,
	keys ...string,
) redis_pool.IntCmd {
	cmd := p.client.Del(ctx, keys...)

	return goredisIntCmd{cmd}
}

func (p *Pool) Close() error {
	return p.client.Close()
}

func (p *Pool) TTL() time.Duration {
	return p.ttl
}
