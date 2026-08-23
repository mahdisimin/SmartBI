package SQLServer

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"intelligentBI/entity"
	"intelligentBI/pkg"
	"intelligentBI/repository/SQLServer"

	mssql "github.com/denisenkom/go-mssqldb"
)

// activityColumns is the column order returned in every entity.DataSet
// produced by this file, and matches the CSV column order downstream.
var activityColumns = []string{
	"EventID", "EventType", "SchemaVersion", "OccurredAt", "Source", "Actor",
	"ResultStatusCode", "ResultDurationSeconds", "ResultError",
	"RequestCountry", "RequestIPAddress", "RequestUserAgent", "RequestBody", "RequestQueryParams",
	"ActivityID", "ActivityName", "ActivityPath", "ActivityMethod",
}

// GetActivityData implements service/export.Repo: given a product, it
// resolves which table holds that product's data and returns the rows in
// [from, to], already flattened into entity.DataSet.
func (s UserActivity) GetActivityData(product pkg.ProductList, from, to time.Time) (entity.DataSet, error) {
	switch product {
	case pkg.SynOps:
		return s.getSynopsActivityData(from, to)
	default:
		return entity.DataSet{}, fmt.Errorf("SQLServer repository: unsupported product %d", product)
	}
}

func (s UserActivity) getSynopsActivityData(from, to time.Time) (entity.DataSet, error) {
	db, err := SQLServer.Connect()
	if err != nil {
		return entity.DataSet{}, err
	}
	defer db.Close()

	sqlRows, err := db.Query(`SELECT EventID, EventType, SchemaVersion, OccurredAt, Source, Actor,
			ResultStatusCode, ResultDurationSeconds, ResultError,
			RequestCountry, RequestIPAddress, RequestUserAgent, RequestBody, RequestQueryParams,
			ActivityID, ActivityName, ActivityPath, ActivityMethod
		FROM synops.UserActivity
		WHERE OccurredAt >= @p1 AND OccurredAt <= @p2
		ORDER BY OccurredAt`, from, to)
	if err != nil {
		return entity.DataSet{}, err
	}
	defer sqlRows.Close()

	var rows [][]any
	for sqlRows.Next() {
		var (
			eventID                          mssql.UniqueIdentifier
			eventType, source                string
			schemaVersion, resultStatusCode  int
			occurredAt                       time.Time
			actor, resultError               sql.NullString
			resultDurationSeconds            float64
			requestCountry, requestIPAddress string
			requestUserAgent                 sql.NullString
			requestBody, requestQueryParams  sql.NullString
			activityID                       int64
			activityName, activityPath       string
			activityMethod                   string
		)

		if err := sqlRows.Scan(
			&eventID, &eventType, &schemaVersion, &occurredAt, &source, &actor,
			&resultStatusCode, &resultDurationSeconds, &resultError,
			&requestCountry, &requestIPAddress, &requestUserAgent, &requestBody, &requestQueryParams,
			&activityID, &activityName, &activityPath, &activityMethod,
		); err != nil {
			return entity.DataSet{}, err
		}

		rows = append(rows, []any{
			eventID.String(),
			eventType,
			strconv.Itoa(schemaVersion),
			occurredAt.Format(time.RFC3339),
			source,
			jsonCell(actor),
			strconv.Itoa(resultStatusCode),
			strconv.FormatFloat(resultDurationSeconds, 'f', -1, 64),
			jsonCell(resultError),
			requestCountry,
			requestIPAddress,
			requestUserAgent.String,
			jsonCell(requestBody),
			jsonCell(requestQueryParams),
			strconv.FormatInt(activityID, 10),
			activityName,
			activityPath,
			activityMethod,
		})
	}
	if err := sqlRows.Err(); err != nil {
		return entity.DataSet{}, err
	}

	return entity.DataSet{Columns: activityColumns, Rows: rows}, nil
}

// jsonCell turns a nullable NVARCHAR(MAX) column that stores JSON text (see
// useractivity.go's nullableJSON, which is what wrote it) back into a value
// that marshals as real, unescaped JSON: nil when the column is NULL,
// otherwise the stored text wrapped as json.RawMessage.
func jsonCell(value sql.NullString) any {
	if !value.Valid {
		return nil
	}
	return json.RawMessage(value.String)
}
