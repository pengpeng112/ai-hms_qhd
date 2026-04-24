import { useState } from 'react'
import type { RefObject } from 'react'
import { useTranslation } from 'react-i18next'
import { UserRole } from '@/types/original'
import type { AppRole } from '@/services/role'
import { Menu, LogOut, ClipboardList, Stethoscope } from 'lucide-react'

interface HeaderProps {
    username?: string
    userRole?: AppRole
    userAvatar?: string
    department?: string
    onLogout?: () => void
    sidebarOpen?: boolean
    onSidebarToggle?: () => void
    taskbarOpen?: boolean
    onTaskbarToggle?: () => void
    taskCount?: number
    toggleBtnRef?: RefObject<HTMLButtonElement | null>
}

export default function Header({
    username = '',
    userRole = UserRole.DOCTOR_SUPERVISOR,
    userAvatar = '',
    department = '肾内透析中心 · 第一病区',
    onLogout,
    sidebarOpen = true,
    onSidebarToggle,
    taskbarOpen = false,
    onTaskbarToggle,
    taskCount = 0,
    toggleBtnRef,
}: HeaderProps) {
    const { t } = useTranslation('nav')
    const [avatarFailed, setAvatarFailed] = useState(false)
    const avatarText = (username || userRole || 'U').trim().slice(0, 1).toUpperCase()

    return (
        <header className="h-16 bg-[#f0f7ff] border-b border-blue-100 flex items-center justify-between px-6 z-10">
            {/* Left: Sidebar Toggle + Department Badge */}
            <div className="flex items-center">
                <button
                    onClick={onSidebarToggle}
                    className={`p-2.5 rounded-xl border transition-all mr-4 ${
                        sidebarOpen
                            ? 'bg-white border-gray-200 text-gray-600 hover:bg-gray-50'
                            : 'bg-blue-600 border-blue-600 text-white hover:bg-blue-700'
                    }`}
                >
                    <Menu size={20} strokeWidth={1.5} />
                </button>
                <div className="hidden lg:flex items-center bg-white px-4 py-2 rounded-full text-sm text-gray-600 border border-gray-200 shadow-sm">
                    <Stethoscope size={16} className="mr-2 text-blue-500" />
                    {t('header.currentDept')}: {department}
                </div>
            </div>

            {/* Right: Task Toggle + User Info */}
            <div className="flex items-center space-x-4">
                <button
                    ref={toggleBtnRef}
                    onClick={onTaskbarToggle}
                    className={`p-2 rounded-lg relative transition-all ${taskbarOpen ? 'bg-blue-600 text-white shadow-lg' : 'bg-gray-100 text-gray-500 hover:bg-gray-200'
                        }`}
                >
                    <ClipboardList size={20} />
                    {taskCount > 0 && !taskbarOpen && (
                        <span className="absolute -top-1 -right-1 w-4 h-4 bg-red-500 text-white text-[10px] flex items-center justify-center rounded-full border-2 border-white animate-bounce">
                            {taskCount}
                        </span>
                    )}
                </button>

                <div className="h-8 w-px bg-gray-200 mx-2"></div>

                <div className="flex items-center gap-3">
                    <div className="text-right hidden sm:block">
                        <p className="text-sm font-bold text-gray-800 leading-none">{username}</p>
                        <p className="text-[10px] text-gray-400 mt-1 uppercase tracking-tight">{userRole}</p>
                    </div>
                    <div className="relative group">
                        {!avatarFailed && userAvatar ? (
                            <img
                                src={userAvatar}
                                className="w-10 h-10 rounded-xl border-2 border-white shadow-sm ring-1 ring-gray-100 cursor-pointer"
                                alt="avatar"
                                onError={() => setAvatarFailed(true)}
                            />
                        ) : (
                            <div className="w-10 h-10 rounded-xl border-2 border-white shadow-sm ring-1 ring-gray-100 bg-slate-700 text-white flex items-center justify-center font-bold cursor-default">
                                {avatarText}
                            </div>
                        )}
                        <div className="absolute top-0 right-0 w-3 h-3 bg-green-500 border-2 border-white rounded-full"></div>
                    </div>
                    <button
                        onClick={onLogout}
                        className="p-2 text-gray-400 hover:text-red-500 hover:bg-red-50 rounded-lg transition-colors"
                    >
                        <LogOut size={18} />
                    </button>
                </div>
            </div>
        </header>
    )
}
