import { useQuery } from '@tanstack/react-query'
import { isAxiosError } from 'axios'
import { authApi, type UserProfileResponse } from '@/api/auth.api'
import type { User } from '@/types/auth.types'

export const SESSION_QUERY_KEY = ['session', 'me'] as const

export function toUser(profile: UserProfileResponse): User {
    return {
        userName: profile.user_name,
        phoneNumber: profile.user_phone,
        webAppList: (profile.user_link_list ?? [])
            .filter((item) => item.WebAppURL?.trim())
            .map((item) => ({
                webAppName: item.WebAppName,
                webAppURL: item.WebAppURL,
            })),
    }
}

/** Fetches the current session: 401 → null (logged out, not an error).
 * Exported so login can prime the cache with the same function. */
export async function fetchSession(): Promise<User | null> {
    try {
        return toUser(await authApi.getMe())
    } catch (err) {
        if (isAxiosError(err) && err.response?.status === 401) return null
        throw err
    }
}

/** The session is the source of truth for "logged in" — the cookie is HttpOnly,
 * so the only way to know is to ask the server. `data` is the user, or null
 * when logged out; `isError` means the check itself failed (e.g. network). */
export function useSession() {
    return useQuery({
        queryKey: SESSION_QUERY_KEY,
        queryFn: fetchSession,
        // Checked once per app load; login/logout update the cache directly,
        // and a mid-session 401 is handled by the axios interceptor.
        staleTime: Infinity,
        retry: false,
    })
}
