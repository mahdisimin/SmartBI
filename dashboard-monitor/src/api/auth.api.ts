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

// UserId uses capital casing — the Go struct has no json tag, so it serializes as-is
export interface RegisterResponse {
    UserId: number
}

export const authApi = {
    login: async (data: LoginRequest): Promise<LoginResponse> => {
        const response = await apiClient.post<LoginResponse>('/user/login', data)
        return response.data
    },

    getUserProfile: async (userId: number): Promise<UserProfileResponse> => {
        const response = await apiClient.get<UserProfileResponse>(
            `/user/user_profile/${userId}`
        )
        return response.data
    },

    register: async (data: RegisterRequest): Promise<RegisterResponse> => {
        const response = await apiClient.post<RegisterResponse>('/user/register', data)
        return response.data
    },
}
