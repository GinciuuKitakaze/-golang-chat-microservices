package service

import (
	"context"
	"errors"
	"fmt"
	"uuid"

	core_logger "github.com/GinciuuKitakaze/gateway/internal/core/logger"
	core_model "github.com/GinciuuKitakaze/gateway/internal/core/model"
)

// GatewayService — основной сервис Gateway.
// Оркестрирует вызовы между Users Service и Auth Service.
type GatewayService struct {
	logger *core_logger.Logger

	usersService UsersService
	authService  AuthService
}

// NewGatewayService — создаёт GatewayService со всеми зависимостями.
func NewGatewayService(
	logger *core_logger.Logger,
	usersService UsersService,
	authService AuthService,
) *GatewayService {
	return &GatewayService{
		logger:       logger,
		usersService: usersService,
		authService:  authService,
	}
}

// UsersService — интерфейс клиента Users Service.
// Gateway зависит от интерфейса, а не от конкретной реализации gRPC-клиента.
type UsersService interface {
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

	GetUserByEmailOrPhone(
		ctx context.Context,
		emailOrPhone string,
	) (core_model.User, error)

	DeleteUser(
		ctx context.Context,
		userID uuid.UUID,
	) (core_model.User, error)

	UpdateUser(
		ctx context.Context,
		updatedUser core_model.UpdateUser,
	) (core_model.User, error)
}

// AuthService — интерфейс клиента Auth Service.
// Через него Gateway работает с регистрацией, логином и токенами.
type AuthService interface {
	Register(
		ctx context.Context,
		req core_model.AuthRegister,
	) error

	Login(
		ctx context.Context,
		req core_model.AuthLogin,
	) (core_model.TokenPair, error)

	Refresh(
		ctx context.Context,
		refreshToken string,
	) (core_model.TokenPair, error)

	Validate(
		ctx context.Context,
		accessToken string,
	) (uuid.UUID, error)

	Logout(
		ctx context.Context,
		req core_model.AuthLogout,
	) error

	DeleteUserAuthData(
		ctx context.Context,
		userID uuid.UUID,
	) error
}

// Register — регистрирует нового пользователя.
//
// Сначала создаём пользователя в Users Service.
// После получения userID создаём credentials в Auth Service.
// После успешной регистрации сразу выполняем Login и возвращаем токены.
func (s *GatewayService) Register(
	ctx context.Context,
	req core_model.Register,
) (core_model.User, core_model.TokenPair, error) {

	// Формируем данные пользователя без пароля.
	// Пароль должен уходить только в Auth Service.
	userRegister := core_model.User{
		Name:  req.Name,
		Email: req.Email,
		Phone: req.Phone,
	}

	// Создаём пользователя в Users Service.
	user, err := s.usersService.CreateUser(ctx, userRegister)
	if err != nil {
		return core_model.User{},
			core_model.TokenPair{},
			fmt.Errorf("create user: %w", err)
	}

	// После создания пользователя у нас появляется userID,
	// который используется как идентификатор credentials в Auth Service.
	authRegister := core_model.AuthRegister{
		UserID:   user.ID,
		Password: req.Password,
	}

	if err := s.authService.Register(ctx, authRegister); err != nil {
		// Пользователь в Users уже создан,
		// но Auth зарегистрировать его не смог.
		//
		// Пытаемся компенсировать предыдущую операцию
		// и удалить созданного пользователя.
		_, compensationErr := s.usersService.DeleteUser(ctx, user.ID)

		if compensationErr != nil {
			// Если упала ещё и компенсация,
			// возвращаем обе ошибки.
			return core_model.User{},
				core_model.TokenPair{},
				errors.Join(
					fmt.Errorf("register auth user: %w", err),
					fmt.Errorf(
						"rollback user %s failed: %w",
						user.ID,
						compensationErr,
					),
				)
		}

		return core_model.User{},
			core_model.TokenPair{},
			fmt.Errorf("register auth user: %w", err)
	}

	// После успешной регистрации сразу выполняем Login,
	// чтобы клиент получил access и refresh токены.
	authLogin := core_model.AuthLogin{
		UserID:   user.ID,
		Password: req.Password,
	}

	tokenPair, err := s.authService.Login(ctx, authLogin)
	if err != nil {
		// Здесь регистрацию уже не откатываем.
		//
		// Пользователь существует в Users Service,
		// credentials существуют в Auth Service.
		// Упал только автоматический Login.
		return core_model.User{},
			core_model.TokenPair{},
			fmt.Errorf("login after registration: %w", err)
	}

	return user, tokenPair, nil
}

// Login — авторизует пользователя по email или телефону.
//
// Сначала через Users Service получаем userID,
// затем передаём userID и пароль в Auth Service.
func (s *GatewayService) Login(
	ctx context.Context,
	req core_model.Login,
) (core_model.TokenPair, error) {

	// Auth Service работает с userID,
	// поэтому сначала находим пользователя по email или телефону.
	user, err := s.usersService.GetUserByEmailOrPhone(
		ctx,
		req.EmailOrPhone,
	)
	if err != nil {
		return core_model.TokenPair{}, fmt.Errorf(
			"get user by email or phone: %w",
			err,
		)
	}

	authLogin := core_model.AuthLogin{
		UserID:   user.ID,
		Password: req.Password,
	}

	// Получаем access и refresh токены из Auth Service.
	tokenPair, err := s.authService.Login(ctx, authLogin)
	if err != nil {
		return core_model.TokenPair{}, fmt.Errorf(
			"login: %w",
			err,
		)
	}

	return tokenPair, nil
}

// Logout — завершает пользовательскую сессию.
//
// Auth Service отзывает refresh token
// и добавляет access token в blacklist.
func (s *GatewayService) Logout(
	ctx context.Context,
	authLogout core_model.AuthLogout,
) error {

	if err := s.authService.Logout(ctx, authLogout); err != nil {
		return fmt.Errorf(
			"logout: %w",
			err,
		)
	}

	return nil
}

// Refresh — обновляет пару access/refresh токенов.
func (s *GatewayService) Refresh(
	ctx context.Context,
	refreshToken string,
) (core_model.TokenPair, error) {

	tokenPair, err := s.authService.Refresh(ctx, refreshToken)
	if err != nil {
		return core_model.TokenPair{}, fmt.Errorf(
			"refresh token: %w",
			err,
		)
	}

	return tokenPair, nil
}

// ValidateToken — проверяет access token.
//
// При успешной проверке Auth Service возвращает userID,
// которому принадлежит токен.
func (s *GatewayService) ValidateToken(
	ctx context.Context,
	accessToken string,
) (uuid.UUID, error) {

	userID, err := s.authService.Validate(ctx, accessToken)
	if err != nil {
		return uuid.Nil(), fmt.Errorf(
			"validate token: %w",
			err,
		)
	}

	return userID, nil
}

// CreateUser — создаёт пользователя напрямую через Users Service.
//
// Этот метод можно использовать отдельно от Register,
// например для внутренних или административных сценариев.
func (s *GatewayService) CreateUser(
	ctx context.Context,
	req core_model.User,
) (core_model.User, error) {

	user, err := s.usersService.CreateUser(ctx, req)
	if err != nil {
		return core_model.User{}, fmt.Errorf(
			"create user: %w",
			err,
		)
	}

	return user, nil
}

// GetUser — получает пользователя по ID.
func (s *GatewayService) GetUser(
	ctx context.Context,
	userID uuid.UUID,
) (core_model.User, error) {

	user, err := s.usersService.GetUser(ctx, userID)
	if err != nil {
		return core_model.User{}, fmt.Errorf(
			"get user: %w",
			err,
		)
	}

	return user, nil
}

// GetUsers — возвращает список пользователей с пагинацией.
func (s *GatewayService) GetUsers(
	ctx context.Context,
	limit,
	offset uint32,
) ([]core_model.User, error) {

	users, err := s.usersService.GetUsers(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf(
			"get users: %w",
			err,
		)
	}

	return users, nil
}

// DeleteUser — полностью удаляет пользовательские данные.
//
// Сначала удаляются credentials и refresh tokens из Auth Service,
// затем профиль пользователя удаляется из Users Service.
func (s *GatewayService) DeleteUser(
	ctx context.Context,
	userID uuid.UUID,
) (core_model.User, error) {

	// Сначала удаляем auth-данные пользователя.
	if err := s.authService.DeleteUserAuthData(ctx, userID); err != nil {
		return core_model.User{}, fmt.Errorf(
			"delete auth data: %w",
			err,
		)
	}

	// После удаления auth-данных удаляем профиль пользователя.
	user, err := s.usersService.DeleteUser(ctx, userID)
	if err != nil {
		// Здесь возможен partial failure:
		// auth-данные уже удалены, а профиль пользователя остался.
		//
		// Без Saga / Outbox нормально откатить эту операцию нельзя.
		return core_model.User{}, fmt.Errorf(
			"delete user after auth data was already deleted: %w",
			err,
		)
	}

	return user, nil
}

// UpdateUser — обновляет данные пользователя в Users Service.
func (s *GatewayService) UpdateUser(
	ctx context.Context,
	req core_model.UpdateUser,
) (core_model.User, error) {

	user, err := s.usersService.UpdateUser(ctx, req)
	if err != nil {
		return core_model.User{}, fmt.Errorf(
			"update user: %w",
			err,
		)
	}

	return user, nil
}
