package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	core_logger "github.com/GinciuuKitakaze/gateway/internal/core/logger"
	http_response "github.com/GinciuuKitakaze/gateway/internal/transport/http/response"
	"go.uber.org/zap"
)

// Recovery — перехватывает panic во время HTTP-запроса
// и не позволяет приложению завершиться целиком.
func Recovery(
	logger *core_logger.Logger,
) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(
			rw http.ResponseWriter,
			r *http.Request,
		) {
			responseHandler := http_response.NewHTTPResponseHandler(
				logger,
				rw,
			)

			defer func() {
				if recovered := recover(); recovered != nil {
					logger.Error(
						"panic recovered",
						zap.String("method", r.Method),
						zap.String("path", r.URL.Path),
						zap.String(
							"request_id",
							RequestIDFromContext(r.Context()),
						),
						zap.Any("panic", recovered),
						zap.ByteString("stack", debug.Stack()),
					)

					responseHandler.ClientErrorResponse(
						fmt.Errorf(
							"unexpected panic: %v",
							recovered,
						),
						"internal server error",
						http.StatusInternalServerError,
					)
				}
			}()

			next.ServeHTTP(rw, r)
		})
	}
}
