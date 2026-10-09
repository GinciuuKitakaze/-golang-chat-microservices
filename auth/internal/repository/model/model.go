package postgres_model

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

type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}
