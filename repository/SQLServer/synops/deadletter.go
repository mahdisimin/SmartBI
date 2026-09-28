package SQLServer

import (
	"intelligentBI/entity"
	"intelligentBI/repository/SQLServer"
)

// PersistDeadLetter stores a message the worker could not process in
// synops.UserActivityDeadLetter. It is idempotent: the table is unique on
// (Topic, KafkaPartition, KafkaOffset), so re-inserting the same message after
// a crash-before-commit hits a duplicate key, which is treated as success.
func (s UserActivity) PersistDeadLetter(msg entity.DeadLetterMessage) error {
	// Payload is NOT NULL, and the driver sends a nil []byte as NULL — a
	// null Kafka value is stored as an empty payload instead.
	payload := msg.Payload
	if payload == nil {
		payload = []byte{}
	}

	// MessageKey is passed as []byte even when nil: the driver sends a nil
	// []byte as a typed VARBINARY NULL, whereas an untyped nil goes as
	// NVARCHAR NULL, which SQL Server refuses to convert to VARBINARY(MAX).
	//
	// MessageTime is nullable: unknown time -> NULL.
	var messageTime any
	if !msg.MessageTime.IsZero() {
		messageTime = msg.MessageTime.UTC()
	}

	// DATETIME2 columns hold no offset, so every time is written as UTC.
	_, err := s.DB.Exec(`INSERT INTO synops.UserActivityDeadLetter
		(Topic, KafkaPartition, KafkaOffset, MessageKey, Payload, Error, MessageTime, FailedAt)
		VALUES (@p1, @p2, @p3, @p4, @p5, @p6, @p7, @p8)`,
		msg.Topic, msg.Partition, msg.Offset, msg.Key, payload, msg.Error, messageTime, msg.FailedAt.UTC(),
	)
	if SQLServer.IsDuplicateKey(err) {
		return nil
	}
	return err
}
