package core_model

import "uuid"

// Register — общая модель регистрации пользователя через Gateway.
type Register struct {
	Name     string
	Email    *string
	Phone    *string
	Password string
}

// User — пользователь внутри Gateway.
type User struct {
	ID    uuid.UUID
	Name  string
	Email *string
	Phone *string
}

// UpdateUser — данные для обновления пользователя.
type UpdateUser struct {
	ID    uuid.UUID
	Name  string
	Email *string
	Phone *string
}

// AuthRegister — данные для регистрации пользователя в Auth Service.
type AuthRegister struct {
	UserID   uuid.UUID
	Password string
}

// AuthLogin — данные для авторизации пользователя в Auth Service.
type AuthLogin struct {
	UserID   uuid.UUID
	Password string
}

// AuthLogout — данные для выхода пользователя из системы.
type AuthLogout struct {
	AccessToken  string
	RefreshToken string
}

// TokenPair — пара access и refresh токенов.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

// Login — данные для авторизации пользователя через Gateway.
type Login struct {
	EmailOrPhone string
	Password     string
}
