package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUser_WithoutSecrets(t *testing.T) {
	testCases := []struct {
		name     string
		user     User
		expected User
	}{
		{"minimal",
			User{
				Username: "test",
				Password: "abc",
			},
			User{
				Username: "test",
			},
		},
		{"maximal",
			User{
				Username:     "test",
				Email:        "user@example.com",
				Password:     "abc",
				PasswordHash: "dfsafdsaf",
				RealName:     "test user",
				Roles:        []string{"foo", "bar"},
			},
			User{
				Username: "test",
				Email:    "user@example.com",
				RealName: "test user",
				Roles:    []string{"foo", "bar"},
			},
		},
		{"empty", User{}, User{}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.user.WithoutSecrets())
		})
	}
}
