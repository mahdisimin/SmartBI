export interface User {
    userName: string
    phoneNumber: string
    avatar?: string
    webAppList: WebApp[]
}

export interface WebApp {
    webAppName: string
    webAppURL: string
}
