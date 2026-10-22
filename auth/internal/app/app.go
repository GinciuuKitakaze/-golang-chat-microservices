package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	pgx_pool "github.com/GinciuuKitakaze/auth/internal/adapter/repository/postgres/pool/pgx"
	goredis_pool "github.com/GinciuuKitakaze/auth/internal/adapter/repository/redis/goredis"
	core_config "github.com/GinciuuKitakaze/auth/internal/core/config"
	core_logger "github.com/GinciuuKitakaze/auth/internal/core/logger"
	"github.com/GinciuuKitakaze/auth/internal/interceptor"
	postgres_repository "github.com/GinciuuKitakaze/auth/internal/repository/postgres"
	"github.com/GinciuuKitakaze/auth/internal/service"
	grpctransport "github.com/GinciuuKitakaze/auth/internal/transport/gRPC"
	authpb "github.com/GinciuuKitakaze/messenger-contracts/auth/go"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
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
	redisPool  *goredis_pool.Pool
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

	logger.Debug("postgres connection pool created")

	// Инициализируем пул соединений с Redis.
	redisPool, err := goredis_pool.NewPool(ctx, cfg.Redis)
	if err != nil {
		logger.Error(
			"Failed to create redis connection pool",
			zap.Error(err),
		)

		pool.Close()
		logger.Close()

		return nil, fmt.Errorf("create redis pool: %w", err)
	}

	logger.Debug("redis connection pool created")

	logger.Debug(
		"initializing service",
		zap.String("service", "auth"),
	)

	// Инициализируем репозиторий.
	authRepository := postgres_repository.NewRepository(pool, logger)

	// Инициализируем сервис.
	authService := service.NewAuthService(
		authRepository,
		redisPool,
		logger,
		cfg.Token,
	)

	// Инициализируем transport.
	authServer := grpctransport.NewServer(authService, logger)

	// Создаём интерсептор
	serverInterceptor := interceptor.NewServerInterceptor(logger)

	// Создаём gRPC сервер.
	gprcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(serverInterceptor.Recovery(),
			serverInterceptor.RequestID(),
			serverInterceptor.Logging()),
	)

	// Регистрируем AuthService в gRPC сервере.
	authpb.RegisterAuthServiceServer(gprcServer, authServer)

	app := &App{
		cfg:        cfg,
		logger:     logger,
		grpcServer: gprcServer,
		pool:       pool,
		redisPool:  redisPool,
	}

	// Запускаем миграции.
	if err := app.runMigrations(ctx); err != nil {
		logger.Error(
			"Failed to run migrations",
			zap.Error(err),
		)

		redisPool.Close()
		pool.Close()
		logger.Close()

		return nil, fmt.Errorf("run migrations: %w", err)
	}
	return app, nil
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
		if err := a.grpcServer.Serve(listener); err != nil {
			if !errors.Is(err, grpc.ErrServerStopped) {
				a.logger.Error(
					"gRPC server stopped",
					zap.Error(err),
				)
			}
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

func (a *App) runMigrations(ctx context.Context) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set postgres dialect: %w", err)
	}

	dbGoose := stdlib.OpenDBFromPool(a.pool.Pool)
	defer dbGoose.Close()

	if err := goose.UpContext(ctx, dbGoose, "./migrations"); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	return nil
}

// Shutdown корректно завершает работу приложения
func (a *App) Shutdown() {
	a.logger.Info("Shutting down gRPC server...")

	// Даём текущим запросам завершиться.
	a.grpcServer.GracefulStop()

	// Закрываем listener.
	if a.listener != nil {
		if err := a.listener.Close(); err != nil {
			a.logger.Error(
				"Failed to close listener",
				zap.Error(err),
			)
		}
	}

	// Закрываем PostgreSQL.
	a.pool.Close()

	// Закрываем Redis.
	a.redisPool.Close()

	// Закрываем логгер.
	a.logger.Close()
}
