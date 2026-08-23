package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	synopsrepo "intelligentBI/repository/SQLServer/synops"

	"github.com/segmentio/kafka-go"
)

// TODO: fill in the real Kafka node config before running.
var (
	kafkaBrokers = []string{"192.168.59.75:9092"} // e.g. []string{"broker1:9092", "broker2:9092"}
	kafkaTopic   = "stinas.user-activities.v1"
)

// kafkaGroupID identifies this worker's consumer group so Kafka tracks its
// offsets and resumes from where it left off after a restart, instead of
// re-reading the whole topic every time.
const kafkaGroupID = "intelligentbi-synops-useractivity-worker"

// fetchRetryDelay avoids a tight retry loop (log/CPU spam) while the broker
// is unreachable.
const fetchRetryDelay = 2 * time.Second

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: kafkaBrokers,
		Topic:   kafkaTopic,
		GroupID: kafkaGroupID,
		// Only takes effect the first time this consumer group ever runs (no
		// committed offset yet): start from the beginning of the topic instead
		// of only messages produced from now on.
		StartOffset: kafka.FirstOffset,
		MinBytes:    1,
		MaxBytes:    10e6,
	})
	defer reader.Close()

	var repo UserActivityRepository = synopsrepo.UserActivity{}

	log.Println("worker started, waiting for messages... (Ctrl+C to stop)")

	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				log.Println("shutdown signal received, stopping worker")
				return
			}
			log.Printf("failed to fetch message from kafka: %v", err)
			time.Sleep(fetchRetryDelay)
			continue
		}

		event, err := ParseUserActivityEvent(msg.Value)
		if err != nil {
			log.Printf("failed to parse message (partition=%d offset=%d): %v — skipping", msg.Partition, msg.Offset, err)
			continue
		}

		if err := repo.PersistUserActivity(event); err != nil {
			log.Printf("failed to persist event %s (partition=%d offset=%d): %v — offset left uncommitted, will retry on restart", event.EventID, msg.Partition, msg.Offset, err)
			continue
		}

		if err := reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("failed to commit offset (partition=%d offset=%d): %v", msg.Partition, msg.Offset, err)
			continue
		}

		fmt.Printf("persisted event_id=%s partition=%d offset=%d\n", event.EventID, msg.Partition, msg.Offset)
	}
}
