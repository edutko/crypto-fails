package auth

import (
	"sync"
	"time"

	"github.com/edutko/crypto-fails/internal/config"
)

func IsSessionRevoked(session string) bool {
	_, ok := revokedSessions.Load(session)
	return ok
}

func RevokeSession(session string) {
	revokedSessions.Store(session, time.Now().Add(config.SessionDuration()))
}

var revokedSessions sync.Map
