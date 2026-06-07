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
                <div className="px-3 mb-6">
                    <h1 className="text-lg font-semibold text-foreground">ActivityMonitor</h1>
                    {user && (
                        <p className="text-xs text-muted-foreground mt-0.5 truncate">{user.email}</p>
                    )}
                </div>

                <nav className="flex-1 flex flex-col gap-1">
                    {navItems.map(({ to, label, icon: Icon }) => (
                        <NavLink
                            key={to}
                            to={to}
                            end
                            className={({ isActive }) =>
                                cn(
                                    'flex items-center gap-3 px-3 py-2 rounded-md text-sm transition-colors',
                                    isActive
                                        ? 'bg-accent text-accent-foreground font-medium'
                                        : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'
                                )
                            }
                        >
                            <Icon size={16} />
                            {label}
                        </NavLink>
                    ))}
                </nav>

                <button
                    onClick={handleLogout}
                    className="flex items-center gap-3 px-3 py-2 rounded-md text-sm text-muted-foreground hover:bg-accent hover:text-accent-foreground transition-colors"
                >
                    <LogOut size={16} />
                    Logout
                </button>
            </aside>

            {/* Main content */}
            <main className="flex-1 flex flex-col min-h-screen">
                <div className="flex-1 p-8">
                    <Outlet />
                </div>
            </main>
        </div>
    )
}