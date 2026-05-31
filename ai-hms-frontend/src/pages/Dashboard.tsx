import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { UserRole } from '@/types/original'
import type { Patient } from '@/types/original'
import { getSelectedRoleUser } from '@/services/role'
import {
  restApi, convertRestPatientList, getActiveShifts, getAllEquipments, getTodayTreatments,
  type Shift as APIShift, type EquipmentInfo, type Treatment,
} from '@/services'
import { TrendingUp, Users, Activity, MoreHorizontal, Package, Link2,
  LayoutGrid, Plus, X, CheckCircle2, AlertTriangle, Library } from 'lucide-react'
import { getDefaultCardKeys, V2_CARD_REGISTRY, LAYOUT_STORAGE_KEY_V2, type V2CardDef } from '@/constants/dashboardDefaults'
import { StatKpiCard, PatientListCard, ChartCard, AlertCard, DeviceGridCard, ShiftListCard, PlaceholderCard } from './dashboard/cards'

const LAYOUT_V1_KEY = 'dashboard_layout_config'

interface LocalCardConfig extends V2CardDef { visible: boolean; colSpan: number; rowSpan: number }

export default function Dashboard() {
  const navigate = useNavigate()
  const { t: tRaw } = useTranslation(['dashboard', 'common', 'role'])
  const t = tRaw as (key: string, fallback?: string) => string
  const selectedRoleUser = getSelectedRoleUser()
  const role = selectedRoleUser?.role ?? UserRole.DOCTOR_SUPERVISOR
  const isAdmin = String(role) === 'ADMIN'
  const nurseRoles = [UserRole.NURSE_HEAD, UserRole.NURSE_MANAGER, UserRole.NURSE_RESPONSIBLE, UserRole.NURSE_SCHEDULER] as const
  const isNurse = role !== 'ADMIN' && nurseRoles.includes(role as typeof nurseRoles[number])

  const [isCustomizing, setIsCustomizing] = useState(false)
  const [showWidgetLibrary, setShowWidgetLibrary] = useState(false)
  const [cardsConfig, setCardsConfig] = useState<LocalCardConfig[]>([])
  const [patients, setPatients] = useState<Partial<Patient>[]>([])
  const [patientTotal, setPatientTotal] = useState<number | null>(null)
  const [shifts, setShifts] = useState<APIShift[]>([])
  const [equipments, setEquipments] = useState<EquipmentInfo[]>([])
  const [treatments, setTreatments] = useState<Treatment[]>([])
  const [apiError, setApiError] = useState<string | null>(null)
  const [dashboardStats, setDashboardStats] = useState<{ activePatients: number; shiftCount: number; equipmentCount: number; todayTreatments: number; treatmentsByHour: { name: string; value: number }[]; qualityByHour: { name: string; value: number }[] } | null>(null)

  // 加载数据
  useEffect(() => {
    Promise.all([
      restApi.getPatientList({ page: 1, pageSize: 50, onlyActive: true }).catch(() => null),
      getActiveShifts().catch(() => []),
      getAllEquipments().catch(() => []),
      getTodayTreatments().catch(() => []),
      restApi.getDashboardStats().catch(() => null),
    ]).then(([patientResult, shiftsData, equipmentsData, treatmentsData, statsData]) => {
      if (patientResult?.data?.items) {
        setPatients(convertRestPatientList(patientResult.data.items))
        setPatientTotal(patientResult.data.pagination?.total ?? patientResult.data.items.length)
      }
      setShifts(shiftsData); setEquipments(equipmentsData); setTreatments(treatmentsData)
      if (statsData) setDashboardStats(statsData)
      setApiError(null)
    }).catch(() => setApiError(t('common:api.notConfigured')))
  }, [t])

  // 初始化卡片配置（v2 迁移）
  useEffect(() => {
    const v2Key = `${LAYOUT_STORAGE_KEY_V2}_${role}`
    const v1Key = `${LAYOUT_V1_KEY}_${role}`
    let saved = localStorage.getItem(v2Key)
    if (!saved) { localStorage.getItem(v1Key); saved = null } // v1 不迁移，走默认
    const defaultKeys = getDefaultCardKeys(String(role))

    if (saved) {
      try {
        const parsed = JSON.parse(saved) as LocalCardConfig[]
        const merged = V2_CARD_REGISTRY.map(def => {
          const s = parsed.find(c => c.id === def.id)
          return s ? { ...def, ...s } : { ...def, visible: defaultKeys.includes(def.id), colSpan: def.size === 'large' ? 6 : 3, rowSpan: def.size === 'large' ? 6 : 5 }
        })
        queueMicrotask(() => setCardsConfig(merged)); return
      } catch { /* fall through */ }
    }

    queueMicrotask(() => setCardsConfig(V2_CARD_REGISTRY.map(def => ({
      ...def, visible: defaultKeys.includes(def.id),
      colSpan: def.size === 'large' ? 6 : 3, rowSpan: def.size === 'large' ? 6 : 5,
    }))))
  }, [role])

  // 保存配置
  useEffect(() => {
    if (cardsConfig.length > 0) localStorage.setItem(`${LAYOUT_STORAGE_KEY_V2}_${role}`, JSON.stringify(cardsConfig))
  }, [cardsConfig, role])

  const toggleVisibility = (id: string, visible: boolean) => setCardsConfig(prev => prev.map(c => c.id === id ? { ...c, visible } : c))
  const totalPatients = dashboardStats?.activePatients ?? patientTotal ?? patients.length

  // 卡片渲染
  const renderCard = (card: LocalCardConfig) => {
    const navigateTo = (route: string) => { if (!isCustomizing) navigate(route) }
    switch (card.id) {
      case 'operationKpis':
      case 'dept_overview':
        return <StatKpiCard items={[
          { labelKey: 'dashboard:stat.patientsInDept', value: totalPatients, color: 'blue' },
          { labelKey: 'dashboard:stat.todayDialysis', value: dashboardStats?.todayTreatments ?? treatments.length, color: 'teal' },
          { labelKey: 'dashboard:stat.totalDevices', value: dashboardStats?.equipmentCount ?? equipments.length, color: 'purple' },
          { labelKey: 'dashboard:stat.shiftCount', value: dashboardStats?.shiftCount ?? shifts.length, color: 'orange' },
        ]} />
      case 'myPatientsToday':
      case 'active_patients':
      case 'my_duty_patients':
        return <PatientListCard patients={patients} onSelect={(id) => navigate(`/patients/${id}`)} onViewAll={() => navigate('/patients')} />
      case 'pendingPrescriptions':
      case 'prescription_adjust':
        return <AlertCard items={[
          { icon: 'alert', titleKey: 'dashboard:alert.bpLow', descKey: 'dashboard:alert.adjustUF', actionKey: 'common:action.handle', actionRoute: '/monitoring', color: 'orange' },
          { icon: 'file', titleKey: 'dashboard:alert.orderRequest', descKey: 'common:order.heparinRequest', actionKey: 'common:action.review', actionRoute: '/dialysis-processing', color: 'blue' },
        ]} onNavigate={navigateTo} />
      case 'recent7dTreatments':
      case 'quality_stats':
      case 'nurse_workload':
      case 'treatmentTrend':
      case 'patientGrowth':
        return <ChartCard data={card.id === 'quality_stats' || card.id === 'patientGrowth' ? (dashboardStats?.qualityByHour ?? []) : (dashboardStats?.treatmentsByHour ?? [])} color={card.id === 'quality_stats' || card.id === 'patientGrowth' ? '#10b981' : '#3b82f6'} />
      case 'abnormalLabs':
        return <PlaceholderCard label={t('dashboard:card.abnormalLabs') || 'Abnormal Labs'} />
      case 'duty_monitor':
      case 'device_status_eng':
      case 'onlineDevices':
      case 'deviceUtilization':
        return <DeviceGridCard devices={equipments as unknown as { Id: string; Name?: string; IDNo?: string; Status?: string }[]} />
      case 'todayShiftMatrix':
      case 'staff_schedule':
        return <ShiftListCard shifts={shifts as unknown as { Id: string; Name?: string; StartTime?: string; EndTime?: string; Status?: string; Type?: string }[]} />
      case 'pendingOrders':
      case 'pendingPreAssessment':
      case 'pendingScheduleQueue':
      case 'bedUtilization':
      case 'contractAnomaly':
      case 'consumables_prep':
      case 'device_binding':
      case 'maintenance_logs':
      case 'schedule_adjust':
        return <PlaceholderCard label={t(`dashboard:card.${card.id}`) || card.id} />
      default:
        return <div className="text-foreground-muted text-sm flex items-center justify-center h-full">{t('loading')}</div>
    }
  }

  const hiddenCards = cardsConfig.filter(c => !c.visible)
  const visibleCards = cardsConfig.filter(c => c.visible)

  const getColSpanClass = (span: number) => {
    const map: Record<number, string> = { 3: 'md:col-span-3', 6: 'md:col-span-6', 12: 'md:col-span-12' }
    return map[span] || 'md:col-span-12'
  }

  const getCardTitle = (cardId: string): string => {
    const def = V2_CARD_REGISTRY.find(d => d.id === cardId)
    return def ? t(def.titleKey) : cardId
  }

  const CardWrapper: React.FC<{ config: LocalCardConfig; children: React.ReactNode }> = ({ config, children }) => (
    <div
      className={`bg-surface rounded-md border border-gray-100 flex flex-col overflow-hidden group relative
        col-span-1 ${getColSpanClass(config.colSpan)}
        ${isCustomizing ? 'ring-2 ring-blue-400 border-blue-400 z-10' : 'hover:shadow-md transition-shadow duration-200 cursor-pointer hover:border-blue-200'}
      `}
      style={{ gridRow: `span ${config.rowSpan}` }}
      onClick={() => { if (!isCustomizing) { const def = V2_CARD_REGISTRY.find(d => d.id === config.id); if (def) { const route = cardRouteMap[def.id]; if (route) navigate(route) } } }}
    >
      {isCustomizing && (
        <button onClick={(e) => { e.stopPropagation(); toggleVisibility(config.id, false) }} className="absolute top-2 right-2 z-20 p-1 bg-red-100 text-red-500 rounded-full hover:bg-red-200 transition-colors">
          <X size={14} />
        </button>
      )}
      <div className="px-4 py-3 border-b border-gray-50 flex justify-between items-center bg-surface select-none shrink-0 h-[52px]">
        <h3 className="font-bold text-foreground flex items-center gap-2 text-sm truncate">
          {config.type === 'stat' && <TrendingUp size={16} className="text-blue-500" />}
          {config.type === 'list' && <Users size={16} className="text-teal-500" />}
          {config.type === 'action' && <AlertTriangle size={16} className="text-orange-500" />}
          {config.type === 'monitor' && <Activity size={16} className="text-purple-500" />}
          {config.type === 'inventory' && <Package size={16} className="text-indigo-500" />}
          {config.type === 'binding' && <Link2 size={16} className="text-slate-500" />}
          {getCardTitle(config.id)}
        </h3>
        {!isCustomizing && (
          <button className="text-gray-400 hover:text-gray-600 opacity-0 group-hover:opacity-100 transition-opacity p-1">
            <MoreHorizontal size={16} />
          </button>
        )}
      </div>
      <div className={`p-4 flex-1 overflow-auto relative min-h-0 ${isCustomizing ? 'pointer-events-none' : ''}`}>
        {children}
      </div>
    </div>
  )

  const cardRouteMap: Record<string, string> = {
    dept_overview: '/statistics', operationKpis: '/statistics',
    active_patients: '/patients', myPatientsToday: '/patients', my_duty_patients: '/patients', prescription_adjust: '/patients',
    quality_stats: '/statistics', nurse_workload: '/statistics', patientGrowth: '/statistics', treatmentTrend: '/statistics',
    duty_monitor: '/monitoring', device_status_eng: '/monitoring', onlineDevices: '/monitoring',
    staff_schedule: '/schedule', todayShiftMatrix: '/schedule', schedule_adjust: '/schedule',
  }

  return (
    <div className="max-w-[1800px] mx-auto relative">
      {apiError && <div className="mb-4 px-4 py-2 bg-yellow-50 border border-yellow-200 rounded-lg text-yellow-700 text-sm">{apiError}</div>}

      <div className="flex justify-between items-center mb-6">
        <div>
          <h2 className="text-h2 font-bold text-foreground">{t('title')}</h2>
          <p className="text-foreground-muted text-sm mt-1">{isAdmin ? t('role:desc.admin') : isNurse ? t('role:desc.nurse') : t('role:desc.doctor')}</p>
        </div>
        <div className="flex space-x-3">
          {isCustomizing && (
            <button onClick={() => setShowWidgetLibrary(true)} className="flex items-center px-4 py-2 bg-white border border-blue-500 text-blue-600 rounded-md text-sm font-medium hover:bg-blue-50 transition-colors">
              <Plus size={16} className="mr-2" /> {t('action.addCard')}
            </button>
          )}
          <button onClick={() => { setIsCustomizing(!isCustomizing); setShowWidgetLibrary(false) }}
            className={`flex items-center px-4 py-2 rounded-md text-sm font-medium transition-colors ${isCustomizing ? 'bg-blue-600 text-white shadow-md' : 'bg-white text-gray-600 border border-gray-200 hover:bg-gray-50'}`}>
            {isCustomizing ? <><CheckCircle2 size={16} className="mr-2" /> {t('action.finishLayout')}</> : <><LayoutGrid size={16} className="mr-2" /> {t('action.customizeLayout')}</>}
          </button>
        </div>
      </div>

      <div className={`grid grid-cols-1 md:grid-cols-12 gap-4 pb-10 content-start auto-rows-[60px] ${isCustomizing ? 'select-none' : ''}`}>
        {visibleCards.map(card => (
          <CardWrapper key={card.id} config={card}>{renderCard(card)}</CardWrapper>
        ))}
      </div>

      {showWidgetLibrary && isCustomizing && (
        <div className="fixed top-16 right-0 bottom-0 w-80 bg-white shadow-2xl border-l border-gray-200 z-50 animate-slide-in-right flex flex-col">
          <div className="p-4 border-b border-gray-100 flex justify-between items-center bg-gray-50">
            <h3 className="font-bold text-gray-800 flex items-center"><Library size={18} className="mr-2 text-blue-600" /> {t('widgetLibrary.title')}</h3>
            <button onClick={() => setShowWidgetLibrary(false)} className="text-gray-400 hover:text-gray-600"><X size={20} /></button>
          </div>
          <div className="flex-1 overflow-y-auto p-4 space-y-4">
            <p className="text-xs text-gray-500 mb-2">{t('widgetLibrary.hint')}</p>
            {hiddenCards.length > 0 ? hiddenCards.map(card => (
              <div key={card.id} onClick={() => toggleVisibility(card.id, true)} className="p-3 border border-gray-200 rounded-md hover:border-blue-400 hover:bg-blue-50 cursor-pointer transition-all group">
                <div className="flex items-center justify-between mb-2">
                  <span className="font-bold text-gray-700 group-hover:text-blue-700 text-sm">{getCardTitle(card.id)}</span>
                  <Plus size={16} className="text-gray-400 group-hover:text-blue-600" />
                </div>
                <div className="text-xs text-gray-400 flex items-center">
                  <span className="bg-gray-100 px-1.5 py-0.5 rounded mr-2">{card.type}</span>
                  <span>{t('widgetLibrary.defaultSize')}: {card.size}</span>
                </div>
              </div>
            )) : (
              <div className="text-center py-10 text-gray-400 text-sm">
                <CheckCircle2 size={32} className="mx-auto mb-2 opacity-50" /><p>{t('widgetLibrary.allAdded')}</p>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
