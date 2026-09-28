package SQLServer

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"testing"
	"time"

	"intelligentBI/entity"
	"intelligentBI/repository/SQLServer"
)

// Session lifecycle against APP.UserSession. Creates a throwaway user (sessions
// reference APP.[USER]), and deletes its sessions and the user afterwards.
// Skipped while SQL Server or the APP.UserSession table is unavailable.
func TestSession_Lifecycle(t *testing.T) {
	db, err := SQLServer.NewDB()
	if err != nil {
		t.Skipf("SQL Server unavailable: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	var tableID *int
	if db.QueryRow("SELECT OBJECT_ID('APP.UserSession')").Scan(&tableID); tableID == nil {
		t.Skip("APP.UserSession does not exist yet")
	}

	phone := fmt.Sprintf("test%d", time.Now().UnixNano())
	var userID int64
	if err := db.QueryRow("INSERT INTO APP.[USER] (UserName, PhoneNumber, Password) OUTPUT INSERTED.ID VALUES ('session-test', @p1, 'x')", phone).Scan(&userID); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Exec("DELETE FROM APP.UserSession WHERE UserID = @p1", userID)
		db.Exec("DELETE FROM APP.[USER] WHERE ID = @p1", userID)
	})

	repo := NewSession(db)
	now := time.Now()
	live := sha256.Sum256([]byte(phone + "-live"))
	expired := sha256.Sum256([]byte(phone + "-expired"))

	if err := repo.CreateSession(live[:], userID, now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateSession(expired[:], userID, now.Add(-time.Minute)); err != nil {
		t.Fatal(err)
	}

	if got, err := repo.GetSessionUserID(live[:], now); err != nil || got != userID {
		t.Fatalf("live session: got %d, %v; want %d", got, err, userID)
	}
	if _, err := repo.GetSessionUserID(expired[:], now); !errors.Is(err, entity.ErrSessionNotFound) {
		t.Fatalf("expired session: want ErrSessionNotFound, got %v", err)
	}

	if err := repo.DeleteExpiredSessions(now); err != nil {
		t.Fatal(err)
	}
	var remaining int
	db.QueryRow("SELECT COUNT(*) FROM APP.UserSession WHERE UserID = @p1", userID).Scan(&remaining)
	if remaining != 1 {
		t.Fatalf("after cleanup: want only the live session left, got %d", remaining)
	}

	if err := repo.DeleteSession(live[:]); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.GetSessionUserID(live[:], now); !errors.Is(err, entity.ErrSessionNotFound) {
		t.Fatalf("deleted session: want ErrSessionNotFound, got %v", err)
	}
}
