package auth

import (
	"errors"

	"github.com/edutko/crypto-fails/internal/crypto"
	"github.com/edutko/crypto-fails/internal/stores"
	"github.com/edutko/crypto-fails/pkg/user"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrIncorrectPassword = errors.New("incorrect password")
)

func AuthenticateWithPassword(username, password string) (*user.User, error) {
	u, err := userStore().Get(username)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if err = crypto.VerifyPassword(password, u.PasswordHash); err != nil {
		return nil, ErrIncorrectPassword
	}

	u = u.WithoutSecrets()
	return &u, nil
}

var userStore = stores.UserStore
