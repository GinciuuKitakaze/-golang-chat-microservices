package core_errors

import "errors"

var (
	//ErrOptimisticLock ошибка при конфликте версий сущности
	ErrOptimisticLock = errors.New("optimistic lock conflict")

	// ErrNotFound ошибка, когда сущность не найдена
	ErrNotFound = errors.New("not found")

	// ErrValidation ошибка валидации входных данных
	ErrValidation = errors.New("validation error")

	// ErrAlreadyExists ошибка, когда значение уже используется
	ErrAlreadyExists = errors.New("already exists")
)
