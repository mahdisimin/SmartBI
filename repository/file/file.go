package file

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
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

func (r Repository) GetActivityEvents(product pkg.ProductList, from, to time.Time) ([]entity.ActivityEvent, error) {
	switch product {
	case pkg.SynOps:
		return r.getSynopsActivityEvents(from, to)
	default:
		return nil, fmt.Errorf("file repository: unsupported product %d", product)
	}
}

func (r Repository) getSynopsActivityEvents(from, to time.Time) ([]entity.ActivityEvent, error) {
	data, err := os.ReadFile(r.FilePath)
	if err != nil {
		return nil, err
	}

	var raw []entity.UserActivityEvent
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	events := make([]entity.ActivityEvent, 0, len(raw))
	for _, event := range raw {
		if event.OccurredAt.Before(from) || event.OccurredAt.After(to) {
			continue
		}
		events = append(events, toActivityEvent(event))
	}
	return events, nil
}

// synopsActor is the subset of the Actor JSON blob the dashboard needs.
type synopsActor struct {
	ID               *int64 `json:"id"`
	OrganizationID   *int64 `json:"organization_id"`
	OrganizationRole string `json:"organization_role"`
}

func toActivityEvent(event entity.UserActivityEvent) entity.ActivityEvent {
	var a synopsActor
	if event.Actor != nil {
		// event.Actor was unmarshaled generically (as `any`), so round-trip it
		// through JSON rather than assuming its dynamic Go type.
		if data, err := json.Marshal(event.Actor); err == nil {
			_ = json.Unmarshal(data, &a)
		}
	}

	return entity.ActivityEvent{
		OccurredAt: event.OccurredAt,
		UserID:     a.ID,
		OrgID:      a.OrganizationID,
		OrgRole:    a.OrganizationRole,
		Status:     event.Result.StatusCode,
		Duration:   event.Result.DurationSeconds,
		Module:     extractModule(event.Activity.Name),
		Method:     event.Activity.Method,
	}
}

// extractModule mirrors repository/SQLServer/synops/export.go's helper of the
// same name: pulls the module out of Synops's "Module | Sub-action"
// activity-name convention.
func extractModule(activityName string) string {
	module := strings.TrimSpace(strings.SplitN(activityName, "|", 2)[0])
	if module == "" || module == "Unknown" {
		return ""
	}
	return module
}
