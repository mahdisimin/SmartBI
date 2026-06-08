import { ExternalLink } from 'lucide-react'
import type { WebApp } from '@/types/auth.types'

interface DashboardCardProps {
    app: WebApp
}

export const DashboardCard = ({ app }: DashboardCardProps) => {
    return (
        <a
            href={app.webAppURL}
            target="_blank"
            rel="noopener noreferrer"
            className="group flex items-center justify-between p-4 rounded-xl border border-border bg-card hover:border-primary/40 hover:shadow-sm transition-all duration-200"
        >
            <div className="flex items-center gap-3">
                {/* App icon — derived from first letter */}
                <div className="w-9 h-9 rounded-lg bg-primary/10 flex items-center justify-center flex-shrink-0">
                    <span className="text-sm font-semibold text-primary">
                        {app.webAppName.charAt(0).toUpperCase()}
                    </span>
                </div>
                <div>
                    <p className="text-sm font-medium text-foreground group-hover:text-primary transition-colors">
                        {app.webAppName}
                    </p>
                    <p className="text-xs text-muted-foreground truncate max-w-[200px]">
                        {app.webAppURL}
                    </p>
                </div>
            </div>

            <ExternalLink
                size={14}
                className="text-muted-foreground group-hover:text-primary transition-colors flex-shrink-0"
            />
        </a>
    )
}
