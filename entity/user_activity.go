package entity

import "time"

// UserActivityEvent mirrors the "user.activity" event envelope produced on the
// stinas.user-activities.v1 Kafka topic.
//
// Actor, Result.Error, Request.Body and Request.QueryParams are null in every
// sample seen so far, so their real shape is unknown — they are typed as `any`
// (whatever JSON value is present) instead of a guessed struct.
type UserActivityEvent struct {
	Actor         any             `json:"actor"`
	Result        ActivityResult  `json:"result"`
	Source        string          `json:"source"`
	Request       ActivityRequest `json:"request"`
	Activity      ActivityInfo    `json:"activity"`
	EventID       string          `json:"event_id"`
	EventType     string          `json:"event_type"`
	OccurredAt    time.Time       `json:"occurred_at"`
	SchemaVersion int             `json:"schema_version"`
}

// ActivityResult describes the outcome of the recorded request.
type ActivityResult struct {
	Error           any     `json:"error"`
	StatusCode      int     `json:"status_code"`
	DurationSeconds float64 `json:"duration_seconds"`
}

// ActivityRequest describes the inbound HTTP request that triggered the activity.
type ActivityRequest struct {
	Body        any    `json:"body"`
	Country     string `json:"country"`
	IPAddress   string `json:"ip_address"`
	UserAgent   string `json:"user_agent"`
	QueryParams any    `json:"query_params"`
}

// ActivityInfo is the normalized business action the source system mapped the request to.
type ActivityInfo struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Path   string `json:"path"`
	Method string `json:"method"`
}
