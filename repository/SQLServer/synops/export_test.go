package SQLServer

import (
	"testing"
	"time"

	"intelligentBI/pkg"
)

func TestUserActivity_GetActivityEvents(t *testing.T) {
	repo := UserActivity{}
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
