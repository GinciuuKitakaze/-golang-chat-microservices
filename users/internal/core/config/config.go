package core_config

import (
	"log"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	GRPC     GRPCConfig
	Postgres PostgresConfig
	//	Redis    RedisConfig
	Logger LoggerConfig
}

type GRPCConfig struct {
	Host string `envconfig:"GRPC_HOST" required:"true" default:"localhost"`
	Port string `envconfig:"GRPC_PORT" required:"true" default:"50051"`
}

type PostgresConfig struct {
	Host     string        `envconfig:"POSTGRES_HOST" required:"true" default:"localhost"`
	Port     string        `envconfig:"POSTGRES_PORT" required:"true" default:"5432"`
	User     string        `envconfig:"POSTGRES_USER" required:"true" default:"postgres"`
	Password string        `envconfig:"POSTGRES_PASSWORD" required:"true" default:"postgres"`
	Database string        `envconfig:"POSTGRES_DB" required:"true" default:"postgres"`
	Timeout  time.Duration `envconfig:"POSTGRES_TIMEOUT" required:"true" default:"30s"`
}

//type RedisConfig struct {
//	RedisPassword string `envconfig:"REDIS_PASSWORD" required:"true" default:""`
//	RedisDB       int    `envconfig:"REDIS_DB" required:"true" default:"0"`
//}

type LoggerConfig struct {
	Level  string `envconfig:"LOGGER_LEVEL" required:"true" default:"DEBUG"`
	Folder string `envconfig:"LOGGER_FOLDER" required:"true" default:"logs"`
}

func Load() (*Config, error) {
	//Загружаем .env файл
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}
	var cfg Config

	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
