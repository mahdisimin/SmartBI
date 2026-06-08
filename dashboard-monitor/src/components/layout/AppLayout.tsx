import { Outlet, NavLink, useNavigate } from 'react-router-dom'
import { LayoutDashboard, MessageSquare, LogOut } from 'lucide-react'
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

    return (
        <div className="min-h-screen flex bg-background">
            {/* Sidebar */}
            <aside className="w-60 border-r border-border flex flex-col py-6 px-3 gap-1">
                {/* Brand */}
                <div className="px-3 mb-6">
                    <div className="flex items-center gap-2">
                        <div className="w-7 h-7 rounded-lg bg-primary flex items-center justify-center">
                            <span className="text-xs text-primary-foreground font-bold">A</span>
                        </div>
                        <h1 className="text-sm font-semibold text-foreground">
                            ActivityMonitor
                        </h1>
                    </div>
                </div>

                {/* Nav */}
                <nav className="flex-1 flex flex-col gap-0.5">
                    {navItems.map(({ to, label, icon: Icon }) => (
                        <NavLink
                            key={to}
                            to={to}
                            end
                            className={({ isActive }) =>
                                cn(
                                    'flex items-center gap-2.5 px-3 py-2 rounded-md text-sm transition-colors',
                                    isActive
                                        ? 'bg-accent text-accent-foreground font-medium'
                                        : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'
                                )
                            }
                        >
                            <Icon size={15} />
                            {label}
                        </NavLink>
                    ))}
                </nav>

                {/* User + logout */}
                <div className="border-t border-border pt-3 mt-2">
                    {user && (
                        <div className="px-3 mb-2">
                            <p className="text-xs font-medium text-foreground truncate">
                                {user.userName}
                            </p>
                            <p className="text-xs text-muted-foreground truncate">
                                {user.phoneNumber}
                            </p>
                        </div>
                    )}
                    <button
                        onClick={handleLogout}
                        className="w-full flex items-center gap-2.5 px-3 py-2 rounded-md text-sm text-muted-foreground hover:bg-accent hover:text-accent-foreground transition-colors"
                    >
                        <LogOut size={15} />
                        Logout
                    </button>
                </div>
            </aside>

            {/* Main */}
            <main className="flex-1 p-8 overflow-auto">
                <Outlet />
            </main>
        </div>
    )
}
