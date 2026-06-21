import { Outlet } from 'react-router-dom'
import { Activity, BarChart2, Users, TrendingUp } from 'lucide-react'

const features = [
    { icon: BarChart2, text: 'Centralized activity analytics across all platforms' },
    { icon: Users, text: 'Unified user behavior intelligence' },
    { icon: TrendingUp, text: 'Data-driven insights for better decisions' },
]

export const AuthLayout = () => {
    return (
        <div className="min-h-screen bg-background flex">
            {/* Left panel — brand & product story */}
            <div className="hidden lg:flex lg:flex-col lg:w-[56%] relative overflow-hidden bg-card border-r border-border p-12">
                {/* Subtle grid background */}
                <div
                    className="absolute inset-0 opacity-[0.04]"
                    style={{
                        backgroundImage: `linear-gradient(hsl(var(--border)) 1px, transparent 1px), linear-gradient(90deg, hsl(var(--border)) 1px, transparent 1px)`,
                        backgroundSize: '40px 40px',
                    }}
                />
                {/* Ambient glow */}
                <div className="absolute -top-32 -left-32 w-[480px] h-[480px] bg-primary/6 rounded-full blur-3xl" />
                <div className="absolute bottom-0 right-0 w-64 h-64 bg-primary/4 rounded-full blur-3xl" />

                <div className="relative z-10 flex flex-col h-full">
                    {/* Brand */}
                    <div className="flex items-center gap-3">
                        <div className="w-9 h-9 rounded-xl bg-primary/15 border border-primary/25 flex items-center justify-center">
                            <Activity size={18} className="text-primary" />
                        </div>
                        <span className="text-[15px] font-semibold text-foreground tracking-tight">
                            ActivityMonitor
                        </span>
                    </div>

                    {/* Headline */}
                    <div className="flex-1 flex flex-col justify-center max-w-sm">
                        <p className="text-[11px] font-semibold text-primary uppercase tracking-widest mb-4">
                            Enterprise Analytics Platform
                        </p>
                        <h2 className="text-[28px] font-bold text-foreground leading-snug tracking-tight mb-4">
                            Unified activity intelligence across your platforms
                        </h2>
                        <p className="text-sm text-muted-foreground leading-relaxed mb-10">
                            Consolidate user activity from all your enterprise systems into one source of truth. Understand engagement, measure adoption, and make confident decisions.
                        </p>

                        <div className="flex flex-col gap-4">
                            {features.map(({ icon: Icon, text }) => (
                                <div key={text} className="flex items-center gap-3">
                                    <div className="w-7 h-7 rounded-lg bg-primary/10 border border-primary/15 flex items-center justify-center flex-shrink-0">
                                        <Icon size={13} className="text-primary" />
                                    </div>
                                    <span className="text-sm text-muted-foreground">{text}</span>
                                </div>
                            ))}
                        </div>
                    </div>

                    <p className="text-xs text-muted-foreground/40">
                        © {new Date().getFullYear()} ActivityMonitor
                    </p>
                </div>
            </div>

            {/* Right panel — auth form */}
            <div className="flex-1 flex flex-col items-center justify-center px-8 py-12">
                {/* Mobile brand mark */}
                <div className="lg:hidden flex items-center gap-2.5 mb-10">
                    <div className="w-8 h-8 rounded-lg bg-primary/15 border border-primary/25 flex items-center justify-center">
                        <Activity size={16} className="text-primary" />
                    </div>
                    <span className="text-sm font-semibold text-foreground">ActivityMonitor</span>
                </div>

                <div className="w-full max-w-[340px]">
                    <Outlet />
                </div>
            </div>
        </div>
    )
}
