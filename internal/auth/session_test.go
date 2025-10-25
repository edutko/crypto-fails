package auth

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetSessionFromRequest(t *testing.T) {
	testCases := []struct {
		name string
		req  *http.Request
		sess Session
		err  error
	}{
		{"valid cookie",
			&http.Request{Header: http.Header{"Cookie": []string{"auth=test"}}}, Session{Username: "test"}, nil,
		},
		{"valid token",
			&http.Request{Header: http.Header{"Authorization": []string{"Bearer test"}}}, Session{Username: "test"}, nil,
		},
		{"error parsing cookie",
			&http.Request{Header: http.Header{"Cookie": []string{"auth=test"}}}, anonymousSession, ErrInvalidCookie,
		},
		{"error parsing token",
			&http.Request{Header: http.Header{"Authorization": []string{"Bearer test"}}}, anonymousSession, ErrInvalidToken,
		},
		{"empty cookie",
			&http.Request{Header: http.Header{"Cookie": []string{""}}}, anonymousSession, nil,
		},
		{"empty token",
			&http.Request{Header: http.Header{"Authorization": []string{""}}}, anonymousSession, nil,
		},
		{"invalid cookie",
			&http.Request{Header: http.Header{"Cookie": []string{"auth=test"}}}, anonymousSession, nil,
		},
		{"invalid token",
			&http.Request{Header: http.Header{"Authorization": []string{"Bearer test"}}}, anonymousSession, nil,
		},
		{"no cookie or token",
			&http.Request{}, anonymousSession, nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parseCookie = func(c *http.Cookie) (Session, error) { return tc.sess, tc.err }
			parseToken = func(token string) (Session, error) { return tc.sess, tc.err }

			s, err := GetSessionFromRequest(tc.req)

			assert.Equal(t, tc.sess, s)
			assert.ErrorIs(t, err, tc.err)
		})
	}
}

func TestGetSessionFromContext(t *testing.T) {
	s := Session{
		Username: "test",
		RealName: "test test",
		Expires:  int(time.Now().Add(4 * time.Hour).Unix()),
	}

	ctx := ContextWithSession(context.Background(), s)

	assert.Equal(t, s, GetSessionFromContext(ctx))
	assert.Equal(t, anonymousSession, GetSessionFromContext(context.Background()))
}

func TestSession_IsAuthenticated(t *testing.T) {
	assert.False(t, anonymousSession.IsAuthenticated())
	assert.True(t, Session{Username: "foo"}.IsAuthenticated())
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
			assert.Equal(t, tc.expected, actual)
		})
	}
}
