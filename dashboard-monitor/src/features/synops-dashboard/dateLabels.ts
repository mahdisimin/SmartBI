import type { TrendLevel } from './types'

const FA_CAL = 'fa-IR-u-ca-persian-nu-latn'

export function monthLabel(key: string): string {
    return new Date(key + '-01T00:00:00Z').toLocaleDateString(FA_CAL, {
        year: 'numeric',
        month: 'long',
        timeZone: 'UTC',
    })
}

export function weekLabel(key: string): string {
    const start = new Date(key + 'T00:00:00Z')
    const end = new Date(start)
    end.setUTCDate(end.getUTCDate() + 6)
    const s = start.toLocaleDateString(FA_CAL, { day: 'numeric', month: 'short', timeZone: 'UTC' })
    const e = end.toLocaleDateString(FA_CAL, { day: 'numeric', month: 'short', timeZone: 'UTC' })
    return `${s} تا ${e}`
}

export function dayLabel(key: string): string {
    return new Date(key + 'T00:00:00Z').toLocaleDateString(FA_CAL, {
        day: 'numeric',
        month: 'long',
        timeZone: 'UTC',
    })
}

export function labelFor(level: TrendLevel, key: string): string {
    if (level === 'month') return monthLabel(key)
    if (level === 'week') return weekLabel(key)
    return dayLabel(key)
}

export function fmt(n: number, decimals = 0): string {
    return Number(n).toLocaleString('en-US', { maximumFractionDigits: decimals, minimumFractionDigits: decimals })
}

export function pct(a: number, b: number): number {
    return b > 0 ? (a / b) * 100 : 0
}
