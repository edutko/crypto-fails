package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/edutko/crypto-fails/internal/auth"
	"github.com/edutko/crypto-fails/internal/responses"
)

func GetCurrentSession(r *http.Request) *auth.Session {
	return auth.GetCurrentSession(r.Context())
}

func Authenticated(next func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		s, err := getCurrentSession(r)
		if err != nil {
			responses.BadRequest(w, err)
			return
		}
		if s == nil {
			responses.Unauthorized(w)
			return
		}
		next(w, r.WithContext(auth.ContextWithSession(r.Context(), s)))
	}
}

func MaybeAuthenticated(next func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		s, _ := getCurrentSession(r)
		if s != nil {
			next(w, r.WithContext(auth.ContextWithSession(r.Context(), s)))
			return
		}
		next(w, r)
	}
}

func RequireAdmin(next func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return Authenticated(func(w http.ResponseWriter, r *http.Request) {
		if s := GetCurrentSession(r); s != nil && s.IsAdmin {
			next(w, r)
		} else {
			responses.Forbidden(w, fmt.Errorf("%q is not an admin", s.Username))
		}
	})
}

func getCurrentSession(r *http.Request) (*auth.Session, error) {
	if authCookie, err := r.Cookie(auth.CookieName); err == nil {
		if !auth.IsSessionRevoked(authCookie.Value) {
			return parseCookie(authCookie)
		}
	}

	if authHeader := r.Header.Get("Authorization"); strings.HasPrefix(authHeader, "Bearer ") {
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if !auth.IsSessionRevoked(token) {
			return parseToken(token)
		}
	}

	return nil, nil
}

var (
	parseCookie = auth.ParseCookie
	parseToken  = auth.ParseToken
)
