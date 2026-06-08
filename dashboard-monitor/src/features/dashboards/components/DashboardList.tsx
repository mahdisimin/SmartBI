import { LayoutDashboard } from 'lucide-react'
import { DashboardCard } from './DashboardCard'
import { useAuthStore } from '@/hooks/useAuthStore'

export const DashboardList = () => {
    const user = useAuthStore((s) => s.user)

    if (!user) return null

    const apps = user.webAppList

    return (
        <div className="max-w-2xl">
            {/* Page header */}
            <div className="mb-8">
                <div className="flex items-center gap-2 mb-1">
                    <LayoutDashboard size={20} className="text-primary" />
                    <h2 className="text-xl font-semibold text-foreground">Dashboards</h2>
                </div>
                <p className="text-sm text-muted-foreground">
                    {apps.length > 0
                        ? `You have access to ${apps.length} dashboard${apps.length > 1 ? 's' : ''}.`
                        : 'No dashboards are assigned to your account yet.'}
                </p>
            </div>

            {/* Cards */}
            {apps.length > 0 ? (
                <div className="flex flex-col gap-3">
                    {apps.map((app) => (
                        <DashboardCard key={app.webAppURL} app={app} />
                    ))}
                </div>
            ) : (
                <div className="flex flex-col items-center justify-center py-16 text-center">
                    <div className="w-12 h-12 rounded-2xl bg-muted flex items-center justify-center mb-4">
                        <LayoutDashboard size={20} className="text-muted-foreground" />
                    </div>
                    <p className="text-sm font-medium text-foreground">No dashboards yet</p>
                    <p className="text-xs text-muted-foreground mt-1">
                        Contact your administrator to get access.
                    </p>
                </div>
            )}
        </div>
    )
}
