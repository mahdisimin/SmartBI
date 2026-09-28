package service

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"intelligentBI/entity"
)

// memSessions is an in-memory SessionRepo keyed by token hash.
type memSessions struct {
	rows map[string]memSession
}

type memSession struct {
	userID    int64
	expiresAt time.Time
}

func newMemSessions() *memSessions { return &memSessions{rows: map[string]memSession{}} }

func (m *memSessions) CreateSession(hash []byte, userID int64, expiresAt time.Time) error {
	m.rows[string(hash)] = memSession{userID, expiresAt}
	return nil
}

func (m *memSessions) GetSessionUserID(hash []byte, now time.Time) (int64, error) {
	s, ok := m.rows[string(hash)]
	if !ok || !s.expiresAt.After(now) {
		return 0, entity.ErrSessionNotFound
	}
	return s.userID, nil
}

func (m *memSessions) DeleteSession(hash []byte) error {
	delete(m.rows, string(hash))
	return nil
}

func (m *memSessions) DeleteExpiredSessions(now time.Time) error {
	for k, s := range m.rows {
		if !s.expiresAt.After(now) {
			delete(m.rows, k)
		}
	}
	return nil
}

func TestSessionService_Lifecycle(t *testing.T) {
	now := time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC)
	repo := newMemSessions()
	svc := SessionService{Repository: repo, TTL: time.Hour, Now: func() time.Time { return now }}

	token, expiresAt, err := svc.Create(42)
	if err != nil {
		t.Fatal(err)
	}
	if !expiresAt.Equal(now.Add(time.Hour)) {
		t.Errorf("expiresAt = %v, want now+TTL", expiresAt)
	}
	if len(token) < 43 { // 32 random bytes, base64url
		t.Errorf("token too short: %d chars", len(token))
	}
	for hash := range repo.rows {
		if bytes.Contains([]byte(hash), []byte(token)) {
			t.Fatal("the raw token must never be stored")
		}
	}

	if userID, err := svc.Authenticate(token); err != nil || userID != 42 {
		t.Fatalf("Authenticate = %d, %v; want 42", userID, err)
	}
	for _, bad := range []string{"", "not-a-token", token + "x"} {
		if _, err := svc.Authenticate(bad); !errors.Is(err, ErrUnauthenticated) {
			t.Errorf("Authenticate(%q): want ErrUnauthenticated, got %v", bad, err)
		}
	}

	// Expired.
	now = now.Add(time.Hour)
	if _, err := svc.Authenticate(token); !errors.Is(err, ErrUnauthenticated) {
		t.Errorf("expired session: want ErrUnauthenticated, got %v", err)
	}

	// Revoked (logout).
	now = now.Add(-30 * time.Minute)
	if err := svc.Revoke(token); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Authenticate(token); !errors.Is(err, ErrUnauthenticated) {
		t.Errorf("revoked session: want ErrUnauthenticated, got %v", err)
	}
}

func TestSessionService_TokensAreUnique(t *testing.T) {
	svc := SessionService{Repository: newMemSessions(), TTL: time.Hour}
	seen := map[string]bool{}
	for i := 0; i < 100; i++ {
		token, _, err := svc.Create(1)
		if err != nil {
			t.Fatal(err)
		}
		if seen[token] {
			t.Fatal("duplicate session token")
		}
		seen[token] = true
	}
}

func TestSessionService_NoRepositoryFailsInsteadOfPanicking(t *testing.T) {
	var svc SessionService
	if _, _, err := svc.Create(1); err == nil {
		t.Error("Create: expected an error")
	}
	if _, err := svc.Authenticate("token"); err == nil || errors.Is(err, ErrUnauthenticated) {
		t.Errorf("Authenticate: expected a configuration error, got %v", err)
	}
	if err := svc.Revoke("token"); err == nil {
		t.Error("Revoke: expected an error")
	}
}
