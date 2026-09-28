import { useCallback, useState } from 'react'
import { emptyFilters, type DashboardFilters, type FilterEntry } from '../types'

export interface UseDashboardFiltersReturn {
    filters: DashboardFilters
    toggle: (entry: FilterEntry) => void
    clearOne: (entry: FilterEntry) => void
    clearAll: () => void
    anyActive: boolean
}

type SetOp = <T>(set: Set<T>, value: T) => Set<T>

const toggled: SetOp = (set, value) => {
    const next = new Set(set)
    if (next.has(value)) next.delete(value)
    else next.add(value)
    return next
}

const without: SetOp = (set, value) => {
    const next = new Set(set)
    next.delete(value)
    return next
}

/** Returns a new filter state with `op` applied to the entry's dimension. */
function apply(prev: DashboardFilters, entry: FilterEntry, op: SetOp): DashboardFilters {
    if (entry.dim === 'userIds') return { ...prev, userIds: op(prev.userIds, entry.value) }
    return { ...prev, [entry.dim]: op(prev[entry.dim], entry.value) }
}

/** Mirrors the sample dashboard's toggleFilter/clearOne/clearAll — cross-filter
 * state lives here; every change re-triggers the query in useDashboardData. */
export function useDashboardFilters(): UseDashboardFiltersReturn {
    const [filters, setFilters] = useState<DashboardFilters>(emptyFilters)

    const toggle = useCallback((entry: FilterEntry) => {
        setFilters((prev) => apply(prev, entry, toggled))
    }, [])

    const clearOne = useCallback((entry: FilterEntry) => {
        setFilters((prev) => apply(prev, entry, without))
    }, [])

    const clearAll = useCallback(() => setFilters(emptyFilters()), [])

    const anyActive = Object.values(filters).some((s) => s.size > 0)

    return { filters, toggle, clearOne, clearAll, anyActive }
}
