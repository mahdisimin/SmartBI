import axios from 'axios'
import { useAuthStore } from '@/hooks/useAuthStore'

export const apiClient = axios.create({
    baseURL: import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080',
    headers: { 'Content-Type': 'application/json' },
})

// JWT injection — token is null for now, will activate automatically
// once setToken() is called after backend implements JWT
apiClient.interceptors.request.use((config) => {
    const token = useAuthStore.getState().token
    if (token) {
        config.headers.Authorization = `Bearer ${token}`
    }
    return config
})

// Global 401 handler — force logout on expired/invalid token
apiClient.interceptors.response.use(
    (response) => response,
    (error) => {
        if (error.response?.status === 401) {
            useAuthStore.getState().clearAuth()
            window.location.replace('/login')
        }
        return Promise.reject(error)
    }
)