import { useState, useMemo, useRef, useEffect } from 'react'
import { Outlet, useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import Sidebar from './Sidebar'
import Header from './Header'
import { UserRole } from '@/types/original'
import { logout } from '@/services/auth'
import { getSelectedRoleUser } from '@/services/role'
import {
    AlertCircle, Zap, FileEdit, CheckCircle2, X, ChevronRight, ClipboardList
} from 'lucide-react'

interface TaskItem {
    id: string
    type: 'PRESCRIPTION' | 'ORDER' | 'ALERT' | 'ASSESSMENT'
    title: string
    patientName: string
    bedNumber: string
    patientId: string
    description: string
    severity: 'high' | 'medium' | 'low'
    roles: UserRole[]
    targetRoute: string
}

// Mock tasks data
const MOCK_TASKS: TaskItem[] = [
    {
        id: 't1', type: 'ALERT', title: '危急值预警', patientName: '张伟', bedNumber: 'A01', patientId: 'P001',
        description: '收缩压 85mmHg，低于阈值', severity: 'high', targetRoute: '/monitoring',
        roles: [UserRole.DOCTOR_CHIEF, UserRole.DOCTOR_SUPERVISOR, UserRole.DOCTOR_DUTY, UserRole.NURSE_RESPONSIBLE]
    },
    {
        id: 't2', type: 'PRESCRIPTION', title: '处方待调整', patientName: '李娜', bedNumber: 'A02', patientId: 'P002',
        description: '体重增量过大，需重新评估超滤量', severity: 'medium', targetRoute: '/monitoring',
        roles: [UserRole.DOCTOR_CHIEF, UserRole.DOCTOR_SUPERVISOR, UserRole.DOCTOR_DUTY]
    },
    {
        id: 't3', type: 'ORDER', title: '新医嘱待执行', patientName: '王强', bedNumber: 'B05', patientId: 'P003',
        description: '追加低分子肝素 1000iu iv', severity: 'medium', targetRoute: '/dialysis',
        roles: [UserRole.NURSE_RESPONSIBLE, UserRole.NURSE_HEAD]
    },
    {
        id: 't4', type: 'ASSESSMENT', title: '透后评估提醒', patientName: '孙行', bedNumber: 'C01', patientId: 'P006',
        description: '治疗即将结束，请及时进行透后评估', severity: 'low', targetRoute: '/dialysis',
        roles: [UserRole.NURSE_RESPONSIBLE]
    },
    {
        id: 't5', type: 'PRESCRIPTION', title: '处方变动确认', patientName: '李娜', bedNumber: 'A02', patientId: 'P002',
        description: '医生已修改超滤方案，请核对', severity: 'medium', targetRoute: '/dialysis',
        roles: [UserRole.NURSE_RESPONSIBLE]
    },
]

// localStorage key for taskbar state
const TASKBAR_STATE_KEY = 'hdis_taskbar_open'

export default function MainLayout() {
    const { t } = useTranslation('common')
    const navigate = useNavigate()
    const [sidebarOpen, setSidebarOpen] = useState(true)
    const [taskbarOpen, setTaskbarOpen] = useState(() => {
        const saved = localStorage.getItem(TASKBAR_STATE_KEY)
        return saved !== null ? saved === 'true' : true
    })
    const taskbarRef = useRef<HTMLDivElement>(null)
    const toggleBtnRef = useRef<HTMLButtonElement>(null)

    // 获取选中的角色用户
    const roleUser = useMemo(() => getSelectedRoleUser(), [])

    // 用户数据 - 从角色选择获取
    const user = {
        name: roleUser?.name || '用户',
        role: roleUser?.role || UserRole.DOCTOR_SUPERVISOR,
        avatar: roleUser?.avatar || 'https://api.dicebear.com/7.x/avataaars/svg?seed=User',
    }

    // Filter tasks visible to current role
    const visibleTasks = MOCK_TASKS.filter(task => task.roles.includes(user.role))

    // Persist taskbar state
    useEffect(() => {
        localStorage.setItem(TASKBAR_STATE_KEY, String(taskbarOpen))
    }, [taskbarOpen])

    // Click outside to close
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

    const handleTaskClick = (task: TaskItem) => {
        // Navigate to target route
        navigate(task.targetRoute)
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
            {/* Sidebar */}
            <Sidebar isOpen={sidebarOpen} />

            {/* Main Content Area */}
            <div className="flex-1 flex flex-col min-w-0">
                {/* Header */}
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

                {/* Content + Taskbar as flex siblings for push/shrink effect */}
                <div className="flex-1 flex overflow-hidden">
                    {/* Main content - automatically shrinks when taskbar opens */}
                    <main className="flex-1 overflow-auto p-6 no-scrollbar">
                        <Outlet />
                    </main>

                    {/* Real-time Task Panel - flex sibling for push effect */}
                    <aside
                        ref={taskbarRef}
                        className={`bg-white border-l border-gray-200 flex flex-col transition-all duration-300 ease-in-out shadow-xl z-20
                            ${taskbarOpen ? 'w-[340px]' : 'w-0 overflow-hidden border-none'}`}
                    >
                        {/* Header */}
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

                        {/* Task List */}
                        <div className="flex-1 p-4 space-y-3 overflow-y-auto no-scrollbar bg-slate-50/50">
                            {visibleTasks.length > 0 ? (
                                visibleTasks.map(task => (
                                    <div
                                        key={task.id}
                                        onClick={() => handleTaskClick(task)}
                                        className={`group p-4 rounded-2xl border-l-4 shadow-sm cursor-pointer transition-all active:scale-[0.98] ${getSeverityStyles(task.severity)} border-l-current`}
                                    >
                                        <div className="flex justify-between items-start mb-2">
                                            <div className="flex items-center font-bold text-sm whitespace-nowrap">
                                                <span className="mr-2 p-1.5 bg-white/50 rounded-lg shrink-0">{getTaskIcon(task.type)}</span>
                                                {task.title}
                                            </div>
                                            <span className="text-[10px] font-bold opacity-60 bg-white/30 px-1.5 py-0.5 rounded whitespace-nowrap shrink-0">
                                                {t('taskbar.bed', { bed: task.bedNumber })}
                                            </span>
                                        </div>
                                        <p className="text-xs font-bold mb-1">{task.patientName}</p>
                                        <p className="text-xs opacity-80 leading-relaxed mb-3">{task.description}</p>
                                        <div className="flex items-center justify-end text-[10px] font-bold uppercase tracking-wider group-hover:translate-x-1 transition-transform">
                                            {t('taskbar.goHandle')} <ChevronRight size={12} className="ml-1" />
                                        </div>
                                    </div>
                                ))
                            ) : (
                                <div className="flex flex-col items-center justify-center h-full text-gray-400 opacity-50 space-y-3">
                                    <CheckCircle2 size={48} className="text-green-500" />
                                    <p className="text-sm font-medium">{t('empty.tasks')}</p>
                                </div>
                            )}
                        </div>

                        {/* Footer */}
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
