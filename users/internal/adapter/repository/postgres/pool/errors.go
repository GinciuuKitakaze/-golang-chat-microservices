package postgres_pool

import "errors"

// Sentinel-ошибки уровня пула соединений.
// Адаптер pgx преобразует специфические ошибки pgx в эти типизированные ошибки,
// позволяя репозиториям работать с ними без зависимости от pgx.
var (
	// ErrNoRows — запрос не вернул строк.
	// Репозиторий преобразует это в core_errors.ErrNotFound.
	ErrNoRows = errors.New("no rows")

	// ErrViolatesForeignKey — нарушение ограничения внешнего ключа (PostgreSQL код 23503).
	// Например: попытка создать задачу с несуществующим user_id.
	ErrViolatesForeignKey = errors.New("violates foreign key")

	// ErrUnknown — любая другая ошибка базы данных, не обработанная явно.
	ErrUnknown = errors.New("unknown")

	// ErrAlreadyExists — нарушение ограничения уникальности (PostgreSQL код 23505).
	// Например: попытка создать пользователя с уже существующим email или phone.
	ErrAlreadyExists = errors.New("already exists")
)
