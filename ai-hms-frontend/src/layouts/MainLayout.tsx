import { useState, useMemo, useRef, useEffect } from 'react'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import Sidebar from './Sidebar'
import Header from './Header'
import TaskCard from './TaskCard'
import TaskbarRail from './TaskbarRail'
import { UserRole } from '@/types/original'
import { logout } from '@/services/auth'
import { getRolePermissionCodes, getSelectedRoleUser } from '@/services/role'
import { restApi, type RestClinicalTask } from '@/services/restClient'
import { getRouteMeta } from './routeMeta'
import { X } from 'lucide-react'

const TASKBAR_STATE_KEY = 'hdis_taskbar_open'
const TASKBAR_LOCKED_KEY = 'hdis_taskbar_locked'

const TASK_PERMISSION_MAP: Record<string, string[]> = {
    ALERT: ['task.alert.view', 'monitoring', 'menu.monitoring'],
    PRESCRIPTION: ['task.prescription.view', 'monitoring', 'menu.monitoring'],
    ORDER: ['task.order.view', 'dialysis_processing', 'menu.dialysis_processing'],
    ASSESSMENT: ['task.assessment.view', 'dialysis_processing', 'menu.dialysis_processing'],
}

const TASK_HANDLE_PERMISSION_MAP: Record<string, string[]> = {
    ALERT: ['task.alert.handle'],
    PRESCRIPTION: ['task.prescription.handle'],
    ORDER: ['task.order.handle'],
    ASSESSMENT: ['task.assessment.handle'],
}

const canViewTask = (permissions: Set<string>, type: string) => {
    const required = TASK_PERMISSION_MAP[type]
    if (!required || required.length === 0) {
        return false
    }
    return required.some(code => permissions.has(code))
}

const canHandleTask = (permissions: Set<string>, type: string) => {
    const required = TASK_HANDLE_PERMISSION_MAP[type]
    if (!required || required.length === 0) {
        return false
    }
    return required.some(code => permissions.has(code))
}

const getTaskRoute = (type: string) => {
    if (type === 'ALERT') return '/monitoring'
    if (type === 'PRESCRIPTION') return '/monitoring'
    if (type === 'ORDER') return '/dialysis-processing'
    if (type === 'ASSESSMENT') return '/dialysis-processing'
    return '/dashboard'
}

export default function MainLayout() {
    const { t } = useTranslation('common')
    const navigate = useNavigate()
    const location = useLocation()
    const [sidebarOpen, setSidebarOpen] = useState(true)
    const [taskbarOpen, setTaskbarOpen] = useState(() => {
        const saved = localStorage.getItem(TASKBAR_STATE_KEY)
        // 改为默认收起，老用户已有显式设置则尊重
        return saved !== null ? saved === 'true' : false
    })
    const [taskbarLocked, setTaskbarLocked] = useState(() => localStorage.getItem(TASKBAR_LOCKED_KEY) === 'true')
    const [tasks, setTasks] = useState<RestClinicalTask[]>([])
    const [permissionCodes, setPermissionCodes] = useState<string[]>([])
    const [taskLoading, setTaskLoading] = useState(false)

    useEffect(() => {
        const meta = getRouteMeta(location.pathname)
        document.title = `${meta.title} - AI-HMS 智能透析`
    }, [location.pathname])

    const taskbarRef = useRef<HTMLDivElement>(null)
    const toggleBtnRef = useRef<HTMLButtonElement | null>(null)

    const roleUser = useMemo(() => getSelectedRoleUser(), [])

    const user = {
        name: roleUser?.name || '用户',
        role: roleUser?.role || UserRole.DOCTOR_SUPERVISOR,
        avatar: roleUser?.avatar || '',
    }

    useEffect(() => {
        setTaskLoading(true)
        restApi.getClinicalTasks({ status: 'pending' })
            .then(res => setTasks(res.data.items))
            .catch(() => setTasks([]))
            .finally(() => setTaskLoading(false))
    }, [])

    useEffect(() => {
        if (!roleUser?.role) {
            setPermissionCodes([])
            return
        }
        getRolePermissionCodes(roleUser.role)
            .then(codes => setPermissionCodes(codes))
            .catch(() => setPermissionCodes([]))
    }, [roleUser?.role])

    const permissionSet = useMemo(() => new Set(permissionCodes), [permissionCodes])
    const visibleTasks = useMemo(
        () => tasks.filter(task => canViewTask(permissionSet, task.type)),
        [permissionSet, tasks]
    )

    // severity 分组计数
    const severityCounts = useMemo(() => {
        let high = 0, medium = 0, low = 0
        for (const task of visibleTasks) {
            if (task.severity === 'high') high++
            else if (task.severity === 'medium') medium++
            else low++
        }
        return { high, medium, low }
    }, [visibleTasks])

    useEffect(() => {
        localStorage.setItem(TASKBAR_STATE_KEY, String(taskbarOpen))
    }, [taskbarOpen])

    useEffect(() => {
        localStorage.setItem(TASKBAR_LOCKED_KEY, String(taskbarLocked))
    }, [taskbarLocked])

    // 锁定时不响应外部点击关闭
    useEffect(() => {
        const handleClickOutside = (event: MouseEvent) => {
            if (
                taskbarOpen &&
                !taskbarLocked &&
                taskbarRef.current && !taskbarRef.current.contains(event.target as Node) &&
                toggleBtnRef.current && !toggleBtnRef.current.contains(event.target as Node)
            ) {
                setTaskbarOpen(false)
            }
        }
        document.addEventListener('mousedown', handleClickOutside)
        return () => document.removeEventListener('mousedown', handleClickOutside)
    }, [taskbarOpen, taskbarLocked])

    const handleLogout = () => {
        logout()
    }

    const handleTaskClick = (task: RestClinicalTask) => {
        navigate(getTaskRoute(task.type))
    }

    return (
        <div className="flex h-screen bg-gray-50 overflow-hidden font-sans">
            <Sidebar isOpen={sidebarOpen} />

            <div className="flex-1 flex flex-col min-w-0">
                <Header
                    username={user.name}
                    userRole={user.role}
                    userAvatar={user.avatar}
                    onLogout={handleLogout}
                    sidebarOpen={sidebarOpen}
                    onSidebarToggle={() => setSidebarOpen(!sidebarOpen)}
                    taskbarOpen={taskbarOpen}
                    onTaskbarToggle={() => setTaskbarOpen(!taskbarOpen)}
                    taskCount={visibleTasks.length}
                    toggleBtnRef={toggleBtnRef}
                />

                <div className="flex-1 flex overflow-hidden">
                    <main className="flex-1 overflow-auto p-4 no-scrollbar">
                        <Outlet />
                    </main>

                    {/* 收起态：窄条 */}
                    {!taskbarOpen && (
                        <TaskbarRail
                            taskCount={visibleTasks.length}
                            highCount={severityCounts.high}
                            mediumCount={severityCounts.medium}
                            lowCount={severityCounts.low}
                            locked={taskbarLocked}
                            onExpand={() => setTaskbarOpen(true)}
                            onToggleLock={() => setTaskbarLocked(prev => !prev)}
                        />
                    )}

                    {/* 展开态 */}
                    {taskbarOpen && (
                        <aside
                            ref={taskbarRef}
                            className="w-[320px] bg-white border-l border-gray-200 flex flex-col transition-all duration-300 ease-in-out shadow-xl z-20 shrink-0"
                        >
                            <div className="p-5 border-b border-gray-100 flex justify-between items-center bg-white shrink-0">
                                <div className="flex items-center">
                                    <span className="w-1.5 h-6 bg-blue-600 rounded-full mr-3"></span>
                                    <h3 className="font-bold text-gray-800 text-lg whitespace-nowrap">{t('taskbar.title')}</h3>
                                </div>
                                <button
                                    onClick={() => setTaskbarOpen(false)}
                                    className="p-1.5 text-gray-400 hover:bg-gray-100 rounded-full transition-colors"
                                >
                                    <X size={18} />
                                </button>
                            </div>

                            <div className="flex-1 p-4 space-y-3 overflow-y-auto no-scrollbar bg-slate-50/50">
                                {taskLoading ? (
                                    Array.from({ length: 3 }).map((_, idx) => (
                                        <div key={idx} className="p-4 rounded-md border border-gray-200 bg-white animate-pulse h-24" />
                                    ))
                                ) : visibleTasks.length > 0 ? (
                                    visibleTasks.map(task => {
                                        const canHandle = canHandleTask(permissionSet, task.type)
                                        return (
                                            <TaskCard
                                                key={task.id}
                                                task={task}
                                                canHandle={canHandle}
                                                onClick={() => handleTaskClick(task)}
                                            />
                                        )
                                    })
                                ) : (
                                    <div className="flex items-center justify-center py-8 text-gray-400">
                                        <p className="text-sm">已清空 ✓</p>
                                    </div>
                                )}
                            </div>

                            {/* TODO U5 历史接口 */}
                            <div className="p-4 border-t border-gray-100 bg-gray-50/80 shrink-0">
                                <button className="w-full py-2.5 text-xs font-bold text-blue-600 bg-white border border-blue-100 rounded-lg hover:bg-blue-50 transition-colors shadow-sm whitespace-nowrap">
                                    {t('taskbar.viewHistory')}
                                </button>
                            </div>
                        </aside>
                    )}
                </div>
            </div>
        </div>
    )
}
