package dbtools

import (
	"errors"
)

var (
	ErrDb     = errors.New("db error")
	ErrNoRows = errors.New("not found")
)
