package app

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	accountpb "github.com/GinciuuKitakaze/messenger-contracts/users/go"
	pgx_pool "github.com/GinciuuKitakaze/users/internal/adapter/repository/postgres/pool/pgx"
	core_config "github.com/GinciuuKitakaze/users/internal/core/config"
	core_logger "github.com/GinciuuKitakaze/users/internal/core/logger"
	"github.com/GinciuuKitakaze/users/internal/interceptor"
	postgres_repository "github.com/GinciuuKitakaze/users/internal/repository/postgres"
	"github.com/GinciuuKitakaze/users/internal/service"
	grpctransport "github.com/GinciuuKitakaze/users/internal/transport/gRPC"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// App содержит основные компоненты приложения.
type App struct {
	cfg        *core_config.Config
	logger     *core_logger.Logger
	grpcServer *grpc.Server
	listener   net.Listener
	pool       *pgx_pool.Pool
}

// NewApp создаёт и инициализирует приложение.
func NewApp(ctx context.Context) (*App, error) {

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

	logger.Debug("initializing postgres connection pool")

	// Инициализируем пул соединений с PostgreSQL.
	pool, err := pgx_pool.NewPool(ctx, cfg.Postgres)
	if err != nil {
		logger.Error(
			"Failed to create postgres connection pool",
			zap.Error(err),
		)
		logger.Close()

		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	logger.Debug(
		"initializing service",
		zap.String("service", "user"),
	)

	// Инициализируем репозиторий.
	usersRepository := postgres_repository.NewRepository(pool, logger)

	// Инициализируем сервис.
	userService := service.NewUserService(usersRepository, logger)

	// Инициализируем transport.
	userServer := grpctransport.NewServer(userService, logger)

	// Создаём интерсептор
	serverInterceptor := interceptor.NewServerInterceptor(logger)

	// Создаём gRPC сервер.
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(serverInterceptor.Recovery(),
			serverInterceptor.RequestID(),
			serverInterceptor.Logging()),
	)

	// Регистрируем UserService в gRPC сервере.
	accountpb.RegisterUserServiceServer(grpcServer, userServer)

	return &App{
		cfg:        cfg,
		logger:     logger,
		grpcServer: grpcServer,
		pool:       pool,
	}, nil
}

// Run запускает приложение и ожидает сигнал завершения.
func (a *App) Run() error {

	// Формируем адрес gRPC сервера.
	listenAddr := fmt.Sprintf(
		"%s:%s",
		a.cfg.GRPC.Host,
		a.cfg.GRPC.Port,
	)

	// Создаём TCP listener.
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", listenAddr, err)
	}

	a.listener = listener

	a.logger.Info(
		"gRPC server listening",
		zap.String("addr", listenAddr),
	)

	// Запускаем gRPC сервер.
	go func() {
		if err := a.grpcServer.Serve(a.listener); err != nil {
			a.logger.Error(
				"gRPC server stopped",
				zap.Error(err),
			)
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

	<-quit

	return nil
}

// Shutdown корректно завершает работу приложения.
func (a *App) Shutdown() {
	a.logger.Info("Shutting down gRPC server...")

	// Даём текущим запросам завершиться.
	a.grpcServer.GracefulStop()

	// Закрываем listener.
	if err := a.listener.Close(); err != nil {
		a.logger.Error(
			"Failed to close listener",
			zap.Error(err),
		)
	}

	// Закрываем PostgreSQL.
	a.pool.Close()

	// Закрываем логгер.
	a.logger.Close()
}
