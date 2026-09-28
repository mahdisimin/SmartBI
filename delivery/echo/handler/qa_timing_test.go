package echowebframework

import (
	"net/http"
	"sort"
	"testing"
	"time"
)

// Login must take about the same time whether or not the phone number is
// registered; otherwise response time reveals which accounts exist even
// though the error message is identical.
func TestQA_Login_TimingDoesNotRevealAccountExistence(t *testing.T) {
	e := qaServer(newQARepo(), qaExportRepo{})
	median := func(body string) time.Duration {
		var d []time.Duration
		for i := 0; i < 7; i++ {
			start := time.Now()
			rec := qaDo(e, http.MethodPost, "/user/login", body)
			d = append(d, time.Since(start))
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("want 401, got %d", rec.Code)
			}
		}
		sort.Slice(d, func(i, j int) bool { return d[i] < d[j] })
		return d[len(d)/2]
	}
	unknown := median(`{"phone_number":"09129999999","password":"wrong"}`)
	wrongPw := median(`{"phone_number":"09120000001","password":"wrong"}`)
	t.Logf("median login latency: unknown phone=%v, known phone+wrong password=%v", unknown, wrongPw)
	if wrongPw > 5*unknown {
		t.Errorf("timing side channel: known phone is %.0fx slower than unknown (%v vs %v) — bcrypt only runs for existing accounts",
			float64(wrongPw)/float64(unknown), wrongPw, unknown)
	}
}
