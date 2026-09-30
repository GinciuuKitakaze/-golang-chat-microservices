package service

import (
	"context"
	"fmt"
	"uuid"

	core_logger "github.com/GinciuuKitakaze/users/internal/core/logger"
	core_model "github.com/GinciuuKitakaze/users/internal/core/model"
)

type UserService struct {
	repo   Repository
	logger *core_logger.Logger
}

func NewUserService(repo Repository, logger *core_logger.Logger) *UserService {
	return &UserService{
		repo:   repo,
		logger: logger,
	}
}

type Repository interface {
	CreateUser(ctx context.Context, user core_model.CreateUser) (core_model.User, error)
	GetUser(ctx context.Context, userID uuid.UUID) (core_model.User, error)
	GetUsers(ctx context.Context, limit, offset int) ([]core_model.User, error)
	GetUserByEmailOrPhone(ctx context.Context, emailOrPhone string) (core_model.User, error)
	DeleteUser(ctx context.Context, userID uuid.UUID, user core_model.User) (core_model.User, error)
	UpdateUser(ctx context.Context, userID uuid.UUID, user core_model.User) (core_model.User, error)
}

func (s *UserService) CreateUser(ctx context.Context, user core_model.CreateUser) (core_model.User, error) {
	if err := user.Validate(); err != nil {
		return core_model.User{}, err
	}
	user.ID = uuid.NewV7()

	return s.repo.CreateUser(ctx, user)
}

func (s *UserService) GetUser(ctx context.Context, userID uuid.UUID) (core_model.User, error) {
	user, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		return core_model.User{}, fmt.Errorf("get user %v: %w", userID, err)
	}
	return user, nil
}

func (s *UserService) GetUsers(ctx context.Context, limit, offset int) ([]core_model.User, error) {
	return s.repo.GetUsers(ctx, limit, offset)
}
func (s *UserService) GetUserByEmailOrPhone(ctx context.Context, emailOrPhone string) (core_model.User, error) {
	return s.repo.GetUserByEmailOrPhone(ctx, emailOrPhone)
}

func (s *UserService) DeleteUser(ctx context.Context, userID uuid.UUID) (core_model.User, error) {
	user, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		return core_model.User{}, fmt.Errorf("get user %v: %w", userID, err)
	}

	userResponse, err := s.repo.DeleteUser(ctx, userID, user)
	if err != nil {
		return core_model.User{}, fmt.Errorf("delete user %v: %w", userID, err)
	}
	return userResponse, nil
}

func (s *UserService) UpdateUser(ctx context.Context, userID uuid.UUID, userUpdate core_model.UpdateUser) (core_model.User, error) {
	if err := userUpdate.Validate(); err != nil {
		return core_model.User{}, err
	}

	user, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		return core_model.User{}, fmt.Errorf("get user %v: %w", userID, err)
	}

	if err := user.ApplyUpdate(userUpdate); err != nil {
		return core_model.User{}, fmt.Errorf("apply update user %v: %w", userID, err)
	}

	userResponse, err := s.repo.UpdateUser(ctx, userID, user)
	if err != nil {
		return core_model.User{}, fmt.Errorf("update user %v: %w", userID, err)
	}

	return userResponse, nil
}
