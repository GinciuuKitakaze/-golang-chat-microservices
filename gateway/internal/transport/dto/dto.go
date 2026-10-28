package dto

import "uuid"

// RegisterRequest — данные для регистрации пользователя.
type RegisterRequest struct {
	Name     string  `json:"name" validate:"required,min=3,max=100"`
	Email    *string `json:"email" validate:"required_without=Phone,omitempty,email,max=100"`
	Phone    *string `json:"phone" validate:"required_without=Email,omitempty,min=10,max=20,startswith=+"`
	Password string  `json:"password" validate:"required,min=8,max=64"`
}

// RegisterResponse — ответ после успешной регистрации.
// Возвращает созданного пользователя и пару токенов.
type RegisterResponse struct {
	User      UserResponse      `json:"user"`
	TokenPair TokenPairResponse `json:"token_pair"`
}

// UserRequest — данные для создания пользователя.
type UserRequest struct {
	Name  string  `json:"name" validate:"required,min=3,max=100"`
	Email *string `json:"email" validate:"required_without=Phone,omitempty,email,max=100"`
	Phone *string `json:"phone" validate:"required_without=Email,omitempty,min=10,max=20,startswith=+"`
}

// UserResponse — пользователь в HTTP-ответе.
type UserResponse struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email *string   `json:"email"`
	Phone *string   `json:"phone"`
}

// TokenPairResponse — access и refresh токены.
type TokenPairResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// LoginRequest — данные для авторизации пользователя.
type LoginRequest struct {
	EmailOrPhone string `json:"email_or_phone" validate:"required,min=3,max=100"`
	Password     string `json:"password" validate:"required,min=8,max=64"`
}

// AuthLogoutRequest — токены для завершения пользовательской сессии.
type AuthLogoutRequest struct {
	AccessToken  string `json:"access_token" validate:"required"`
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshTokenRequest — refresh token для получения новой пары токенов.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// UpdateUserRequest — изменяемые данные пользователя.
// ID пользователя передаётся отдельно через path.
type UpdateUserRequest struct {
	Name  string  `json:"name" validate:"required,min=3,max=100"`
	Email *string `json:"email" validate:"required_without=Phone,omitempty,email,max=100"`
	Phone *string `json:"phone" validate:"required_without=Email,omitempty,min=10,max=20,startswith=+"`
}

// AccessTokenRequest — access token для проверки валидности.
//
// Позже этот DTO может стать не нужен,
// когда проверка токена полностью уйдёт в middleware.
type AccessTokenRequest struct {
	AccessToken string `json:"access_token" validate:"required"`
}
