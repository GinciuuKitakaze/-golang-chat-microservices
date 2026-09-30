package main

import (
	"context"
	"fmt"
	"log"
	"net"

	accountpb "github.com/GinciuuKitakaze/messenger-contracts/users/go"
	"github.com/GinciuuKitakaze/users/internal/adapter/repository/postgres/pool/pgx"
	core_config "github.com/GinciuuKitakaze/users/internal/core/config"
	core_logger "github.com/GinciuuKitakaze/users/internal/core/logger"
	postgres_repository "github.com/GinciuuKitakaze/users/internal/repository/postgres"
	"github.com/GinciuuKitakaze/users/internal/service"
	grpctransport "github.com/GinciuuKitakaze/users/internal/transport/gRPC"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()
	//Загружаем конфигурацию
	cfg, err := core_config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}
	//Инициализируем общий логгер
	logger, err := core_logger.NewLogger(cfg.Logger)
	if err != nil {
		log.Fatal("Failed to create logger:", err)
	}

	logger.Debug("initializing postgres connection pool")

	pool, err := pgx_pool.NewPool(ctx, cfg.Postgres)
	if err != nil {
		logger.Fatal("Failed to create postgres connection pool", zap.Error(err))
	}
	defer pool.Close()

	logger.Debug("initializing service", zap.String("service", "user"))
	// Инициализируем репозиторий
	usersRepository := postgres_repository.NewRepository(pool, logger)
	// Инициализируем сервис
	userService := service.NewUserService(usersRepository, logger)
	// Инициализируем сервер
	server := grpctransport.NewServer(userService, logger)
	// Создание gRPC сервера
	s := grpc.NewServer()
	accountpb.RegisterUserServiceServer(s, server)

	listenAddr := fmt.Sprintf("%s:%s", cfg.GRPC.Host, cfg.GRPC.Port)
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		logger.Fatal("Failed to listen", zap.String("addr", listenAddr), zap.Error(err))
	}

	logger.Info(fmt.Sprintf("gRPC server listening %s", listenAddr))
	if err = s.Serve(lis); err != nil {
		logger.Fatal("Failed to serve", zap.Error(err))
		return
	}
}
