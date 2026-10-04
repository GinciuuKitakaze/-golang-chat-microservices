package postgres_repository

import (
	"context"
	"errors"
	"fmt"
	"uuid"

	"github.com/GinciuuKitakaze/users/internal/adapter/repository/postgres/pool"
	core_errors "github.com/GinciuuKitakaze/users/internal/core/errors"
	core_logger "github.com/GinciuuKitakaze/users/internal/core/logger"
	core_model "github.com/GinciuuKitakaze/users/internal/core/model"
	repository_mapper "github.com/GinciuuKitakaze/users/internal/repository/mapper"
	repository_model "github.com/GinciuuKitakaze/users/internal/repository/model"
)

type Repository struct {
	pool   postgres_pool.Pool
	logger *core_logger.Logger
}

func NewRepository(pool postgres_pool.Pool, logger *core_logger.Logger) *Repository {
	return &Repository{
		pool:   pool,
		logger: logger,
	}
}

func (r *Repository) CreateUser(ctx context.Context, user core_model.CreateUser) (core_model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	INSERT INTO users (id, email, phone, name)
	VALUES ($1, $2, $3, $4)
	RETURNING id, version, email, phone, name, is_deleted, created_at, updated_at;`

	row := r.pool.QueryRow(ctx, query, user.ID, user.Email, user.Phone, user.Name)

	var userModel repository_model.User
	err := row.Scan(
		&userModel.ID,
		&userModel.Version,
		&userModel.Email,
		&userModel.Phone,
		&userModel.Name,
		&userModel.IsDeleted,
		&userModel.Created,
		&userModel.Updated,
	)
	if err != nil {
		if errors.Is(err, postgres_pool.ErrAlreadyExists) {
			return core_model.User{}, core_errors.ErrAlreadyExists
		}

		return core_model.User{}, fmt.Errorf("create user: %w", err)
	}

	return repository_mapper.ToDomain(userModel), nil
}

func (r *Repository) GetUser(ctx context.Context, userID uuid.UUID) (core_model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	SELECT id, version, email, phone, name, is_deleted, created_at, updated_at
	FROM users
	WHERE id = $1 AND is_deleted = $2;`

	row := r.pool.QueryRow(ctx, query, userID, false)

	var userModel repository_model.User
	err := row.Scan(
		&userModel.ID,
		&userModel.Version,
		&userModel.Email,
		&userModel.Phone,
		&userModel.Name,
		&userModel.IsDeleted,
		&userModel.Created,
		&userModel.Updated,
	)
	if err != nil {
		if errors.Is(err, postgres_pool.ErrNoRows) {
			return core_model.User{}, core_errors.ErrNotFound
		}
		return core_model.User{}, fmt.Errorf("scan user: %w", err)
	}

	return repository_mapper.ToDomain(userModel), nil
}

func (r *Repository) GetUsers(ctx context.Context, limit, offset int) ([]core_model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	SELECT id, version, email, phone, name, is_deleted, created_at, updated_at
	FROM users
	WHERE is_deleted = false
	LIMIT $1 OFFSET $2;`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("select users: %w", err)
	}
	defer rows.Close()

	var userModels []repository_model.User
	for rows.Next() {
		var userModel repository_model.User

		err := rows.Scan(
			&userModel.ID,
			&userModel.Version,
			&userModel.Email,
			&userModel.Phone,
			&userModel.Name,
			&userModel.IsDeleted,
			&userModel.Created,
			&userModel.Updated,
		)
		if err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		userModels = append(userModels, userModel)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows users: %w", err)
	}

	return repository_mapper.ToDomains(userModels), nil
}

func (r *Repository) GetUserByEmailOrPhone(ctx context.Context, emailOrPhone string) (core_model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	SELECT id, version, email, phone, name, is_deleted, created_at, updated_at
	FROM users
	WHERE (phone = $1 or email = $1) AND is_deleted = $2;`

	row := r.pool.QueryRow(ctx, query, emailOrPhone, false)

	var userModel repository_model.User
	err := row.Scan(
		&userModel.ID,
		&userModel.Version,
		&userModel.Email,
		&userModel.Phone,
		&userModel.Name,
		&userModel.IsDeleted,
		&userModel.Created,
		&userModel.Updated,
	)
	if err != nil {
		if errors.Is(err, postgres_pool.ErrNoRows) {
			return core_model.User{}, core_errors.ErrNotFound
		}
		return core_model.User{}, fmt.Errorf("scan user: %w", err)
	}

	return repository_mapper.ToDomain(userModel), nil
}

func (r *Repository) DeleteUser(ctx context.Context, userID uuid.UUID, user core_model.User) (core_model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	UPDATE users
	SET is_deleted = $1, version = version + 1, updated_at = NOW()
	WHERE id = $2 AND version = $3
	RETURNING id, version, email, phone, name, is_deleted, created_at, updated_at;`

	row := r.pool.QueryRow(ctx, query, true, userID, user.Version)

	var userModel repository_model.User
	err := row.Scan(
		&userModel.ID,
		&userModel.Version,
		&userModel.Email,
		&userModel.Phone,
		&userModel.Name,
		&userModel.IsDeleted,
		&userModel.Created,
		&userModel.Updated,
	)
	if err != nil {
		if errors.Is(err, postgres_pool.ErrNoRows) {
			return core_model.User{}, core_errors.ErrOptimisticLock
		}
		return core_model.User{}, fmt.Errorf("scan user: %w", err)
	}

	return repository_mapper.ToDomain(userModel), nil
}

func (r *Repository) UpdateUser(ctx context.Context, userID uuid.UUID, user core_model.User) (core_model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	UPDATE users
	SET version = version + 1, email = $1, phone = $2, name = $3, updated_at = NOW()
	WHERE id = $4 AND version = $5
	RETURNING id, version, email, phone, name, is_deleted, created_at, updated_at;`

	row := r.pool.QueryRow(ctx, query, user.Email, user.Phone, user.Name, userID, user.Version)

	var userModel repository_model.User
	err := row.Scan(
		&userModel.ID,
		&userModel.Version,
		&userModel.Email,
		&userModel.Phone,
		&userModel.Name,
		&userModel.IsDeleted,
		&userModel.Created,
		&userModel.Updated,
	)
	if err != nil {
		if errors.Is(err, postgres_pool.ErrNoRows) {
			return core_model.User{}, core_errors.ErrOptimisticLock
		}
		if errors.Is(err, postgres_pool.ErrAlreadyExists) {
			return core_model.User{}, core_errors.ErrAlreadyExists
		}
		return core_model.User{}, fmt.Errorf("update user: %w", err)
	}

	return repository_mapper.ToDomain(userModel), nil
}
