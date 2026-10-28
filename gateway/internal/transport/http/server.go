package httptransport

import (
	"context"
	"errors"
	"net/http"
	"uuid"

	core_logger "github.com/GinciuuKitakaze/gateway/internal/core/logger"
	core_model "github.com/GinciuuKitakaze/gateway/internal/core/model"
	"github.com/GinciuuKitakaze/gateway/internal/middleware"
)

// Server — HTTP transport Gateway.
type Server struct {
	gatewayService GatewayService
	logger         *core_logger.Logger
	server         *http.Server
}

// NewServer — создаёт HTTP transport Gateway.
func NewServer(
	gatewayService GatewayService,
	logger *core_logger.Logger,
	server *http.Server,
) *Server {
	return &Server{
		gatewayService: gatewayService,
		logger:         logger,
		server:         server,
	}
}

type GatewayService interface {
	Register(
		ctx context.Context,
		req core_model.Register,
	) (core_model.User, core_model.TokenPair, error)

	Login(
		ctx context.Context,
		req core_model.Login,
	) (core_model.TokenPair, error)

	Logout(
		ctx context.Context,
		authLogout core_model.AuthLogout,
	) error

	Refresh(
		ctx context.Context,
		refreshToken string,
	) (core_model.TokenPair, error)

	ValidateToken(
		ctx context.Context,
		accessToken string,
	) (uuid.UUID, error)

	CreateUser(
		ctx context.Context,
		req core_model.User,
	) (core_model.User, error)

	GetUser(
		ctx context.Context,
		userID uuid.UUID,
	) (core_model.User, error)

	GetUsers(
		ctx context.Context,
		limit,
		offset uint32,
	) ([]core_model.User, error)

	DeleteUser(
		ctx context.Context,
		userID uuid.UUID,
	) (core_model.User, error)

	UpdateUser(
		ctx context.Context,
		req core_model.UpdateUser,
	) (core_model.User, error)
}

// RegisterRoutes — регистрирует HTTP-маршруты Gateway.
func (s *Server) RegisterRoutes() {
	mux := http.NewServeMux()

	// Auth routes.
	mux.HandleFunc("POST /api/v1/auth/register", s.Register)
	mux.HandleFunc("POST /api/v1/auth/login", s.Login)
	mux.HandleFunc("POST /api/v1/auth/logout", s.Logout)
	mux.HandleFunc("POST /api/v1/auth/refresh", s.Refresh)

	// Users routes.
	mux.HandleFunc("GET /api/v1/users/{id}", s.GetUser)
	mux.HandleFunc("GET /api/v1/users", s.GetUsers)
	mux.HandleFunc("PATCH /api/v1/users/{id}", s.UpdateUser)
	mux.HandleFunc("DELETE /api/v1/users/{id}", s.DeleteUser)

	// Маршруты, которые доступны без access token.
	publicRoutes := []string{
		"POST:/api/v1/auth/register",
		"POST:/api/v1/auth/login",
		"POST:/api/v1/auth/refresh",
	}

	var handler http.Handler = mux

	// Middleware применяются снизу вверх.
	handler = middleware.JWTMiddleware(
		publicRoutes,
		s.gatewayService,
		s.logger,
	)(handler)

	handler = middleware.Recovery(
		s.logger,
	)(handler)

	handler = middleware.Logging(
		s.logger,
	)(handler)

	handler = middleware.RequestID()(handler)

	s.server.Handler = handler
}

// Run — запускает HTTP-сервер.
func (s *Server) Run() error {
	err := s.server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

// Shutdown — корректно останавливает HTTP-сервер.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
