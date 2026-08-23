package main

import (
	"os"
	"testing"
)

// TestParseUserActivityEvent_SampleData verifies the struct against the real
// sample pulled from Kafka, without needing a live broker connection.
func TestParseUserActivityEvent_SampleData(t *testing.T) {
	data, err := os.ReadFile("sample_data.json")
	if err != nil {
		t.Fatalf("read sample_data.json: %v", err)
	}

	event, err := ParseUserActivityEvent(data)
	if err != nil {
		t.Fatalf("parse sample data: %v", err)
	}

	//var repo UserActivityRepository = synopsrepo.UserActivity{}

	//if err := repo.PersistUserActivity(event); err != nil {
	//	log.Fatalf("failed to persist user activity: %v", err)
	//}

	PrintUserActivityEvent(event)
}
