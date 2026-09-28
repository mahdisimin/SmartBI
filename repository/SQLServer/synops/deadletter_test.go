package SQLServer

import (
	"testing"
	"time"

	"intelligentBI/entity"
)

// Writes rows to synops.UserActivityDeadLetter (with and without a message
// key), verifies a duplicate insert is treated as success, then deletes them.
func TestUserActivity_PersistDeadLetter(t *testing.T) {
	for name, key := range map[string][]byte{"nil key": nil, "with key": []byte("k-1")} {
		t.Run(name, func(t *testing.T) { testPersistDeadLetter(t, key) })
	}
}

func testPersistDeadLetter(t *testing.T, key []byte) {
	db := requireDB(t)
	repo := NewUserActivity(db)

	msg := entity.DeadLetterMessage{
		Topic:       "intelligentbi.test.deadletter",
		Partition:   0,
		Offset:      time.Now().UnixNano(), // unique per run
		Key:         key,
		Payload:     []byte("{not json"),
		Error:       "test: invalid character 'n' looking for beginning of object key string",
		MessageTime: time.Now(),
		FailedAt:    time.Now(),
	}
	t.Cleanup(func() {
		db.Exec("DELETE FROM synops.UserActivityDeadLetter WHERE Topic = @p1 AND KafkaPartition = @p2 AND KafkaOffset = @p3",
			msg.Topic, msg.Partition, msg.Offset)
	})

	if err := repo.PersistDeadLetter(msg); err != nil {
		t.Fatalf("first insert: %v", err)
	}
	if err := repo.PersistDeadLetter(msg); err != nil {
		t.Fatalf("duplicate insert should be treated as success, got: %v", err)
	}

	var (
		payload, storedKey []byte
		rows               int
	)
	if err := db.QueryRow("SELECT Payload, MessageKey, COUNT(*) OVER () FROM synops.UserActivityDeadLetter WHERE Topic = @p1 AND KafkaPartition = @p2 AND KafkaOffset = @p3",
		msg.Topic, msg.Partition, msg.Offset).Scan(&payload, &storedKey, &rows); err != nil {
		t.Fatal(err)
	}
	if rows != 1 {
		t.Fatalf("expected exactly 1 row after duplicate insert, got %d", rows)
	}
	if string(payload) != string(msg.Payload) {
		t.Fatalf("payload mismatch: got %q", payload)
	}
	if (key == nil) != (storedKey == nil) || string(storedKey) != string(key) {
		t.Fatalf("key mismatch: want %q, got %q", key, storedKey)
	}
}
