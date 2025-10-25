package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/edutko/crypto-fails/internal/net/urlquery"
)

type Session struct {
	IsAdmin  bool   `query:"adm"`
	Expires  int    `query:"exp"`
	RealName string `query:"name,omitempty"`
	Username string `query:"uid"`
}

func GetSessionFromRequest(r *http.Request) (Session, error) {
	if authCookie, err := r.Cookie(CookieName); err == nil {
		if !IsSessionRevoked(authCookie.Value) {
			return parseCookie(authCookie)
		}
	}

	if authHeader := r.Header.Get("Authorization"); strings.HasPrefix(authHeader, "Bearer ") {
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if !IsSessionRevoked(token) {
			return parseToken(token)
		}
	}

	return anonymousSession, nil
}

func GetSessionFromContext(ctx context.Context) Session {
	if s := ctx.Value(sessionCtxKey); s != nil {
		return s.(Session)
	}
	return anonymousSession
}

func ContextWithSession(ctx context.Context, s Session) context.Context {
	return context.WithValue(ctx, sessionCtxKey, s)
}

func (s Session) IsAuthenticated() bool {
	return s != anonymousSession
}

func (s Session) QueryString() string {
	qs, _ := urlquery.Marshal(s)
	return string(qs)
}

func ParseSession(qs string) Session {
	var s Session
	_ = urlquery.Unmarshal([]byte(qs), &s)
	return s
}

const sessionCtxKey = "session"

var anonymousSession = Session{}

var (
	parseCookie = ParseCookie
	parseToken  = ParseToken
)
