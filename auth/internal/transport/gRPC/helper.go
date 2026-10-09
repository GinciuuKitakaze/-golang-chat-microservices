package grpctransport

import (
	"fmt"
	"uuid"

	core_errors "github.com/GinciuuKitakaze/auth/internal/core/errors"
)

func parseUserID(value string) (uuid.UUID, error) {
	userID, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil(), fmt.Errorf(
			"%w: invalid user id: %v",
			core_errors.ErrValidation,
			err,
		)
	}

	return userID, nil
}
