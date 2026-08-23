package export_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"intelligentBI/entity"
	"intelligentBI/pkg"
	filerepo "intelligentBI/repository/file"
	"intelligentBI/service/export"
)

func TestExportService_Export(t *testing.T) {
	now := time.Now()

	events := []entity.UserActivityEvent{
		sampleEvent("in-range-recent", now.AddDate(0, 0, -10)), // 10 days ago
		sampleEvent("in-range-older", now.AddDate(0, -1, -20)), // ~50 days ago
		sampleEvent("out-of-range", now.AddDate(0, -3, 0)),     // 3 months ago
	}

	dir, err := os.MkdirTemp(".", "export_test_")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })

	fixturePath := filepath.Join(dir, "synops_user_activity.json")
	writeFixture(t, fixturePath, events)

	repo := filerepo.NewRepository(fixturePath)
	svc := export.NewExportService(repo)

	response, err := svc.Export(export.ExportRequest{Product: pkg.SynOps})
	if err != nil {
		t.Fatalf("Export returned error: %v", err)
	}

	if response.RowCount != 2 {
		t.Fatalf("expected 2 rows within the last 2 months, got %d", response.RowCount)
	}
	if len(response.Data) != 2 {
		t.Fatalf("expected 2 data records, got %d", len(response.Data))
	}

	gotEventIDs := map[string]bool{}
	for _, record := range response.Data {
		eventID, _ := record["EventID"].(string)
		gotEventIDs[eventID] = true
	}
	if !gotEventIDs["in-range-recent"] || !gotEventIDs["in-range-older"] {
		t.Fatalf("expected in-range events in response, got %v", gotEventIDs)
	}
	if gotEventIDs["out-of-range"] {
		t.Fatalf("did not expect the out-of-range event in response")
	}
}

func sampleEvent(eventID string, occurredAt time.Time) entity.UserActivityEvent {
	return entity.UserActivityEvent{
		EventID:       eventID,
		EventType:     "user.activity",
		SchemaVersion: 1,
		OccurredAt:    occurredAt,
		Source:        "server_backend",
		Result: entity.ActivityResult{
			StatusCode:      200,
			DurationSeconds: 0.01,
		},
		Request: entity.ActivityRequest{
			Country:   "unknown",
			IPAddress: "10.0.0.1",
			UserAgent: "test-agent",
		},
		Activity: entity.ActivityInfo{
			ID:     1,
			Name:   "Test",
			Path:   "/test",
			Method: "GET",
		},
	}
}

func writeFixture(t *testing.T, path string, events []entity.UserActivityEvent) {
	t.Helper()
	data, err := json.Marshal(events)
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
}
