package middleware

import (
	"net/http"
	"time"

	core_logger "github.com/GinciuuKitakaze/gateway/internal/core/logger"
	"go.uber.org/zap"
)

// Logging — логирует HTTP-запрос,
// статус ответа и время выполнения.
func Logging(
	logger *core_logger.Logger,
) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(
			rw http.ResponseWriter,
			r *http.Request,
		) {
			start := time.Now()

			writer := &responseWriter{
				ResponseWriter: rw,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(writer, r)

			fields := []zap.Field{
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", writer.statusCode),
				zap.Duration(
					"latency",
					time.Since(start),
				),
				zap.String(
					"request_id",
					RequestIDFromContext(r.Context()),
				),
			}

			switch {
			case writer.statusCode >= http.StatusInternalServerError:
				logger.Error(
					"HTTP request",
					fields...,
				)

			case writer.statusCode >= http.StatusBadRequest:
				logger.Warn(
					"HTTP request",
					fields...,
				)

			default:
				logger.Debug(
					"HTTP request",
					fields...,
				)
			}
		})
	}
}
