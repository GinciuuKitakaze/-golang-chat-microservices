package middleware

import (
	"context"
	"uuid"
)

// contextKey — собственный тип ключей context,
// чтобы избежать конфликтов с другими пакетами.
type contextKey string

const (
	userIDKey    contextKey = "userID"
	requestIDKey contextKey = "requestID"
)

// UserIDFromContext — возвращает ID авторизованного пользователя.
func UserIDFromContext(
	ctx context.Context,
) (uuid.UUID, bool) {
	userID, ok := ctx.Value(userIDKey).(uuid.UUID)

	return userID, ok
}

// RequestIDFromContext — возвращает request ID из context.
func RequestIDFromContext(
	ctx context.Context,
) string {
	requestID, _ := ctx.Value(requestIDKey).(string)

	return requestID
}
