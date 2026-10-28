package core_config

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	HTTP   HTTPConfig
	Logger LoggerConfig
	Auth   AuthConfig
	Users  UsersConfig
}

type HTTPConfig struct {
	Host string `envconfig:"HTTP_HOST" default:"localhost"`
	Port string `envconfig:"HTTP_PORT" default:"50053"`
}

type AuthConfig struct {
	GrpcHost string `envconfig:"AUTH_GRPC_HOST" default:"localhost:50052"`
}

type UsersConfig struct {
	GrpcHost string `envconfig:"USERS_GRPC_HOST" default:"localhost:50051"`
}

type LoggerConfig struct {
	Level  string `envconfig:"LOGGER_LEVEL" default:"DEBUG"`
	Folder string `envconfig:"LOGGER_FOLDER" default:"logs"`
}

func Load() (*Config, error) {
	// Загружаем .env файл.
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	var cfg Config

	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
