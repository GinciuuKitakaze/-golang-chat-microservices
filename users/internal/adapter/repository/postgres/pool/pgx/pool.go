package pgx_pool

import (
	"context"
	"fmt"
	"time"

	postgres_pool "github.com/GinciuuKitakaze/users/internal/adapter/repository/postgres/pool"
	core_config "github.com/GinciuuKitakaze/users/internal/core/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool — конкретная реализация интерфейса postgres_pool.Pool
// на базе pgxpool.Pool (пул соединений pgx).
//
// Встраивание *pgxpool.Pool даёт доступ ко всем методам pgx,
// но мы переопределяем Query/QueryRow/Exec, чтобы:
//  1. Оборачивать результаты в наши интерфейсы (Rows, Row, CommandTag)
//  2. Скрывать зависимость от pgx от слоя репозиториев
type Pool struct {
	*pgxpool.Pool
	opTimeout time.Duration
}

// NewPool создаёт и проверяет пул соединений с PostgreSQL.
// Ping() при инициализации гарантирует, что БД доступна до начала работы сервера.
func NewPool(ctx context.Context, cfg core_config.PostgresConfig) (*Pool, error) {
	connetstionString := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database)

	pgxConfig, err := pgxpool.ParseConfig(connetstionString)
	if err != nil {
		return nil, fmt.Errorf("parse pgxconfig: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, pgxConfig)
	if err != nil {
		return nil, fmt.Errorf("create pgxpool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()

		return nil, fmt.Errorf("pgxpool ping: %w", err)
	}

	return &Pool{
		Pool:      pool,
		opTimeout: cfg.Timeout,
	}, nil
}

// QueryRow выполняет запрос, возвращающий одну строку.
// Оборачивает pgx.Row в pgxRow (с маппингом ошибок).
func (p *Pool) QueryRow(ctx context.Context, sql string, args ...any) postgres_pool.Row {
	row := p.Pool.QueryRow(ctx, sql, args...)
	return pgxRow{row}
}

// Query выполняет SELECT-запрос и возвращает интерфейс Rows.
// Оборачивает pgx.Rows в pgxRows для соответствия интерфейсу пула.
func (p *Pool) Query(ctx context.Context, sql string, args ...any) (postgres_pool.Rows, error) {
	rows, err := p.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return pgxRows{rows}, nil
}

// Exec выполняет DML-запрос (INSERT, UPDATE, DELETE).
// Оборачивает pgconn.CommandTag в pgxCommandTag.
func (p *Pool) Exec(ctx context.Context, sql string, arguments ...any) (postgres_pool.CommandTag, error) {
	tag, err := p.Pool.Exec(ctx, sql, arguments...)
	if err != nil {
		return nil, err
	}

	return pgxCommandTag{tag}, nil
}

// OpTimeout возвращает максимальное время выполнения одного запроса.
func (p *Pool) OpTimeout() time.Duration {
	return p.opTimeout
}
