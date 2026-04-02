import { createBrowserRouter } from 'react-router-dom'
import { MainLayout } from '@/layouts'
import {
    Dashboard,
    PatientList,
    PatientDetail,
    Monitoring,
    Schedule,
    Statistics,
    Settings,
    RoleSelect,
    WardOverview,
    MasterData,
    Inventory,
    DialysisProcessing,
    DeviceManagement,
    TreatmentConfig,
    DictConfig,
} from '@/pages'
import Login from '@/pages/Login'
import AuthGuard from '@/components/AuthGuard'
import LoginGuard from '@/components/LoginGuard'

const router = createBrowserRouter([
    {
        path: '/login',
        element: <Login />,
    },
    {
        path: '/role-select',
        element: (
            <LoginGuard>
                <RoleSelect />
            </LoginGuard>
        ),
    },
    {
        path: '/',
        element: (
            <AuthGuard>
                <MainLayout />
            </AuthGuard>
        ),
        children: [
            {
                index: true,
                element: <Dashboard />,
            },
            {
                path: 'dashboard',
                element: <Dashboard />,
            },
            {
                path: 'patients',
                element: <PatientList />,
            },
            {
                path: 'patients/:id',
                element: <PatientDetail />,
            },
            {
                path: 'monitoring',
                element: <Monitoring />,
            },
            {
                path: 'schedule',
                element: <Schedule />,
            },
            {
                path: 'statistics',
                element: <Statistics />,
            },
            {
                path: 'settings',
                element: <Settings />,
            },
            {
                path: 'ward-overview',
                element: <WardOverview />,
            },
            // 以下为待开发页面，使用占位组件
            {
                path: 'dialysis-processing',
                element: <DialysisProcessing />,
            },
            {
                path: 'inventory',
                element: <Inventory/>,
            },
            {
                path: 'device-binding',
                element: <DeviceManagement />,
            },
            {
                path: 'master-data',
                element: <MasterData />,
            },
            {
                path: 'treatment-config',
                element: <TreatmentConfig />,
            },
            {
                path: 'dict-config',
                element: <DictConfig />,
            },
        ],
    },
])

export default router
