package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"
	"uuid"

	redis_pool "github.com/GinciuuKitakaze/auth/internal/adapter/repository/redis"
	core_config "github.com/GinciuuKitakaze/auth/internal/core/config"
	core_errors "github.com/GinciuuKitakaze/auth/internal/core/errors"
	core_logger "github.com/GinciuuKitakaze/auth/internal/core/logger"
	core_model "github.com/GinciuuKitakaze/auth/internal/core/model"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo        Repository
	redis       redis_pool.Pool
	logger      *core_logger.Logger
	tokenConfig core_config.TokenConfig
}

func NewAuthService(repo Repository, redis redis_pool.Pool, logger *core_logger.Logger, tokenConfig core_config.TokenConfig) *AuthService {
	return &AuthService{
		repo:        repo,
		redis:       redis,
		logger:      logger,
		tokenConfig: tokenConfig,
	}
}

type Repository interface {
	CreateCredentials(ctx context.Context, credentials core_model.CreateCredentials) error
	GetCredentialsByUserID(ctx context.Context, userID uuid.UUID) (core_model.Credentials, error)

	SaveRefreshToken(ctx context.Context, token core_model.CreateRefreshToken) error
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
	ConsumeRefreshToken(ctx context.Context, tokenHash string) (uuid.UUID, error)

	DeleteUserAuthData(ctx context.Context, userID uuid.UUID) error
}

// Register - регистрация нового пользователя
func (s *AuthService) Register(ctx context.Context,
	req core_model.Register,
) error {

	if req.UserID == uuid.Nil() {
		return fmt.Errorf("%w: user id is required", core_errors.ErrValidation)
	}

	if err := validatePassword(req.Password); err != nil {
		return err
	}

	// Хешируем пароль с помощью bcrypt
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	createCredentials := core_model.CreateCredentials{
		UserID:       req.UserID,
		PasswordHash: string(hash),
	}
	return s.repo.CreateCredentials(ctx, createCredentials)
}

// Login - аутентификация пользователя
func (s *AuthService) Login(ctx context.Context,
	req core_model.Login,
) (core_model.TokenPair, error) {

	if req.UserID == uuid.Nil() {
		return core_model.TokenPair{}, core_errors.ErrUnauthenticated
	}

	// Ищем Credentials по id
	credentials, err := s.repo.GetCredentialsByUserID(ctx, req.UserID)
	if err != nil {
		if errors.Is(err, core_errors.ErrNotFound) {
			return core_model.TokenPair{}, core_errors.ErrUnauthenticated
		}

		return core_model.TokenPair{}, fmt.Errorf(
			"get credentials by user id: %w",
			err,
		)
	}

	key := "login:attempts:" + req.UserID.String()

	data, err := s.redis.Get(ctx, key).Bytes()
	if err != nil {
		if !errors.Is(err, redis_pool.NotFound) {
			return core_model.TokenPair{}, fmt.Errorf(
				"get login attempts: %w",
				err,
			)
		}
	} else {
		attempts, err := strconv.Atoi(string(data))
		if err != nil {
			return core_model.TokenPair{}, fmt.Errorf(
				"parse login attempts: %w",
				err,
			)
		}

		if attempts >= 5 {
			return core_model.TokenPair{}, core_errors.ErrUnauthenticated
		}
	}

	// Проверяем,что введённый пароль совпадает с хешем в бд
	if err := bcrypt.CompareHashAndPassword(
		[]byte(credentials.PasswordHash),
		[]byte(req.Password),
	); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {

			cmd := s.redis.Incr(ctx, key)
			if err := cmd.Err(); err != nil {
				return core_model.TokenPair{}, fmt.Errorf(
					"increment login attempts: %w",
					err,
				)
			}

			attempts := cmd.Val()

			if attempts == 1 {
				if err := s.redis.Expire(ctx, key, time.Minute*5).Err(); err != nil {
					return core_model.TokenPair{}, fmt.Errorf(
						"set login attempts ttl: %w",
						err,
					)
				}
			}

			return core_model.TokenPair{}, core_errors.ErrUnauthenticated
		}

		return core_model.TokenPair{}, fmt.Errorf(
			"compare password: %w",
			err,
		)
	}

	// Пароль правильный — сбрасываем счётчик неудачных попыток.
	if err := s.redis.Del(ctx, key).Err(); err != nil {
		return core_model.TokenPair{}, fmt.Errorf(
			"delete login attempts: %w",
			err,
		)
	}

	// Генерируем пару токенов (access + refresh)
	return s.issueTokens(ctx, credentials.UserID)
}

// Refresh — обновление пары токенов по действующему refresh-токену
func (s *AuthService) Refresh(
	ctx context.Context,
	refreshToken string,
) (core_model.TokenPair, error) {

	// Хешируем refresh-токен для поиска в БД
	hash := sha256.Sum256([]byte(refreshToken))
	refreshHash := hex.EncodeToString(hash[:])

	userID, err := s.repo.ConsumeRefreshToken(ctx, refreshHash)
	if err != nil {
		if errors.Is(err, core_errors.ErrNotFound) {
			return core_model.TokenPair{}, core_errors.ErrUnauthenticated
		}

		return core_model.TokenPair{}, err
	}

	return s.issueTokens(ctx, userID)
}

// Validate - проверка access-токена и извлечение userID
func (s *AuthService) Validate(
	ctx context.Context,
	accessToken string,
) (uuid.UUID, error) {

	// Создаём мапу для парсинга claims
	claims := jwt.MapClaims{}

	// Парсим токен, проверяем подпись и срок действия
	_, err := jwt.ParseWithClaims(
		accessToken,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			return []byte(s.tokenConfig.JwtSecret), nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		return uuid.Nil(), core_errors.ErrUnauthenticated
	}

	//Формируем ключ для проверки access-токена в blacklist.
	hash := sha256.Sum256([]byte(accessToken))
	key := "blacklist:access:" + hex.EncodeToString(hash[:])

	//Проверяем, находится ли access-токен в blacklist.
	_, err = s.redis.Get(ctx, key).Bytes()
	if err == nil {
		return uuid.Nil(), core_errors.ErrUnauthenticated
	}

	if !errors.Is(err, redis_pool.NotFound) {
		return uuid.Nil(), fmt.Errorf(
			"check access token blacklist: %w",
			err,
		)
	}

	// Извлекаем userID
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return uuid.Nil(), core_errors.ErrUnauthenticated
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil(), core_errors.ErrUnauthenticated
	}

	return userID, nil
}

func (s *AuthService) Logout(
	ctx context.Context,
	logout core_model.Logout,
) error {

	// Парсим access-токен, проверяем подпись и срок действия.
	claims := jwt.MapClaims{}

	_, err := jwt.ParseWithClaims(
		logout.AccessToken,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			return []byte(s.tokenConfig.JwtSecret), nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		return core_errors.ErrUnauthenticated
	}

	// Получаем время истечения access-токена.
	expiration, err := claims.GetExpirationTime()
	if err != nil || expiration == nil {
		return core_errors.ErrUnauthenticated
	}

	// Вычисляем, сколько ещё должен жить access-токен.
	ttl := time.Until(expiration.Time)
	if ttl <= 0 {
		return core_errors.ErrUnauthenticated
	}

	// Формируем ключ blacklist по хешу access-токена.
	hashAccessToken := sha256.Sum256([]byte(logout.AccessToken))
	blacklistKey := "blacklist:access:" + hex.EncodeToString(hashAccessToken[:])

	// Хешируем refresh-токен для поиска в PostgreSQL.
	hashRefreshToken := sha256.Sum256([]byte(logout.RefreshToken))
	refreshHash := hex.EncodeToString(hashRefreshToken[:])

	// Отзываем refresh-токен.
	if err := s.repo.RevokeRefreshToken(ctx, refreshHash); err != nil {
		if errors.Is(err, core_errors.ErrNotFound) {
			return core_errors.ErrUnauthenticated
		}

		return err
	}

	// Добавляем access-токен в blacklist до момента его естественного истечения.
	if err := s.redis.Set(
		ctx,
		blacklistKey,
		"1",
		ttl,
	).Err(); err != nil {
		return fmt.Errorf("blacklist access token: %w", err)
	}

	return nil
}

// DeleteUserAuthData — удаление авторизационных данных пользователя.
func (s *AuthService) DeleteUserAuthData(
	ctx context.Context,
	userID uuid.UUID,
) error {

	return s.repo.DeleteUserAuthData(ctx, userID)
}

func (s *AuthService) issueTokens(
	ctx context.Context,
	userID uuid.UUID,
) (core_model.TokenPair, error) {

	now := time.Now()

	// Вычисляем время истечения access-токена
	accessExp := now.Add(time.Duration(s.tokenConfig.AccessTokenTTLMinutes) * time.Minute)

	// Формируем claims для access-токена
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     accessExp.Unix(),
	}

	// Создаём и подписываем access-токен
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessStr, err := accessToken.SignedString([]byte(s.tokenConfig.JwtSecret))
	if err != nil {
		return core_model.TokenPair{}, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Генерируем refresh-токен
	refreshBytes := make([]byte, 32)
	if _, err := rand.Read(refreshBytes); err != nil {
		return core_model.TokenPair{},
			fmt.Errorf("generate refresh token: %w", err)
	}

	refreshStr := base64.RawURLEncoding.EncodeToString(refreshBytes)
	// Хешируем refresh-токен перед сохранением в БД
	h := sha256.Sum256([]byte(refreshStr))
	refreshHash := hex.EncodeToString(h[:])

	// Сохраняем hash refresh-токена в БД
	refresh := core_model.CreateRefreshToken{
		ID:        uuid.NewV7(),
		UserID:    userID,
		TokenHash: refreshHash,
		ExpiresAt: now.Add(time.Duration(s.tokenConfig.RefreshTokenTTLDays) * time.Hour * 24),
	}
	if err := s.repo.SaveRefreshToken(ctx, refresh); err != nil {
		return core_model.TokenPair{}, fmt.Errorf("failed to save refresh token: %w", err)
	}

	return core_model.TokenPair{
		AccessToken:  accessStr,
		RefreshToken: refreshStr,
	}, nil
}
