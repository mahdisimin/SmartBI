package SQLServer

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"unicode/utf16"

	"intelligentBI/entity"
	"intelligentBI/repository/SQLServer"
)

// uuidPattern is the canonical 8-4-4-4-12 hex form SQL Server accepts for a
// UNIQUEIDENTIFIER.
var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// validateUserActivity checks an event against synops.UserActivity's column
// types and sizes before inserting it. SQL Server would reject such a row on
// every retry anyway; catching it here gives a precise reason and returns it
// as a SQLServer.PermanentError so the worker dead-letters it instead of
// retrying forever.
func validateUserActivity(e entity.UserActivityEvent) error {
	var problems []string
	add := func(format string, args ...any) { problems = append(problems, fmt.Sprintf(format, args...)) }

	if !uuidPattern.MatchString(e.EventID) {
		add("event_id %q is not a UUID", truncateForMessage(e.EventID))
	}
	if strings.TrimSpace(e.EventType) == "" {
		add("event_type is empty")
	}
	if e.OccurredAt.IsZero() {
		add("occurred_at is missing")
	}
	if !fitsSmallInt(e.SchemaVersion) {
		add("schema_version %d out of SMALLINT range", e.SchemaVersion)
	}
	if !fitsSmallInt(e.Result.StatusCode) {
		add("result.status_code %d out of SMALLINT range", e.Result.StatusCode)
	}
	if math.IsNaN(e.Result.DurationSeconds) || math.IsInf(e.Result.DurationSeconds, 0) {
		add("result.duration_seconds is not a finite number")
	}

	checkNVarChar := func(field, value string, max int) {
		if n := nvarcharLen(value); n > max {
			add("%s is %d characters, max %d", field, n, max)
		}
	}
	checkVarChar := func(field, value string, max int) {
		if n := len(value); n > max {
			add("%s is %d bytes, max %d", field, n, max)
		}
	}
	checkNVarChar("event_type", e.EventType, 50)
	checkNVarChar("source", e.Source, 100)
	checkNVarChar("request.country", e.Request.Country, 100)
	checkVarChar("request.ip_address", e.Request.IPAddress, 45)
	checkNVarChar("request.user_agent", e.Request.UserAgent, 1000)
	checkNVarChar("activity.name", e.Activity.Name, 200)
	checkNVarChar("activity.path", e.Activity.Path, 1000)
	checkVarChar("activity.method", e.Activity.Method, 10)

	if len(problems) == 0 {
		return nil
	}
	return SQLServer.PermanentError{Err: fmt.Errorf("invalid user activity event: %s", strings.Join(problems, "; "))}
}

func fitsSmallInt(v int) bool { return v >= math.MinInt16 && v <= math.MaxInt16 }

// nvarcharLen is a string's length as NVARCHAR counts it: UTF-16 code units.
func nvarcharLen(s string) int { return len(utf16.Encode([]rune(s))) }

func truncateForMessage(s string) string {
	const max = 64
	if r := []rune(s); len(r) > max {
		return string(r[:max]) + "…"
	}
	return s
}
