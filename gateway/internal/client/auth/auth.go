package auth

import (
	"context"
	"fmt"
	"uuid"

	core_model "github.com/GinciuuKitakaze/gateway/internal/core/model"
	authpb "github.com/GinciuuKitakaze/messenger-contracts/auth/go"
)

type Service struct {
	client authpb.AuthServiceClient
}

func NewService(client authpb.AuthServiceClient) *Service {
	return &Service{
		client: client,
	}
}

// Register — регистрация пользователя в Auth Service.
func (s *Service) Register(
	ctx context.Context,
	req core_model.AuthRegister,
) error {

	_, err := s.client.Register(ctx, &authpb.RegisterRequest{
		UserId:   req.UserID.String(),
		Password: req.Password,
	})
	if err != nil {
		return fmt.Errorf("register failed: %w", err)
	}

	return nil
}

// Login — авторизация пользователя и получение пары токенов.
func (s *Service) Login(
	ctx context.Context,
	req core_model.AuthLogin,
) (core_model.TokenPair, error) {

	res, err := s.client.Login(ctx, &authpb.LoginRequest{
		UserId:   req.UserID.String(),
		Password: req.Password,
	})
	if err != nil {
		return core_model.TokenPair{}, fmt.Errorf(
			"login failed: %w",
			err,
		)
	}

	tokenPair := res.GetTokenPair()

	return core_model.TokenPair{
		AccessToken:  tokenPair.GetAccessToken(),
		RefreshToken: tokenPair.GetRefreshToken(),
	}, nil
}

// Refresh — обновление пары токенов по refresh-токену.
func (s *Service) Refresh(
	ctx context.Context,
	refreshToken string,
) (core_model.TokenPair, error) {

	res, err := s.client.Refresh(ctx, &authpb.RefreshRequest{
		RefreshToken: refreshToken,
	})
	if err != nil {
		return core_model.TokenPair{}, fmt.Errorf(
			"refresh failed: %w",
			err,
		)
	}

	tokenPair := res.GetTokenPair()

	return core_model.TokenPair{
		AccessToken:  tokenPair.GetAccessToken(),
		RefreshToken: tokenPair.GetRefreshToken(),
	}, nil
}

// Validate — проверка access-токена и получение userID.
func (s *Service) Validate(
	ctx context.Context,
	accessToken string,
) (uuid.UUID, error) {

	res, err := s.client.Validate(ctx, &authpb.ValidateRequest{
		AccessToken: accessToken,
	})
	if err != nil {
		return uuid.Nil(), fmt.Errorf(
			"validate failed: %w",
			err,
		)
	}

	userID, err := uuid.Parse(res.GetUserId())
	if err != nil {
		return uuid.Nil(), fmt.Errorf(
			"parse user id: %w",
			err,
		)
	}

	return userID, nil
}

// Logout — отзыв access и refresh токенов.
func (s *Service) Logout(
	ctx context.Context,
	req core_model.AuthLogout,
) error {

	_, err := s.client.Logout(ctx, &authpb.LogoutRequest{
		AccessToken:  req.AccessToken,
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		return fmt.Errorf("logout failed: %w", err)
	}

	return nil
}

// DeleteUserAuthData — удаление данных пользователя из Auth Service.
func (s *Service) DeleteUserAuthData(
	ctx context.Context,
	userID uuid.UUID,
) error {

	_, err := s.client.DeleteUserAuthData(ctx, &authpb.DeleteRequest{
		UserId: userID.String(),
	})
	if err != nil {
		return fmt.Errorf(
			"delete user auth data failed: %w",
			err,
		)
	}

	return nil
}
