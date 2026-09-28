package main

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"intelligentBI/entity"

	"github.com/segmentio/kafka-go"
)

// fakeRepo fails the first `failures` writes of any kind, then succeeds, and
// records what was stored.
type fakeRepo struct {
	failures    int
	calls       int
	events      []entity.UserActivityEvent
	deadLetters []entity.DeadLetterMessage
}

func (r *fakeRepo) fail() error {
	r.calls++
	if r.calls <= r.failures {
		return errors.New("dial tcp: connectex: Only one usage of each socket address")
	}
	return nil
}

func (r *fakeRepo) PersistUserActivity(e entity.UserActivityEvent) error {
	if err := r.fail(); err != nil {
		return err
	}
	r.events = append(r.events, e)
	return nil
}

func (r *fakeRepo) PersistDeadLetter(m entity.DeadLetterMessage) error {
	if err := r.fail(); err != nil {
		return err
	}
	r.deadLetters = append(r.deadLetters, m)
	return nil
}

func shrinkBackoff(t *testing.T) {
	t.Helper()
	initial, max := persistRetryInitialDelay, persistRetryMaxDelay
	persistRetryInitialDelay, persistRetryMaxDelay = time.Millisecond, 4*time.Millisecond
	t.Cleanup(func() { persistRetryInitialDelay, persistRetryMaxDelay = initial, max })
}

func TestRetryUntilSuccess_RetriesUntilSuccess(t *testing.T) {
	shrinkBackoff(t)
	calls := 0
	err := retryUntilSuccess(context.Background(), "test op", func() error {
		calls++
		if calls <= 5 {
			return errors.New("boom")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if calls != 6 {
		t.Fatalf("expected 6 attempts (5 failures + 1 success), got %d", calls)
	}
}

func TestRetryUntilSuccess_StopsOnShutdown(t *testing.T) {
	shrinkBackoff(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	calls := 0
	err := retryUntilSuccess(ctx, "test op", func() error { calls++; return errors.New("boom") })
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context error, got %v", err)
	}
	if calls < 2 {
		t.Fatalf("expected multiple attempts before shutdown, got %d", calls)
	}
}

func TestHandleMessage_ValidEventIsPersisted(t *testing.T) {
	shrinkBackoff(t)
	data, err := os.ReadFile("sample_data.json")
	if err != nil {
		t.Fatalf("read sample_data.json: %v", err)
	}
	repo := &fakeRepo{failures: 2}

	if err := handleMessage(context.Background(), repo, kafka.Message{Topic: "t", Partition: 0, Offset: 7, Value: data}); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if len(repo.events) != 1 || len(repo.deadLetters) != 0 {
		t.Fatalf("expected 1 event and 0 dead letters, got %d / %d", len(repo.events), len(repo.deadLetters))
	}
}

func TestHandleMessage_MalformedMessageIsDeadLettered(t *testing.T) {
	shrinkBackoff(t)
	repo := &fakeRepo{failures: 2}
	msgTime := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)
	msg := kafka.Message{Topic: "stinas.user-activities.v1", Partition: 3, Offset: 366344, Key: []byte("k"), Value: []byte("{not json"), Time: msgTime}

	if err := handleMessage(context.Background(), repo, msg); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if len(repo.events) != 0 || len(repo.deadLetters) != 1 {
		t.Fatalf("expected 0 events and 1 dead letter, got %d / %d", len(repo.events), len(repo.deadLetters))
	}

	dl := repo.deadLetters[0]
	if dl.Topic != msg.Topic || dl.Partition != 3 || dl.Offset != 366344 {
		t.Errorf("wrong position: %+v", dl)
	}
	if string(dl.Payload) != "{not json" || string(dl.Key) != "k" {
		t.Errorf("raw bytes not preserved: key=%q payload=%q", dl.Key, dl.Payload)
	}
	if dl.Error == "" {
		t.Error("expected parse error text")
	}
	if !dl.MessageTime.Equal(msgTime) || dl.FailedAt.IsZero() {
		t.Errorf("wrong timestamps: message=%v failed=%v", dl.MessageTime, dl.FailedAt)
	}
}

func TestHandleMessage_DeadLetterFailureBlocksUntilShutdown(t *testing.T) {
	shrinkBackoff(t)
	repo := &fakeRepo{failures: 1 << 30}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	err := handleMessage(ctx, repo, kafka.Message{Value: []byte("{not json")})
	if err == nil {
		t.Fatal("expected an error so the caller does not commit the offset")
	}
	if len(repo.deadLetters) != 0 {
		t.Fatal("nothing should be stored")
	}
}

// rejectingRepo permanently rejects every event, like SQL Server rejecting a
// row whose data can never fit, and records dead letters.
type rejectingRepo struct {
	fakeRepo
	persistCalls int
}

type permanentErr struct{}

func (permanentErr) Error() string   { return "string or binary data would be truncated" }
func (permanentErr) Permanent() bool { return true }

func (r *rejectingRepo) PersistUserActivity(entity.UserActivityEvent) error {
	r.persistCalls++
	return permanentErr{}
}

func TestHandleMessage_PermanentlyRejectedEventIsDeadLettered(t *testing.T) {
	shrinkBackoff(t)
	data, err := os.ReadFile("sample_data.json")
	if err != nil {
		t.Fatal(err)
	}
	repo := &rejectingRepo{}

	if err := handleMessage(context.Background(), repo, kafka.Message{Topic: "t", Offset: 9, Value: data}); err != nil {
		t.Fatalf("expected the message to be stored as a dead letter, got %v", err)
	}
	if repo.persistCalls != 1 {
		t.Errorf("a permanent error must not be retried: %d attempts", repo.persistCalls)
	}
	if len(repo.deadLetters) != 1 || !strings.Contains(repo.deadLetters[0].Error, "truncated") {
		t.Fatalf("expected 1 dead letter carrying the rejection reason, got %+v", repo.deadLetters)
	}
}
