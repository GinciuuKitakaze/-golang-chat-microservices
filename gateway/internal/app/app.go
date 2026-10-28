package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/GinciuuKitakaze/gateway/internal/client/auth"
	"github.com/GinciuuKitakaze/gateway/internal/client/users"
	core_config "github.com/GinciuuKitakaze/gateway/internal/core/config"
	core_logger "github.com/GinciuuKitakaze/gateway/internal/core/logger"
	"github.com/GinciuuKitakaze/gateway/internal/service"
	httptransport "github.com/GinciuuKitakaze/gateway/internal/transport/http"
	authpb "github.com/GinciuuKitakaze/messenger-contracts/auth/go"
	userspb "github.com/GinciuuKitakaze/messenger-contracts/users/go"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// App содержит основные компоненты приложения.
type App struct {
	cfg    *core_config.Config
	logger *core_logger.Logger

	httpServer *httptransport.Server

	authConn  *grpc.ClientConn
	usersConn *grpc.ClientConn
}

// NewApp создаёт и инициализирует приложение.
func NewApp() (*App, error) {
	// Загружаем конфигурацию.
	cfg, err := core_config.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	// Инициализируем общий логгер.
	logger, err := core_logger.NewLogger(cfg.Logger)
	if err != nil {
		return nil, fmt.Errorf("create logger: %w", err)
	}

	// Подключаемся к Users Service по gRPC.
	usersConn, err := grpc.NewClient(
		cfg.Users.GrpcHost,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("create users connection: %w", err)
	}

	// Создаём gRPC-клиент Users Service
	// и оборачиваем его в клиент Gateway.
	usersClient := userspb.NewUserServiceClient(usersConn)
	usersService := users.NewService(usersClient)

	// Подключаемся к Auth Service по gRPC.
	authConn, err := grpc.NewClient(
		cfg.Auth.GrpcHost,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		_ = usersConn.Close()

		return nil, fmt.Errorf("create auth connection: %w", err)
	}

	// Создаём gRPC-клиент Auth Service
	// и оборачиваем его в клиент Gateway.
	authClient := authpb.NewAuthServiceClient(authConn)
	authService := auth.NewService(authClient)

	// Создаём основной сервис Gateway,
	// который оркестрирует вызовы Users и Auth.
	gatewayService := service.NewGatewayService(
		logger,
		usersService,
		authService,
	)

	httpServer := &http.Server{
		Addr: cfg.HTTP.Host + ":" + cfg.HTTP.Port,

		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Создаём HTTP server Gateway.
	gatewayHTTPServer := httptransport.NewServer(
		gatewayService,
		logger,
		httpServer,
	)

	gatewayHTTPServer.RegisterRoutes()

	app := &App{
		cfg:        cfg,
		logger:     logger,
		httpServer: gatewayHTTPServer,
		authConn:   authConn,
		usersConn:  usersConn,
	}

	return app, nil
}

// Run — запускает приложение и ожидает сигнал завершения.
func (a *App) Run() error {

	listenAddr := fmt.Sprintf(
		"%s:%s",
		a.cfg.HTTP.Host,
		a.cfg.HTTP.Port,
	)

	a.logger.Info(
		"HTTP server listening",
		zap.String("addr", listenAddr),
	)

	// Запускаем HTTP server в отдельной goroutine,
	// чтобы основной поток мог ожидать системный сигнал.
	errCh := make(chan error, 1)

	go func() {
		if err := a.httpServer.Run(); err != nil {
			errCh <- err
		}
	}()

	// Ожидаем сигнал завершения приложения.
	quit := make(chan os.Signal, 1)

	signal.Notify(
		quit,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	defer signal.Stop(quit)

	select {
	case <-quit:
		a.logger.Info("shutdown signal received")
		return nil

	case err := <-errCh:
		return fmt.Errorf("HTTP server stopped: %w", err)
	}
}

// Shutdown — корректно завершает работу приложения.
func (a *App) Shutdown(ctx context.Context) error {
	a.logger.Info("shutting down application...")

	// Даём текущим HTTP-запросам корректно завершиться.
	if err := a.httpServer.Shutdown(ctx); err != nil {
		a.logger.Error(
			"shutdown HTTP server",
			zap.Error(err),
		)

		return fmt.Errorf("shutdown HTTP server: %w", err)
	}

	// Закрываем соединение с Auth Service.
	if err := a.authConn.Close(); err != nil {
		a.logger.Error(
			"close auth connection",
			zap.Error(err),
		)

		return fmt.Errorf("close auth connection: %w", err)
	}

	// Закрываем соединение с Users Service.
	if err := a.usersConn.Close(); err != nil {
		a.logger.Error(
			"close users connection",
			zap.Error(err),
		)

		return fmt.Errorf("close users connection: %w", err)
	}

	// Логгер закрываем последним,
	// чтобы до этого момента можно было записывать ошибки shutdown.
	a.logger.Close()

	return nil
}
