package SQLServer

import (
	"testing"
	"time"

	"intelligentBI/pkg"
)

func TestUserActivity_GetActivityData(t *testing.T) {
	repo := UserActivity{}
	to := time.Now()
	from := to.AddDate(0, -2, 0)

	data, err := repo.GetActivityData(pkg.SynOps, from, to)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("columns: %v", data.Columns)
	for _, row := range data.Rows {
		t.Log(row)
	}
}
