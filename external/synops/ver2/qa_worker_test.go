package main

// QA: zero-loss / liveness tests for the Kafka worker against the real
// SQL Server repository. Each case feeds handleMessage one message and checks
// whether it is stored (event or dead letter) within a deadline. A case that
// times out means the worker would retry that message forever, stalling its
// partition — every later message on it is blocked.

import (
	"context"
	"crypto/rand"
	"fmt"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"intelligentBI/repository/SQLServer"
	synopsrepo "intelligentBI/repository/SQLServer/synops"

	"github.com/jmoiron/sqlx"
	"github.com/segmentio/kafka-go"
)

const qaTopic = "qa.worker.test"

func qaDB(t *testing.T) *sqlx.DB {
	t.Helper()
	db, err := SQLServer.NewDB()
	if err != nil {
		t.Skipf("SQL Server unavailable: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// qaEvent returns the sample event as a map so a case can mutate one field.
func qaEvent(t *testing.T, eventID string) map[string]any {
	t.Helper()
	data, err := os.ReadFile("sample_data.json")
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	m["event_id"] = eventID
	return m
}

func qaUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	b[6] = b[6]&0x0f | 0x40
	b[8] = b[8]&0x3f | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func TestQA_Worker_PoisonMessages(t *testing.T) {
	db := qaDB(t)
	shrinkBackoff(t)
	repo := synopsrepo.NewUserActivity(db)
	t.Cleanup(func() {
		db.Exec("DELETE FROM synops.UserActivityDeadLetter WHERE Topic = @p1", qaTopic)
		db.Exec("DELETE FROM synops.UserActivity WHERE Source LIKE 'qa%'")
	})

	cases := []struct {
		name   string
		mutate func(m map[string]any) // nil => raw payload below
		raw    string
	}{
		{name: "missing event_id", mutate: func(m map[string]any) { delete(m, "event_id") }},
		{name: "non-uuid event_id", mutate: func(m map[string]any) { m["event_id"] = "evt-123" }},
		{name: "json null payload", raw: "null"},
		{name: "empty json object", raw: "{}"},
		{name: "event_type > 50 chars", mutate: func(m map[string]any) { m["event_type"] = strings.Repeat("x", 51) }},
		{name: "method > 10 chars", mutate: func(m map[string]any) {
			m["activity"].(map[string]any)["method"] = "PROPFINDXYZ"
		}},
		{name: "status_code > smallint", mutate: func(m map[string]any) {
			m["result"].(map[string]any)["status_code"] = 70000
		}},
		{name: "ip_address > 45 chars", mutate: func(m map[string]any) {
			m["request"].(map[string]any)["ip_address"] = strings.Repeat("1", 46)
		}},
		{name: "activity.name > 200 chars", mutate: func(m map[string]any) {
			m["activity"].(map[string]any)["name"] = strings.Repeat("a", 201)
		}},
		{name: "source > 100 chars", mutate: func(m map[string]any) { m["source"] = strings.Repeat("s", 101) }},
		{name: "emoji name: 101 runes = 202 UTF-16 units > 200", mutate: func(m map[string]any) {
			m["activity"].(map[string]any)["name"] = strings.Repeat("😀", 101)
		}},
		{name: "missing occurred_at", mutate: func(m map[string]any) { delete(m, "occurred_at") }},
		{name: "schema_version > smallint", mutate: func(m map[string]any) { m["schema_version"] = 40000 }},
		{name: "wrong type: status_code string", mutate: func(m map[string]any) {
			m["result"].(map[string]any)["status_code"] = "200"
		}},
		{name: "uuid with braces", mutate: func(m map[string]any) { m["event_id"] = "{a544f4eb-41a0-4458-abba-7f3f35a31220}" }},
		// controls: these are expected to succeed
		{name: "CONTROL valid event", mutate: func(m map[string]any) {}},
		{name: "CONTROL malformed json -> dead letter", raw: "{not json"},
		{name: "CONTROL tombstone (nil value) -> dead letter", raw: "\x00nil"},
	}

	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var payload []byte
			eventID := ""
			switch {
			case tc.raw == "\x00nil":
				payload = nil
			case tc.raw != "":
				payload = []byte(tc.raw)
			default:
				eventID = qaUUID()
				m := qaEvent(t, eventID)
				m["source"] = "qa_worker"
				tc.mutate(m)
				if id, ok := m["event_id"].(string); !ok || id != eventID {
					eventID = "" // case replaced/removed the id
				}
				payload, _ = json.Marshal(m)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
			defer cancel()
			msg := kafka.Message{Topic: qaTopic, Partition: 0, Offset: int64(1000 + i), Value: payload, Time: time.Now()}

			err := handleMessage(ctx, repo, msg)
			if errors.Is(err, context.DeadlineExceeded) {
				t.Errorf("STALL: message never stored, worker retries forever (payload=%.80q)", payload)
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			// Zero-loss: the message must actually be stored somewhere.
			var events, dls int
			if eventID != "" {
				db.QueryRow("SELECT COUNT(*) FROM synops.UserActivity WHERE Source LIKE 'qa%' AND CONVERT(varchar(36), EventID) = @p1", eventID).Scan(&events)
			}
			var dlErr string
			db.QueryRow("SELECT COUNT(*), MAX(Error) FROM synops.UserActivityDeadLetter WHERE Topic=@p1 AND KafkaPartition=0 AND KafkaOffset=@p2", qaTopic, msg.Offset).Scan(&dls, &dlErr)
			if strings.HasPrefix(tc.name, "CONTROL valid") {
				if events != 1 || dls != 0 {
					t.Errorf("valid event: want 1 event row / 0 dead letters, got %d / %d", events, dls)
				}
				return
			}
			if dls != 1 || events != 0 {
				t.Errorf("want exactly 1 dead letter and 0 event rows, got %d / %d", dls, events)
				return
			}
			t.Logf("dead-lettered: %.140s", dlErr)
		})
	}
}
