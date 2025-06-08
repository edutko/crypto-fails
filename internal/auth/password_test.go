package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/edutko/crypto-fails/internal/crypto"
	"github.com/edutko/crypto-fails/internal/store/kv"
	"github.com/edutko/crypto-fails/pkg/user"
)

func TestAuthenticateWithPassword(t *testing.T) {
	testCases := []struct {
		name         string
		username     string
		password     string
		expectedUser *user.User
		expectedErr  error
	}{
		{"correct password", "test", "password", &user.User{Username: "test"}, nil},
		{"incorrect password", "test", "WRONG!", nil, ErrIncorrectPassword},
		{"nonexistent user", "whodis", "password", nil, ErrUserNotFound},
	}

	userStore = mockUserStore

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			u, err := AuthenticateWithPassword(tc.username, tc.password)

			assert.Equal(t, tc.expectedErr, err)
			assert.Equal(t, tc.expectedUser, u)
		})
	}
}

func mockUserStore() kv.Store[user.User] {
	s := kv.NewInMemoryStore[user.User]()

	ph, _ := crypto.HashPassword("password")
	_ = s.Put("test", user.User{
		Username:     "test",
		PasswordHash: ph,
	})

	return s
}
