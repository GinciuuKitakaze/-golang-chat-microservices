package service

import (
	"fmt"

	core_errors "github.com/GinciuuKitakaze/auth/internal/core/errors"
)

func validatePassword(password string) error {
	passwordLen := len([]byte(password))

	if passwordLen < 8 {
		return fmt.Errorf(
			"%w: password is too short",
			core_errors.ErrValidation,
		)
	}

	if passwordLen > 72 {
		return fmt.Errorf(
			"%w: password is too long",
			core_errors.ErrValidation,
		)
	}

	return nil
}
