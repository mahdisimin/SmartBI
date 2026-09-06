import { useQuery } from '@tanstack/react-query'
import { synopsDashboardApi } from '../api/synopsDashboard.api'
import type { DashboardFilters } from '../types'

/** A stable, serializable key for the current filter state so TanStack Query
 * refetches whenever any dimension changes (Sets aren't comparable by value). */
function filtersKey(filters: DashboardFilters) {
    return [
        Array.from(filters.days).sort(),
        Array.from(filters.modules).sort(),
        Array.from(filters.methods).sort(),
        Array.from(filters.userIds).sort((a, b) => a - b),
    ]
}

export function useDashboardData(filters: DashboardFilters) {
    return useQuery({
        queryKey: ['synops-dashboard', filtersKey(filters)],
        queryFn: () => synopsDashboardApi.getDashboard(filters),
        placeholderData: (previousData) => previousData,
    })
}
