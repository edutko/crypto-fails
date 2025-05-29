package route

import (
	"net/http"
	"strings"

	"github.com/edutko/crypto-fails/internal/auth"
	"github.com/edutko/crypto-fails/internal/responses"
)

func Logout(w http.ResponseWriter, r *http.Request) {
	if c, _ := r.Cookie(auth.CookieName); c != nil {
		auth.RevokeSession(c.Value)
	}

	if t := r.Header.Get("Authorization"); strings.HasPrefix(t, "Bearer ") {
		auth.RevokeSession(strings.TrimPrefix(t, "Bearer "))
	}

	http.SetCookie(w, &http.Cookie{Name: auth.CookieName, MaxAge: -1})
	responses.Found(w, "/")
}
