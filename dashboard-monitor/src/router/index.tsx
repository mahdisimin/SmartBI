import { createBrowserRouter } from 'react-router-dom'
import { AuthGuard } from '@/guards/AuthGuard'
import { AppLayout } from '@/components/layout/AppLayout'
import { AuthLayout } from '@/components/layout/AuthLayout'
import LoginPage from '@/pages/LoginPage'
import DashboardsPage from '@/pages/DashboardsPage'
import AiChatPage from '@/pages/AiChatPage'

export const router = createBrowserRouter([
    {
        element: <AuthLayout />,
        children: [
            { path: '/login', element: <LoginPage /> },
        ],
    },
    {
        element: <AuthGuard />,
        children: [
            {
                element: <AppLayout />,
                children: [
                    { path: '/', element: <DashboardsPage /> },
                    { path: '/chat', element: <AiChatPage /> },
                ],
            },
        ],
    },
])