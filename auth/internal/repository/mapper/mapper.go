package repository_mapper

import (
	core_model "github.com/GinciuuKitakaze/auth/internal/core/model"
	postgres_model "github.com/GinciuuKitakaze/auth/internal/repository/model"
)

func ToDomainCredentials(credentials postgres_model.Credentials) core_model.Credentials {
	return core_model.Credentials{
		UserID:       credentials.UserID,
		PasswordHash: credentials.PasswordHash,
		Created:      credentials.Created,
		Updated:      credentials.Updated,
	}
}

func ToDomainRefreshToken(token postgres_model.RefreshToken) core_model.RefreshToken {
	return core_model.RefreshToken{
		ID:        token.ID,
		UserID:    token.UserID,
		TokenHash: token.TokenHash,
		ExpiresAt: token.ExpiresAt,
		RevokedAt: token.RevokedAt,
		CreatedAt: token.CreatedAt,
	}
}
