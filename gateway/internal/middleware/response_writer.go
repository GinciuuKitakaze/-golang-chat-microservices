package middleware

import "net/http"

// responseWriter — оборачивает http.ResponseWriter
// и сохраняет HTTP статус ответа.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader — сохраняет HTTP статус ответа.
func (rw *responseWriter) WriteHeader(
	statusCode int,
) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}
