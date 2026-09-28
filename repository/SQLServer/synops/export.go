package SQLServer

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"intelligentBI/entity"
	"intelligentBI/pkg"
)

// GetActivityEvents implements service/export.Repo: given a product, it
// resolves which table holds that product's data and returns the normalized
// per-event facts needed to build a dashboard. Only the columns a dashboard
// can actually use are selected — no EventID, request body/IP/user-agent,
// query params, or anything else leaves the database.
func (s UserActivity) GetActivityEvents(product pkg.ProductList, from, to time.Time) ([]entity.ActivityEvent, error) {
	switch product {
	case pkg.SynOps:
		return s.getSynopsActivityEvents(from, to)
	default:
		return nil, fmt.Errorf("SQLServer repository: unsupported product %d", product)
	}
}

// synopsHealthCheckFilter excludes uptime-monitoring probes, which are not
// user activity. In the data every probe carries all of: ActivityName
// 'health-check' (Synops's own classification), path '/healthy', user agent
// "Blackbox Exporter/...", and no actor — the three request markers select
// exactly the same rows (5,538 of 15,740 at the time of writing), and no
// request with an actor matches. The name is the primary signal; the path is
// a backstop in case the source renames the activity.
const synopsHealthCheckFilter = `ActivityName <> 'health-check' AND ActivityPath <> '/healthy'`

func (s UserActivity) getSynopsActivityEvents(from, to time.Time) ([]entity.ActivityEvent, error) {
	query := `SELECT OccurredAt, Actor, ResultStatusCode, ResultDurationSeconds, ActivityName, ActivityMethod
		FROM synops.UserActivity
		WHERE OccurredAt <= @p1 AND ` + synopsHealthCheckFilter
	args := []any{to}
	if !from.IsZero() { // zero `from` = full history, no lower bound
		query += ` AND OccurredAt >= @p2`
		args = append(args, from)
	}
	query += ` ORDER BY OccurredAt`

	sqlRows, err := s.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer sqlRows.Close()

	var events []entity.ActivityEvent
	for sqlRows.Next() {
		var (
			occurredAt   time.Time
			actor        sql.NullString
			statusCode   int
			durationSecs float64
			activityName string
			method       string
		)
		if err := sqlRows.Scan(&occurredAt, &actor, &statusCode, &durationSecs, &activityName, &method); err != nil {
			return nil, err
		}

		userID, orgID, orgRole := parseActor(actor)

		events = append(events, entity.ActivityEvent{
			OccurredAt: occurredAt,
			UserID:     userID,
			OrgID:      orgID,
			OrgRole:    orgRole,
			Status:     statusCode,
			Duration:   durationSecs,
			Module:     extractModule(activityName),
			Method:     method,
		})
	}
	if err := sqlRows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

// synopsActor is the subset of the Actor JSON blob the dashboard needs.
type synopsActor struct {
	ID               *int64 `json:"id"`
	OrganizationID   *int64 `json:"organization_id"`
	OrganizationRole string `json:"organization_role"`
}

func parseActor(value sql.NullString) (userID, orgID *int64, orgRole string) {
	if !value.Valid {
		return nil, nil, ""
	}
	var a synopsActor
	if err := json.Unmarshal([]byte(value.String), &a); err != nil {
		return nil, nil, ""
	}
	return a.ID, a.OrganizationID, a.OrganizationRole
}

// extractModule pulls the module/feature name out of Synops's
// "Module | Sub-action" activity-name convention, e.g.
// "User_Account | Profile_Retrieve" -> "User_Account". Returns "" when the
// activity name is blank or the source system couldn't classify it
// ("Unknown"), matching how the dashboard already treats those two cases.
func extractModule(activityName string) string {
	module := strings.TrimSpace(strings.SplitN(activityName, "|", 2)[0])
	if module == "" || module == "Unknown" {
		return ""
	}
	return module
}
