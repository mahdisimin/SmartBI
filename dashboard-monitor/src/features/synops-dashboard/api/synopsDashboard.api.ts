import { apiClient } from '@/api/client'
import type { DashboardData, DashboardFilters } from '../types'

function csv<T>(set: Set<T>): string | undefined {
    return set.size > 0 ? Array.from(set).join(',') : undefined
}

export const synopsDashboardApi = {
    getDashboard: async (filters: DashboardFilters): Promise<DashboardData> => {
        const response = await apiClient.get<DashboardData>('/export/synops', {
            params: {
                days: csv(filters.days),
                modules: csv(filters.modules),
                methods: csv(filters.methods),
                users: csv(filters.userIds),
            },
        })
        return response.data
    },
}
