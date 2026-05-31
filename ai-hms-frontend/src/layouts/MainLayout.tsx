import { useState, useMemo, useEffect } from 'react'
import { Outlet, useLocation } from 'react-router-dom'
import Sidebar from './Sidebar'
import Header from './Header'
import { UserRole } from '@/types/original'
import { logout } from '@/services/auth'
import { getSelectedRoleUser } from '@/services/role'
import { getRouteMeta } from './routeMeta'

export default function MainLayout() {
    const location = useLocation()
    const [sidebarOpen, setSidebarOpen] = useState(true)

    useEffect(() => {
        const meta = getRouteMeta(location.pathname)
        document.title = `${meta.title} - AI-HMS 智能透析`
    }, [location.pathname])

    const roleUser = useMemo(() => getSelectedRoleUser(), [])

    const user = {
        name: roleUser?.name || '用户',
        role: roleUser?.role || UserRole.DOCTOR_SUPERVISOR,
        avatar: roleUser?.avatar || '',
    }

    const handleLogout = () => {
        logout()
    }

    return (
        <div className="flex h-screen bg-slate-50 overflow-hidden font-sans">
            <Sidebar isOpen={sidebarOpen} />

            <div className="flex-1 flex flex-col min-w-0">
                <Header
                    username={user.name}
                    userRole={user.role}
                    userAvatar={user.avatar}
                    onLogout={handleLogout}
                    sidebarOpen={sidebarOpen}
                    onSidebarToggle={() => setSidebarOpen(!sidebarOpen)}
                />

                <div className="flex-1 flex overflow-hidden">
                    <main className="flex-1 overflow-auto bg-slate-50">
                        <Outlet />
                    </main>
                </div>
            </div>
        </div>
    )
}
