package users

import (
	"time"
	"uuid"
)

type User struct {
	ID             uuid.UUID
	Name           string
	LastSeen       time.Time
	Active         bool
	PassHash       string
	RefreshTokenID string
}
