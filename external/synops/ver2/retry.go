package main

import (
	"context"
	"errors"
	"log"
	"time"
)

// Backoff bounds for retrying a failed write. Variables (not consts) so tests
// can shrink them.
var (
	persistRetryInitialDelay = 1 * time.Second
	persistRetryMaxDelay     = 30 * time.Second
)

// retryUntilSuccess keeps calling fn, with exponential backoff, until it
// succeeds or ctx is cancelled. `what` describes the operation for logs.
//
// The worker must never move past a message it failed to store: Kafka commits
// are cumulative per partition, so committing any later offset would
// implicitly commit — and permanently lose — the failed one. Blocking here
// keeps the partition stalled on that message until it lands.
//
// It returns nil once fn succeeds, ctx.Err() on shutdown (the offset is then
// left uncommitted and the message is redelivered on restart), or — without
// retrying — an error that isPermanent reports as never able to succeed.
func retryUntilSuccess(ctx context.Context, what string, fn func() error) error {
	delay := persistRetryInitialDelay
	for attempt := 1; ; attempt++ {
		err := fn()
		if err == nil {
			if attempt > 1 {
				log.Printf("%s succeeded after %d attempts", what, attempt)
			}
			return nil
		}
		if isPermanent(err) {
			return err
		}

		log.Printf("%s failed (attempt %d): %v — retrying in %s", what, attempt, err, delay)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}

		delay *= 2
		if delay > persistRetryMaxDelay {
			delay = persistRetryMaxDelay
		}
	}
}

// isPermanent reports whether err will fail identically on every retry (e.g.
// the database rejects the data itself). Repositories mark such errors with a
// Permanent() bool method; the worker checks for that method rather than a
// concrete type, so it never depends on a repository package.
func isPermanent(err error) bool {
	var p interface{ Permanent() bool }
	return errors.As(err, &p) && p.Permanent()
}
