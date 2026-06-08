import { apiClient } from '@/api/client'

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
}

export interface User {
    id: number
    userName: string
    phoneNumber: string
    avatar?: string
    webAppList: WebApp[]
}


export interface WebApp {
    webAppName: string
    webAppURL: string
}