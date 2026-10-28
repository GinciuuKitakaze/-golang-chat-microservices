package middleware

import (
	"context"
	"net/http"
	"uuid"
)

// RequestID — добавляет request ID в header и context запроса.
func RequestID() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(
			rw http.ResponseWriter,
			r *http.Request,
		) {
			requestID := r.Header.Get(requestIDHeader)

			if requestID == "" {
				requestID = uuid.New().String()
			}

			rw.Header().Set(
				requestIDHeader,
				requestID,
			)

			ctx := context.WithValue(
				r.Context(),
				requestIDKey,
				requestID,
			)

			next.ServeHTTP(
				rw,
				r.WithContext(ctx),
			)
		})
	}
}
