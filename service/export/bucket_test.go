package export

import (
	"intelligentBI/entity"
	"testing"
	"time"
)

// Buckets use Tehran local time (UTC+03:30), Saturday-start weeks and
// Solar Hijri (Jalali) months.
func TestBucketKeys_TehranTime(t *testing.T) {
	for _, tc := range []struct {
		name             string
		utc              string
		day, week, month string
	}{
		// 21:00 UTC Mon 31 Aug = 00:30 Tue 1 Sep in Tehran: next day. Both
		// dates are in Shahrivar 1405 (23 Aug – 22 Sep): a Gregorian month
		// boundary is not a Jalali one.
		{"crosses Gregorian month, same Jalali month", "2026-08-31T21:00:00Z", "2026-09-01", "2026-08-29", "1405-06"},
		// 20:29 UTC is still 23:59 the same Tehran day.
		{"just before Tehran midnight", "2026-08-31T20:29:00Z", "2026-08-31", "2026-08-29", "1405-06"},
		// 31 Shahrivar -> 1 Mehr at Tehran midnight (20:30 UTC on 22 Sep).
		{"last minute of Shahrivar", "2026-09-22T20:29:00Z", "2026-09-22", "2026-09-19", "1405-06"},
		{"first minute of Mehr", "2026-09-22T20:30:00Z", "2026-09-23", "2026-09-19", "1405-07"},
		// Nowruz: 29 Esfand 1404 (1404 is not a leap year) -> 1 Farvardin 1405
		// on 21 Mar 2026; the Jalali year changes too.
		{"last day of 1404", "2026-03-20T12:00:00Z", "2026-03-20", "2026-03-14", "1404-12"},
		{"Nowruz 1405", "2026-03-21T12:00:00Z", "2026-03-21", "2026-03-21", "1405-01"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ts, err := time.Parse(time.RFC3339, tc.utc)
			if err != nil {
				t.Fatal(err)
			}
			if got := dayKey(ts); got != tc.day {
				t.Errorf("dayKey = %s, want %s", got, tc.day)
			}
			if got := weekStartKey(ts); got != tc.week {
				t.Errorf("weekStartKey = %s, want %s", got, tc.week)
			}
			if got := monthKey(ts); got != tc.month {
				t.Errorf("monthKey = %s, want %s", got, tc.month)
			}
		})
	}
}

// Weeks run Saturday..Friday in Tehran time: every day of one week shares a
// key, and the next Saturday starts a new one.
func TestWeekStartKey_SaturdayStart(t *testing.T) {
	sat := time.Date(2026, 9, 19, 0, 0, 0, 0, dashboardLocation) // a Saturday
	if sat.Weekday() != time.Saturday {
		t.Fatalf("fixture: %s is a %s", sat.Format("2006-01-02"), sat.Weekday())
	}
	for d := 0; d < 7; d++ {
		day := sat.AddDate(0, 0, d).Add(23*time.Hour + 59*time.Minute) // late in the day
		if got := weekStartKey(day); got != "2026-09-19" {
			t.Errorf("%s (%s): week = %s, want 2026-09-19", day.Format("2006-01-02"), day.Weekday(), got)
		}
	}
	if got := weekStartKey(sat.AddDate(0, 0, 7)); got != "2026-09-26" {
		t.Errorf("next Saturday: week = %s, want 2026-09-26", got)
	}
}

func TestMonthKey_OutOfRangeDoesNotPanic(t *testing.T) {
	if got := monthKey(time.Date(9999, 12, 31, 0, 0, 0, 0, time.UTC)); got != invalidMonthKey {
		t.Errorf("year 9999: got %q, want %q", got, invalidMonthKey)
	}
}

// The day filter must use the same Tehran day as the daily trend, or picking
// a bar on the chart would select a different set of events.
func TestDayFilter_UsesTehranDay(t *testing.T) {
	ts, _ := time.Parse(time.RFC3339, "2026-08-31T21:00:00Z") // 2026-09-01 in Tehran
	af := newActiveFilters(Filters{Days: []string{"2026-09-01"}})
	if !af.matches(eventAt(ts), "") {
		t.Error("event on Tehran day 2026-09-01 not matched by days=2026-09-01")
	}
	af = newActiveFilters(Filters{Days: []string{"2026-08-31"}})
	if af.matches(eventAt(ts), "") {
		t.Error("event matched by its UTC day instead of its Tehran day")
	}
}

func eventAt(ts time.Time) entity.ActivityEvent { return entity.ActivityEvent{OccurredAt: ts} }
