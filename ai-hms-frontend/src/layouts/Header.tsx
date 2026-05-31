import { useState } from 'react'
import { useLocation } from 'react-router-dom'
import { Popover } from 'antd'
import { UserRole } from '@/types/original'
import type { AppRole } from '@/services/role'
import { Menu, Stethoscope, ChevronRight } from 'lucide-react'
import { getRouteMeta } from './routeMeta'
import HeaderUserMenu from './HeaderUserMenu'

interface HeaderProps {
    username?: string
    userRole?: AppRole
    userAvatar?: string
    department?: string
    wardName?: string
    onLogout?: () => void
    sidebarOpen?: boolean
    onSidebarToggle?: () => void
}

export default function Header({
    username = '',
    userRole = UserRole.DOCTOR_SUPERVISOR,
    userAvatar = '',
    department = '肾内透析中心',
    wardName = '第一病区',
    onLogout,
    sidebarOpen = true,
    onSidebarToggle,
}: HeaderProps) {
    const location = useLocation()
    const [avatarFailed, setAvatarFailed] = useState(false)
    const avatarText = (username || userRole || 'U').trim().slice(0, 1).toUpperCase()

    const routeMeta = getRouteMeta(location.pathname)
    const showBreadcrumb = routeMeta.breadcrumb.length > 1

    return (
        <header className="shrink-0 z-10">
            {/* 主行 */}
            <div className="h-14 bg-surface border-b border-gray-200 flex items-center justify-between px-4">
                {/* 左侧：侧边栏切换 + 科室信息 */}
                <div className="flex items-center">
                    <button
                        onClick={onSidebarToggle}
                        className={`p-2.5 rounded-lg border transition-all mr-4 ${
                            sidebarOpen
                                ? 'bg-white border-gray-200 text-gray-600 hover:bg-gray-50'
                                : 'bg-blue-600 border-blue-600 text-white hover:bg-blue-700'
                        }`}
                    >
                        <Menu size={20} strokeWidth={1.5} />
                    </button>
                    <div className="hidden lg:flex flex-col">
                        <div className="flex items-center text-body font-medium text-gray-800">
                            <Stethoscope size={16} className="mr-2 text-blue-500" />
                            {department}
                        </div>
                        <p className="text-meta text-foreground-muted ml-6">{wardName}</p>
                    </div>
                </div>

                {/* 右侧：头像弹出菜单 */}
                <div className="flex items-center">

                    <Popover
                        content={<HeaderUserMenu username={username} userRole={userRole} onLogout={onLogout || (() => {})} />}
                        trigger="click"
                        placement="bottomRight"
                        arrow={false}
                    >
                        <div className="flex items-center gap-2 cursor-pointer hover:bg-gray-50 rounded-lg px-2 py-1 transition-colors">
                            <div className="relative">
                                {!avatarFailed && userAvatar ? (
                                    <img
                                        src={userAvatar}
                                        className="w-9 h-9 rounded-lg border-2 border-white shadow-sm ring-1 ring-gray-100"
                                        alt="avatar"
                                        onError={() => setAvatarFailed(true)}
                                    />
                                ) : (
                                    <div className="w-9 h-9 rounded-lg border-2 border-white shadow-sm ring-1 ring-gray-100 bg-slate-700 text-white flex items-center justify-center font-bold text-sm">
                                        {avatarText}
                                    </div>
                                )}
                                <div className="absolute -bottom-0.5 -right-0.5 w-2.5 h-2.5 bg-green-500 border-2 border-white rounded-full"></div>
                            </div>
                        </div>
                    </Popover>
                </div>
            </div>

            {/* 面包屑行（仅深页面显示） */}
            {showBreadcrumb && (
                <div className="h-8 bg-surface border-b border-gray-100 flex items-center px-4">
                    {routeMeta.breadcrumb.map((crumb, idx) => (
                        <span key={idx} className="flex items-center">
                            {idx > 0 && <ChevronRight size={12} className="mx-1.5 text-gray-300" />}
                            <span className={`text-meta ${idx === routeMeta.breadcrumb.length - 1 ? 'text-gray-800 font-medium' : 'text-foreground-muted'}`}>
                                {crumb}
                            </span>
                        </span>
                    ))}
                </div>
            )}
        </header>
    )
}
