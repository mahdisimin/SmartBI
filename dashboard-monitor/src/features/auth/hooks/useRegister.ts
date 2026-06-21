import { useState } from 'react'
import { authApi } from '@/api/auth.api'
import type { RegisterRequest } from '@/api/auth.api'

interface UseRegisterReturn {
    register: (form: RegisterRequest) => Promise<void>
    isLoading: boolean
    error: string | null
    successUserId: number | null
}

export const useRegister = (): UseRegisterReturn => {
    const [isLoading, setIsLoading] = useState(false)
    const [error, setError] = useState<string | null>(null)
    const [successUserId, setSuccessUserId] = useState<number | null>(null)

    const register = async (form: RegisterRequest) => {
        setIsLoading(true)
        setError(null)

        try {
            const response = await authApi.register(form)
            setSuccessUserId(Number(response.UserId))
        } catch (err: any) {
            // TODO: Map backend error codes to user-friendly messages here when codes become available
            const message =
                err?.response?.data?.message ??
                'Registration failed. Please check your information and try again.'
            setError(message)
        } finally {
            setIsLoading(false)
        }
    }

    return { register, isLoading, error, successUserId }
}
