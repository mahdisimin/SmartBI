package file

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"intelligentBI/entity"
	"intelligentBI/pkg"
)

// Repository implements service/export.Repo by reading activity data from a
// local JSON file instead of a live database. It exists so the export
// service can be exercised end-to-end without a SQL Server connection; a
// SQLServer-backed implementation can replace it later without the service
// changing at all.
type Repository struct {
	// FilePath points at a JSON file containing a []entity.UserActivityEvent
	// array — the same shape external/synops/ver2's Kafka worker parses.
	FilePath string
}

func NewRepository(filePath string) *Repository {
	return &Repository{FilePath: filePath}
}

func (r Repository) GetActivityData(product pkg.ProductList, from, to time.Time) (entity.DataSet, error) {
	switch product {
	case pkg.SynOps:
		return r.getSynopsActivityData(from, to)
	default:
		return entity.DataSet{}, fmt.Errorf("file repository: unsupported product %d", product)
	}
}

func (r Repository) getSynopsActivityData(from, to time.Time) (entity.DataSet, error) {
	data, err := os.ReadFile(r.FilePath)
	if err != nil {
		return entity.DataSet{}, err
	}

	var events []entity.UserActivityEvent
	if err := json.Unmarshal(data, &events); err != nil {
		return entity.DataSet{}, err
	}

	columns := []string{
		"EventID", "EventType", "SchemaVersion", "OccurredAt", "Source", "Actor",
		"ResultStatusCode", "ResultDurationSeconds", "ResultError",
		"RequestCountry", "RequestIPAddress", "RequestUserAgent", "RequestBody", "RequestQueryParams",
		"ActivityID", "ActivityName", "ActivityPath", "ActivityMethod",
	}

	rows := make([][]any, 0, len(events))
	for _, event := range events {
		if event.OccurredAt.Before(from) || event.OccurredAt.After(to) {
			continue
		}
		rows = append(rows, []any{
			event.EventID,
			event.EventType,
			strconv.Itoa(event.SchemaVersion),
			event.OccurredAt.Format(time.RFC3339),
			event.Source,
			jsonCell(event.Actor),
			strconv.Itoa(event.Result.StatusCode),
			strconv.FormatFloat(event.Result.DurationSeconds, 'f', -1, 64),
			jsonCell(event.Result.Error),
			event.Request.Country,
			event.Request.IPAddress,
			event.Request.UserAgent,
			jsonCell(event.Request.Body),
			jsonCell(event.Request.QueryParams),
			strconv.FormatInt(event.Activity.ID, 10),
			event.Activity.Name,
			event.Activity.Path,
			event.Activity.Method,
		})
	}

	return entity.DataSet{Columns: columns, Rows: rows}, nil
}

// jsonCell mirrors the encoding used when this data is persisted to SQL
// Server (see repository/SQLServer/synops/useractivity.go's nullableJSON):
// nil stays nil (-> JSON null), anything else is embedded as raw JSON so it
// round-trips as real JSON, not an escaped string.
func jsonCell(value any) any {
	if value == nil {
		return nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	return json.RawMessage(data)
}
