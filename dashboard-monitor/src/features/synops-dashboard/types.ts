// Mirrors entity.DashboardData and friends from the Go backend
// (service/export). Every number here is already aggregated server-side —
// nothing in this feature re-derives numbers from raw events.

export interface DashboardKPIs {
    total_events: number
    unique_users: number
    unique_orgs: number
    success_count: number
    error_count: number
    success_rate: number
    avg_duration_seconds: number
    module_count: number
    days_covered: number
}

export interface TrendPoint {
    key: string
    count: number
}

export interface RankedCount {
    name: string
    count: number
}

export interface UserStat {
    user_id: number
    org_id?: number
    org_role?: string
    actions: number
    module_breadth: number
    success_rate: number
    avg_duration_seconds: number
}

export type VerdictSignal = 'ok' | 'warn' | 'bad'

export interface DashboardVerdict {
    reliability_signal: VerdictSignal
    breadth_signal: VerdictSignal
    sample_signal: VerdictSignal
}

export interface DashboardData {
    kpis: DashboardKPIs
    daily_trend: TrendPoint[]
    weekly_trend: TrendPoint[]
    monthly_trend: TrendPoint[]
    top_modules: RankedCount[]
    method_breakdown: RankedCount[]
    users: UserStat[]
    verdict: DashboardVerdict
}

// Cross-filter state — one Set per dimension, matching the sample dashboard's
// own `filters` object. Empty set = no restriction on that dimension.
export interface DashboardFilters {
    days: Set<string>
    modules: Set<string>
    methods: Set<string>
    userIds: Set<number>
}

export const emptyFilters = (): DashboardFilters => ({
    days: new Set(),
    modules: new Set(),
    methods: new Set(),
    userIds: new Set(),
})

export type TrendLevel = 'day' | 'week' | 'month'
