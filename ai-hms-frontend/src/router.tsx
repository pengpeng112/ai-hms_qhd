import { createBrowserRouter } from 'react-router-dom'
import { MainLayout } from '@/layouts'
import {
    Dashboard,
    Cockpit,
    PatientList,
    PatientDetail,
    Monitoring,
    SmartSchedulePage,
    StaffSchedulePage,
    QCScoringPage,
    Statistics,
    Settings,
    SyncCenterPage,
    RoleSelect,
    WardOverview,
    MasterData,
    Inventory,
    DialysisProcessing,
    DeviceManagement,
    TreatmentConfig,
    DictConfig,
    WardManagement,
    BedManagement,
    WardBedManagement,
    UserManagement,
    RoleManagement,
    EducationManagement,
    CnrdsReportPage,
    MonitoringThresholds,
} from '@/pages'
import Login from '@/pages/Login'
import AuthGuard from '@/components/AuthGuard'
import LoginGuard from '@/components/LoginGuard'
import PermissionGuard from '@/components/PermissionGuard'

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
                path: 'cockpit',
                element: <Cockpit />,
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
                element: <SmartSchedulePage />,
            },
            {
                path: 'staff-schedule',
                element: <StaffSchedulePage />,
            },
            {
                path: 'statistics',
                element: <Statistics />,
            },
            {
                path: 'qc-scoring',
                element: <QCScoringPage />,
            },
            {
                path: 'settings',
                element: <PermissionGuard><Settings /></PermissionGuard>,
            },
            {
                path: 'sync-center',
                element: <PermissionGuard><SyncCenterPage /></PermissionGuard>,
            },
            {
                path: 'cnrds-report',
                element: <PermissionGuard><CnrdsReportPage /></PermissionGuard>,
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
                element: <PermissionGuard><DictConfig /></PermissionGuard>,
            },
            {
                path: 'ward-bed-management',
                element: <WardBedManagement />,
            },
            {
                path: 'ward-management',
                element: <WardManagement />,
            },
            {
                path: 'bed-management',
                element: <BedManagement />,
            },
            {
                path: 'user-management',
                element: <PermissionGuard><UserManagement /></PermissionGuard>,
            },
            {
                path: 'role-management',
                element: <PermissionGuard><RoleManagement /></PermissionGuard>,
            },
            {
                path: 'education-management',
                element: <EducationManagement />,
            },
            {
                path: 'monitoring-thresholds',
                element: <PermissionGuard><MonitoringThresholds /></PermissionGuard>,
            },
        ],
    },
])

export default router
