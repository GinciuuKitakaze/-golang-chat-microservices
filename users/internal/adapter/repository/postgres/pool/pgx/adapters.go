// Package pgx_pool содержит конкретную реализацию интерфейса Pool
// на базе библиотеки jackc/pgx/v5 — одного из самых популярных
// и производительных драйверов PostgreSQL для Go.
package pgx_pool

import (
	"errors"
	"fmt"

	postgres_pool "github.com/GinciuuKitakaze/users/internal/adapter/repository/postgres/pool"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// pgxRows оборачивает pgx.Rows для реализации интерфейса core_postgres_pool.Rows.
// Встраивание pgx.Rows даёт все методы (Next, Close, Err, Scan) «бесплатно».
type pgxRows struct {
	pgx.Rows
}

// pgxRow оборачивает pgx.Row для реализации интерфейса postgres_pool.Row.
// Переопределяем Scan, чтобы преобразовывать pgx-ошибки в наши типизированные ошибк
type pgxRow struct {
	pgx.Row
}

// Scan вызывает оригинальный pgx Scan и преобразует ошибки через mapErrors.
func (r pgxRow) Scan(dest ...any) error {
	err := r.Row.Scan(dest...)
	if err != nil {
		return mapErrors(err)
	}
	return nil
}

// pgxCommandTag оборачивает pgconn.CommandTag для реализации интерфейса CommandTag.
type pgxCommandTag struct {
	pgconn.CommandTag
}

// mapErrors преобразует специфические ошибки pgx в типизированные ошибки пакета pool.
// Это «антикоррупционный слой» (Anti-Corruption Layer) — изолирует детали pgx
// от остального кода приложения.
func mapErrors(err error) error {
	const (
		// Код PostgreSQL для ошибки нарушения внешнего ключа.
		// Полный список кодов: https://www.postgresql.org/docs/current/errcodes-appendix.html
		pgxViolatesForeignKeyErrorCode = "23503"

		// Код PostgreSQL для нарушения ограничения уникальности.
		pgxUniqueViolationErrorCode = "23505"
	)

	// pgx.ErrNoRows → наш ErrNoRows (запись не найдена)
	if errors.Is(err, pgx.ErrNoRows) {
		return postgres_pool.ErrNoRows
	}

	// Проверяем, является ли ошибка структурированной PostgreSQL-ошибкой.
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == pgxViolatesForeignKeyErrorCode {
			return fmt.Errorf(
				"%v: %w",
				err,
				postgres_pool.ErrViolatesForeignKey,
			)
		}
		if pgErr.Code == pgxUniqueViolationErrorCode {
			return fmt.Errorf(
				"%v: %w",
				err,
				postgres_pool.ErrAlreadyExists,
			)
		}
	}

	// Все остальные ошибки оборачиваем в ErrUnknown.
	return fmt.Errorf(
		"%v: %w",
		err,
		postgres_pool.ErrUnknown,
	)
}
