package auth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetCurrentSession(t *testing.T) {
	s := Session{
		Username: "test",
		RealName: "test test",
		Expires:  int(time.Now().Add(4 * time.Hour).Unix()),
	}

	ctx := ContextWithSession(context.Background(), &s)

	assert.Equal(t, s, *GetCurrentSession(ctx))
	assert.Nil(t, GetCurrentSession(context.Background()))
}

func TestSession_QueryString(t *testing.T) {
	testCases := []struct {
		name     string
		session  Session
		expected string
	}{
		{"minimal", Session{Username: "alice"}, "adm=false&exp=0&uid=alice"},
		{"maximal",
			Session{IsAdmin: true, Expires: 1746996601, RealName: "root", Username: "admin"},
			"adm=true&exp=1746996601&name=root&uid=admin",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			qs := tc.session.QueryString()
			assert.Equal(t, tc.expected, qs)
		})
	}
}

func TestParseSession(t *testing.T) {
	testCases := []struct {
		name        string
		cookieValue string
		expected    Session
	}{
		{"minimal", "adm=false&uid=alice", Session{Username: "alice"}},
		{"maximal", "adm=true&exp=1746996601&name=root&uid=admin",
			Session{Username: "admin", Expires: 1746996601, IsAdmin: true, RealName: "root"},
		},
		{"empty", "", Session{}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := ParseSession(tc.cookieValue)
			assert.Equal(t, tc.expected, *actual)
		})
	}
}
