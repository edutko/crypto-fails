package store

import (
	"errors"
)

var ErrNotFound = errors.New("not found")
var ErrInvalidKey = errors.New("invalid key")
