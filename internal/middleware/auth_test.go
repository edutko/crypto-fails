package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/edutko/crypto-fails/internal/auth"
)

func TestGetCurrentSession(t *testing.T) {
	expected := auth.Session{Username: "test"}
	r := &http.Request{}
	r = r.WithContext(context.WithValue(context.Background(), "session", expected))

	actual := GetCurrentSession(r)

	assert.Equal(t, expected, actual)
}

func TestAuthenticated(t *testing.T) {
	testCases := []struct {
		name               string
		req                *http.Request
		sess               auth.Session
		err                error
		expectedStatusCode int
	}{
		{"valid cookie",
			&http.Request{Header: http.Header{"Cookie": []string{"auth=test"}}}, auth.Session{Username: "test"}, nil, 200,
		},
		{"valid token",
			&http.Request{Header: http.Header{"Authorization": []string{"Bearer test"}}}, auth.Session{Username: "test"}, nil, 200,
		},
		{"no cookie or token",
			&http.Request{}, auth.Session{}, nil, 401,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			getSessionFromRequest = func(r *http.Request) (auth.Session, error) { return tc.sess, tc.err }
			w := httptest.NewRecorder()

			var m mockHandler
			if tc.sess.IsAuthenticated() {
				m.On("handle", w, tc.req.WithContext(auth.ContextWithSession(tc.req.Context(), tc.sess)))
			}

			Authenticated(m.handle)(w, tc.req)

			m.AssertExpectations(t)
			assert.Equal(t, tc.expectedStatusCode, w.Result().StatusCode)
		})
	}
}

func TestMaybeAuthenticated(t *testing.T) {
	testCases := []struct {
		name string
		req  *http.Request
		sess auth.Session
		err  error
	}{
		{"valid cookie",
			&http.Request{Header: http.Header{"Cookie": []string{"auth=test"}}}, auth.Session{Username: "test"}, nil,
		},
		{"valid token",
			&http.Request{Header: http.Header{"Authorization": []string{"Bearer test"}}}, auth.Session{Username: "test"}, nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			getSessionFromRequest = func(r *http.Request) (auth.Session, error) { return tc.sess, tc.err }
			w := httptest.NewRecorder()

			var m mockHandler
			if tc.sess.IsAuthenticated() {
				m.On("handle", w, tc.req.WithContext(auth.ContextWithSession(tc.req.Context(), tc.sess)))
			}

			MaybeAuthenticated(m.handle)(w, tc.req)

			m.AssertExpectations(t)
		})
	}
}

func TestRequireAdmin(t *testing.T) {
	testCases := []struct {
		name               string
		req                *http.Request
		sess               auth.Session
		err                error
		expectedStatusCode int
	}{
		{"admin cookie",
			&http.Request{Header: http.Header{"Cookie": []string{"auth=test"}}}, auth.Session{Username: "test", IsAdmin: true}, nil, 200,
		},
		{"admin token",
			&http.Request{Header: http.Header{"Authorization": []string{"Bearer test"}}}, auth.Session{Username: "test", IsAdmin: true}, nil, 200,
		},
		{"non-admin cookie",
			&http.Request{Header: http.Header{"Cookie": []string{"auth=test"}}}, auth.Session{Username: "test"}, nil, 403,
		},
		{"non-admin token",
			&http.Request{Header: http.Header{"Authorization": []string{"Bearer test"}}}, auth.Session{Username: "test"}, nil, 403,
		},
		{"no cookie or token",
			&http.Request{}, auth.Session{}, nil, 401,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			getSessionFromRequest = func(r *http.Request) (auth.Session, error) { return tc.sess, tc.err }
			w := httptest.NewRecorder()

			var m mockHandler
			if tc.expectedStatusCode == 200 {
				m.On("handle", w, tc.req.WithContext(auth.ContextWithSession(tc.req.Context(), tc.sess)))
			}

			RequireAdmin(m.handle)(w, tc.req)

			m.AssertExpectations(t)
			assert.Equal(t, tc.expectedStatusCode, w.Result().StatusCode)
		})
	}
}

type mockHandler struct {
	mock.Mock
}

func (m *mockHandler) handle(w http.ResponseWriter, r *http.Request) {
	m.Called(w, r)
}
