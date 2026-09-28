import axios from 'axios'

export const apiClient = axios.create({
    baseURL: import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8091',
    headers: { 'Content-Type': 'application/json' },
    // Send and accept the HttpOnly session cookie on cross-origin API calls
    withCredentials: true,
})

// A 401 from these means "bad credentials" or "not logged in yet" — the
// caller handles it, so it must not trigger the logout redirect.
const NO_REDIRECT_ENDPOINTS = ['/user/login', '/user/register', '/user/user_profile/me']
const AUTH_PAGES = ['/login', '/register']

// Global 401 handler — the session expired or was revoked mid-use, so go back
// to login. A full-page navigation also drops every cached query.
apiClient.interceptors.response.use(
    (response) => response,
    (error) => {
        const isExempt = NO_REDIRECT_ENDPOINTS.includes(error.config?.url ?? '')
        const onAuthPage = AUTH_PAGES.includes(window.location.pathname)
        if (error.response?.status === 401 && !isExempt && !onAuthPage) {
            window.location.replace('/login')
        }
        return Promise.reject(error)
    }
)
