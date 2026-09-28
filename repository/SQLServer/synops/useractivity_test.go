package SQLServer

import (
	"crypto/rand"
	"fmt"
	"testing"
	"time"

	"intelligentBI/entity"
)

// Inserts one event into synops.UserActivity, inserts it again to verify the
// duplicate EventID is treated as success (exactly one row), then deletes it.
func TestUserActivity_PersistUserActivity_Idempotent(t *testing.T) {
	db := requireDB(t)
	repo := NewUserActivity(db)

	event := entity.UserActivityEvent{
		EventID:       newUUID(t),
		EventType:     "user.activity",
		SchemaVersion: 1,
		OccurredAt:    time.Now().UTC(),
		Source:        "intelligentbi_test",
		Result:        entity.ActivityResult{StatusCode: 200, DurationSeconds: 0.01},
		Request:       entity.ActivityRequest{Country: "unknown", IPAddress: "127.0.0.1", UserAgent: "go-test"},
		Activity:      entity.ActivityInfo{ID: 1, Name: "Test | Idempotency", Path: "/test", Method: "GET"},
	}
	t.Cleanup(func() {
		db.Exec("DELETE FROM synops.UserActivity WHERE EventID = @p1", event.EventID)
	})

	if err := repo.PersistUserActivity(event); err != nil {
		t.Fatalf("first insert: %v", err)
	}
	if err := repo.PersistUserActivity(event); err != nil {
		t.Fatalf("duplicate insert should be treated as success, got: %v", err)
	}

	var rows int
	if err := db.QueryRow("SELECT COUNT(*) FROM synops.UserActivity WHERE EventID = @p1", event.EventID).Scan(&rows); err != nil {
		t.Fatal(err)
	}
	if rows != 1 {
		t.Fatalf("expected exactly 1 row after duplicate insert, got %d", rows)
	}
}

// newUUID returns a random RFC 4122 v4 UUID string.
func newUUID(t *testing.T) string {
	t.Helper()
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(err)
	}
	b[6] = b[6]&0x0f | 0x40
	b[8] = b[8]&0x3f | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
