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
	svc := newTestService(t)

	data, err := svc.Export(export.ExportRequest{Product: pkg.SynOps})
	if err != nil {
		t.Fatalf("Export returned error: %v", err)
	}

	if data.KPIs.TotalEvents != 4 {
		t.Fatalf("expected 4 in-range events, got %d", data.KPIs.TotalEvents)
	}
	if data.KPIs.UniqueUsers != 2 {
		t.Fatalf("expected 2 unique users, got %d", data.KPIs.UniqueUsers)
	}
	if data.KPIs.UniqueOrgs != 2 {
		t.Fatalf("expected 2 unique orgs, got %d", data.KPIs.UniqueOrgs)
	}
	if data.KPIs.SuccessCount != 3 {
		t.Fatalf("expected 3 successful events, got %d", data.KPIs.SuccessCount)
	}
	if data.KPIs.ErrorCount != 1 {
		t.Fatalf("expected 1 error event, got %d", data.KPIs.ErrorCount)
	}
	if data.KPIs.ModuleCount != 2 {
		t.Fatalf("expected 2 distinct modules (Unknown excluded), got %d", data.KPIs.ModuleCount)
	}
	if data.KPIs.DaysCovered != 4 {
		t.Fatalf("expected 4 distinct days, got %d", data.KPIs.DaysCovered)
	}

	if len(data.TopModules) != 2 || data.TopModules[0].Name != "Alpha" || data.TopModules[0].Count != 2 {
		t.Fatalf("expected Alpha ranked first with count 2, got %+v", data.TopModules)
	}

	if len(data.MethodBreakdown) != 2 || data.MethodBreakdown[0].Name != "GET" || data.MethodBreakdown[0].Count != 3 {
		t.Fatalf("expected GET ranked first with count 3, got %+v", data.MethodBreakdown)
	}

	if len(data.Users) != 2 {
		t.Fatalf("expected 2 users in the table, got %d", len(data.Users))
	}
	u1 := data.Users[0]
	if u1.UserID != 1 || u1.Actions != 2 || u1.ModuleBreadth != 1 || u1.SuccessRate != 100 {
		t.Fatalf("unexpected stats for user 1: %+v", u1)
	}

	var dailySum int
	for _, p := range data.DailyTrend {
		dailySum += p.Count
	}
	if dailySum != data.KPIs.TotalEvents {
		t.Fatalf("daily trend total %d does not match TotalEvents %d", dailySum, data.KPIs.TotalEvents)
	}

	for _, signal := range []string{data.Verdict.ReliabilitySignal, data.Verdict.BreadthSignal, data.Verdict.SampleSignal} {
		if signal != "ok" && signal != "warn" && signal != "bad" {
			t.Fatalf("unexpected verdict signal: %q", signal)
		}
	}
}

// TestExportService_CrossFilter verifies the dashboard's own cross-filter
// rule: every panel is filtered by every active dimension EXCEPT its own.
// Filtering by module=Alpha should narrow KPIs/methods/users/trend down to
// Alpha's events, while the module ranking itself stays unfiltered (still
// shows Beta too) — otherwise you could never see what else exists to filter by.
func TestExportService_CrossFilter(t *testing.T) {
	svc := newTestService(t)

	data, err := svc.Export(export.ExportRequest{
		Product: pkg.SynOps,
		Filters: export.Filters{Modules: []string{"Alpha"}},
	})
	if err != nil {
		t.Fatalf("Export returned error: %v", err)
	}

	// fAll applies the module filter: only user 1's 2 Alpha events remain.
	if data.KPIs.TotalEvents != 2 {
		t.Fatalf("expected 2 events once filtered to module=Alpha, got %d", data.KPIs.TotalEvents)
	}
	if data.KPIs.UniqueUsers != 1 {
		t.Fatalf("expected 1 unique user once filtered to module=Alpha, got %d", data.KPIs.UniqueUsers)
	}
	// ModuleCount stays global (2), unaffected by the active module filter.
	if data.KPIs.ModuleCount != 2 {
		t.Fatalf("expected ModuleCount to stay global at 2, got %d", data.KPIs.ModuleCount)
	}

	// The module panel excludes its own dimension: both modules still show,
	// in their global-count order, with their true (unfiltered-by-module) counts.
	if len(data.TopModules) != 2 || data.TopModules[0].Name != "Alpha" || data.TopModules[0].Count != 2 ||
		data.TopModules[1].Name != "Beta" || data.TopModules[1].Count != 1 {
		t.Fatalf("expected module panel to stay unfiltered by its own dimension, got %+v", data.TopModules)
	}

	// Every other panel DOES apply the module filter.
	if len(data.MethodBreakdown) != 2 || data.MethodBreakdown[0].Count+data.MethodBreakdown[1].Count != 2 {
		t.Fatalf("expected method breakdown restricted to Alpha's 2 events, got %+v", data.MethodBreakdown)
	}

	if len(data.Users) != 2 {
		t.Fatalf("expected both known users still listed, got %d", len(data.Users))
	}
	var user1, user2 *entity.UserStat
	for i := range data.Users {
		switch data.Users[i].UserID {
		case 1:
			user1 = &data.Users[i]
		case 2:
			user2 = &data.Users[i]
		}
	}
	if user1 == nil || user1.Actions != 2 {
		t.Fatalf("expected user 1 to keep both Alpha actions, got %+v", user1)
	}
	if user2 == nil || user2.Actions != 0 {
		t.Fatalf("expected user 2 zeroed out (their only action was Beta), got %+v", user2)
	}
}

func newTestService(t *testing.T) *export.ExportService {
	t.Helper()
	now := time.Now()

	events := []entity.UserActivityEvent{
		// user 1: two Alpha actions, both successful.
		sampleEvent(1, 10, "owner", "Alpha | DoStuff", "GET", 200, 0.1, now.AddDate(0, 0, -5)),
		sampleEvent(1, 10, "owner", "Alpha | DoOtherStuff", "POST", 200, 0.2, now.AddDate(0, 0, -3)),
		// user 2: one Beta action that failed.
		sampleEvent(2, 20, "member", "Beta | Retrieve", "GET", 404, 0.05, now.AddDate(0, 0, -1)),
		// anonymous, unclassified activity — should not count toward modules.
		anonymousEvent("Unknown", "GET", 200, 0.01, now.AddDate(0, 0, -2)),
		// outside the lookback window — must be excluded entirely.
		sampleEvent(3, 30, "owner", "Gamma | Old", "GET", 200, 0.3, now.AddDate(0, -3, 0)),
	}

	dir, err := os.MkdirTemp(".", "export_test_")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })

	fixturePath := filepath.Join(dir, "synops_user_activity.json")
	writeFixture(t, fixturePath, events)

	repo := filerepo.NewRepository(fixturePath)
	return export.NewExportService(repo)
}

func sampleEvent(userID, orgID int64, orgRole, activityName, method string, status int, duration float64, occurredAt time.Time) entity.UserActivityEvent {
	e := baseEvent(activityName, method, status, duration, occurredAt)
	e.Actor = map[string]any{
		"id":                userID,
		"organization_id":   orgID,
		"organization_role": orgRole,
	}
	return e
}

func anonymousEvent(activityName, method string, status int, duration float64, occurredAt time.Time) entity.UserActivityEvent {
	return baseEvent(activityName, method, status, duration, occurredAt)
}

func baseEvent(activityName, method string, status int, duration float64, occurredAt time.Time) entity.UserActivityEvent {
	return entity.UserActivityEvent{
		EventType:     "user.activity",
		SchemaVersion: 1,
		OccurredAt:    occurredAt,
		Source:        "server_backend",
		Result: entity.ActivityResult{
			StatusCode:      status,
			DurationSeconds: duration,
		},
		Request: entity.ActivityRequest{
			Country:   "unknown",
			IPAddress: "10.0.0.1",
			UserAgent: "test-agent",
		},
		Activity: entity.ActivityInfo{
			ID:     1,
			Name:   activityName,
			Path:   "/test",
			Method: method,
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
