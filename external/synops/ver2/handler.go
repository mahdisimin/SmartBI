package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"intelligentBI/entity"

	"github.com/segmentio/kafka-go"
)

// WorkerRepository is everything the worker needs to store: valid events and
// dead-lettered messages.
type WorkerRepository interface {
	UserActivityRepository
	DeadLetterRepository
}

// handleMessage stores msg durably, retrying transient failures until the
// write succeeds:
//   - a message that parses and is accepted is stored as a user activity event;
//   - a message that cannot be parsed, or that the repository permanently
//     rejects (invalid EventID, value too long/out of range, ...), is stored
//     as a dead letter — retrying it would stall the partition forever.
//
// When it returns nil the message is safely stored and its offset may be
// committed. A non-nil error means it could not be stored (shutdown, or even
// the dead-letter write was permanently rejected) and the offset must be left
// uncommitted.
func handleMessage(ctx context.Context, repo WorkerRepository, msg kafka.Message) error {
	event, err := ParseUserActivityEvent(msg.Value)
	if err != nil {
		return deadLetter(ctx, repo, msg, err)
	}

	what := fmt.Sprintf("persist event %s (partition=%d offset=%d)", event.EventID, msg.Partition, msg.Offset)
	err = retryUntilSuccess(ctx, what, func() error { return repo.PersistUserActivity(event) })
	if err != nil && isPermanent(err) {
		return deadLetter(ctx, repo, msg, err)
	}
	return err
}

// deadLetter stores msg with the reason it could not be processed.
func deadLetter(ctx context.Context, repo DeadLetterRepository, msg kafka.Message, reason error) error {
	dl := entity.DeadLetterMessage{
		Topic:       msg.Topic,
		Partition:   msg.Partition,
		Offset:      msg.Offset,
		Key:         msg.Key,
		Payload:     msg.Value,
		Error:       reason.Error(),
		MessageTime: msg.Time,
		FailedAt:    time.Now().UTC(),
	}
	what := fmt.Sprintf("dead-letter message (partition=%d offset=%d)", msg.Partition, msg.Offset)
	if err := retryUntilSuccess(ctx, what, func() error { return repo.PersistDeadLetter(dl) }); err != nil {
		return err
	}
	log.Printf("message dead-lettered (partition=%d offset=%d): %v", msg.Partition, msg.Offset, reason)
	return nil
}
