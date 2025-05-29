package crypto

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashPassword(t *testing.T) {
	testCases := []struct {
		name     string
		password string
	}{
		{"password", "password"},
		{"empty string", ""},
		{"very long", strings.Repeat("a", 100)},
	}

	for _, tc := range testCases {
		t.Run(tc.password, func(t *testing.T) {
			h, err := HashPassword(tc.password)

			assert.NoError(t, err)
			assert.NoError(t, VerifyPassword(tc.password, h))
		})
	}
}

func TestVerifyPassword(t *testing.T) {
	testCases := []struct {
		name        string
		password    string
		hash        string
		expectedErr error
	}{
		{"valid", "password", mustHashPassword("password"), nil},
		{"empty string", "", mustHashPassword(""), nil},
		{"collision", strings.Repeat("a", 72), mustHashPassword(strings.Repeat("a", 72) + "bcdef"), nil},

		{"incorrect", "wrong!", mustHashPassword("password"), ErrMismatchedHashAndPassword},
	}

	for _, tc := range testCases {
		t.Run(tc.password, func(t *testing.T) {
			err := VerifyPassword(tc.password, tc.hash)

			assert.Equal(t, tc.expectedErr, err)
		})
	}
}
