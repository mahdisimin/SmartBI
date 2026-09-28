package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"intelligentBI/repository/SQLServer"
	synopsrepo "intelligentBI/repository/SQLServer/synops"

	"github.com/segmentio/kafka-go"
)

// TODO: fill in the real Kafka node config before running.
var (
	kafkaBrokers = []string{"192.168.59.75:9092", "192.168.59.76:9092", "192.168.59.77:9092"} // e.g. []string{"broker1:9092", "broker2:9092"}
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
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
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

	// One shared connection pool for the worker's lifetime — opening a new
	// connection per message exhausts ephemeral ports during backlog replay.
	db, err := SQLServer.NewDB()
	if err != nil {
		log.Fatalf("failed to connect to SQL Server: %v", err)
	}
	defer db.Close()

	var repo WorkerRepository = synopsrepo.NewUserActivity(db)

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

		// Blocks until the message is stored (as an event, or as a dead letter
		// if it cannot be parsed or is permanently rejected); never moves past
		// an unstored message, since committing a later offset would drop it.
		if err := handleMessage(ctx, repo, msg); err != nil {
			log.Printf("stopping: could not store message (partition=%d offset=%d): %v — offset left uncommitted, will be redelivered on restart", msg.Partition, msg.Offset, err)
			return
		}

		if err := reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("failed to commit offset (partition=%d offset=%d): %v", msg.Partition, msg.Offset, err)
			continue
		}

		fmt.Printf("committed partition=%d offset=%d\n", msg.Partition, msg.Offset)
	}
}
