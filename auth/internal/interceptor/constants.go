package interceptor

import "context"

type ctxKey string

const requestIDKey ctxKey = "x-request-id"

func RequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}
