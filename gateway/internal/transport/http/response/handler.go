package response

import (
	"encoding/json"
	"net/http"

	core_logger "github.com/GinciuuKitakaze/gateway/internal/core/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HTTPResponseHandler — формирует единообразные HTTP-ответы Gateway.
type HTTPResponseHandler struct {
	logger *core_logger.Logger
	rw     http.ResponseWriter
}

// NewHTTPResponseHandler — создаёт обработчик HTTP-ответов.
func NewHTTPResponseHandler(
	logger *core_logger.Logger,
	rw http.ResponseWriter,
) *HTTPResponseHandler {
	return &HTTPResponseHandler{
		logger: logger,
		rw:     rw,
	}
}

// JSONResponse — отправляет JSON-ответ с указанным HTTP статусом.
func (h *HTTPResponseHandler) JSONResponse(
	responseBody any,
	statusCode int,
) {
	h.rw.Header().Set("Content-Type", "application/json")
	h.rw.WriteHeader(statusCode)

	if err := json.NewEncoder(h.rw).Encode(responseBody); err != nil {
		h.logger.Error(
			"write HTTP response",
			zap.Error(err),
		)
	}
}

// NoContentResponse — отправляет HTTP-ответ без body.
func (h *HTTPResponseHandler) NoContentResponse() {
	h.rw.WriteHeader(http.StatusNoContent)
}

// ErrorResponse — преобразует ошибку внутреннего сервиса
// в соответствующий HTTP статус и возвращает ответ клиенту.
func (h *HTTPResponseHandler) ErrorResponse(
	err error,
	message string,
) {
	statusCode := mapErrorToHTTPStatus(err)

	// Ожидаемые клиентские ошибки логируем как Warn.
	switch statusCode {
	case http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusNotFound,
		http.StatusConflict,
		http.StatusTooManyRequests:

		h.logger.Warn(
			message,
			zap.Error(err),
		)

	default:
		h.logger.Error(
			message,
			zap.Error(err),
		)
	}

	h.JSONResponse(
		ErrorResponse{
			Message: message,
		},
		statusCode,
	)
}

// ClientErrorResponse — отправляет ошибку,
// которая возникла при обработке HTTP-запроса внутри Gateway.
func (h *HTTPResponseHandler) ClientErrorResponse(
	err error,
	message string,
	statusCode int,
) {
	h.logger.Warn(
		message,
		zap.Error(err),
	)

	h.JSONResponse(
		ErrorResponse{
			Message: message,
		},
		statusCode,
	)
}

// mapErrorToHTTPStatus — преобразует gRPC статус
// внутреннего сервиса в HTTP статус Gateway.
func mapErrorToHTTPStatus(err error) int {
	switch status.Code(err) {
	case codes.InvalidArgument:
		return http.StatusBadRequest

	case codes.Unauthenticated:
		return http.StatusUnauthorized

	case codes.PermissionDenied:
		return http.StatusForbidden

	case codes.NotFound:
		return http.StatusNotFound

	case codes.AlreadyExists:
		return http.StatusConflict

	case codes.ResourceExhausted:
		return http.StatusTooManyRequests

	default:
		return http.StatusInternalServerError
	}
}
