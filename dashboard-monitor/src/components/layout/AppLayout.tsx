import { Outlet, NavLink, useNavigate } from 'react-router-dom'
import { Activity, LayoutDashboard, MessageSquare, LogOut } from 'lucide-react'
import { useAuthStore } from '@/hooks/useAuthStore'
import { cn } from '@/lib/utils'

const navItems = [
    { to: '/', label: 'Dashboards', icon: LayoutDashboard },
    { to: '/chat', label: 'AI Chat', icon: MessageSquare },
]

export const AppLayout = () => {
    const clearAuth = useAuthStore((s) => s.clearAuth)
    const user = useAuthStore((s) => s.user)
    const navigate = useNavigate()

    const handleLogout = () => {
        clearAuth()
        navigate('/login', { replace: true })
    }

    const initials = user?.userName
        ? user.userName.split(' ').map((n) => n[0]).join('').toUpperCase().slice(0, 2)
        : '?'

    return (
        <div className="min-h-screen flex bg-background">
            {/* Sidebar */}
            <aside className="w-60 border-r border-border flex flex-col bg-card flex-shrink-0">
                {/* Brand */}
                <div className="px-4 py-5 border-b border-border">
                    <div className="flex items-center gap-2.5">
                        <div className="w-8 h-8 rounded-lg bg-primary/15 border border-primary/25 flex items-center justify-center flex-shrink-0">
                            <Activity size={16} className="text-primary" />
                        </div>
                        <div className="min-w-0">
                            <p className="text-[13px] font-semibold text-foreground truncate tracking-tight">
                                ActivityMonitor
                            </p>
                            <p className="text-[10px] text-muted-foreground/70 tracking-wide">
                                Analytics Platform
                            </p>
                        </div>
                    </div>
                </div>

                {/* Nav */}
                <nav className="flex-1 px-3 py-4">
                    <p className="text-[10px] uppercase tracking-widest font-semibold text-muted-foreground/40 px-3 mb-2">
                        Menu
                    </p>
                    <div className="flex flex-col gap-0.5">
                        {navItems.map(({ to, label, icon: Icon }) => (
                            <NavLink key={to} to={to} end>
                                {({ isActive }) => (
                                    <div
                                        className={cn(
                                            'relative flex items-center gap-2.5 px-3 py-2 rounded-lg text-sm transition-all cursor-pointer select-none',
                                            isActive
                                                ? 'bg-primary/10 text-primary font-medium'
                                                : 'text-muted-foreground hover:bg-accent hover:text-foreground'
                                        )}
                                    >
                                        {isActive && (
                                            <span className="absolute left-0 top-1/2 -translate-y-1/2 w-[3px] h-[18px] bg-primary rounded-r-full" />
                                        )}
                                        <Icon size={15} />
                                        {label}
                                    </div>
                                )}
                            </NavLink>
                        ))}
                    </div>
                </nav>

                {/* User & logout */}
                <div className="border-t border-border p-3">
                    {user && (
                        <div className="flex items-center gap-2.5 px-2 py-2 mb-1 rounded-lg">
                            <div className="w-7 h-7 rounded-full bg-primary/15 border border-primary/20 flex items-center justify-center flex-shrink-0">
                                <span className="text-[10px] font-bold text-primary">{initials}</span>
                            </div>
                            <div className="min-w-0">
                                <p className="text-xs font-medium text-foreground truncate">{user.userName}</p>
                                <p className="text-[10px] text-muted-foreground truncate">{user.phoneNumber}</p>
                            </div>
                        </div>
                    )}
                    <button
                        onClick={handleLogout}
                        className="w-full flex items-center gap-2.5 px-3 py-2 rounded-lg text-sm text-muted-foreground hover:bg-accent hover:text-foreground transition-colors"
                    >
                        <LogOut size={14} />
                        Sign out
                    </button>
                </div>
            </aside>

            {/* Main content */}
            <main className="flex-1 overflow-auto">
                <Outlet />
            </main>
        </div>
    )
}
