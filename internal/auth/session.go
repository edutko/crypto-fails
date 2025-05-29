package auth

import (
	"context"

	"github.com/edutko/crypto-fails/internal/net/urlquery"
)

type Session struct {
	IsAdmin  bool   `query:"adm"`
	Expires  int    `query:"exp"`
	RealName string `query:"name,omitempty"`
	Username string `query:"uid"`
}

func GetCurrentSession(ctx context.Context) *Session {
	if s := ctx.Value(sessionCtxKey); s != nil {
		return s.(*Session)
	}
	return nil
}

func ContextWithSession(ctx context.Context, s *Session) context.Context {
	return context.WithValue(ctx, sessionCtxKey, s)
}

func (s *Session) QueryString() string {
	qs, _ := urlquery.Marshal(s)
	return string(qs)
}

func ParseSession(qs string) *Session {
	var s Session
	_ = urlquery.Unmarshal([]byte(qs), &s)
	return &s
}

const sessionCtxKey = "session"
