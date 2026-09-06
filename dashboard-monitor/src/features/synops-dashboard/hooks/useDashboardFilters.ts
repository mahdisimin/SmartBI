import { useCallback, useState } from 'react'
import { emptyFilters, type DashboardFilters } from '../types'

type Dimension = keyof DashboardFilters

export interface UseDashboardFiltersReturn {
    filters: DashboardFilters
    toggle: (dim: Dimension, value: string | number) => void
    clearOne: (dim: Dimension, value: string | number) => void
    clearAll: () => void
    anyActive: boolean
}

/** Mirrors the sample dashboard's toggleFilter/clearOne/clearAll — cross-filter
 * state lives here; every change re-triggers the query in useDashboardData. */
export function useDashboardFilters(): UseDashboardFiltersReturn {
    const [filters, setFilters] = useState<DashboardFilters>(emptyFilters)

    const toggle = useCallback((dim: Dimension, value: string | number) => {
        setFilters((prev) => {
            const next = { ...prev, [dim]: new Set(prev[dim] as Set<any>) }
            const set = next[dim] as Set<any>
            if (set.has(value)) set.delete(value)
            else set.add(value)
            return next
        })
    }, [])

    const clearOne = useCallback((dim: Dimension, value: string | number) => {
        setFilters((prev) => {
            const next = { ...prev, [dim]: new Set(prev[dim] as Set<any>) }
            ;(next[dim] as Set<any>).delete(value)
            return next
        })
    }, [])

    const clearAll = useCallback(() => setFilters(emptyFilters()), [])

    const anyActive = Object.values(filters).some((s) => s.size > 0)

    return { filters, toggle, clearOne, clearAll, anyActive }
}
