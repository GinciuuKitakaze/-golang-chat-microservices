package core_errors

import "errors"

var (
	// ErrNotFound ошибка, когда сущность не найдена
	ErrNotFound = errors.New("not found")

	// ErrValidation ошибка валидации входных данных
	ErrValidation = errors.New("validation error")

	// ErrAlreadyExists ошибка, когда значение уже используется
	ErrAlreadyExists = errors.New("already exists")

	// ErrUnauthenticated ошибка, когда не удалось подтвердить личность
	ErrUnauthenticated = errors.New("unauthenticated")
)
