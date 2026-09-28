import { apiClient } from './client'

export interface LoginRequest {
    phone_number: string
    password: string
}

export interface LoginResponse {
    user_id: number
}

export interface UserProfileResponse {
    user_name: string
    user_phone: string
    user_link_list: {
        WebAppName: string
        WebAppURL: string
    }[]
}

export interface RegisterRequest {
    user_name: string
    phone_number: string
    password: string
}

export interface RegisterResponse {
    user_id: number
}

// The session lives in an HttpOnly cookie set by /user/login — the browser
// sends it automatically (withCredentials), so none of these take a token.
export const authApi = {
    login: async (data: LoginRequest): Promise<LoginResponse> => {
        const response = await apiClient.post<LoginResponse>('/user/login', data)
        return response.data
    },

    /** The logged-in user's profile. 401 = not logged in. */
    getMe: async (): Promise<UserProfileResponse> => {
        const response = await apiClient.get<UserProfileResponse>('/user/user_profile/me')
        return response.data
    },

    logout: async (): Promise<void> => {
        await apiClient.post('/user/logout')
    },

    register: async (data: RegisterRequest): Promise<RegisterResponse> => {
        const response = await apiClient.post<RegisterResponse>('/user/register', data)
        return response.data
    },
}
