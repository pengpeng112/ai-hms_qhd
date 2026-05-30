import { useState, useEffect, useTransition } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { UserRole } from '@/types/original'
import type { Patient as OriginalPatient, DashboardCardConfig } from '@/types/original'
import { DASHBOARD_CARDS } from '@/constants'
import { getSelectedRoleUser, type AppRole } from '@/services/role'
import {
    restApi,
    convertRestPatientList,
    getActiveShifts,
    getAllEquipments,
    getTodayTreatments,
    type Shift as APIShift,
    type EquipmentInfo,
    type Treatment
} from '@/services'
import {
    TrendingUp, Users, Activity, MoreHorizontal, Monitor, Package, Link2, ArrowRight,
    LayoutGrid, Library, Plus, X, CheckCircle2, Settings2, AlertTriangle, FilePlus, Clock
} from 'lucide-react'
import { BarChart, Bar, XAxis, Tooltip, ResponsiveContainer } from 'recharts'
import { PatientListItem } from '@/components'

// Extended config type for local state
interface LocalCardConfig extends DashboardCardConfig {
    visible: boolean
    colSpan: number
    rowSpan: number
}

// localStorage key for persisting layout
const LAYOUT_STORAGE_KEY = 'dashboard_layout_config'

interface DashboardProps {
    userRole?: AppRole
}

interface DashboardStatsSummary {
    activePatients: number
    shiftCount: number
    equipmentCount: number
    todaySchedules: number
    todayTreatments: number
    alertItems: number
    treatmentsByHour: { name: string; value: number }[]
    qualityByHour: { name: string; value: number }[]
}

export default function Dashboard({ userRole }: DashboardProps) {
    const navigate = useNavigate()
    const { t } = useTranslation(['dashboard', 'common', 'role'])
    const [, startTransition] = useTransition()
    const selectedRoleUser = getSelectedRoleUser()
    const currentUserRole = userRole ?? selectedRoleUser?.role ?? UserRole.DOCTOR_SUPERVISOR
    const isAdminRole = String(currentUserRole) === 'ADMIN'
    const nurseRoles: UserRole[] = [
        UserRole.NURSE_HEAD,
        UserRole.NURSE_MANAGER,
        UserRole.NURSE_RESPONSIBLE,
        UserRole.NURSE_SCHEDULER,
    ]
    const isNurseRole = currentUserRole !== 'ADMIN' && nurseRoles.includes(currentUserRole)
    const [isCustomizing, setIsCustomizing] = useState(false)
    const [showWidgetLibrary, setShowWidgetLibrary] = useState(false)
    const [cardsConfig, setCardsConfig] = useState<LocalCardConfig[]>([])
    const [patients, setPatients] = useState<Partial<OriginalPatient>[]>([])
    const [patientTotal, setPatientTotal] = useState<number | null>(null)
    const [shifts, setShifts] = useState<APIShift[]>([])
    const [equipments, setEquipments] = useState<EquipmentInfo[]>([])
    const [treatments, setTreatments] = useState<Treatment[]>([])
    const [apiError, setApiError] = useState<string | null>(null)
    const [dashboardStats, setDashboardStats] = useState<DashboardStatsSummary | null>(null)
    const [treatmentsByHour, setTreatmentsByHour] = useState<{ name: string; value: number }[]>([])
    const [qualityByHour, setQualityByHour] = useState<{ name: string; value: number }[]>([])

    // 加载数据：患者列表走 REST API，其他走 GraphQL/HDIS
    useEffect(() => {
        const loadData = async () => {
            try {
                const [patientResult, shiftsData, equipmentsData, treatmentsData, statsData] = await Promise.all([
                    restApi.getPatientList({ page: 1, pageSize: 50, onlyActive: true }).catch(() => null),
                    getActiveShifts().catch(() => []),
                    getAllEquipments().catch(() => []),
                    getTodayTreatments().catch(() => []),
                    restApi.getDashboardStats().catch(() => null),
                ])

                if (patientResult?.data?.items) {
                    setPatients(convertRestPatientList(patientResult.data.items))
                    setPatientTotal(patientResult.data.pagination?.total ?? patientResult.data.items.length)
                }
                if (shiftsData.length > 0) {
                    setShifts(shiftsData)
                }
                if (equipmentsData.length > 0) {
                    setEquipments(equipmentsData)
                }
                if (treatmentsData.length > 0) {
                    setTreatments(treatmentsData)
                }
                if (statsData) {
                    setDashboardStats(statsData)
                    setTreatmentsByHour(statsData.treatmentsByHour ?? [])
                    setQualityByHour(statsData.qualityByHour ?? [])
                }
                setApiError(null)
            } catch (error) {
                console.log('Dashboard data loading error:', error)
                setApiError(t('common:api.notConfigured'))
            }
        }
        loadData()
    }, [t])

    // Initialize cards based on user role - load from localStorage if available
    useEffect(() => {
        // Try to load from localStorage first
        const storageKey = `${LAYOUT_STORAGE_KEY}_${currentUserRole}`
        const savedConfig = localStorage.getItem(storageKey)

        if (savedConfig) {
            try {
                const parsed = JSON.parse(savedConfig) as LocalCardConfig[]
                // Merge saved config with current DASHBOARD_CARDS to handle new cards
                const mergedConfig = DASHBOARD_CARDS.map(card => {
                    const saved = parsed.find(c => c.id === card.id)
                    if (saved) {
                        return { ...card, ...saved }
                    }
                    // New card not in saved config - use defaults
                    let defaultCol = 3
                    let defaultRow = 3
                    if (card.size === 'medium') { defaultCol = 6; defaultRow = 5 }
                    if (card.size === 'large') { defaultCol = 6; defaultRow = 6 }
                    if (card.id === 'dept_overview') { defaultCol = 12; defaultRow = 3 }
                    if (card.id === 'duty_monitor') { defaultCol = 12; defaultRow = 5 }
                    if (card.id === 'device_status_eng') { defaultCol = 12; defaultRow = 5 }
                    return {
                        ...card,
                        visible: isAdminRole ? card.roles.includes('ADMIN' as UserRole) : card.roles.includes(currentUserRole as UserRole),
                        colSpan: defaultCol,
                        rowSpan: defaultRow
                    }
                })
                startTransition(() => {
                    setCardsConfig(mergedConfig)
                })
                return
            } catch (e) {
                console.warn('Failed to parse saved layout config:', e)
            }
        }

        // No saved config or parse failed - use defaults
        const initialConfig = DASHBOARD_CARDS.map(card => {
            let defaultCol = 3
            let defaultRow = 3

            if (card.size === 'medium') {
                defaultCol = 6
                defaultRow = 5
            }
            if (card.size === 'large') {
                defaultCol = 6
                defaultRow = 6
            }

            if (card.id === 'dept_overview') { defaultCol = 12; defaultRow = 3 }
            if (card.id === 'duty_monitor') { defaultCol = 12; defaultRow = 5 }
            if (card.id === 'device_status_eng') { defaultCol = 12; defaultRow = 5 }

            const isAllowed = isAdminRole ? card.roles.includes('ADMIN' as UserRole) : card.roles.includes(currentUserRole as UserRole)

            return {
                ...card,
                visible: isAllowed,
                colSpan: defaultCol,
                rowSpan: defaultRow
            }
        })
        startTransition(() => {
            setCardsConfig(initialConfig)
        })
    }, [currentUserRole])

    // Save layout to localStorage when it changes
    useEffect(() => {
        if (cardsConfig.length > 0) {
            const storageKey = `${LAYOUT_STORAGE_KEY}_${currentUserRole}`
            localStorage.setItem(storageKey, JSON.stringify(cardsConfig))
        }
    }, [cardsConfig, currentUserRole])

    const toggleVisibility = (id: string, visible: boolean) => {
        setCardsConfig(prev => prev.map(c => c.id === id ? { ...c, visible } : c))
    }

    const handleCardClick = (card: LocalCardConfig) => {
        if (isCustomizing) return
        switch (card.id) {
            case 'dept_overview':
                navigate('/statistics')
                break
            case 'active_patients':
            case 'my_duty_patients':
            case 'prescription_adjust':
                navigate('/patients')
                break
            case 'quality_stats':
            case 'nurse_workload':
                navigate('/statistics')
                break
            case 'duty_monitor':
            case 'device_status_eng':
                navigate('/monitoring')
                break
            case 'staff_schedule':
            case 'schedule_adjust':
                navigate('/schedule')
                break
            default:
                break
        }
    }

    const handlePatientSelect = (patientId: string) => {
        navigate(`/patients/${patientId}`)
    }

    // Helper to generate Tailwind classes for responsiveness
    const getColSpanClass = (span: number) => {
        const map: Record<number, string> = {
            1: 'md:col-span-1', 2: 'md:col-span-2', 3: 'md:col-span-3', 4: 'md:col-span-4',
            5: 'md:col-span-5', 6: 'md:col-span-6', 7: 'md:col-span-7', 8: 'md:col-span-8',
            9: 'md:col-span-9', 10: 'md:col-span-10', 11: 'md:col-span-11', 12: 'md:col-span-12'
        }
        return map[span] || 'md:col-span-12'
    }

    // Helper to get translated card title
    const getCardTitle = (cardId: string): string => {
        const titleMap: Record<string, string> = {
            'dept_overview': t('dashboard:card.deptOverview'),
            'active_patients': t('dashboard:card.activePatients'),
            'my_duty_patients': t('dashboard:card.myDutyPatients'),
            'prescription_adjust': t('dashboard:card.prescriptionAdjust'),
            'quality_stats': t('dashboard:card.qualityStats'),
            'nurse_workload': t('dashboard:card.nurseWorkload'),
            'duty_monitor': t('dashboard:card.dutyMonitor'),
            'device_status_eng': t('dashboard:card.deviceStatusEng'),
            'staff_schedule': t('dashboard:card.staffSchedule'),
            'schedule_adjust': t('dashboard:card.scheduleAdjust'),
            'maintenance_logs': t('dashboard:card.maintenanceLogs'),
            'consumables_prep': t('dashboard:card.consumablesPrep'),
            'device_binding': t('dashboard:card.deviceBinding'),
        }
        return titleMap[cardId] || cardId
    }

    // Card Container Component
    const CardWrapper: React.FC<{ config: LocalCardConfig; children: React.ReactNode }> = ({ config, children }) => (
        <div
            className={`
                bg-white rounded-lg shadow-sm border border-gray-100 flex flex-col overflow-hidden group relative
                col-span-1 ${getColSpanClass(config.colSpan)}
                ${isCustomizing ? 'ring-2 ring-blue-400 border-blue-400 z-10' : 'hover:shadow-md transition-shadow duration-200 cursor-pointer hover:border-blue-200'}
            `}
            style={{ gridRow: `span ${config.rowSpan}` }}
            onClick={() => handleCardClick(config)}
        >
            {isCustomizing && (
                <button
                    onClick={(e) => { e.stopPropagation(); toggleVisibility(config.id, false) }}
                    className="absolute top-2 right-2 z-20 p-1 bg-red-100 text-red-500 rounded-full hover:bg-red-200 transition-colors"
                >
                    <X size={14} />
                </button>
            )}

            <div className="px-5 py-4 border-b border-gray-50 flex justify-between items-center bg-white select-none shrink-0 h-[60px]">
                <h3 className="font-bold text-gray-800 flex items-center gap-2 text-base truncate">
                    {config.type === 'stat' && <TrendingUp size={18} className="text-blue-500" />}
                    {config.type === 'list' && <Users size={18} className="text-teal-500" />}
                    {config.type === 'action' && <AlertTriangle size={18} className="text-orange-500" />}
                    {config.type === 'monitor' && <Activity size={18} className="text-purple-500" />}
                    {config.type === 'inventory' && <Package size={18} className="text-indigo-500" />}
                    {config.type === 'binding' && <Link2 size={18} className="text-slate-500" />}
                    {getCardTitle(config.id)}
                </h3>
                {!isCustomizing && (
                    <div className="flex items-center space-x-2">
                        <button className="text-gray-400 hover:text-blue-600 opacity-0 group-hover:opacity-100 transition-opacity p-1">
                            <Settings2 size={14} />
                        </button>
                        <button className="text-gray-400 hover:text-gray-600">
                            <MoreHorizontal size={20} />
                        </button>
                    </div>
                )}
            </div>

            <div className={`p-5 flex-1 overflow-auto relative min-h-0 ${isCustomizing ? 'pointer-events-none' : ''}`}>
                {children}
            </div>
        </div>
    )

    const renderCardContent = (card: DashboardCardConfig) => {
        const totalPatients = dashboardStats?.activePatients ?? patientTotal ?? patients.length

        switch (card.id) {
            case 'dept_overview': {
                const todayTreatmentCount = dashboardStats?.todayTreatments ?? treatments.length
                const equipmentCount = dashboardStats?.equipmentCount ?? equipments.length
                const shiftCount = dashboardStats?.shiftCount ?? shifts.length
                return (
                    <div className="grid grid-cols-2 md:grid-cols-4 gap-4 h-full">
                        <div className="p-4 bg-blue-50/50 rounded-lg border border-blue-100 flex flex-col justify-center">
                            <p className="text-xs text-blue-600 font-medium uppercase mb-1">{t('dashboard:stat.patientsInDept')}</p>
                            <div className="flex items-baseline">
                                <p className="text-3xl font-bold text-gray-800">{totalPatients}</p>
                                <span className="ml-2 text-xs text-green-600 flex items-center"><TrendingUp size={10} className="mr-0.5" /> +2</span>
                            </div>
                        </div>
                        <div className="p-4 bg-teal-50/50 rounded-lg border border-teal-100 flex flex-col justify-center">
                            <p className="text-xs text-teal-600 font-medium uppercase mb-1">{t('dashboard:stat.todayDialysis')}</p>
                            <p className="text-3xl font-bold text-gray-800">{todayTreatmentCount || '--'}</p>
                        </div>
                        <div className="p-4 bg-purple-50/50 rounded-lg border border-purple-100 flex flex-col justify-center">
                            <p className="text-xs text-purple-600 font-medium uppercase mb-1">{t('dashboard:stat.totalDevices')}</p>
                            <p className="text-3xl font-bold text-gray-800">{equipmentCount || '--'}</p>
                        </div>
                        <div className="p-4 bg-orange-50/50 rounded-lg border border-orange-100 flex flex-col justify-center">
                            <p className="text-xs text-orange-600 font-medium uppercase mb-1">{t('dashboard:stat.shiftCount')}</p>
                            <p className="text-3xl font-bold text-gray-800">{shiftCount || '--'}</p>
                        </div>
                    </div>
                )
            }

            case 'nurse_workload':
            case 'quality_stats':
                return (
                    <div className="h-full min-h-[220px] w-full">
                        <ResponsiveContainer width="100%" height="100%">
                            <BarChart data={card.id === 'quality_stats' ? qualityByHour : treatmentsByHour}>
                                <XAxis dataKey="name" fontSize={10} tickLine={false} axisLine={false} tick={{ fill: '#9ca3af' }} />
                                <Tooltip
                                    contentStyle={{ borderRadius: '8px', border: 'none', boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.1)' }}
                                    cursor={{ fill: '#f3f4f6' }}
                                />
                                <Bar dataKey="value" fill={card.id === 'quality_stats' ? "#10b981" : "#3b82f6"} radius={[4, 4, 0, 0]} barSize={24} />
                            </BarChart>
                        </ResponsiveContainer>
                    </div>
                )

            case 'active_patients':
            case 'my_duty_patients':
                return (
                    <div className="space-y-3">
                        {patients.slice(0, 5).map(patient => (
                            <div key={patient.id} onClick={(e) => e.stopPropagation()}>
                                <PatientListItem
                                    patient={patient as OriginalPatient}
                                    variant="compact"
                                    onClick={(p) => handlePatientSelect(p.id)}
                                />
                            </div>
                        ))}
                        {patients.length === 0 && (
                            <div className="text-center py-4 text-gray-400 text-sm">{t('common:noData.patient') || '\u6682\u65e0\u60a3\u8005\u6570\u636e'}</div>
                        )}
                        <button
                            onClick={(e) => { e.stopPropagation(); navigate('/patients') }}
                            className="w-full py-2 text-xs text-center text-gray-400 hover:text-blue-600 transition-colors border-t border-gray-50 mt-1"
                        >
                            {t('common:action.viewAll')} <ArrowRight size={10} className="inline ml-1" />
                        </button>
                    </div>
                )

            case 'prescription_adjust':
            case 'schedule_adjust':
                return (
                    <div className="space-y-3">
                        <div className="p-3 bg-white rounded-lg border border-orange-200 shadow-sm relative overflow-hidden">
                            <div className="absolute top-0 left-0 w-1 h-full bg-orange-500"></div>
                            <div className="flex justify-between items-start mb-1">
                                <div className="flex items-center">
                                    <AlertTriangle size={16} className="text-orange-500 mr-2" />
                                    <h4 className="font-bold text-gray-800 text-sm">{t('dashboard:alert.bpLow')}</h4>
                                </div>
                            </div>
                            <p className="text-xs text-gray-600 pl-6 mb-2">{t('dashboard:alert.adjustUF')}</p>
                            <button onClick={(e) => { e.stopPropagation(); navigate('/monitoring') }} className="ml-6 px-2 py-1 bg-orange-50 text-orange-600 text-xs rounded hover:bg-orange-100">
                                {t('common:action.handle')}
                            </button>
                        </div>
                        <div className="p-3 bg-white rounded-lg border border-blue-200 shadow-sm relative overflow-hidden">
                            <div className="absolute top-0 left-0 w-1 h-full bg-blue-500"></div>
                            <div className="flex justify-between items-start mb-1">
                                <div className="flex items-center">
                                    <FilePlus size={16} className="text-blue-500 mr-2" />
                                    <h4 className="font-bold text-gray-800 text-sm">{t('dashboard:alert.orderRequest')}</h4>
                                </div>
                            </div>
                            <p className="text-xs text-gray-600 pl-6 mb-2">{t('common:order.heparinRequest')}</p>
                            <button onClick={(e) => { e.stopPropagation(); navigate('/dialysis-processing') }} className="ml-6 px-2 py-1 bg-blue-50 text-blue-600 text-xs rounded hover:bg-blue-100">
                                {t('common:action.review')}
                            </button>
                        </div>
                    </div>
                )

            case 'duty_monitor':
            case 'device_status_eng': {
                // ????????????? 24 ?
                const displayEquipments = equipments.slice(0, 24)
                return (
                    <div className="grid grid-cols-4 lg:grid-cols-6 gap-2">
                        {displayEquipments.length > 0 ? displayEquipments.map((eq, i) => {
                            const status = (eq.Status || '').toLowerCase()
                            const isAlarm = status === 'alarm' || status === 'error' || status === '报警'
                            const isOffline = status === 'offline' || status === 'inactive' || status === '离线'
                            return (
                                <div key={eq.Id} className={`p-2 rounded border flex flex-col items-center text-center relative ${isAlarm ? 'bg-red-50 border-red-200' :
                                        isOffline ? 'bg-gray-100 border-gray-200' : 'bg-white border-gray-100'
                                    }`}>
                                    {/* eslint-disable-next-line no-restricted-syntax -- density:strict 故意小字（设备序号） */}
                                    <span className="text-[10px] font-bold text-gray-500 absolute top-1 left-1">{i + 1}</span>
                                    <div className={`mt-3 mb-1 ${isAlarm ? 'text-red-500' : isOffline ? 'text-gray-400' : 'text-green-500'}`}>
                                        <Monitor size={16} />
                                    </div>
                                    <span className="text-[8px] text-gray-400 truncate w-full">{eq.Name || eq.IDNo}</span>
                                </div>
                            )
                        }) : (
                            // 无数据时显示占位
                            Array.from({ length: 12 }).map((_, i) => (
                                <div key={i} className="p-2 rounded border border-gray-100 bg-gray-50 flex flex-col items-center text-center relative">
                                    {/* eslint-disable-next-line no-restricted-syntax -- density:strict 故意小字（设备序号） */}
                                    <span className="text-[10px] font-bold text-gray-300 absolute top-1 left-1">{i + 1}</span>
                                    <div className="mt-3 mb-1 text-gray-300">
                                        <Monitor size={16} />
                                    </div>
                                </div>
                            ))
                        )}
                    </div>
                )
            }

            case 'staff_schedule':
                // 使用真实班次数据
                return (
                    <div className="space-y-2">
                        {shifts.length > 0 ? shifts.map(shift => (
                            <div key={shift.Id} className="flex items-center justify-between p-3 bg-gray-50 rounded-lg border border-gray-100">
                                <div className="flex items-center gap-3">
                                    <div className="w-8 h-8 rounded-full bg-blue-100 flex items-center justify-center">
                                        <Clock size={14} className="text-blue-600" />
                                    </div>
                                    <div>
                                        <p className="font-medium text-gray-800 text-sm">{shift.Name || `${t('common:shift')} ${shift.Id}`}</p>
                                        <p className="text-xs text-gray-500">
                                            {shift.StartTime && shift.EndTime ? `${shift.StartTime} - ${shift.EndTime}` : t('common:timeNotSet')}
                                        </p>
                                    </div>
                                </div>
                                <span className={`text-xs px-2 py-1 rounded-full ${shift.Status === '1' ? 'bg-green-100 text-green-600' : 'bg-gray-100 text-gray-500'}`}>
                                    {shift.Type || t('common:regular')}
                                </span>
                            </div>
                        )) : (
                            <div className="text-center py-8 text-gray-400 text-sm">{t('common:noData.shift')}</div>
                        )}
                    </div>
                )

            case 'maintenance_logs':
                // 显示设备数量统计
                return (
                    <div className="space-y-3">
                        <div className="flex items-center justify-between p-3 bg-green-50 rounded-lg border border-green-100">
                            <span className="text-sm text-gray-700">{t('dashboard:stat.totalDevices')}</span>
                            <span className="text-lg font-bold text-green-600">{equipments.length}</span>
                        </div>
                        <div className="flex items-center justify-between p-3 bg-blue-50 rounded-lg border border-blue-100">
                            <span className="text-sm text-gray-700">{t('dashboard:stat.todayDialysis')}</span>
                            <span className="text-lg font-bold text-blue-600">{treatments.length}</span>
                        </div>
                        <div className="flex items-center justify-between p-3 bg-purple-50 rounded-lg border border-purple-100">
                            <span className="text-sm text-gray-700">{t('dashboard:stat.shiftCount')}</span>
                            <span className="text-lg font-bold text-purple-600">{shifts.length}</span>
                        </div>
                    </div>
                )

            case 'consumables_prep':
                return (
                    <div className="flex items-center justify-center h-full text-gray-400 text-xs">{t('dashboard:card.consumablesPrep')}</div>
                )
            case 'device_binding':
                return (
                    <div className="flex items-center justify-center h-full text-gray-400 text-xs">{t('dashboard:card.deviceBinding')}</div>
                )

            default:
                return <div className="text-gray-400 text-sm flex items-center justify-center h-full">{t('dashboard:loading')}</div>
        }
    }

    const hiddenCards = cardsConfig.filter(c => !c.visible)
    const visibleCards = cardsConfig.filter(c => c.visible)

    return (
        <div className="max-w-[1800px] mx-auto relative">
            {/* API Status Banner */}
            {apiError && (
                <div className="mb-4 px-4 py-2 bg-yellow-50 border border-yellow-200 rounded-lg text-yellow-700 text-sm">
                    {apiError}
                </div>
            )}

            <div className="flex justify-between items-center mb-6">
                <div>
                    <h2 className="text-2xl font-bold text-gray-800">{t('dashboard:title')}</h2>
                    <p className="text-gray-500 text-sm mt-1">
                        {isAdminRole ? t('role:desc.admin') :
                            isNurseRole ? t('role:desc.nurse') :
                                t('role:desc.doctor')}
                    </p>
                </div>
                <div className="flex space-x-3">
                    {isCustomizing && (
                        <button
                            onClick={() => setShowWidgetLibrary(true)}
                            className="flex items-center px-4 py-2 bg-white border border-blue-500 text-blue-600 rounded-lg text-sm font-medium hover:bg-blue-50 transition-colors shadow-sm"
                        >
                            <Plus size={16} className="mr-2" /> {t('dashboard:action.addCard')}
                        </button>
                    )}

                    <button
                        onClick={() => {
                            setIsCustomizing(!isCustomizing)
                            setShowWidgetLibrary(false)
                        }}
                        className={`flex items-center px-4 py-2 rounded-lg text-sm font-medium transition-colors ${isCustomizing ? 'bg-blue-600 text-white shadow-md' : 'bg-white text-gray-600 border border-gray-200 hover:bg-gray-50 hover:border-gray-300'}`}
                    >
                        {isCustomizing ? (
                            <>
                                <CheckCircle2 size={16} className="mr-2" /> {t('dashboard:action.finishLayout')}
                            </>
                        ) : (
                            <>
                                <LayoutGrid size={16} className="mr-2" /> {t('dashboard:action.customizeLayout')}
                            </>
                        )}
                    </button>
                </div>
            </div>

            {/* Grid Layout System */}
            <div className={`grid grid-cols-1 md:grid-cols-12 gap-4 pb-10 content-start auto-rows-[60px] ${isCustomizing ? 'select-none' : ''}`}>
                {visibleCards.map(card => (
                    <CardWrapper key={card.id} config={card}>
                        {renderCardContent(card)}
                    </CardWrapper>
                ))}
            </div>

            {/* Widget Library Drawer */}
            {showWidgetLibrary && isCustomizing && (
                <div className="fixed top-16 right-0 bottom-0 w-80 bg-white shadow-2xl border-l border-gray-200 z-50 animate-slide-in-right flex flex-col">
                    <div className="p-4 border-b border-gray-100 flex justify-between items-center bg-gray-50">
                        <h3 className="font-bold text-gray-800 flex items-center">
                            <Library size={18} className="mr-2 text-blue-600" /> {t('dashboard:widgetLibrary.title')}
                        </h3>
                        <button onClick={() => setShowWidgetLibrary(false)} className="text-gray-400 hover:text-gray-600">
                            <X size={20} />
                        </button>
                    </div>
                    <div className="flex-1 overflow-y-auto p-4 space-y-4">
                        <p className="text-xs text-gray-500 mb-2">{t('dashboard:widgetLibrary.hint')}</p>
                        {hiddenCards.length > 0 ? hiddenCards.map(card => (
                            <div
                                key={card.id}
                                onClick={() => toggleVisibility(card.id, true)}
                                className="p-3 border border-gray-200 rounded-lg hover:border-blue-400 hover:bg-blue-50 cursor-pointer transition-all group"
                            >
                                <div className="flex items-center justify-between mb-2">
                                    <span className="font-bold text-gray-700 group-hover:text-blue-700 text-sm">{getCardTitle(card.id)}</span>
                                    <Plus size={16} className="text-gray-400 group-hover:text-blue-600" />
                                </div>
                                <div className="text-xs text-gray-400 flex items-center">
                                    <span className="bg-gray-100 px-1.5 py-0.5 rounded mr-2">{card.type}</span>
                                    <span>{t('dashboard:widgetLibrary.defaultSize')}: {card.size}</span>
                                </div>
                            </div>
                        )) : (
                            <div className="text-center py-10 text-gray-400 text-sm">
                                <CheckCircle2 size={32} className="mx-auto mb-2 opacity-50" />
                                <p>{t('dashboard:widgetLibrary.allAdded')}</p>
                            </div>
                        )}
                    </div>
                </div>
            )}
        </div>
    )
}
