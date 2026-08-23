package main

import (
	"encoding/json"
	"fmt"

	"intelligentBI/entity"
)

// ParseUserActivityEvent unmarshals a single raw user.activity JSON message.
func ParseUserActivityEvent(data []byte) (entity.UserActivityEvent, error) {
	var event entity.UserActivityEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return event, fmt.Errorf("parse user activity event: %w", err)
	}
	return event, nil
}

// PrintUserActivityEvent writes every field of the event to stdout, one per line.
func PrintUserActivityEvent(e entity.UserActivityEvent) {
	fmt.Println("event_id                :", e.EventID)
	fmt.Println("event_type              :", e.EventType)
	fmt.Println("schema_version          :", e.SchemaVersion)
	fmt.Println("occurred_at             :", e.OccurredAt)
	fmt.Println("source                  :", e.Source)
	fmt.Println("actor                   :", e.Actor)
	fmt.Println("result.error            :", e.Result.Error)
	fmt.Println("result.status_code      :", e.Result.StatusCode)
	fmt.Println("result.duration_seconds :", e.Result.DurationSeconds)
	fmt.Println("request.body            :", e.Request.Body)
	fmt.Println("request.country         :", e.Request.Country)
	fmt.Println("request.ip_address      :", e.Request.IPAddress)
	fmt.Println("request.user_agent      :", e.Request.UserAgent)
	fmt.Println("request.query_params    :", e.Request.QueryParams)
	fmt.Println("activity.id             :", e.Activity.ID)
	fmt.Println("activity.name           :", e.Activity.Name)
	fmt.Println("activity.path           :", e.Activity.Path)
	fmt.Println("activity.method         :", e.Activity.Method)
}
