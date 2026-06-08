import { Outlet } from 'react-router-dom'

export const AuthLayout = () => {
    return (
    <div className="min-h-screen bg-background flex items-center justify-center">
        <div className="w-full max-w-md px-6">
            <Outlet />
        </div>
    </div>
    )
}