package SQLServer

import (
	"encoding/json"

	"intelligentBI/entity"
	"intelligentBI/repository/SQLServer"
)

type UserActivity struct {
}

func (s UserActivity) PersistUserActivity(event entity.UserActivityEvent) error {
	db, err := SQLServer.Connect()
	if err != nil {
		return err
	}
	defer db.Close()

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

	_, err = db.Exec(`INSERT INTO synops.UserActivity
		(EventID, EventType, SchemaVersion, OccurredAt, Source, Actor,
		 ResultStatusCode, ResultDurationSeconds, ResultError,
		 RequestCountry, RequestIPAddress, RequestUserAgent, RequestBody, RequestQueryParams,
		 ActivityID, ActivityName, ActivityPath, ActivityMethod)
		VALUES (@p1, @p2, @p3, @p4, @p5, @p6, @p7, @p8, @p9, @p10, @p11, @p12, @p13, @p14, @p15, @p16, @p17, @p18)`,
		event.EventID, event.EventType, event.SchemaVersion, event.OccurredAt, event.Source, actor,
		event.Result.StatusCode, event.Result.DurationSeconds, resultError,
		event.Request.Country, event.Request.IPAddress, event.Request.UserAgent, requestBody, requestQueryParams,
		event.Activity.ID, event.Activity.Name, event.Activity.Path, event.Activity.Method,
	)
	return err
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
