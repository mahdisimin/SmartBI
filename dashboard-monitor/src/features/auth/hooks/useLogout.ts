import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQueryClient } from '@tanstack/react-query'
import { authApi } from '@/api/auth.api'
import { SESSION_QUERY_KEY } from './useSession'

export const useLogout = () => {
    const [isLoading, setIsLoading] = useState(false)
    const navigate = useNavigate()
    const queryClient = useQueryClient()

    const logout = async () => {
        setIsLoading(true)
        try {
            await authApi.logout()
        } catch {
            // Best effort — clear local state even if the request failed, so
            // the user is never stuck logged in on this screen
        } finally {
            setIsLoading(false)
            // Leave the guarded route first, so no mounted query refetches
            // after its cache is dropped
            navigate('/login', { replace: true })
            queryClient.removeQueries()
            queryClient.setQueryData(SESSION_QUERY_KEY, null)
        }
    }

    return { logout, isLoading }
}
