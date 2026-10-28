package response

// ErrorResponse — стандартный HTTP-ответ Gateway при ошибке.
type ErrorResponse struct {
	Message string `json:"message"`
}
