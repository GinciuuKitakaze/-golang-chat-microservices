package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"uuid"

	core_logger "github.com/GinciuuKitakaze/gateway/internal/core/logger"
	http_response "github.com/GinciuuKitakaze/gateway/internal/transport/http/response"
)

// Middleware — HTTP middleware Gateway.
type Middleware func(http.Handler) http.Handler

// TokenValidator — сервис для проверки access token.
type TokenValidator interface {
	ValidateToken(
		ctx context.Context,
		accessToken string,
	) (uuid.UUID, error)
}

// JWTMiddleware — проверяет access token
// для всех маршрутов, кроме публичных.
func JWTMiddleware(
	publicRoutesList []string,
	service TokenValidator,
	logger *core_logger.Logger,
) Middleware {

	publicRoutes := make(
		map[string]struct{},
		len(publicRoutesList),
	)

	for _, route := range publicRoutesList {
		publicRoutes[route] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(
			rw http.ResponseWriter,
			r *http.Request,
		) {
			responseHandler := http_response.NewHTTPResponseHandler(
				logger,
				rw,
			)

			key := r.Method + ":" + r.URL.Path

			if _, ok := publicRoutes[key]; ok {
				next.ServeHTTP(rw, r)
				return
			}

			authHeader := r.Header.Get(authorizationHeader)

			if authHeader == "" {
				responseHandler.ClientErrorResponse(
					errors.New("authorization header is missing"),
					"unauthorized",
					http.StatusUnauthorized,
				)
				return
			}

			parts := strings.Fields(authHeader)

			if len(parts) != 2 ||
				!strings.EqualFold(parts[0], "Bearer") {

				responseHandler.ClientErrorResponse(
					errors.New("invalid authorization header"),
					"unauthorized",
					http.StatusUnauthorized,
				)
				return
			}

			userID, err := service.ValidateToken(
				r.Context(),
				parts[1],
			)
			if err != nil {
				responseHandler.ErrorResponse(
					err,
					"unauthorized",
				)
				return
			}

			ctx := context.WithValue(
				r.Context(),
				userIDKey,
				userID,
			)

			next.ServeHTTP(
				rw,
				r.WithContext(ctx),
			)
		})
	}
}
