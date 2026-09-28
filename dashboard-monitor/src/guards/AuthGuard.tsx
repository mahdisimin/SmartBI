import { Navigate, Outlet, useLocation } from 'react-router-dom'
import { Loader2 } from 'lucide-react'
import { useSession } from '@/features/auth/hooks/useSession'

export const AuthGuard = () => {
    // The server is the source of truth: GET /user/user_profile/me → 200/401
    const { data: user, isPending, isError, refetch } = useSession()
    const location = useLocation()

    if (isPending) {
        return (
            <div className="min-h-screen flex items-center justify-center bg-background">
                <Loader2 size={20} className="animate-spin text-muted-foreground" aria-label="Checking session" />
            </div>
        )
    }

    if (isError) {
        return (
            <div className="min-h-screen flex flex-col items-center justify-center gap-3 bg-background text-center px-4">
                <p className="text-sm text-muted-foreground">Could not reach the server.</p>
                <button
                    onClick={() => refetch()}
                    className="text-sm font-medium text-primary hover:underline"
                >
                    Try again
                </button>
            </div>
        )
    }

    if (!user) {
        return <Navigate to="/login" replace state={{ from: location.pathname + location.search }} />
    }

    return <Outlet />
}
