package session

import (
	"time"
)

const (
	sessionPath   = "/api/session"
	exchangeRoute = "POST " + sessionPath
	revokeRoute   = "DELETE " + sessionPath
)

type Credential struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
}
