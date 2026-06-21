import { ArrowUpRight } from 'lucide-react'
import type { WebApp } from '@/types/auth.types'

interface DashboardCardProps {
    app: WebApp
}

const palette = [
    { bg: 'bg-indigo-500/12 border-indigo-500/20', text: 'text-indigo-400' },
    { bg: 'bg-violet-500/12 border-violet-500/20', text: 'text-violet-400' },
    { bg: 'bg-sky-500/12 border-sky-500/20', text: 'text-sky-400' },
    { bg: 'bg-emerald-500/12 border-emerald-500/20', text: 'text-emerald-400' },
    { bg: 'bg-amber-500/12 border-amber-500/20', text: 'text-amber-400' },
    { bg: 'bg-rose-500/12 border-rose-500/20', text: 'text-rose-400' },
    { bg: 'bg-teal-500/12 border-teal-500/20', text: 'text-teal-400' },
    { bg: 'bg-cyan-500/12 border-cyan-500/20', text: 'text-cyan-400' },
]

function getColor(name: string) {
    return palette[(name.charCodeAt(0) || 0) % palette.length]
}

export const DashboardCard = ({ app }: DashboardCardProps) => {
    const color = getColor(app.webAppName)
    const initial = app.webAppName.charAt(0).toUpperCase()

    return (
        <a
            href={app.webAppURL}
            target="_blank"
            rel="noopener noreferrer"
            className="group flex flex-col p-5 rounded-xl border border-border bg-card hover:border-primary/35 hover:bg-accent/40 transition-all duration-200"
        >
            {/* Icon row */}
            <div className="flex items-start justify-between mb-4">
                <div className={`w-10 h-10 rounded-xl border flex items-center justify-center flex-shrink-0 ${color.bg}`}>
                    <span className={`text-sm font-bold ${color.text}`}>{initial}</span>
                </div>
                <ArrowUpRight
                    size={15}
                    className="text-muted-foreground/30 group-hover:text-primary group-hover:translate-x-0.5 group-hover:-translate-y-0.5 transition-all duration-200"
                />
            </div>

            {/* Name & URL */}
            <p className="text-sm font-semibold text-foreground group-hover:text-primary transition-colors mb-1 truncate">
                {app.webAppName}
            </p>
            <p className="text-xs text-muted-foreground truncate">
                {app.webAppURL}
            </p>
        </a>
    )
}
