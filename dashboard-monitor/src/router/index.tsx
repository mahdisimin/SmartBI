import { createBrowserRouter } from 'react-router-dom'
import { AuthGuard } from '@/guards/AuthGuard'
import { AppLayout } from '@/components/layout/AppLayout'
import { AuthLayout } from '@/components/layout/AuthLayout'
import LoginPage from '@/pages/LoginPage'
import RegisterPage from '@/pages/RegisterPage'
import DashboardsPage from '@/pages/DashboardsPage'
import SynopsDashboardPage from '@/pages/SynopsDashboardPage'
import AiChatPage from '@/pages/AiChatPage'

export const router = createBrowserRouter([
    {
        element: <AuthLayout />,
        children: [
            { path: '/login', element: <LoginPage /> },
            { path: '/register', element: <RegisterPage /> },
        ],
    },
    {
        element: <AuthGuard />,
        children: [
            // Full-bleed page with its own visual theme/topbar — deliberately
            // outside AppLayout so it isn't nested inside the sidebar shell.
            { path: '/dashboards/synops', element: <SynopsDashboardPage /> },
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