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
				Username:     "test",
				Password:     "abc",
				PasswordHash: "dfsafdsaf",
			},
			User{
				Username: "test",
			},
		},
		{"maximal",
			User{
				Username:     "test",
				Password:     "abc",
				PasswordHash: "dfsafdsaf",
				Roles:        []string{"foo", "bar"},
				RealName:     "test user",
			},
			User{
				Username: "test",
				Roles:    []string{"foo", "bar"},
				RealName: "test user",
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
