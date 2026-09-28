package SQLServer

import (
	"testing"
	"time"

	"intelligentBI/pkg"
)

func TestUserActivity_GetActivityEvents(t *testing.T) {
	repo := NewUserActivity(requireDB(t))
	to := time.Now()
	from := to.AddDate(0, -2, 0)

	events, err := repo.GetActivityEvents(pkg.SynOps, from, to)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("event count: %d", len(events))
	for _, e := range events {
		t.Logf("%+v", e)
	}
}

// Full history, minus uptime-monitoring probes: GetActivityEvents with a zero
// `from` returns every stored event except health checks. Read-only.
func TestUserActivity_GetActivityEvents_FullHistoryExcludesHealthChecks(t *testing.T) {
	db := requireDB(t)
	repo := NewUserActivity(db)

	// Snapshot `to` first so rows written concurrently can't skew the counts.
	to := time.Now()
	var all, probes int
	if err := db.QueryRow(`SELECT COUNT(*),
		SUM(CASE WHEN ActivityName = 'health-check' OR ActivityPath = '/healthy' THEN 1 ELSE 0 END)
		FROM synops.UserActivity WHERE OccurredAt <= @p1`, to).Scan(&all, &probes); err != nil {
		t.Fatal(err)
	}

	events, err := repo.GetActivityEvents(pkg.SynOps, time.Time{}, to)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != all-probes {
		t.Fatalf("got %d events, want %d (all %d minus %d probes)", len(events), all-probes, all, probes)
	}
	for _, e := range events {
		if e.Module == "health-check" {
			t.Fatal("a health-check probe leaked into the export")
		}
	}
	t.Logf("%d events returned, %d probes excluded", len(events), probes)
}
