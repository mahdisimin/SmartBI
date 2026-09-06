package entity

// DashboardKPIs are the top-line numbers shown as KPI cards.
type DashboardKPIs struct {
	TotalEvents  int     `json:"total_events"`
	UniqueUsers  int     `json:"unique_users"`
	UniqueOrgs   int     `json:"unique_orgs"`
	SuccessCount int     `json:"success_count"`
	ErrorCount   int     `json:"error_count"`
	SuccessRate  float64 `json:"success_rate"`
	AvgDuration  float64 `json:"avg_duration_seconds"`
	ModuleCount  int     `json:"module_count"`
	DaysCovered  int     `json:"days_covered"`
}

// TrendPoint is one bucket of a time-series (a day, an ISO week start, or a
// month, depending on which trend it appears in).
type TrendPoint struct {
	Key   string `json:"key"`
	Count int    `json:"count"`
}

// RankedCount is a named bucket (a module or an HTTP method) with its count,
// pre-sorted descending.
type RankedCount struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// UserStat is one row of the per-user activity table.
type UserStat struct {
	UserID        int64   `json:"user_id"`
	OrgID         *int64  `json:"org_id,omitempty"`
	OrgRole       string  `json:"org_role,omitempty"`
	Actions       int     `json:"actions"`
	ModuleBreadth int     `json:"module_breadth"`
	SuccessRate   float64 `json:"success_rate"`
	AvgDuration   float64 `json:"avg_duration_seconds"`
}

// DashboardVerdict holds the pre-computed ok/warn/bad signal for each summary
// card — the raw thresholds a UI would otherwise have to evaluate itself.
type DashboardVerdict struct {
	ReliabilitySignal string `json:"reliability_signal"`
	BreadthSignal     string `json:"breadth_signal"`
	SampleSignal      string `json:"sample_signal"`
}

// DashboardData is the complete, ready-to-render payload for a product's
// activity dashboard — every number already aggregated server-side. Nothing
// in here requires the client to iterate raw events again.
type DashboardData struct {
	KPIs            DashboardKPIs    `json:"kpis"`
	DailyTrend      []TrendPoint     `json:"daily_trend"`
	WeeklyTrend     []TrendPoint     `json:"weekly_trend"`
	MonthlyTrend    []TrendPoint     `json:"monthly_trend"`
	TopModules      []RankedCount    `json:"top_modules"`
	MethodBreakdown []RankedCount    `json:"method_breakdown"`
	Users           []UserStat       `json:"users"`
	Verdict         DashboardVerdict `json:"verdict"`
}
