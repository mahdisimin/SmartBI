package SQLServer

import (
	"errors"
	"strings"
	"testing"
	"time"

	"intelligentBI/entity"
)

func validEvent() entity.UserActivityEvent {
	return entity.UserActivityEvent{
		EventID:    "a544f4eb-41a0-4458-abba-7f3f35a31220",
		EventType:  "user.activity",
		OccurredAt: time.Now(),
		Source:     "server_backend",
		Result:     entity.ActivityResult{StatusCode: 200},
		Request:    entity.ActivityRequest{Country: "unknown", IPAddress: "10.0.0.1"},
		Activity:   entity.ActivityInfo{Name: "User_Account | Profile_Retrieve", Path: "/x", Method: "GET"},
	}
}

func TestValidateUserActivity(t *testing.T) {
	if err := validateUserActivity(validEvent()); err != nil {
		t.Fatalf("valid event rejected: %v", err)
	}
	// Exactly at the limit is fine, including characters NVARCHAR counts as 2.
	atLimit := validEvent()
	atLimit.Activity.Name = strings.Repeat("😀", 100) // 200 UTF-16 units
	if err := validateUserActivity(atLimit); err != nil {
		t.Fatalf("event at column limit rejected: %v", err)
	}

	for name, mutate := range map[string]func(e *entity.UserActivityEvent){
		"non-uuid event_id":       func(e *entity.UserActivityEvent) { e.EventID = "evt-123" },
		"missing event_id":        func(e *entity.UserActivityEvent) { e.EventID = "" },
		"empty event_type":        func(e *entity.UserActivityEvent) { e.EventType = " " },
		"missing occurred_at":     func(e *entity.UserActivityEvent) { e.OccurredAt = time.Time{} },
		"status_code overflow":    func(e *entity.UserActivityEvent) { e.Result.StatusCode = 70000 },
		"schema_version overflow": func(e *entity.UserActivityEvent) { e.SchemaVersion = -40000 },
		"activity.name too long":  func(e *entity.UserActivityEvent) { e.Activity.Name = strings.Repeat("😀", 101) },
		"ip_address too long":     func(e *entity.UserActivityEvent) { e.Request.IPAddress = strings.Repeat("1", 46) },
		"method too long":         func(e *entity.UserActivityEvent) { e.Activity.Method = "PROPFINDXYZ" },
	} {
		t.Run(name, func(t *testing.T) {
			e := validEvent()
			mutate(&e)
			err := validateUserActivity(e)
			var p interface{ Permanent() bool }
			if err == nil || !errors.As(err, &p) || !p.Permanent() {
				t.Fatalf("want a permanent validation error, got %v", err)
			}
		})
	}
}
