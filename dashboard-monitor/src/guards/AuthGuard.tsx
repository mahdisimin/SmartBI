import { Navigate, Outlet } from 'react-router-dom'
import { useAuthStore } from '@/hooks/useAuthStore'

export const AuthGuard = () => {
    // Currently guarding by userId — will switch to token once JWT is implemented
    const userId = useAuthStore((s) => s.userId)
    return userId !== null ? <Outlet /> : <Navigate to="/login" replace />
}
