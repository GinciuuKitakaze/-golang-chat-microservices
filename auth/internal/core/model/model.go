package core_model

import (
	"time"
	"uuid"
)

type Credentials struct {
	UserID       uuid.UUID
	PasswordHash string
	Created      time.Time
	Updated      time.Time
}

type CreateCredentials struct {
	UserID       uuid.UUID
	PasswordHash string
}

type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

type CreateRefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
}

type Register struct {
	UserID   uuid.UUID
	Password string
}

type Login struct {
	UserID   uuid.UUID
	Password string
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}
