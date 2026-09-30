package repository_model

import (
	"time"
	"uuid"
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
