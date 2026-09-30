package core_model

import (
	"fmt"
	"net/mail"
	"strings"
	"time"
	"uuid"

	core_errors "github.com/GinciuuKitakaze/users/internal/core/errors"
	"github.com/nyaruka/phonenumbers/v2"
)

type User struct {
	ID        uuid.UUID
	Version   int64
	Name      string
	Email     *string
	Phone     *string
	IsDeleted bool
	Created   time.Time
	Updated   time.Time
}

func (u *User) ApplyUpdate(update UpdateUser) error {
	u.Name = strings.TrimSpace(update.Name)

	if update.Email != nil {
		email := strings.TrimSpace(*update.Email)

		if email == "" {
			u.Email = nil
		} else {
			u.Email = &email
		}
	}

	if update.Phone != nil {
		phone := strings.TrimSpace(*update.Phone)

		if phone == "" {
			u.Phone = nil
		} else {
			u.Phone = &phone
		}
	}

	if u.Email == nil && u.Phone == nil {
		return fmt.Errorf("%w: email or phone is required", core_errors.ErrValidation)
	}

	return nil
}

type CreateUser struct {
	ID    uuid.UUID
	Name  string
	Email *string
	Phone *string
}

func (u CreateUser) Validate() error {
	name := strings.TrimSpace(u.Name)

	if name == "" {
		return fmt.Errorf("%w: name is required", core_errors.ErrValidation)
	}
	if len(name) > 255 {
		return fmt.Errorf("%w: name is too long", core_errors.ErrValidation)
	}
	if u.Email == nil && u.Phone == nil {
		return fmt.Errorf("%w: email or phone is required", core_errors.ErrValidation)
	}
	if u.Email != nil {
		email := strings.TrimSpace(*u.Email)

		if email == "" {
			return fmt.Errorf("%w: email cannot be empty", core_errors.ErrValidation)
		}
		if len(email) > 255 {
			return fmt.Errorf("%w: email is too long", core_errors.ErrValidation)
		}
		if _, err := mail.ParseAddress(email); err != nil {
			return fmt.Errorf("%w: invalid email", core_errors.ErrValidation)
		}
	}
	if u.Phone != nil {
		phone := strings.TrimSpace(*u.Phone)
		if phone == "" {
			return fmt.Errorf("%w: phone cannot be empty", core_errors.ErrValidation)
		}
		if len(phone) > 32 {
			return fmt.Errorf("%w: phone is too long", core_errors.ErrValidation)
		}
		number, err := phonenumbers.Parse(phone, "RU")
		if err != nil || !phonenumbers.IsValidNumber(number) {
			return fmt.Errorf("%w: invalid phone number", core_errors.ErrValidation)
		}
	}
	return nil
}

type UpdateUser struct {
	Name  string
	Email *string
	Phone *string
}

func (u UpdateUser) Validate() error {
	name := strings.TrimSpace(u.Name)

	if name == "" {
		return fmt.Errorf("%w: name is required", core_errors.ErrValidation)
	}

	if len(name) > 255 {
		return fmt.Errorf("%w: name is too long", core_errors.ErrValidation)
	}

	if u.Email != nil {
		email := strings.TrimSpace(*u.Email)

		if email != "" {
			if len(email) > 255 {
				return fmt.Errorf("%w: email is too long", core_errors.ErrValidation)
			}

			if _, err := mail.ParseAddress(email); err != nil {
				return fmt.Errorf("%w: invalid email", core_errors.ErrValidation)
			}
		}
	}

	if u.Phone != nil {
		phone := strings.TrimSpace(*u.Phone)

		if phone != "" {
			if len(phone) > 32 {
				return fmt.Errorf("%w: phone is too long", core_errors.ErrValidation)
			}

			number, err := phonenumbers.Parse(phone, "RU")
			if err != nil || !phonenumbers.IsValidNumber(number) {
				return fmt.Errorf("%w: invalid phone", core_errors.ErrValidation)
			}
		}
	}

	return nil
}
