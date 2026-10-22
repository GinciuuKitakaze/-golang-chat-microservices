package redis_pool

import (
	"context"
	"time"
)

// Pool — абстракция над клиентом Redis.
// Инкапсулирует только команды, используемые приложением, и изолирует
// код приложения от конкретной Redis-библиотеки (go-redis, redigo и т.д.).
type Pool interface {
	Get(ctx context.Context, key string) StringCmd
	Set(ctx context.Context, key string, value any, ttl time.Duration) StatusCmd
	Exists(ctx context.Context, keys ...string) IntCmd
	Incr(ctx context.Context, key string) IntCmd
	Expire(ctx context.Context, key string, ttl time.Duration) BoolCmd
	Del(ctx context.Context, keys ...string) IntCmd

	Close() error

	// TTL возвращает значение времени жизни по умолчанию,
	// заданное в конфигурации Redis.
	TTL() time.Duration
}

// StringCmd — результат Redis-команды, возвращающей строковое значение.
type StringCmd interface {
	Bytes() ([]byte, error)
}

// StatusCmd — результат Redis-команды, возвращающей статус выполнения операции.
type StatusCmd interface {
	Err() error
}

// IntCmd — результат Redis-команды, возвращающей целочисленное значение.
// Используется для команд EXISTS, INCR и DEL.
type IntCmd interface {
	Err() error
	Val() int64
}

// BoolCmd — результат Redis-команды, возвращающей логическое значение.
type BoolCmd interface {
	Err() error
	Val() bool
}
