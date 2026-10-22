package grpctransport

import (
	"context"
	"uuid"

	core_logger "github.com/GinciuuKitakaze/auth/internal/core/logger"
	core_model "github.com/GinciuuKitakaze/auth/internal/core/model"
	transport_mapper "github.com/GinciuuKitakaze/auth/internal/transport/mapper"
	authpb "github.com/GinciuuKitakaze/messenger-contracts/auth/go"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Server реализует gRPC AuthService.
type Server struct {
	authpb.UnimplementedAuthServiceServer

	authService AuthService
	logger      *core_logger.Logger
}

// NewServer создаёт новый gRPC сервер Auth Service.
func NewServer(
	authService AuthService,
	logger *core_logger.Logger,
) *Server {
	return &Server{
		authService: authService,
		logger:      logger,
	}
}

// AuthService описывает методы service-слоя,
// которые используются gRPC transport.
type AuthService interface {
	Register(ctx context.Context, req core_model.Register) error
	Login(ctx context.Context, req core_model.Login) (core_model.TokenPair, error)
	Refresh(ctx context.Context, refreshToken string) (core_model.TokenPair, error)
	Validate(ctx context.Context, accessToken string) (uuid.UUID, error)
	Logout(ctx context.Context, req core_model.Logout) error
	DeleteUserAuthData(ctx context.Context, userID uuid.UUID) error
}

// Register создаёт credentials пользователя.
func (s *Server) Register(
	ctx context.Context,
	req *authpb.RegisterRequest,
) (*emptypb.Empty, error) {

	// Парсим user_id из protobuf.
	userID, err := parseUserID(req.UserId)
	if err != nil {
		return &emptypb.Empty{}, toGRPCError(err)
	}

	// Преобразуем protobuf модель в доменную.
	register := transport_mapper.ToDomainRegister(req, userID)

	// Регистрируем credentials пользователя.
	if err := s.authService.Register(ctx, register); err != nil {
		return &emptypb.Empty{}, toGRPCError(err)
	}

	return &emptypb.Empty{}, nil
}

// Login выполняет аутентификацию пользователя.
func (s *Server) Login(
	ctx context.Context,
	req *authpb.LoginRequest,
) (*authpb.LoginResponse, error) {

	// Парсим user_id из protobuf.
	userID, err := parseUserID(req.UserId)
	if err != nil {
		return &authpb.LoginResponse{}, toGRPCError(err)
	}

	// Преобразуем protobuf модель в доменную.
	login := transport_mapper.ToDomainLogin(req, userID)

	// Проверяем credentials и получаем пару токенов.
	tokenPair, err := s.authService.Login(ctx, login)
	if err != nil {
		return &authpb.LoginResponse{}, toGRPCError(err)
	}

	// Преобразуем доменную модель токенов в protobuf.
	return transport_mapper.ToProtoLogin(tokenPair), nil
}

// Refresh обновляет пару токенов по действующему refresh-токену.
func (s *Server) Refresh(
	ctx context.Context,
	req *authpb.RefreshRequest,
) (*authpb.RefreshResponse, error) {

	// Обновляем пару токенов.
	tokenPair, err := s.authService.Refresh(
		ctx,
		req.RefreshToken,
	)
	if err != nil {
		return &authpb.RefreshResponse{}, toGRPCError(err)
	}

	// Преобразуем доменную модель токенов в protobuf.
	return transport_mapper.ToProtoRefresh(tokenPair), nil
}

// Validate проверяет access-токен и возвращает user_id.
func (s *Server) Validate(
	ctx context.Context,
	req *authpb.ValidateRequest,
) (*authpb.ValidateResponse, error) {

	// Проверяем access-токен.
	userID, err := s.authService.Validate(
		ctx,
		req.AccessToken,
	)
	if err != nil {
		return &authpb.ValidateResponse{}, toGRPCError(err)
	}

	// Формируем protobuf ответ.
	return transport_mapper.ToProtoValidate(
		userID.String(),
	), nil
}

// Logout завершает сессию пользователя,
// отзывая refresh-токен и добавляя access-токен в blacklist.
func (s *Server) Logout(
	ctx context.Context,
	req *authpb.LogoutRequest,
) (*emptypb.Empty, error) {
	// Преобразуем protobuf модель в доменную.
	logout := transport_mapper.ToDomainLogout(req)

	// Завершаем пользовательскую сессию.
	if err := s.authService.Logout(ctx, logout); err != nil {
		return &emptypb.Empty{}, toGRPCError(err)
	}

	return &emptypb.Empty{}, nil
}

// DeleteUserAuthData удаляет авторизационные данные пользователя.
func (s *Server) DeleteUserAuthData(
	ctx context.Context,
	req *authpb.DeleteRequest,
) (*emptypb.Empty, error) {

	// Парсим user_id из protobuf.
	userID, err := parseUserID(req.UserId)
	if err != nil {
		return &emptypb.Empty{}, toGRPCError(err)
	}

	// Удаляем credentials и связанные refresh-токены.
	if err := s.authService.DeleteUserAuthData(
		ctx,
		userID,
	); err != nil {
		return &emptypb.Empty{}, toGRPCError(err)
	}

	return &emptypb.Empty{}, nil
}
