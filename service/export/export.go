package export

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"intelligentBI/entity"
	"intelligentBI/pkg"

	_ "time/tzdata" // embed the tz database for dashboardLocation

	"github.com/jalaali/go-jalaali"
)

// Repo is implemented by the repository layer. Given a product/system, it
// resolves which table holds that system's data and returns the normalized
// per-event facts for the given time range (a zero `from` means no lower
// bound). The repository owns interpreting
// that system's raw schema (parsing an actor blob, extracting a module name
// from an activity name, ...); this service owns turning those facts into
// the aggregated numbers a dashboard actually renders — no raw event ever
// reaches the caller.
type Repo interface {
	GetActivityEvents(product pkg.ProductList, from, to time.Time) ([]entity.ActivityEvent, error)
}

// dashboardLocation is the time zone trend buckets (day / week / month) and
// the day filter are computed in: the dashboard's users are in Iran, so a
// "day" is a Tehran calendar day, not a UTC one. The tz database is embedded
// (time/tzdata) so this works on hosts without one, e.g. Windows servers.
var dashboardLocation = mustLoadLocation("Asia/Tehran")

func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic(fmt.Sprintf("load time zone %q: %v", name, err))
	}
	return loc
}

// AccessRepo answers whether a user may see a product's dashboard.
type AccessRepo interface {
	HasDashboardAccess(userID int64, link string) (bool, error)
}

// ErrForbidden is returned when the requesting user has not been granted
// access to the product.
var ErrForbidden = errors.New("no access to this product")

type ExportService struct {
	Repository Repo
	// Access authorizes every export. When nil, every export is denied
	// (fail closed) — wire it up in production code.
	Access AccessRepo
}

func NewExportService(repo Repo) *ExportService {
	return &ExportService{Repository: repo}
}

// Filters mirrors the dashboard's own cross-filter state: the set of values
// currently selected per dimension. Empty means "no restriction on this
// dimension" (matches filters[dim].size===0 in the original JS).
type Filters struct {
	Days    []string
	Modules []string
	Methods []string
	UserIDs []int64
}

type ExportRequest struct {
	// UserID is the authenticated user the export is for.
	UserID  int64
	Product pkg.ProductList
	Filters Filters
}

// Export returns ErrForbidden unless request.UserID has been granted the
// product's dashboard.
func (e ExportService) Export(request ExportRequest) (entity.DashboardData, error) {
	if err := e.authorize(request.UserID, request.Product); err != nil {
		return entity.DashboardData{}, err
	}

	// The export covers the product's full history: a zero `from` means no
	// lower bound.
	events, err := e.Repository.GetActivityEvents(request.Product, time.Time{}, time.Now())
	if err != nil {
		return entity.DashboardData{}, fmt.Errorf("fetch activity events: %w", err)
	}

	return buildDashboardData(events, request.Filters), nil
}

func (e ExportService) authorize(userID int64, product pkg.ProductList) error {
	if e.Access == nil {
		return ErrForbidden
	}
	link, ok := pkg.DashboardLink(product)
	if !ok {
		return ErrForbidden
	}
	granted, err := e.Access.HasDashboardAccess(userID, link)
	if err != nil {
		return fmt.Errorf("check access: %w", err)
	}
	if !granted {
		return ErrForbidden
	}
	return nil
}

// ---------------------------------------------------------------------------
// Cross-filtering
//
// The original dashboard keeps every event in the browser and, for each
// panel, filters by every active dimension EXCEPT the one that panel itself
// represents (matchFilters/getFiltered(excludeDim)) — so clicking a module
// dims everything else without collapsing the module ranking itself to one
// bar. Replicating that server-side means: compute the panel-independent
// "categories" (the fixed universe of days/modules/methods/users, and their
// display order) once from the unfiltered event set, then compute each
// panel's counts from its own excluded-dimension-filtered subset.
// ---------------------------------------------------------------------------

type filterSet[T comparable] map[T]struct{}

func newFilterSet[T comparable](values []T) filterSet[T] {
	set := make(filterSet[T], len(values))
	for _, v := range values {
		set[v] = struct{}{}
	}
	return set
}

func (s filterSet[T]) has(v T) bool {
	_, ok := s[v]
	return ok
}

type activeFilters struct {
	days    filterSet[string]
	modules filterSet[string]
	methods filterSet[string]
	userIDs filterSet[int64]
}

func newActiveFilters(f Filters) activeFilters {
	return activeFilters{
		days:    newFilterSet(f.Days),
		modules: newFilterSet(f.Modules),
		methods: newFilterSet(f.Methods),
		userIDs: newFilterSet(f.UserIDs),
	}
}

// matches reports whether ev satisfies every active filter dimension except
// `exclude`.
func (af activeFilters) matches(ev entity.ActivityEvent, exclude string) bool {
	if exclude != "day" && len(af.days) > 0 {
		if !af.days.has(dayKey(ev.OccurredAt)) {
			return false
		}
	}
	if exclude != "module" && len(af.modules) > 0 {
		if ev.Module == "" || !af.modules.has(ev.Module) {
			return false
		}
	}
	if exclude != "method" && len(af.methods) > 0 {
		if ev.Method == "" || !af.methods.has(ev.Method) {
			return false
		}
	}
	if exclude != "userId" && len(af.userIDs) > 0 {
		if ev.UserID == nil || !af.userIDs.has(*ev.UserID) {
			return false
		}
	}
	return true
}

func filterEvents(events []entity.ActivityEvent, af activeFilters, exclude string) []entity.ActivityEvent {
	out := make([]entity.ActivityEvent, 0, len(events))
	for _, ev := range events {
		if af.matches(ev, exclude) {
			out = append(out, ev)
		}
	}
	return out
}

// categories is the fixed universe computed once from every event in the
// lookback window (never affected by the active filters), so panel ordering
// and which rows/bars exist at all stays stable as filters are toggled —
// only the counts next to them change.
type categories struct {
	days    []string // ascending
	weeks   []string // ascending
	months  []string // ascending
	modules []string // ranked by global count, descending
	methods []string // ranked by global count, descending
	userIDs []int64  // ranked by global action count, descending
}

func buildCategories(events []entity.ActivityEvent) categories {
	daySet := map[string]struct{}{}
	weekSet := map[string]struct{}{}
	monthSet := map[string]struct{}{}

	moduleCounts := map[string]int{}
	var moduleOrder []string
	methodCounts := map[string]int{}
	var methodOrder []string
	userCounts := map[int64]int{}
	var userOrder []int64

	for _, ev := range events {
		daySet[dayKey(ev.OccurredAt)] = struct{}{}
		weekSet[weekStartKey(ev.OccurredAt)] = struct{}{}
		monthSet[monthKey(ev.OccurredAt)] = struct{}{}

		if ev.Module != "" {
			if _, seen := moduleCounts[ev.Module]; !seen {
				moduleOrder = append(moduleOrder, ev.Module)
			}
			moduleCounts[ev.Module]++
		}
		if ev.Method != "" {
			if _, seen := methodCounts[ev.Method]; !seen {
				methodOrder = append(methodOrder, ev.Method)
			}
			methodCounts[ev.Method]++
		}
		if ev.UserID != nil {
			uid := *ev.UserID
			if _, seen := userCounts[uid]; !seen {
				userOrder = append(userOrder, uid)
			}
			userCounts[uid]++
		}
	}

	sort.SliceStable(moduleOrder, func(i, j int) bool { return moduleCounts[moduleOrder[i]] > moduleCounts[moduleOrder[j]] })
	sort.SliceStable(methodOrder, func(i, j int) bool { return methodCounts[methodOrder[i]] > methodCounts[methodOrder[j]] })
	sort.SliceStable(userOrder, func(i, j int) bool { return userCounts[userOrder[i]] > userCounts[userOrder[j]] })

	return categories{
		days:    sortedStringKeys(daySet),
		weeks:   sortedStringKeys(weekSet),
		months:  sortedStringKeys(monthSet),
		modules: moduleOrder,
		methods: methodOrder,
		userIDs: userOrder,
	}
}

func sortedStringKeys(set map[string]struct{}) []string {
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func buildDashboardData(events []entity.ActivityEvent, filters Filters) entity.DashboardData {
	cats := buildCategories(events)
	af := newActiveFilters(filters)

	fAll := filterEvents(events, af, "")
	fDay := filterEvents(events, af, "day")
	fModule := filterEvents(events, af, "module")
	fMethod := filterEvents(events, af, "method")
	fUser := filterEvents(events, af, "userId")

	userStats := userStatsFor(cats.userIDs, fUser)
	kpis, verdict := buildKPIsAndVerdict(fAll, cats, userStats)

	return entity.DashboardData{
		KPIs: kpis,
		DailyTrend: trendFor(cats.days, fDay, func(ev entity.ActivityEvent) string {
			return dayKey(ev.OccurredAt)
		}),
		WeeklyTrend: trendFor(cats.weeks, fDay, func(ev entity.ActivityEvent) string {
			return weekStartKey(ev.OccurredAt)
		}),
		MonthlyTrend: trendFor(cats.months, fDay, func(ev entity.ActivityEvent) string {
			return monthKey(ev.OccurredAt)
		}),
		TopModules:      rankedFor(cats.modules, fModule, func(ev entity.ActivityEvent) string { return ev.Module }),
		MethodBreakdown: rankedFor(cats.methods, fMethod, func(ev entity.ActivityEvent) string { return ev.Method }),
		Users:           userStats,
		Verdict:         verdict,
	}
}

func trendFor(keys []string, events []entity.ActivityEvent, keyFn func(entity.ActivityEvent) string) []entity.TrendPoint {
	counts := map[string]int{}
	for _, ev := range events {
		counts[keyFn(ev)]++
	}
	points := make([]entity.TrendPoint, 0, len(keys))
	for _, key := range keys {
		points = append(points, entity.TrendPoint{Key: key, Count: counts[key]})
	}
	return points
}

// rankedFor always returns one entry per name in `names` (the fixed,
// globally-ranked order from categories) — even a name with 0 matches in the
// currently-filtered `events` still appears, with Count: 0.
func rankedFor(names []string, events []entity.ActivityEvent, keyFn func(entity.ActivityEvent) string) []entity.RankedCount {
	counts := map[string]int{}
	for _, ev := range events {
		if key := keyFn(ev); key != "" {
			counts[key]++
		}
	}
	ranked := make([]entity.RankedCount, 0, len(names))
	for _, name := range names {
		ranked = append(ranked, entity.RankedCount{Name: name, Count: counts[name]})
	}
	return ranked
}

// userStatsFor always returns one row per id in `userIDs` (the fixed,
// globally-ranked order) — a user filtered out of `events` still appears,
// with zeroed-out stats, matching the dashboard's own always-visible table.
func userStatsFor(userIDs []int64, events []entity.ActivityEvent) []entity.UserStat {
	type agg struct {
		orgID        *int64
		orgRole      string
		actions      int
		modules      map[string]struct{}
		successCount int
		durationSum  float64
	}

	stats := make(map[int64]*agg, len(userIDs))
	for _, uid := range userIDs {
		stats[uid] = &agg{modules: map[string]struct{}{}}
	}

	for _, ev := range events {
		if ev.UserID == nil {
			continue
		}
		a, ok := stats[*ev.UserID]
		if !ok {
			continue
		}
		a.actions++
		a.orgID = ev.OrgID
		a.orgRole = ev.OrgRole
		if ev.Module != "" {
			a.modules[ev.Module] = struct{}{}
		}
		if ev.Status >= 200 && ev.Status < 400 {
			a.successCount++
		}
		a.durationSum += ev.Duration
	}

	result := make([]entity.UserStat, 0, len(userIDs))
	for _, uid := range userIDs {
		a := stats[uid]
		successRate, avgDuration := 0.0, 0.0
		if a.actions > 0 {
			successRate = float64(a.successCount) / float64(a.actions) * 100
			avgDuration = a.durationSum / float64(a.actions)
		}
		result = append(result, entity.UserStat{
			UserID:        uid,
			OrgID:         a.orgID,
			OrgRole:       a.orgRole,
			Actions:       a.actions,
			ModuleBreadth: len(a.modules),
			SuccessRate:   successRate,
			AvgDuration:   avgDuration,
		})
	}
	return result
}

// dayKey, weekStartKey and monthKey bucket an event by its calendar day,
// week and month in dashboardLocation, following Iranian conventions:
//
//   - day:   Gregorian date, "2006-01-02".
//   - week:  weeks start on Saturday; keyed by that Saturday's Gregorian
//     date, "2006-01-02".
//   - month: Solar Hijri (Jalali) month, keyed "YYYY-MM" in the Jalali
//     calendar, e.g. "1405-06" for Shahrivar 1405. Buckets follow real
//     Persian month boundaries (1 Shahrivar = 23 Aug), not Gregorian ones.
func dayKey(t time.Time) string { return t.In(dashboardLocation).Format("2006-01-02") }

// invalidMonthKey buckets timestamps the Jalali conversion cannot handle; it
// sorts before every real month.
const invalidMonthKey = "0000-00"

func monthKey(t time.Time) string {
	t = t.In(dashboardLocation)
	jy, jm, _, err := jalaali.ToJalaali(t.Year(), t.Month(), t.Day())
	if err != nil {
		// Only outside the algorithm's range (Jalali years -61..3177), i.e. a
		// corrupt timestamp — DATETIMEOFFSET itself allows years 1..9999.
		// Bucket it visibly instead of failing the whole export.
		return invalidMonthKey
	}
	return fmt.Sprintf("%04d-%02d", jy, int(jm))
}

func weekStartKey(t time.Time) string {
	t = t.In(dashboardLocation)
	// Days since the most recent Saturday: Saturday=0, Sunday=1 .. Friday=6.
	sinceSaturday := (int(t.Weekday()) + 1) % 7
	start := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, dashboardLocation).AddDate(0, 0, -sinceSaturday)
	return start.Format("2006-01-02")
}

// buildKPIsAndVerdict computes the KPI cards from fAll (every filter
// applied), except ModuleCount and DaysCovered which — like the dashboard's
// own kpiModules/rangeLabel cards — always reflect the full, unfiltered
// lookback window, not the current selection.
func buildKPIsAndVerdict(fAll []entity.ActivityEvent, cats categories, userStats []entity.UserStat) (entity.DashboardKPIs, entity.DashboardVerdict) {
	kpis := entity.DashboardKPIs{
		TotalEvents: len(fAll),
		ModuleCount: len(cats.modules),
		DaysCovered: len(cats.days),
	}

	uniqueUsers := map[int64]struct{}{}
	uniqueOrgs := map[int64]struct{}{}
	var durationSum float64
	for _, ev := range fAll {
		if ev.UserID != nil {
			uniqueUsers[*ev.UserID] = struct{}{}
		}
		if ev.OrgID != nil {
			uniqueOrgs[*ev.OrgID] = struct{}{}
		}
		if ev.Status >= 200 && ev.Status < 400 {
			kpis.SuccessCount++
		}
		if ev.Status >= 400 {
			kpis.ErrorCount++
		}
		durationSum += ev.Duration
	}
	kpis.UniqueUsers = len(uniqueUsers)
	kpis.UniqueOrgs = len(uniqueOrgs)
	if kpis.TotalEvents > 0 {
		kpis.SuccessRate = float64(kpis.SuccessCount) / float64(kpis.TotalEvents) * 100
		kpis.AvgDuration = durationSum / float64(kpis.TotalEvents)
	}

	var breadthSum float64
	for _, u := range userStats {
		breadthSum += float64(u.ModuleBreadth)
	}
	avgBreadth := 0.0
	if len(userStats) > 0 {
		avgBreadth = breadthSum / float64(len(userStats))
	}

	return kpis, buildVerdict(kpis, avgBreadth)
}

func buildVerdict(kpis entity.DashboardKPIs, avgBreadth float64) entity.DashboardVerdict {
	errorShare := 0.0
	if kpis.TotalEvents > 0 {
		errorShare = float64(kpis.ErrorCount) / float64(kpis.TotalEvents) * 100
	}

	reliability := "bad"
	switch {
	case errorShare < 3:
		reliability = "ok"
	case errorShare < 10:
		reliability = "warn"
	}

	breadth := "bad"
	switch {
	case avgBreadth >= float64(kpis.ModuleCount)*0.4:
		breadth = "ok"
	case avgBreadth >= float64(kpis.ModuleCount)*0.2:
		breadth = "warn"
	}

	sample := "ok"
	if kpis.UniqueUsers < 10 || kpis.DaysCovered < 14 {
		sample = "warn"
	}

	return entity.DashboardVerdict{
		ReliabilitySignal: reliability,
		BreadthSignal:     breadth,
		SampleSignal:      sample,
	}
}
