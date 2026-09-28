package entity

import "time"

// DeadLetterMessage is a Kafka message the worker could not process (e.g.
// malformed JSON). It is stored verbatim instead of being dropped, so no
// message is ever lost — it can be inspected and replayed later.
type DeadLetterMessage struct {
	Topic       string
	Partition   int
	Offset      int64
	Key         []byte    // Kafka message key; nil when the message has none
	Payload     []byte    // message value exactly as received
	Error       string    // why the message could not be processed
	MessageTime time.Time // Kafka's own message timestamp; zero if unknown
	FailedAt    time.Time // when the worker rejected it
}
