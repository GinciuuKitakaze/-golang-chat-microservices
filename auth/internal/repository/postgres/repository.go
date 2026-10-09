package postgres_repository

import (
	"context"
	"errors"
	"fmt"
	"time"
	"uuid"

	postgres_pool "github.com/GinciuuKitakaze/auth/internal/adapter/repository/postgres/pool"
	core_errors "github.com/GinciuuKitakaze/auth/internal/core/errors"
	core_logger "github.com/GinciuuKitakaze/auth/internal/core/logger"
	core_model "github.com/GinciuuKitakaze/auth/internal/core/model"
	repository_mapper "github.com/GinciuuKitakaze/auth/internal/repository/mapper"
	postgres_model "github.com/GinciuuKitakaze/auth/internal/repository/model"
)

// Repository реализует доступ Auth Service к PostgreSQL.
type Repository struct {
	// pool — интерфейс PostgreSQL-пула.
	// Благодаря интерфейсу Repository не зависит напрямую от pgx.
	pool postgres_pool.Pool

	// logger — логгер для записи ошибок и служебной информации.
	logger *core_logger.Logger
}

// NewRepository создаёт новый PostgreSQL Repository.
func NewRepository(
	pool postgres_pool.Pool,
	logger *core_logger.Logger,
) *Repository {
	return &Repository{
		pool:   pool,
		logger: logger,
	}
}

// CreateCredentials создаёт авторизационные данные пользователя.
//
// В Auth Service пользователь уже должен существовать в Users Service.
// Здесь мы создаём только credentials: user_id + password_hash.
func (r *Repository) CreateCredentials(
	ctx context.Context,
	credentials core_model.CreateCredentials,
) error {

	// Ограничиваем время выполнения запроса.
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	// Создаём запись в credentials.
	// created_at и updated_at PostgreSQL заполнит сам через DEFAULT NOW().
	query := `
		INSERT INTO credentials (
			user_id,
			password_hash
		)
		VALUES ($1, $2);
	`

	// Выполняем INSERT.
	_, err := r.pool.Exec(
		ctx,
		query,
		credentials.UserID,
		credentials.PasswordHash,
	)
	if err != nil {

		// Credentials для этого user_id уже существуют.
		if errors.Is(err, postgres_pool.ErrAlreadyExists) {
			return core_errors.ErrAlreadyExists
		}

		// Остальные ошибки передаём дальше с контекстом.
		return fmt.Errorf("create credentials: %w", err)
	}

	return nil
}

// GetCredentialsByUserID получает credentials пользователя по его UUID.
func (r *Repository) GetCredentialsByUserID(
	ctx context.Context,
	userID uuid.UUID,
) (core_model.Credentials, error) {

	// Ограничиваем время выполнения запроса.
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	// Ищем credentials конкретного пользователя.
	query := `
		SELECT user_id, password_hash, created_at, updated_at
		FROM credentials
		WHERE user_id = $1;
	`

	// Выполняем SELECT одной строки.
	row := r.pool.QueryRow(ctx, query, userID)

	var credentialsModel postgres_model.Credentials

	// Считываем найденную запись.
	err := row.Scan(
		&credentialsModel.UserID,
		&credentialsModel.PasswordHash,
		&credentialsModel.Created,
		&credentialsModel.Updated,
	)
	if err != nil {

		// pgx.ErrNoRows уже был преобразован нашим pool
		// в postgres_pool.ErrNoRows.
		// На уровне core превращаем его в ErrNotFound.
		if errors.Is(err, postgres_pool.ErrNoRows) {
			return core_model.Credentials{}, core_errors.ErrNotFound
		}

		return core_model.Credentials{}, fmt.Errorf(
			"get credentials by user id: %w",
			err,
		)
	}

	// Преобразуем модель БД в доменную модель.
	return repository_mapper.ToDomainCredentials(credentialsModel), nil
}

// SaveRefreshToken сохраняет refresh token в PostgreSQL.
//
// Сам refresh token мы не храним.
// В БД сохраняется только его hash.
func (r *Repository) SaveRefreshToken(
	ctx context.Context,
	token core_model.CreateRefreshToken,
) error {

	// Ограничиваем время выполнения запроса.
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	// Сохраняем hash refresh token.
	query := `
		INSERT INTO refresh_tokens (
			id,
			user_id,
			token_hash,
			expires_at
		)
		VALUES ($1, $2, $3, $4);
	`

	// Выполняем INSERT.
	_, err := r.pool.Exec(
		ctx,
		query,
		token.ID,
		token.UserID,
		token.TokenHash,
		token.ExpiresAt,
	)
	if err != nil {

		// Такой refresh token уже существует.
		if errors.Is(err, postgres_pool.ErrAlreadyExists) {
			return core_errors.ErrAlreadyExists
		}

		// Credentials для пользователя не найдены.
		if errors.Is(err, postgres_pool.ErrViolatesForeignKey) {
			return core_errors.ErrNotFound
		}

		// Остальные ошибки передаём дальше с контекстом.
		return fmt.Errorf("save refresh token: %w", err)
	}

	return nil
}

// RevokeRefreshToken отзывает refresh token.
//
// Для этого мы записываем время отзыва в revoked_at.
func (r *Repository) RevokeRefreshToken(
	ctx context.Context,
	tokenHash string,
) error {

	// Ограничиваем время выполнения запроса.
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	// Обновляем revoked_at.
	query := `
		UPDATE refresh_tokens
		SET revoked_at = $1
		WHERE token_hash = $2;
	`

	// Текущее время будет временем отзыва токена.
	now := time.Now()

	cmdTag, err := r.pool.Exec(
		ctx,
		query,
		now,
		tokenHash,
	)
	if err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return core_errors.ErrNotFound
	}

	return nil
}

func (r *Repository) ConsumeRefreshToken(
	ctx context.Context,
	tokenHash string,
) (uuid.UUID, error) {

	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		UPDATE refresh_tokens
		SET revoked_at = $1
		WHERE token_hash = $2
			AND revoked_at IS NULL
			AND expires_at > $1
		RETURNING user_id;
	`

	row := r.pool.QueryRow(ctx, query, time.Now(), tokenHash)

	var userID uuid.UUID
	err := row.Scan(&userID)
	if err != nil {
		if errors.Is(err, postgres_pool.ErrNoRows) {
			return uuid.Nil(), core_errors.ErrNotFound
		}

		return uuid.Nil(), fmt.Errorf("consume refresh token: %w", err)
	}

	return userID, nil
}

// DeleteUserAuthData удаляет все авторизационные данные пользователя.
//
// Сначала удаляется credentials.
// Благодаря ON DELETE CASCADE PostgreSQL автоматически
// удаляет связанные refresh_tokens.
func (r *Repository) DeleteUserAuthData(
	ctx context.Context,
	userID uuid.UUID,
) error {

	// Ограничиваем время выполнения запроса.
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	// Удаляем credentials пользователя.
	// refresh_tokens удалятся автоматически благодаря FK + CASCADE.
	query := `
		DELETE FROM credentials
		WHERE user_id = $1;
	`

	// Выполняем DELETE.
	cmdTag, err := r.pool.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf(
			"delete user auth data: %w",
			err,
		)
	}

	// Если ничего не удалили,
	// значит auth-данных такого пользователя нет.
	if cmdTag.RowsAffected() == 0 {
		return core_errors.ErrNotFound
	}

	return nil
}
