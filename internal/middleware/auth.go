package middleware

import (
	"fmt"
	"net/http"

	"github.com/edutko/crypto-fails/internal/auth"
	"github.com/edutko/crypto-fails/internal/route/responses"
)

func GetCurrentSession(r *http.Request) auth.Session {
	return auth.GetSessionFromContext(r.Context())
}

func Authenticated(next func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		s, err := getSessionFromRequest(r)
		if err != nil {
			responses.BadRequest(w, err)
			return
		}
		if !s.IsAuthenticated() {
			responses.Unauthorized(w)
			return
		}
		next(w, r.WithContext(auth.ContextWithSession(r.Context(), s)))
	}
}

func MaybeAuthenticated(next func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		s, _ := getSessionFromRequest(r)
		if s.IsAuthenticated() {
			next(w, r.WithContext(auth.ContextWithSession(r.Context(), s)))
			return
		}
		next(w, r)
	}
}

func RequireAdmin(next func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return Authenticated(func(w http.ResponseWriter, r *http.Request) {
		if s := GetCurrentSession(r); s.IsAdmin {
			next(w, r)
		} else {
			responses.Forbidden(w, fmt.Errorf("%q is not an admin", s.Username))
		}
	})
}

var getSessionFromRequest = auth.GetSessionFromRequest
