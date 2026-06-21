import { Activity } from 'lucide-react'
import { DashboardCard } from './DashboardCard'
import { useAuthStore } from '@/hooks/useAuthStore'

export const DashboardList = () => {
    const user = useAuthStore((s) => s.user)

    if (!user) return null

    const apps = user.webAppList

    return (
        <div className="p-8">
            {/* Page header */}
            <div className="mb-8">
                <p className="text-[11px] font-semibold text-primary uppercase tracking-widest mb-2">
                    Platform Access
                </p>
                <h2 className="text-2xl font-bold text-foreground tracking-tight mb-1.5">
                    Analytics Hub
                </h2>
                <p className="text-sm text-muted-foreground">
                    {apps.length > 0
                        ? `${apps.length} connected platform${apps.length > 1 ? 's' : ''} — select a dashboard to explore`
                        : 'No platforms have been assigned to your account yet.'}
                </p>
            </div>

            {/* Grid */}
            {apps.length > 0 ? (
                <div className="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-3 gap-4">
                    {apps.map((app) => (
                        <DashboardCard key={app.webAppURL} app={app} />
                    ))}
                </div>
            ) : (
                <EmptyState />
            )}
        </div>
    )
}

const EmptyState = () => (
    <div className="flex flex-col items-center justify-center py-20 text-center border border-dashed border-border rounded-2xl max-w-lg">
        <div className="w-14 h-14 rounded-2xl bg-muted border border-border flex items-center justify-center mb-5">
            <Activity size={22} className="text-muted-foreground" />
        </div>
        <p className="text-sm font-semibold text-foreground mb-1.5">No platforms connected</p>
        <p className="text-xs text-muted-foreground max-w-[240px] leading-relaxed">
            Contact your administrator to configure platform access for your account.
        </p>
    </div>
)
