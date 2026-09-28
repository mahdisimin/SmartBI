package SQLServer

import (
	"encoding/json"
	"time"

	"intelligentBI/entity"
	"intelligentBI/repository/SQLServer"

	"github.com/jmoiron/sqlx"
)

// occurredAtPrecision matches synops.UserActivity.OccurredAt, a
// DATETIMEOFFSET(6). Truncating before insert stores a predictable instant;
// otherwise SQL Server rounds the 7th fractional digit, which can move the
// stored time up to 0.5µs later than the event's.
const occurredAtPrecision = time.Microsecond

type UserActivity struct {
	DB *sqlx.DB
}

func NewUserActivity(db *sqlx.DB) UserActivity {
	return UserActivity{DB: db}
}

// PersistUserActivity inserts one event into synops.UserActivity. It is
// idempotent: EventID is UNIQUE (UQ_UserActivity_EventID), so an event
// redelivered by Kafka (e.g. after a crash between insert and offset commit)
// hits a duplicate key, which is treated as success — it is already stored.
//
// An event SQL Server can never accept (invalid EventID, a value too long or
// out of range for its column, ...) is returned as a SQLServer.PermanentError
// — whether caught by validation up front or raised by the insert itself — so
// the caller can dead-letter it instead of retrying forever.
func (s UserActivity) PersistUserActivity(event entity.UserActivityEvent) error {
	if err := validateUserActivity(event); err != nil {
		return err
	}

	actor, err := nullableJSON(event.Actor)
	if err != nil {
		return err
	}
	resultError, err := nullableJSON(event.Result.Error)
	if err != nil {
		return err
	}
	requestBody, err := nullableJSON(event.Request.Body)
	if err != nil {
		return err
	}
	requestQueryParams, err := nullableJSON(event.Request.QueryParams)
	if err != nil {
		return err
	}

	_, err = s.DB.Exec(`INSERT INTO synops.UserActivity
		(EventID, EventType, SchemaVersion, OccurredAt, Source, Actor,
		 ResultStatusCode, ResultDurationSeconds, ResultError,
		 RequestCountry, RequestIPAddress, RequestUserAgent, RequestBody, RequestQueryParams,
		 ActivityID, ActivityName, ActivityPath, ActivityMethod)
		VALUES (@p1, @p2, @p3, @p4, @p5, @p6, @p7, @p8, @p9, @p10, @p11, @p12, @p13, @p14, @p15, @p16, @p17, @p18)`,
		event.EventID, event.EventType, event.SchemaVersion, event.OccurredAt.Truncate(occurredAtPrecision), event.Source, actor,
		event.Result.StatusCode, event.Result.DurationSeconds, resultError,
		event.Request.Country, event.Request.IPAddress, event.Request.UserAgent, requestBody, requestQueryParams,
		event.Activity.ID, event.Activity.Name, event.Activity.Path, event.Activity.Method,
	)
	if SQLServer.IsDuplicateKey(err) {
		return nil
	}
	return SQLServer.ClassifyError(err)
}

// nullableJSON marshals value to a JSON string for storage in an NVARCHAR(MAX)
// column, or returns nil (SQL NULL) when value itself is nil.
func nullableJSON(value any) (any, error) {
	if value == nil {
		return nil, nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return string(data), nil
}
