package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"time"

	"intelligentBI/entity"
)

// ErrUnauthenticated is returned when a request carries no valid session.
var ErrUnauthenticated = errors.New("not logged in")

// errNoSessionStore guards against a SessionService built without a
// repository: it fails every operation instead of panicking.
var errNoSessionStore = errors.New("session store not configured")

type SessionRepo interface {
	CreateSession(tokenHash []byte, userID int64, expiresAt time.Time) error
	// GetSessionUserID returns entity.ErrSessionNotFound for an unknown or
	// expired session.
	GetSessionUserID(tokenHash []byte, now time.Time) (int64, error)
	DeleteSession(tokenHash []byte) error
	DeleteExpiredSessions(now time.Time) error
}

// SessionService issues and checks opaque login-session tokens. A token is
// 32 random bytes (base64url); only its SHA-256 is stored server-side, so
// sessions can be revoked (logout) and a database leak exposes no tokens.
type SessionService struct {
	Repository SessionRepo
	TTL        time.Duration
	// Now is the clock; nil means time.Now. Tests override it.
	Now func() time.Time
}

func (s SessionService) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now()
}

// Create starts a session for userID and returns its token and expiry.
func (s SessionService) Create(userID int64) (token string, expiresAt time.Time, err error) {
	if s.Repository == nil {
		return "", time.Time{}, errNoSessionStore
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", time.Time{}, fmt.Errorf("generate session token: %w", err)
	}
	token = base64.RawURLEncoding.EncodeToString(raw)
	now := s.now()
	expiresAt = now.Add(s.TTL)

	if err := s.Repository.CreateSession(hashToken(token), userID, expiresAt); err != nil {
		return "", time.Time{}, err
	}

	// Opportunistic cleanup keeps the table small without a scheduled job. A
	// failure here does not affect the login, so it is only logged.
	if err := s.Repository.DeleteExpiredSessions(now); err != nil {
		log.Printf("session cleanup failed: %v", err)
	}
	return token, expiresAt, nil
}

// Authenticate returns the user of a valid session, or ErrUnauthenticated.
func (s SessionService) Authenticate(token string) (int64, error) {
	if token == "" {
		return 0, ErrUnauthenticated
	}
	if s.Repository == nil {
		return 0, errNoSessionStore
	}
	userID, err := s.Repository.GetSessionUserID(hashToken(token), s.now())
	if errors.Is(err, entity.ErrSessionNotFound) {
		return 0, ErrUnauthenticated
	}
	if err != nil {
		return 0, err
	}
	return userID, nil
}

// Revoke ends a session (logout). Revoking an unknown token is not an error.
func (s SessionService) Revoke(token string) error {
	if token == "" {
		return nil
	}
	if s.Repository == nil {
		return errNoSessionStore
	}
	return s.Repository.DeleteSession(hashToken(token))
}

func hashToken(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}
