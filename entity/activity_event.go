package entity

import "time"

// ActivityEvent is the normalized, per-event fact a repository extracts from
// a product's raw activity data — just enough for the export service to
// aggregate into dashboard-ready summaries. Deliberately narrow: no request
// body, query params, IP/user-agent, or IDs the dashboard never uses — those
// never need to leave the database.
type ActivityEvent struct {
	OccurredAt time.Time
	UserID     *int64
	OrgID      *int64
	OrgRole    string
	Status     int
	Duration   float64
	// Module is the feature/module a request belongs to, extracted from the
	// source system's activity name (e.g. "User_Account | Profile_Retrieve"
	// -> "User_Account"). Empty when the source activity name is blank/unknown.
	Module string
	Method string
}
