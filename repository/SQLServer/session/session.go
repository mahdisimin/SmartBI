package SQLServer

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"intelligentBI/entity"

	"github.com/jmoiron/sqlx"
)

// Session stores login sessions in APP.UserSession. Only a hash of each
// session token is stored — never the token itself — so a database leak does
// not hand out usable sessions.
//
// Expected table (owned by the Database team):
//
//	APP.UserSession (
//	    TokenHash BINARY(32)   NOT NULL PRIMARY KEY,
//	    UserID    INT          NOT NULL REFERENCES APP.[USER](ID) ON DELETE CASCADE,
//	    CreatedAt DATETIME2(3) NOT NULL DEFAULT SYSUTCDATETIME(),
//	    ExpiresAt DATETIME2(3) NOT NULL
//	)  + indexes on UserID and ExpiresAt
//
// All times are UTC.
type Session struct {
	DB *sqlx.DB
}

func NewSession(db *sqlx.DB) Session {
	return Session{DB: db}
}

func (s Session) CreateSession(tokenHash []byte, userID int64, expiresAt time.Time) error {
	_, err := s.DB.Exec("INSERT INTO APP.UserSession (TokenHash, UserID, ExpiresAt) VALUES (@p1, @p2, @p3)",
		tokenHash, userID, expiresAt.UTC())
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

// GetSessionUserID returns the user of an unexpired session, or
// entity.ErrSessionNotFound.
func (s Session) GetSessionUserID(tokenHash []byte, now time.Time) (int64, error) {
	var userID int64
	err := s.DB.QueryRow("SELECT UserID FROM APP.UserSession WHERE TokenHash = @p1 AND ExpiresAt > @p2",
		tokenHash, now.UTC()).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, entity.ErrSessionNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("get session: %w", err)
	}
	return userID, nil
}

func (s Session) DeleteSession(tokenHash []byte) error {
	if _, err := s.DB.Exec("DELETE FROM APP.UserSession WHERE TokenHash = @p1", tokenHash); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

// DeleteExpiredSessions removes sessions that expired at or before now.
func (s Session) DeleteExpiredSessions(now time.Time) error {
	if _, err := s.DB.Exec("DELETE FROM APP.UserSession WHERE ExpiresAt <= @p1", now.UTC()); err != nil {
		return fmt.Errorf("delete expired sessions: %w", err)
	}
	return nil
}
