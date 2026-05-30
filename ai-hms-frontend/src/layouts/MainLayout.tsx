import { useState, useMemo, useRef, useEffect } from 'react'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import Sidebar from './Sidebar'
import Header from './Header'
import PageBreadcrumb from './PageBreadcrumb'
import { UserRole } from '@/types/original'
import { logout } from '@/services/auth'
import { getRolePermissionCodes, getSelectedRoleUser } from '@/services/role'
import { restApi, type RestClinicalTask } from '@/services/restClient'
import { getRouteMeta } from './routeMeta'
import {
    AlertCircle, Zap, FileEdit, CheckCircle2, X, ChevronRight, ClipboardList
} from 'lucide-react'

const TASKBAR_STATE_KEY = 'hdis_taskbar_open'

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
        return saved !== null ? saved === 'true' : true
    })
    const [tasks, setTasks] = useState<RestClinicalTask[]>([])
    const [permissionCodes, setPermissionCodes] = useState<string[]>([])
    const [taskLoading, setTaskLoading] = useState(false)

    useEffect(() => {
        const meta = getRouteMeta(location.pathname)
        document.title = `${meta.title} - AI-HMS 智能透析`
    }, [location.pathname])

    const taskbarRef = useRef<HTMLDivElement>(null)
    const toggleBtnRef = useRef<HTMLButtonElement>(null)

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

    useEffect(() => {
        localStorage.setItem(TASKBAR_STATE_KEY, String(taskbarOpen))
    }, [taskbarOpen])

    useEffect(() => {
        const handleClickOutside = (event: MouseEvent) => {
            if (
                taskbarOpen &&
                taskbarRef.current && !taskbarRef.current.contains(event.target as Node) &&
                toggleBtnRef.current && !toggleBtnRef.current.contains(event.target as Node)
            ) {
                setTaskbarOpen(false)
            }
        }
        document.addEventListener('mousedown', handleClickOutside)
        return () => document.removeEventListener('mousedown', handleClickOutside)
    }, [taskbarOpen])

    const handleLogout = () => {
        logout()
    }

    const handleTaskClick = (task: RestClinicalTask) => {
        navigate(getTaskRoute(task.type))
    }

    const getSeverityStyles = (severity: string) => {
        switch (severity) {
            case 'high': return 'bg-red-50 border-red-200 text-red-700 hover:bg-red-100'
            case 'medium': return 'bg-orange-50 border-orange-200 text-orange-700 hover:bg-orange-100'
            case 'low': return 'bg-blue-50 border-blue-200 text-blue-700 hover:bg-blue-100'
            default: return 'bg-gray-50 border-gray-200 text-gray-700 hover:bg-gray-100'
        }
    }

    const getTaskIcon = (type: string) => {
        switch (type) {
            case 'ALERT': return <AlertCircle size={16} />
            case 'PRESCRIPTION': return <FileEdit size={16} />
            case 'ORDER': return <Zap size={16} />
            case 'ASSESSMENT': return <CheckCircle2 size={16} />
            default: return <ClipboardList size={16} />
        }
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
                        <PageBreadcrumb />
                        <Outlet />
                    </main>

                    <aside
                        ref={taskbarRef}
                        className={`bg-white border-l border-gray-200 flex flex-col transition-all duration-300 ease-in-out shadow-xl z-20
                            ${taskbarOpen ? 'w-[340px]' : 'w-0 overflow-hidden border-none'}`}
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
                                    <div key={idx} className="p-4 rounded-2xl border border-gray-200 bg-white animate-pulse h-24" />
                                ))
                            ) : visibleTasks.length > 0 ? (
                                visibleTasks.map(task => {
                                    const canHandle = canHandleTask(permissionSet, task.type)
                                    return (
                                    <div
                                        key={task.id}
                                        onClick={canHandle ? () => handleTaskClick(task) : undefined}
                                        className={`group p-4 rounded-2xl border-l-4 shadow-sm transition-all ${canHandle ? 'cursor-pointer active:scale-[0.98]' : 'cursor-not-allowed opacity-70'} ${getSeverityStyles(task.severity)} border-l-current`}
                                    >
                                        <div className="flex justify-between items-start mb-2">
                                            <div className="flex items-center font-bold text-sm whitespace-nowrap">
                                                <span className="mr-2 p-1.5 bg-white/50 rounded-lg shrink-0">{getTaskIcon(task.type)}</span>
                                                {task.title}
                                            </div>
                                            {/* density:strict 故意小字 */}
                                            <span className="text-[10px] font-bold opacity-60 bg-white/30 px-1.5 py-0.5 rounded whitespace-nowrap shrink-0">
                                                {t('taskbar.bed', { bed: task.bedNumber || '--' })}
                                            </span>
                                        </div>
                                        <p className="text-xs font-bold mb-1">{task.patientName || '--'}</p>
                                        <p className="text-xs opacity-80 leading-relaxed mb-3">{task.description || ''}</p>
                                        <div className={`flex items-center justify-end text-meta font-bold uppercase tracking-wider transition-transform ${canHandle ? 'group-hover:translate-x-1' : ''}`}>
                                            {canHandle ? t('taskbar.goHandle') : '无处理权限'} <ChevronRight size={12} className="ml-1" />
                                        </div>
                                    </div>
                                    )
                                })
                            ) : (
                                <div className="flex flex-col items-center justify-center h-full text-gray-400 opacity-50 space-y-3">
                                    <CheckCircle2 size={48} className="text-green-500" />
                                    <p className="text-sm font-medium">暂无待处理任务</p>
                                </div>
                            )}
                        </div>

                        <div className="p-4 border-t border-gray-100 bg-gray-50/80 shrink-0">
                            <button className="w-full py-2.5 text-xs font-bold text-blue-600 bg-white border border-blue-100 rounded-xl hover:bg-blue-50 transition-colors shadow-sm whitespace-nowrap">
                                {t('taskbar.viewHistory')}
                            </button>
                        </div>
                    </aside>
                </div>
            </div>
        </div>
    )
}
