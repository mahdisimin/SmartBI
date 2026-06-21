import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { authApi } from '@/api/auth.api'
import { useAuthStore } from '@/hooks/useAuthStore'
import type { User } from '@/types/auth.types'

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
    const { setUserId, setUser, clearAuth } = useAuthStore()

    const login = async (form: LoginForm) => {
        setIsLoading(true)
        setError(null)

        try {
            // Step 1 — Login and get user_id
            const loginResponse = await authApi.login(form)

            if (!loginResponse.user_id) {
                setError('Username or password is incorrect.')
                return
            }

            setUserId(loginResponse.user_id)

            // Step 2 — Fetch full user profile
            const profile = await authApi.getUserProfile(loginResponse.user_id)

            const user: User = {
                id: loginResponse.user_id,
                userName: profile.user_name,
                phoneNumber: profile.user_phone,
                webAppList: (profile.user_link_list ?? [])
                    .filter((item) => item.WebAppURL?.trim())
                    .map((item) => ({
                        webAppName: item.WebAppName,
                        webAppURL: item.WebAppURL,
                    })),
            }

            setUser(user)
            navigate('/', { replace: true })
        } catch {
            clearAuth()
            setError('Username or password is incorrect.')
        } finally {
            setIsLoading(false)
        }
    }

    return { login, isLoading, error }
}