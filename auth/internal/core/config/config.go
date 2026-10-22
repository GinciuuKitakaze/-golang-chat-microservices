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
	Logger   LoggerConfig
	Token    TokenConfig
	Redis    RedisConfig
}

type GRPCConfig struct {
	Host string `envconfig:"GRPC_HOST" default:"localhost"`
	Port string `envconfig:"GRPC_PORT" default:"50052"`
}

type PostgresConfig struct {
	Host     string        `envconfig:"POSTGRES_HOST" default:"localhost"`
	Port     string        `envconfig:"POSTGRES_PORT" default:"5432"`
	User     string        `envconfig:"POSTGRES_USER" default:"postgres"`
	Password string        `envconfig:"POSTGRES_PASSWORD" default:"postgres"`
	Database string        `envconfig:"POSTGRES_DB" default:"postgres"`
	Timeout  time.Duration `envconfig:"POSTGRES_TIMEOUT" default:"30s"`
}

type LoggerConfig struct {
	Level  string `envconfig:"LOGGER_LEVEL" default:"DEBUG"`
	Folder string `envconfig:"LOGGER_FOLDER" default:"logs"`
}

type TokenConfig struct {
	JwtSecret             string `envconfig:"JWT_SECRET" required:"true"`
	AccessTokenTTLMinutes int    `envconfig:"ACCESS_TOKEN_TTL_MINUTES" default:"60"`
	RefreshTokenTTLDays   int    `envconfig:"REFRESH_TOKEN_TTL_DAYS" default:"30"`
}

type RedisConfig struct {
	Host     string        `envconfig:"REDIS_HOST" default:"localhost"`
	Port     string        `envconfig:"REDIS_PORT" default:"6379"`
	Password string        `envconfig:"REDIS_PASSWORD" default:"redis"`
	DB       int           `envconfig:"REDIS_DB" default:"0"`
	TTL      time.Duration `envconfig:"REDIS_TTL" default:"5m"`
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
