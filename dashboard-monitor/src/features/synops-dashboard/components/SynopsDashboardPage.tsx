import { useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { ArrowRight } from 'lucide-react'
import { Line, Bar } from 'react-chartjs-2'
import {
    Chart as ChartJS,
    CategoryScale,
    LinearScale,
    PointElement,
    LineElement,
    BarElement,
    Tooltip,
    Filler,
} from 'chart.js'
import { useDashboardData } from '../hooks/useDashboardData'
import { useDashboardFilters } from '../hooks/useDashboardFilters'
import { fmt, pct, labelFor } from '../dateLabels'
import type { DashboardData, TrendLevel } from '../types'
import '../synops-dashboard.css'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, BarElement, Tooltip, Filler)

const COLORS = {
    amber: '#F5A623',
    amberDim: 'rgba(245,166,35,.3)',
    teal: '#4EC9B0',
    tealDim: 'rgba(78,201,176,.25)',
    grid: '#1E2A47',
}

const DIM_LABEL: Record<string, string> = {
    days: 'روز',
    modules: 'قابلیت',
    methods: 'عملیات',
    userIds: 'کاربر',
}

export const SynopsDashboardPage = () => {
    const { filters, toggle, clearOne, clearAll, anyActive } = useDashboardFilters()
    const { data, isLoading, isError, error, isFetching } = useDashboardData(filters)
    const [trendLevel, setTrendLevel] = useState<TrendLevel>('day')

    const trendPoints = useMemo(() => {
        if (!data) return []
        if (trendLevel === 'month') return data.monthly_trend
        if (trendLevel === 'week') return data.weekly_trend
        return data.daily_trend
    }, [data, trendLevel])

    if (isLoading) {
        return (
            <div className="synops-dashboard">
                <div className="wrap">
                    <div className="empty">در حال بارگذاری داشبورد...</div>
                </div>
            </div>
        )
    }

    if (isError || !data) {
        return (
            <div className="synops-dashboard">
                <div className="wrap">
                    <div className="empty">
                        خطا در دریافت داده از سرور{error instanceof Error ? `: ${error.message}` : ''}
                    </div>
                </div>
            </div>
        )
    }

    return (
        <div className="synops-dashboard">
            <div className="sweep-track" />
            <div className="wrap">
                <Header isFetching={isFetching} />

                <FilterBar filters={filters} anyActive={anyActive} clearOne={clearOne} clearAll={clearAll} />

                {data.kpis.total_events === 0 ? (
                    <div className="empty">داده‌ای برای نمایش وجود ندارد.</div>
                ) : (
                    <>
                        <KpiGrid data={data} />

                        <div className="sec-label">
                            <span className="code">SIG-01</span>
                            <h2>روند و الگوی استفاده</h2>
                            <span className="sub">حجم فعالیت در طول زمان و پرکاربردترین قابلیت‌ها</span>
                        </div>
                        <div className="grid-2-1">
                            <TrendPanel
                                points={trendPoints}
                                level={trendLevel}
                                setLevel={setTrendLevel}
                                activeDays={filters.days}
                                onPointClick={(key) => toggle('days', key)}
                            />
                            <ModuleRankPanel data={data} activeModules={filters.modules} onToggle={(name) => toggle('modules', name)} />
                        </div>

                        <div className="sec-label">
                            <span className="code">SIG-02</span>
                            <h2>سطح مشارکت کاربران</h2>
                            <span className="sub">چه کسانی، چقدر و با چه نوع عملیاتی از سامانه استفاده می‌کنند</span>
                        </div>
                        <div className="grid-2-1">
                            <UserTablePanel data={data} activeUsers={filters.userIds} onToggle={(id) => toggle('userIds', id)} />
                            <MethodPanel data={data} activeMethods={filters.methods} onToggle={(name) => toggle('methods', name)} />
                        </div>

                        <div className="sec-label">
                            <span className="code">VERDICT</span>
                            <h2>جمع‌بندی: آیا سامانه کاربردی است؟</h2>
                            <span className="sub">{anyActive ? 'بر اساس فیلترهای فعال محاسبه شده' : 'بر اساس کل داده'}</span>
                        </div>
                        <VerdictPanel data={data} />
                    </>
                )}
            </div>
        </div>
    )
}

// ---------------------------------------------------------------------------

const Header = ({ isFetching }: { isFetching: boolean }) => (
    <div className="topbar">
        <div className="brand">
            <h1>اتاق پایش</h1>
            <span className="tag">SYNOPS · ACTIVITY MONITOR</span>
        </div>
        <div className="status-cluster">
            <div className="status-item">
                <span className="pulse-dot" />
                {isFetching ? 'در حال به‌روزرسانی...' : 'زنده'}
            </div>
            <Link to="/" className="back-link">
                <ArrowRight size={14} />
                بازگشت به داشبوردها
            </Link>
        </div>
    </div>
)

const FilterBar = ({
    filters,
    anyActive,
    clearOne,
    clearAll,
}: {
    filters: ReturnType<typeof useDashboardFilters>['filters']
    anyActive: boolean
    clearOne: ReturnType<typeof useDashboardFilters>['clearOne']
    clearAll: () => void
}) => {
    const chips: { dim: 'days' | 'modules' | 'methods' | 'userIds'; value: string | number }[] = []
    filters.days.forEach((v) => chips.push({ dim: 'days', value: v }))
    filters.modules.forEach((v) => chips.push({ dim: 'modules', value: v }))
    filters.methods.forEach((v) => chips.push({ dim: 'methods', value: v }))
    filters.userIds.forEach((v) => chips.push({ dim: 'userIds', value: v }))

    return (
        <div className="filter-bar">
            {!anyActive ? (
                <div className="hint">
                    ↳ برای فیلتر متقابل، روی هر بخش از نمودارها، ردیف‌های جدول یا قابلیت‌ها کلیک کنید
                </div>
            ) : (
                <>
                    {chips.map(({ dim, value }) => (
                        <div className="filter-chip" key={`${dim}-${value}`}>
                            <span className="dim">{DIM_LABEL[dim]}</span>
                            {dim === 'userIds' ? `#${value}` : value}
                            <button onClick={() => clearOne(dim, value)}>✕</button>
                        </div>
                    ))}
                    <button className="clear-all" onClick={clearAll}>
                        پاک‌کردن همه فیلترها ✕
                    </button>
                </>
            )}
        </div>
    )
}

const KpiGrid = ({ data }: { data: DashboardData }) => {
    const { kpis } = data
    return (
        <div className="kpi-grid">
            <div className="kpi-card">
                <div className="lbl">کل رویدادها</div>
                <div className="val mono">{fmt(kpis.total_events)}</div>
                <div className="delta">{kpis.days_covered} روز پوشش داده</div>
            </div>
            <div className="kpi-card">
                <div className="lbl">کاربران فعال</div>
                <div className="val mono">
                    {fmt(kpis.unique_users)}
                    <span className="unit">کاربر</span>
                </div>
                <div className="delta">{fmt(kpis.unique_orgs)} سازمان</div>
            </div>
            <div className="kpi-card">
                <div className="lbl">نرخ موفقیت درخواست‌ها</div>
                <div className="val mono">
                    {fmt(kpis.success_rate, 1)}
                    <span className="unit">%</span>
                </div>
                <div className="delta">
                    {kpis.error_count} خطا از {kpis.total_events} رویداد
                </div>
            </div>
            <div className="kpi-card">
                <div className="lbl">میانگین زمان پاسخ</div>
                <div className="val mono">
                    {fmt(kpis.avg_duration_seconds, 2)}
                    <span className="unit">ثانیه</span>
                </div>
                <div className="delta">شاخص کارایی سیستم</div>
            </div>
            <div className="kpi-card">
                <div className="lbl">تنوع قابلیت‌های استفاده‌شده</div>
                <div className="val mono">
                    {kpis.module_count}
                    <span className="unit">ماژول</span>
                </div>
                <div className="delta">عمق استفاده از سامانه</div>
            </div>
        </div>
    )
}

const TrendPanel = ({
    points,
    level,
    setLevel,
    activeDays,
    onPointClick,
}: {
    points: DashboardData['daily_trend']
    level: TrendLevel
    setLevel: (l: TrendLevel) => void
    activeDays: Set<string>
    onPointClick: (key: string) => void
}) => {
    const isDayLevel = level === 'day'
    const labels = points.map((p) => labelFor(level, p.key))
    const counts = points.map((p) => p.count)
    const pointColors = points.map((p) => (activeDays.has(p.key) ? COLORS.amber : activeDays.size ? COLORS.amberDim : COLORS.amber))

    return (
        <div className="panel">
            <h3>
                روند فعالیت <span className="cnt">({points.length} {level === 'month' ? 'ماه' : level === 'week' ? 'هفته' : 'روز'})</span>
            </h3>
            <div className="trend-controls">
                <div />
                <div className="gran-toggle">
                    <button className={level === 'month' ? 'active' : ''} onClick={() => setLevel('month')}>ماه</button>
                    <button className={level === 'week' ? 'active' : ''} onClick={() => setLevel('week')}>هفته</button>
                    <button className={level === 'day' ? 'active' : ''} onClick={() => setLevel('day')}>روز</button>
                </div>
            </div>
            <div className="hint-sm">{isDayLevel ? 'روی یک نقطه کلیک کنید تا فقط همان روز فیلتر شود' : 'نمای کلی — برای فیلتر به نمای روزانه بروید'}</div>
            {isDayLevel ? (
                <Line
                    data={{
                        labels,
                        datasets: [
                            {
                                label: 'رویدادها',
                                data: counts,
                                borderColor: COLORS.amber,
                                backgroundColor: 'rgba(245,166,35,.12)',
                                fill: true,
                                tension: 0.35,
                                pointRadius: points.map((p) => (activeDays.has(p.key) ? 6 : 3)),
                                pointBackgroundColor: pointColors,
                                pointHoverRadius: 8,
                            },
                        ],
                    }}
                    options={{
                        responsive: true,
                        animation: { duration: 300 },
                        plugins: { legend: { display: false } },
                        interaction: { mode: 'index', intersect: false },
                        scales: {
                            x: { grid: { color: COLORS.grid } },
                            y: { grid: { color: COLORS.grid }, beginAtZero: true, ticks: { precision: 0 } },
                        },
                        onClick: (_evt, elements) => {
                            if (!elements.length) return
                            onPointClick(points[elements[0].index].key)
                        },
                    }}
                    height={230}
                />
            ) : (
                <Bar
                    data={{ labels, datasets: [{ label: 'رویدادها', data: counts, backgroundColor: COLORS.amber, borderRadius: 5, maxBarThickness: 46 }] }}
                    options={{
                        responsive: true,
                        animation: { duration: 300 },
                        plugins: { legend: { display: false } },
                        scales: {
                            x: { grid: { color: COLORS.grid } },
                            y: { grid: { color: COLORS.grid }, beginAtZero: true, ticks: { precision: 0 } },
                        },
                    }}
                    height={230}
                />
            )}
        </div>
    )
}

const ModuleRankPanel = ({
    data,
    activeModules,
    onToggle,
}: {
    data: DashboardData
    activeModules: Set<string>
    onToggle: (name: string) => void
}) => {
    const max = Math.max(1, ...data.top_modules.map((m) => m.count))
    return (
        <div className="panel">
            <h3>پرکاربردترین قابلیت‌ها</h3>
            <div className="hint-sm">کلیک روی هر ردیف = فیلتر متقابل</div>
            <div className="rank-list">
                {data.top_modules.length === 0 ? (
                    <div className="empty">داده‌ای برای این فیلتر یافت نشد</div>
                ) : (
                    data.top_modules.map((m) => {
                        const isActive = activeModules.has(m.name)
                        const isDimmed = activeModules.size > 0 && !isActive
                        return (
                            <div
                                key={m.name}
                                className={`rank-row ${isActive ? 'active' : ''} ${isDimmed ? 'dimmed' : ''}`}
                                onClick={() => onToggle(m.name)}
                            >
                                <div className="name" title={m.name}>{m.name}</div>
                                <div className="bar-bg"><div className="bar-fill" style={{ width: `${pct(m.count, max)}%` }} /></div>
                                <div className="num">{m.count}</div>
                            </div>
                        )
                    })
                )}
            </div>
        </div>
    )
}

const MethodPanel = ({
    data,
    activeMethods,
    onToggle,
}: {
    data: DashboardData
    activeMethods: Set<string>
    onToggle: (name: string) => void
}) => {
    const labels = data.method_breakdown.map((m) => m.name)
    const counts = data.method_breakdown.map((m) => m.count)
    const colors = data.method_breakdown.map((m) => (activeMethods.has(m.name) ? COLORS.teal : activeMethods.size ? COLORS.tealDim : COLORS.teal))

    return (
        <div className="panel">
            <h3>نوع عملیات (Method)</h3>
            <div className="hint-sm">GET = دریافت اطلاعات · POST = ایجاد/اجرای عملیات</div>
            <Bar
                data={{ labels, datasets: [{ data: counts, backgroundColor: colors, borderRadius: 4, barThickness: 26 }] }}
                options={{
                    indexAxis: 'y',
                    responsive: true,
                    animation: { duration: 300 },
                    plugins: { legend: { display: false } },
                    scales: {
                        x: { grid: { color: COLORS.grid }, beginAtZero: true, ticks: { precision: 0 } },
                        y: { grid: { display: false } },
                    },
                    onClick: (_evt, elements) => {
                        if (!elements.length) return
                        onToggle(labels[elements[0].index])
                    },
                }}
                height={200}
            />
        </div>
    )
}

const UserTablePanel = ({
    data,
    activeUsers,
    onToggle,
}: {
    data: DashboardData
    activeUsers: Set<number>
    onToggle: (id: number) => void
}) => (
    <div className="panel">
        <h3>
            جدول فعالیت کاربران <span className="cnt">({data.users.length} کاربر شناسایی‌شده)</span>
        </h3>
        <div className="hint-sm">کلیک روی هر ردیف برای فیلتر بر اساس آن کاربر</div>
        <div style={{ overflowX: 'auto' }}>
            <table>
                <thead>
                    <tr>
                        <th>کاربر</th>
                        <th>سازمان</th>
                        <th>نقش</th>
                        <th className="mono">اقدامات</th>
                        <th>تنوع قابلیت</th>
                        <th className="mono">نرخ موفقیت</th>
                        <th className="mono">میانگین پاسخ</th>
                    </tr>
                </thead>
                <tbody>
                    {data.users.length === 0 ? (
                        <tr><td colSpan={7} className="empty">داده‌ای یافت نشد</td></tr>
                    ) : (
                        data.users.map((u) => {
                            const isActive = activeUsers.has(u.user_id)
                            const isDimmed = activeUsers.size > 0 && !isActive
                            return (
                                <tr
                                    key={u.user_id}
                                    className={`${isActive ? 'active-row' : ''} ${isDimmed ? 'dimmed-row' : ''}`}
                                    onClick={() => onToggle(u.user_id)}
                                >
                                    <td className="mono">#{u.user_id}</td>
                                    <td className="mono">{u.org_id != null ? `org-${u.org_id}` : '—'}</td>
                                    <td>{u.org_role ? <span className="pill owner">{u.org_role}</span> : '—'}</td>
                                    <td className="mono">{u.actions}</td>
                                    <td>{u.module_breadth} <span className="bar-mini"><i style={{ width: `${pct(u.module_breadth, data.kpis.module_count)}%` }} /></span></td>
                                    <td className="mono">{fmt(u.success_rate, 0)}%</td>
                                    <td className="mono">{fmt(u.avg_duration_seconds, 2)}s</td>
                                </tr>
                            )
                        })
                    )}
                </tbody>
            </table>
        </div>
    </div>
)

const VerdictPanel = ({ data }: { data: DashboardData }) => {
    const { verdict, kpis } = data
    return (
        <div className="verdict">
            <div className="verdict-head">
                <div className="score-badge mono">{fmt(kpis.success_rate, 0)}%</div>
                <div>
                    <h3>در حال آماده‌سازی...</h3>
                    <p>
                        بر اساس {fmt(kpis.total_events)} رویداد ثبت‌شده در {kpis.days_covered} روز از {fmt(kpis.unique_users)} کاربر شناسایی‌شده
                    </p>
                </div>
            </div>
            <div className="verdict-grid">
                <div className="vcard">
                    <div className="t"><span className={`dot ${verdict.reliability_signal}`} />پایداری فنی</div>
                    <p>در حال آماده‌سازی...</p>
                </div>
                <div className="vcard">
                    <div className="t"><span className={`dot ${verdict.breadth_signal}`} />عمق استفاده</div>
                    <p>در حال آماده‌سازی...</p>
                </div>
                <div className="vcard">
                    <div className="t"><span className={`dot ${verdict.sample_signal}`} />کفایت داده برای تصمیم‌گیری</div>
                    <p>در حال آماده‌سازی...</p>
                </div>
            </div>
        </div>
    )
}
