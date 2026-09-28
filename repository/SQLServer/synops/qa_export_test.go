package SQLServer

// QA: export query window / timezone handling against the real DB.

import (
	"testing"
	"time"

	"intelligentBI/entity"
	"intelligentBI/pkg"
)

func TestQA_GetActivityEvents_WindowAcrossTimezones(t *testing.T) {
	db := requireDB(t)
	repo := NewUserActivity(db)

	// Event stored 30 min ago with a +03:30 offset (Tehran) — the query window
	// is built from time.Now() in local time. It must be found regardless of
	// the offset it was written with.
	tehran := time.FixedZone("IRST", 3*3600+1800)
	occurred := time.Now().Add(-30 * time.Minute).In(tehran)
	ev := entity.UserActivityEvent{
		EventID: newUUID(t), EventType: "user.activity", SchemaVersion: 1, OccurredAt: occurred,
		Source: "qa_export", Result: entity.ActivityResult{StatusCode: 200},
		Request:  entity.ActivityRequest{Country: "ir", IPAddress: "127.0.0.1"},
		Activity: entity.ActivityInfo{ID: 1, Name: "QA_Module | X", Path: "/qa", Method: "GET"},
	}
	t.Cleanup(func() { db.Exec("DELETE FROM synops.UserActivity WHERE EventID = @p1", ev.EventID) })
	if err := repo.PersistUserActivity(ev); err != nil {
		t.Fatal(err)
	}

	for _, loc := range []*time.Location{time.UTC, time.Local, tehran, time.FixedZone("PST", -8*3600)} {
		to := time.Now().In(loc)
		events, err := repo.GetActivityEvents(pkg.SynOps, to.Add(-time.Hour), to)
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, e := range events {
			if e.Module == "QA_Module" && e.OccurredAt.Equal(occurred.Truncate(time.Microsecond)) {
				found = true
			} else if e.Module == "QA_Module" {
				t.Errorf("[%s] OccurredAt round-trip changed instant: stored %v, read %v", loc, occurred, e.OccurredAt)
				found = true
			}
		}
		if !found {
			t.Errorf("[%s] event inside window not returned", loc)
		}
	}
}

func TestQA_GetActivityEvents_UnsupportedProduct(t *testing.T) {
	repo := NewUserActivity(requireDB(t))
	if _, err := repo.GetActivityEvents(pkg.ProductList(99), time.Now().Add(-time.Hour), time.Now()); err == nil {
		t.Error("unsupported product should error")
	}
}
