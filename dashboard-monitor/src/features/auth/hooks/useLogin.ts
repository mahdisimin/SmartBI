import { useState } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { useQueryClient } from '@tanstack/react-query'
import { isAxiosError } from 'axios'
import { authApi } from '@/api/auth.api'
import { SESSION_QUERY_KEY, fetchSession } from './useSession'

interface LoginForm {
    phone_number: string
    password: string
}

interface UseLoginReturn {
    login: (form: LoginForm) => Promise<void>
    isLoading: boolean
    error: string | null
}

export const useLogin = (): UseLoginReturn => {
    const [isLoading, setIsLoading] = useState(false)
    const [error, setError] = useState<string | null>(null)
    const navigate = useNavigate()
    const location = useLocation()
    const queryClient = useQueryClient()

    const login = async (form: LoginForm) => {
        setIsLoading(true)
        setError(null)

        try {
            // Step 1 — Login; the response sets the session cookie
            await authApi.login(form)

            // Step 2 — Load the profile through the session, so the guard's
            // cache is filled and it doesn't show a loader after navigating
            queryClient.removeQueries() // drop any previous user's cached data
            const user = await queryClient.fetchQuery({
                queryKey: SESSION_QUERY_KEY,
                queryFn: fetchSession,
            })
            if (!user) {
                setError('Could not start a session. Please try again.')
                return
            }

            // Return to the page the guard bounced the user from, if any
            const from = (location.state as { from?: string } | null)?.from
            navigate(from ?? '/', { replace: true })
        } catch (err) {
            setError(
                isAxiosError(err) && err.response?.status === 401
                    ? 'Phone number or password is incorrect.'
                    : 'Could not sign in. Please check your connection and try again.'
            )
        } finally {
            setIsLoading(false)
        }
    }

    return { login, isLoading, error }
}
