import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { User } from '@/types/auth.types'

interface AuthState {
    // Current auth state — no token yet, will be added in future JWT phase
    userId: number | null
    user: User | null

    // Actions
    setUserId: (userId: number) => void
    setUser: (user: User) => void
    clearAuth: () => void

    // Future JWT integration point — token is stored but unused for now
    token: string | null
    setToken: (token: string) => void
}

export const useAuthStore = create<AuthState>()(
    persist(
        (set) => ({
            userId: null,
            user: null,
            token: null,

            setUserId: (userId) => set({ userId }),
            setUser: (user) => set({ user }),
            setToken: (token) => set({ token }),
            clearAuth: () => set({ userId: null, user: null, token: null }),
        }),
        { name: 'auth-storage' }
    )
)