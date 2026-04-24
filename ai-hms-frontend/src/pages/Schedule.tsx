import React, { useState, useMemo, useCallback, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import type { PatientScheduleItem, StaffMember } from '../types/original'
import { restApi, convertRestPatientList } from '@/services'
import type { RestShift, RestPatientShift } from '@/services'
import type { Patient } from '@/types/original'
import { getSelectedRoleUser } from '@/services/role'
import {
  Calendar as CalendarIcon, ChevronLeft, ChevronRight,
  Plus, Download, Users, Moon, Sun, Sunset,
  X, Search, CheckCircle2,
  BedDouble, ChevronDown,
  UserCog, Settings, Timer, Zap, Copy, ArrowRightLeft
} from 'lucide-react'

type TabType = 'PATIENT' | 'STAFF'

// 扩展班次定义，支持自定义岗位和时间
interface ShiftTemplate {
  id: string
  name: string
  timeRange: string
  color: string
  category: 'RESPONSIBLE' | 'SUPERVISOR' | 'MANAGER'
}

interface SchedulingModalData {
  bedNumber?: string
  staffId?: string
  date: string
  shift?: 'Morning' | 'Afternoon' | 'Evening'
}

// 预设岗位模板
const DEFAULT_TEMPLATES: ShiftTemplate[] = [
  { id: 't1', name: '总管班', timeRange: '08:00-17:00', color: 'blue', category: 'MANAGER' },
  { id: 't2', name: '主管班', timeRange: '08:00-12:00', color: 'indigo', category: 'SUPERVISOR' },
  { id: 't3', name: '责任-A1', timeRange: '08:00-12:00', color: 'emerald', category: 'RESPONSIBLE' },
  { id: 't4', name: '责任-A2', timeRange: '13:00-17:00', color: 'emerald', category: 'RESPONSIBLE' },
  { id: 't5', name: '责任-B', timeRange: '08:00-12:00', color: 'teal', category: 'RESPONSIBLE' },
]

// 患者班次卡片：极致紧凑型
const PatientShiftCard = React.memo(({
  item,
  shiftType,
  bedNumber,
  date,
  onOpen
}: {
  item?: PatientScheduleItem
  shiftType: 'Morning' | 'Afternoon' | 'Evening'
  bedNumber: string
  date: string
  onOpen: (bed: string, date: string, shift: 'Morning' | 'Afternoon' | 'Evening', pid?: string) => void
}) => {
  const styles = {
    'Morning': {
      bg: 'bg-blue-50/50',
      text: 'text-blue-700',
      border: 'border-blue-100',
      icon: <Sun size={12} className="text-blue-500 shrink-0"/>,
      pill: 'bg-blue-100 text-blue-600'
    },
    'Afternoon': {
      bg: 'bg-amber-50/50',
      text: 'text-amber-700',
      border: 'border-amber-100',
      icon: <Sunset size={12} className="text-orange-500 shrink-0"/>,
      pill: 'bg-orange-100 text-orange-600'
    },
    'Evening': {
      bg: 'bg-indigo-50/50',
      text: 'text-indigo-700',
      border: 'border-indigo-100',
      icon: <Moon size={12} className="text-indigo-500 shrink-0"/>,
      pill: 'bg-indigo-100 text-indigo-600'
    }
  }[shiftType]

  if (!item) {
    return (
      <div
        onClick={() => onOpen(bedNumber, date, shiftType)}
        className="h-[32px] w-10 flex items-center justify-center group/add cursor-pointer rounded-lg border border-dashed border-slate-100 hover:bg-slate-50 hover:border-blue-300 transition-all"
      >
        <Plus size={12} className="text-slate-200 group-hover/add:text-blue-400 font-bold"/>
      </div>
    )
  }

  return (
    <div
      onClick={() => {
        onOpen(bedNumber, date, shiftType, item.patientId)
      }}
      className={`inline-flex items-center justify-start gap-1.5 px-2 h-[32px] rounded-lg border shadow-sm transition-all hover:shadow-md cursor-pointer group/item ${styles.bg} ${styles.border} w-fit max-w-full overflow-hidden`}
    >
      <div className="flex items-center shrink-0">
        <span className="mr-1 shrink-0">{styles.icon}</span>
        <span className="font-bold text-[12px] text-slate-700 whitespace-nowrap tracking-tight">{item.patientName}</span>
      </div>
      <span className={`text-[8px] font-black px-1 py-0.5 rounded-md shrink-0 ${styles.pill}`}>
        {item.mode}
      </span>
    </div>
  )
})

// 护士岗位班次卡片：显示自定义岗位和时间
const StaffShiftCard = React.memo(({
  shift,
  onOpen,
  staffId,
  date
}: {
  shift?: {
    label?: string
    timeRange?: string
    category?: string
    type?: string
  }
  staffId: string
  date: string
  onOpen: (sid: string, date: string) => void
}) => {
  if (!shift || shift.type === 'OFF') return (
    <div
      onClick={() => onOpen(staffId, date)}
      className="h-full w-full min-h-[56px] flex items-center justify-center group cursor-pointer hover:bg-slate-50 rounded-xl transition-all border border-dashed border-transparent hover:border-slate-200"
    >
      <Plus size={14} className="text-slate-100 group-hover:text-blue-400"/>
    </div>
  )

  // 根据类别定义样式
  const getStyle = (category: string) => {
    switch(category) {
      case 'MANAGER': return 'bg-blue-600 text-white border-blue-700'
      case 'SUPERVISOR': return 'bg-indigo-500 text-white border-indigo-600'
      case 'RESPONSIBLE': return 'bg-emerald-50 text-emerald-700 border-emerald-100'
      default: return 'bg-slate-100 text-slate-600 border-slate-200'
    }
  }

  const styleClass = getStyle(shift.category || '')

  return (
    <div
      onClick={() => onOpen(staffId, date)}
      className={`h-full w-full min-h-[56px] flex flex-col items-center justify-center rounded-xl border shadow-sm cursor-pointer hover:shadow-md transition-all p-2 ${styleClass}`}
    >
      <span className="text-[12px] font-black tracking-tight mb-0.5">{shift.label}</span>
      <div className={`flex items-center gap-1 text-[9px] font-bold opacity-80 ${shift.category === 'RESPONSIBLE' ? 'text-emerald-500' : 'text-white/80'}`}>
        <Timer size={10}/>
        {shift.timeRange}
      </div>
    </div>
  )
})

export default function Schedule() {
  const { t } = useTranslation(['schedule'])
  // 获取当前用户角色
  const roleUser = useMemo(() => getSelectedRoleUser(), [])
  const isNurse = roleUser?.role?.includes('NURSE') ?? false

  const [activeTab, setActiveTab] = useState<TabType>('PATIENT')
  const [viewStartDate, setViewStartDate] = useState(new Date())
  const [selectedZone, setSelectedZone] = useState<string>('A')

  // REST API 数据
  const [shifts, setShifts] = useState<RestShift[]>([])
  const [patients, setPatients] = useState<Partial<Patient>[]>([])
  const [searchResults, setSearchResults] = useState<Partial<Patient>[]>([])
  const [staff, setStaff] = useState<StaffMember[]>([])

  const [patientSchedule, setPatientSchedule] = useState<PatientScheduleItem[]>([])
  const [staffSchedule, setStaffSchedule] = useState<Array<{
    staffId: string
    date: string
    type: string
    area: string
    label?: string
    timeRange?: string
    category?: string
  }>>( [])

  const [shiftTemplates, setShiftTemplates] = useState<ShiftTemplate[]>(DEFAULT_TEMPLATES)
  const [isTemplateModalOpen, setIsTemplateModalOpen] = useState(false)
  const [newTemplate, setNewTemplate] = useState<Partial<ShiftTemplate>>({ category: 'RESPONSIBLE', color: 'emerald' })

  // 快速排班状态
  const [isQuickScheduleModalOpen, setIsQuickScheduleModalOpen] = useState(false)
  const [quickTargetStaffId, setQuickTargetStaffId] = useState<string>('')
  const [quickSourceType, setQuickSourceType] = useState<'SELF' | 'OTHER'>('SELF')
  const [quickSourceStaffId, setQuickSourceStaffId] = useState<string>('')

  const scheduleMap = useMemo(() => {
    const map = new Map<string, PatientScheduleItem>()
    patientSchedule.forEach(item => {
      map.set(`${item.bedNumber}-${item.date}-${item.shift}`, item)
    })
    return map
  }, [patientSchedule])

  const staffMap = useMemo(() => {
    const map = new Map<string, typeof staffSchedule[0]>()
    staffSchedule.forEach(s => {
      map.set(`${s.staffId}-${s.date}`, s)
    })
    return map
  }, [staffSchedule])

  const [isScheduling, setIsScheduling] = useState(false)
  const [isStaffScheduling, setIsStaffScheduling] = useState(false)
  const [schedulingData, setSchedulingData] = useState<SchedulingModalData | null>(null)

  const [searchPatientQuery, setSearchPatientQuery] = useState('')
  const [showSearchDropdown, setShowSearchDropdown] = useState(false)
  const [selectedPatientId, setSelectedPatientId] = useState<string | null>(null)
  const [selectedTreatmentMode, setSelectedTreatmentMode] = useState('HD')

  // 生成展示的7天日期（从 viewStartDate 开始）
  const dates = useMemo(() => {
    const res = []
    for (let i = 0; i < 7; i++) {
      const d = new Date(viewStartDate)
      d.setDate(viewStartDate.getDate() + i)
      res.push(d)
    }
    return res
  }, [viewStartDate])

  const dateStrings = useMemo(() => dates.map(d => d.toISOString().split('T')[0]), [dates])

  // 从后端加载班次定义
  useEffect(() => {
    restApi.getShifts()
      .then(res => setShifts(res.data))
      .catch(err => console.error('加载班次失败:', err))
  }, [])

  // 将后端班次名称映射到前端 shift 类型
  const shiftNameToType = useCallback((shiftName: string): 'Morning' | 'Afternoon' | 'Evening' => {
    const name = shiftName.toLowerCase()
    if (name.includes('早') || name.includes('上午') || name.includes('morning')) return 'Morning'
    if (name.includes('晚') || name.includes('夜') || name.includes('evening')) return 'Evening'
    return 'Afternoon' // 默认中班
  }, [])

  // 从后端加载患者排班（按当前视图 7 天）
  useEffect(() => {
    if (dateStrings.length === 0) return
    const startDate = dateStrings[0]
    const endDate = dateStrings[dateStrings.length - 1]

    restApi.getPatientShifts({ startDate, endDate, page: 1, pageSize: 500 })
      .then(res => {
        const items: PatientScheduleItem[] = res.data.items.map((ps: RestPatientShift) => ({
          id: String(ps.id),
          bedNumber: ps.bed?.name || `B${ps.bedId || 0}`,
          date: ps.scheduleDate.split('T')[0],
          shift: ps.shift ? shiftNameToType(ps.shift.name) : 'Morning',
          patientName: ps.patient?.name || `患者${ps.patientId}`,
          mode: (ps.patient as unknown as Record<string, unknown>)?.defaultMode as string || 'HD',
          patientId: ps.patient?.id ? String(ps.patient.id) : String(ps.patientId)
        }))
        setPatientSchedule(items)
      })
      .catch(err => console.error('加载患者排班失败:', err))
  }, [dateStrings, shiftNameToType])

  // 加载患者列表（用于搜索）
  useEffect(() => {
    restApi.getPatientList({ page: 1, pageSize: 200 })
      .then(res => setPatients(convertRestPatientList(res.data.items)))
      .catch(err => console.error('加载患者列表失败:', err))
  }, [])

  // 加载护理人员列表
  useEffect(() => {
    restApi.getUserList({ status: 'active' })
      .then(users => {
        const mapped: StaffMember[] = users.map(u => ({
          id: u.id,
          name: u.realName || u.username,
          role: u.role,
          level: u.role,
        }))
        setStaff(mapped)
        if (mapped.length > 0) {
          setQuickTargetStaffId(mapped[0].id)
          setQuickSourceStaffId(mapped[0].id)
        }
      })
      .catch((err) => console.error('[Schedule] 患者搜索加载失败', err))
  }, [])

  // 患者搜索防抖：输入 300ms 后查询后端
  useEffect(() => {
    if (!searchPatientQuery.trim()) {
      setSearchResults([])
      return
    }
    const timer = setTimeout(() => {
      // 先从已加载的患者列表中本地搜索
      const local = patients.filter(p =>
        p.name?.includes(searchPatientQuery) ||
        p.id?.includes(searchPatientQuery) ||
        p.bedNumber?.includes(searchPatientQuery)
      )
      if (local.length > 0) {
        setSearchResults(local)
      } else {
        // 本地无结果时请求后端
        restApi.getPatientList({ name: searchPatientQuery, page: 1, pageSize: 20 })
          .then(res => setSearchResults(convertRestPatientList(res.data.items)))
          .catch(() => setSearchResults([]))
      }
    }, 300)
    return () => clearTimeout(timer)
  }, [searchPatientQuery, patients])

  const filteredBeds = useMemo(() => {
    const counts: Record<string, number> = { 'A': 16, 'B': 5, 'C': 4 }
    return Array.from({ length: counts[selectedZone] }).map((_, i) => `${selectedZone}${String(i + 1).padStart(2, '0')}`)
  }, [selectedZone])

  const handleOpenScheduling = useCallback((bedNumber: string, date: string, shift: 'Morning' | 'Afternoon' | 'Evening', pid?: string) => {
    setSchedulingData({ bedNumber, date, shift })
    if (pid) {
      const p = patients.find(x => x.id === pid)
      if (p) {
        setSelectedPatientId(p.id!)
        setSearchPatientQuery(p.name || '')
        setSelectedTreatmentMode(p.defaultMode || 'HD')
      }
    } else {
      setSearchPatientQuery('')
      setSelectedPatientId(null)
    }
    setIsScheduling(true)
  }, [patients])

  const handleOpenStaffScheduling = (staffId: string, date: string) => {
    setSchedulingData({ staffId, date })
    setIsStaffScheduling(true)
  }

  const handleAddTemplate = () => {
    if (!newTemplate.name || !newTemplate.timeRange) return
    const template: ShiftTemplate = {
      id: `t-${Date.now()}`,
      name: newTemplate.name,
      timeRange: newTemplate.timeRange,
      color: newTemplate.color || 'blue',
      category: newTemplate.category || 'RESPONSIBLE'
    }
    setShiftTemplates(prev => [...prev, template])
    setNewTemplate({ category: 'RESPONSIBLE', color: 'emerald' })
    setIsTemplateModalOpen(false)
  }

  // 执行快速排班：复制逻辑
  const handleApplyQuickSchedule = () => {
    const sourceId = quickSourceType === 'SELF' ? quickTargetStaffId : quickSourceStaffId

    // 获取当前视图 7 天的日期
    const currentWeekDates = dateStrings

    // 获取"上周"对应的日期（当前日期减 7 天）
    const lastWeekDates = currentWeekDates.map(d => {
      const dateObj = new Date(d)
      dateObj.setDate(dateObj.getDate() - 7)
      return dateObj.toISOString().split('T')[0]
    })

    const newEntries: typeof staffSchedule = []

    currentWeekDates.forEach((currentDate, index) => {
      const lastWeekDate = lastWeekDates[index]
      const sourceShift = staffMap.get(`${sourceId}-${lastWeekDate}`)

      if (sourceShift) {
        newEntries.push({
          ...sourceShift,
          staffId: quickTargetStaffId,
          date: currentDate
        })
      }
    })

    if (newEntries.length > 0) {
      setStaffSchedule(prev => {
        // 移除目标人员在当前视图日期内的旧排班
        const filtered = prev.filter(s => !(s.staffId === quickTargetStaffId && currentWeekDates.includes(s.date)))
        return [...filtered, ...newEntries]
      })
      alert(t('schedule:alert.quickScheduleSuccess', { name: staff.find(s => s.id === quickTargetStaffId)?.name }))
    } else {
      alert(t('schedule:alert.noScheduleFound'))
    }

    setIsQuickScheduleModalOpen(false)
  }

  return (
    <div className="h-full flex flex-col bg-white rounded-3xl shadow-xl border border-slate-200 overflow-hidden relative">
      {/* 顶部主控制栏 */}
      <div className="px-8 py-5 border-b border-slate-100 flex items-center justify-between bg-white shrink-0 z-30">
        <div className="flex items-center">
          <div className="p-3 bg-blue-600 rounded-2xl text-white shadow-lg shadow-blue-200 mr-4">
            <CalendarIcon size={24} />
          </div>
          <div className="mr-12">
            <h2 className="text-xl font-black text-slate-800 tracking-tight">
              {activeTab === 'PATIENT' ? t('schedule:title.patientSchedule') : t('schedule:title.staffSchedule')}
            </h2>
            <p className="text-[10px] text-slate-300 font-bold uppercase tracking-widest leading-none mt-1">{t('schedule:subtitle')}</p>
          </div>

          <div className="flex items-center bg-slate-50 rounded-2xl border border-slate-200 p-1 shadow-inner shrink-0">
            <button onClick={() => setViewStartDate(prev => { const d = new Date(prev); d.setDate(d.getDate() - 1); return d })} className="p-2 hover:bg-white hover:shadow-sm rounded-xl text-slate-400 hover:text-slate-900 transition-all"><ChevronLeft size={16}/></button>
            <span className="px-6 text-[15px] font-black text-slate-700 whitespace-nowrap">{t('schedule:nav.year', { year: dates[0].getFullYear() })}{t('schedule:nav.month', { month: dates[0].getMonth() + 1 })}</span>
            <button onClick={() => setViewStartDate(prev => { const d = new Date(prev); d.setDate(d.getDate() + 1); return d })} className="p-2 hover:bg-white hover:shadow-sm rounded-xl text-slate-400 hover:text-slate-900 transition-all"><ChevronRight size={16}/></button>
            <button onClick={() => setViewStartDate(new Date())} className="ml-2 px-3 py-1 bg-white border border-slate-200 rounded-lg text-[10px] font-bold text-blue-600 hover:bg-blue-50 shadow-sm transition-all">{t('schedule:nav.backToToday')}</button>
          </div>
        </div>

        <div className="flex items-center gap-4">
          <div className="flex bg-slate-100/80 p-1.5 rounded-2xl shadow-inner">
            <button
              onClick={() => setActiveTab('PATIENT')}
              className={`flex items-center px-6 py-2.5 rounded-xl text-sm font-black transition-all ${activeTab === 'PATIENT' ? 'bg-white text-blue-600 shadow-md ring-1 ring-slate-200' : 'text-slate-500 hover:text-slate-700'}`}
            >
              <BedDouble size={14} className="mr-2"/> {t('schedule:tab.patient')}
            </button>
            {isNurse && (
              <button
                onClick={() => setActiveTab('STAFF')}
                className={`flex items-center px-6 py-2.5 rounded-xl text-sm font-black transition-all ${activeTab === 'STAFF' ? 'bg-white text-blue-600 shadow-md ring-1 ring-slate-200' : 'text-slate-500 hover:text-slate-700'}`}
              >
                <UserCog size={14} className="mr-2"/> {t('schedule:tab.staff')}
              </button>
            )}
          </div>

          {activeTab === 'STAFF' && (
            <button
              onClick={() => setIsTemplateModalOpen(true)}
              className="flex items-center gap-2 px-5 py-2.5 bg-white border border-slate-200 rounded-2xl text-sm font-black text-slate-600 hover:bg-slate-50 transition-all shadow-sm"
            >
              <Settings size={16} className="text-blue-500"/> {t('schedule:action.positionDef')}
            </button>
          )}

          <button
            onClick={() => {
              if (activeTab === 'STAFF') setIsQuickScheduleModalOpen(true)
              else alert(t('schedule:alert.patientNoQuickSchedule'))
            }}
            className="px-6 py-3 bg-blue-600 text-white rounded-2xl text-sm font-black hover:bg-blue-700 shadow-xl shadow-blue-100 flex items-center transition-all"
          >
            <Plus size={18} className="mr-2 stroke-[3px]"/> {t('schedule:action.quickSchedule')}
          </button>
        </div>
      </div>

      {/* 辅助筛选栏 */}
      <div className="px-8 py-3 border-b border-slate-100 flex items-center justify-between bg-white shrink-0 z-20">
        <div className="flex items-center gap-8 shrink-0">
          {activeTab === 'PATIENT' ? (
            <>
              <span className="text-[11px] font-black text-slate-400 uppercase tracking-widest whitespace-nowrap">{t('schedule:filter.wardLabel')}</span>
              <div className="flex items-center gap-2">
                {[{ id: 'A', label: t('schedule:filter.wardA', { count: 16 }) }, { id: 'B', label: t('schedule:filter.wardB', { count: 5 }) }, { id: 'C', label: t('schedule:filter.wardC', { count: 4 }) }].map(z => (
                  <button
                    key={z.id}
                    onClick={() => setSelectedZone(z.id)}
                    className={`px-5 py-2 rounded-xl border transition-all font-black text-[12px] ${selectedZone === z.id ? 'bg-blue-50 border-blue-500 text-blue-700 shadow-sm' : 'border-slate-100 text-slate-400 hover:border-slate-200 bg-white'}`}
                  >
                    {z.label}
                  </button>
                ))}
              </div>
              <div className="h-6 w-px bg-slate-100 mx-2"></div>
              <div className="flex items-center gap-6">
                {[{id:'m', label: t('schedule:shift.morning'), color:'blue'}, {id:'a', label: t('schedule:shift.afternoon'), color:'orange'}, {id:'e', label: t('schedule:shift.evening'), color:'indigo'}].map(s => (
                  <label key={s.id} className="flex items-center gap-2.5 cursor-pointer group">
                    <input type="checkbox" defaultChecked className="w-5 h-5 rounded border-slate-200 text-blue-500 focus:ring-blue-400 transition-all"/>
                    <span className="text-[12px] font-black text-slate-500 group-hover:text-slate-800 transition-colors">{s.label}</span>
                  </label>
                ))}
              </div>
            </>
          ) : (
            <>
              <span className="text-[11px] font-black text-slate-400 uppercase tracking-widest whitespace-nowrap">{t('schedule:filter.quickView')}</span>
              <div className="flex items-center gap-2">
                {[t('schedule:filter.selectAll'), t('schedule:filter.headNurse'), t('schedule:filter.responsibleNurse'), t('schedule:filter.mainNurse')].map(label => (
                  <button key={label} className="px-4 py-2 border border-slate-100 rounded-xl text-xs font-bold text-slate-500 hover:bg-slate-50">{label}</button>
                ))}
              </div>
              <div className="h-6 w-px bg-slate-100 mx-2"></div>
              <div className="flex items-center gap-4">
                <span className="text-[11px] font-black text-slate-400 uppercase tracking-widest">{t('schedule:filter.positionLegend')}</span>
                <div className="flex items-center gap-4">
                  <div className="flex items-center gap-1.5"><div className="w-3 h-3 rounded-full bg-blue-600"></div><span className="text-[11px] font-bold text-slate-500">{t('schedule:legend.manager')}</span></div>
                  <div className="flex items-center gap-1.5"><div className="w-3 h-3 rounded-full bg-indigo-500"></div><span className="text-[11px] font-bold text-slate-500">{t('schedule:legend.supervisor')}</span></div>
                  <div className="flex items-center gap-1.5"><div className="w-3 h-3 rounded-full bg-emerald-400"></div><span className="text-[11px] font-bold text-slate-500">{t('schedule:legend.responsible')}</span></div>
                </div>
              </div>
            </>
          )}
        </div>

        <button className="flex items-center gap-2 text-[12px] font-black text-blue-600 hover:bg-blue-50 px-4 py-2 rounded-xl transition-all">
          <Download size={16}/> {t('schedule:action.exportSchedule')}
        </button>
      </div>

      {/* 排班表格主体 - 支持左右平铺横向滑动 */}
      <div className="flex-1 overflow-x-auto overflow-y-auto bg-white no-scrollbar relative">
        {activeTab === 'PATIENT' ? (
          <table className="min-w-max border-separate border-spacing-0 table-fixed">
            <thead>
              <tr>
                <th className="sticky left-0 top-0 z-50 bg-white border-b border-r border-slate-100 p-4 w-[100px] text-center text-[11px] font-black text-slate-400 uppercase tracking-widest shadow-[4px_0_12px_-6px_rgba(0,0,0,0.1)]">{t('schedule:table.bed')}</th>
                {dates.map((date, i) => (
                  <th key={i} colSpan={3} className="sticky top-0 z-40 bg-white border-b border-r border-slate-100 p-4 text-center min-w-[360px]">
                    <div className="flex items-center justify-center gap-4">
                      <span className="text-[14px] font-black text-slate-800">{[t('schedule:day.sun'),t('schedule:day.mon'),t('schedule:day.tue'),t('schedule:day.wed'),t('schedule:day.thu'),t('schedule:day.fri'),t('schedule:day.sat')][date.getDay()]}</span>
                      <span className={`inline-block px-4 py-1 rounded-xl text-[12px] font-black ${date.toDateString() === new Date().toDateString() ? 'bg-blue-600 text-white shadow-lg ring-4 ring-blue-100' : 'bg-slate-100 text-slate-400'}`}>
                        {t('schedule:date.format', { month: date.getMonth() + 1, day: date.getDate() })}
                        {date.toDateString() === new Date().toDateString() && <span className="ml-1 text-[9px] opacity-70">{t('schedule:day.today')}</span>}
                      </span>
                    </div>
                  </th>
                ))}
              </tr>
              <tr className="z-30">
                <th className="sticky left-0 top-[58px] z-50 bg-white border-b border-r border-slate-100 p-0 w-[100px] shadow-[4px_0_12px_-6px_rgba(0,0,0,0.1)]"></th>
                {dates.map((_, i) => (
                  <React.Fragment key={i}>
                    <th className="sticky top-[58px] z-20 bg-slate-50/50 backdrop-blur-sm border-b border-r border-slate-50 p-2 text-center">
                      <span className="text-[11px] font-black text-blue-400/80 uppercase tracking-tighter">{t('schedule:shift.morning')}</span>
                    </th>
                    <th className="sticky top-[58px] z-20 bg-slate-50/50 backdrop-blur-sm border-b border-r border-slate-50 p-2 text-center">
                      <span className="text-[11px] font-black text-orange-400/80 uppercase tracking-tighter">{t('schedule:shift.afternoon')}</span>
                    </th>
                    <th className="sticky top-[58px] z-20 bg-slate-50/50 backdrop-blur-sm border-b border-r border-slate-100 p-2 text-center">
                      <span className="text-[11px] font-black text-indigo-400/80 uppercase tracking-tighter">{t('schedule:shift.evening')}</span>
                    </th>
                  </React.Fragment>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-50">
              {filteredBeds.map(bed => (
                <tr key={bed} className="group">
                  <td className="sticky left-0 z-40 bg-white group-hover:bg-blue-50/20 border-r border-slate-100 p-4 shadow-[4px_0_12px_-6px_rgba(0,0,0,0.1)] transition-colors">
                    <div className="flex items-center justify-center">
                      <div className="w-14 h-10 rounded-xl flex items-center justify-center font-black text-sm border-2 bg-blue-50 text-blue-600 border-blue-100 shadow-sm group-hover:scale-105 transition-all">
                        {bed}
                      </div>
                    </div>
                  </td>
                  {dateStrings.map((dateStr, i) => (
                    <React.Fragment key={i}>
                      <td className="border-r border-slate-50 p-1.5 align-middle"><PatientShiftCard item={scheduleMap.get(`${bed}-${dateStr}-Morning`)} shiftType="Morning" bedNumber={bed} date={dateStr} onOpen={handleOpenScheduling} /></td>
                      <td className="border-r border-slate-50 p-1.5 align-middle"><PatientShiftCard item={scheduleMap.get(`${bed}-${dateStr}-Afternoon`)} shiftType="Afternoon" bedNumber={bed} date={dateStr} onOpen={handleOpenScheduling} /></td>
                      <td className="border-r border-slate-100 p-1.5 align-middle"><PatientShiftCard item={scheduleMap.get(`${bed}-${dateStr}-Evening`)} shiftType="Evening" bedNumber={bed} date={dateStr} onOpen={handleOpenScheduling} /></td>
                    </React.Fragment>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        ) : (
          <table className="min-w-max border-separate border-spacing-0 table-fixed">
            <thead>
              <tr>
                <th className="sticky left-0 top-0 z-50 bg-white border-b border-r border-slate-100 p-6 w-[200px] text-left shadow-[4px_0_12px_-6px_rgba(0,0,0,0.1)]">
                  <div className="flex items-center text-[11px] font-black text-slate-400 uppercase tracking-[0.2em]"><Users size={14} className="mr-2"/> {t('schedule:table.staff')}</div>
                </th>
                {dates.map((date, i) => (
                  <th key={i} className="sticky top-0 z-40 bg-white border-b border-r border-slate-100 p-4 text-center min-w-[160px]">
                    <div className="flex flex-col items-center">
                      <span className="text-[13px] font-black text-slate-800 mb-1">{[t('schedule:day.sun'),t('schedule:day.mon'),t('schedule:day.tue'),t('schedule:day.wed'),t('schedule:day.thu'),t('schedule:day.fri'),t('schedule:day.sat')][date.getDay()]}</span>
                      <span className={`inline-block px-3 py-0.5 rounded-lg text-[11px] font-black ${date.toDateString() === new Date().toDateString() ? 'bg-blue-600 text-white shadow-lg' : 'bg-slate-100 text-slate-400'}`}>
                        {t('schedule:date.format', { month: date.getMonth() + 1, day: date.getDate() })}
                      </span>
                    </div>
                  </th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-50">
              {staff.map(staff => (
                <tr key={staff.id} className="group">
                  <td className="sticky left-0 z-40 bg-white group-hover:bg-slate-50 border-r border-slate-100 p-4 shadow-[4px_0_12px_-6px_rgba(0,0,0,0.1)] transition-colors">
                    <div className="flex items-center gap-4">
                      <div className="w-10 h-10 rounded-2xl bg-slate-100 flex items-center justify-center font-black text-slate-400 border border-slate-200 shrink-0">{staff.name[0]}</div>
                      <div className="truncate">
                        <p className="text-sm font-black text-slate-800 truncate">{staff.name}</p>
                        <p className="text-[10px] text-slate-400 font-bold uppercase tracking-tighter">{staff.role} · {staff.level}</p>
                      </div>
                    </div>
                  </td>
                  {dateStrings.map((dateStr, i) => (
                    <td key={i} className="border-r border-slate-100 p-2 h-[84px] align-middle">
                      <StaffShiftCard
                        shift={staffMap.get(`${staff.id}-${dateStr}`)}
                        staffId={staff.id}
                        date={dateStr}
                        onOpen={handleOpenStaffScheduling}
                      />
                    </td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {/* 岗位定义弹窗 */}
      {isTemplateModalOpen && (
        <div className="fixed inset-0 z-[110] flex items-center justify-center bg-black/50 backdrop-blur-md animate-fade-in p-4">
          <div className="bg-white rounded-[32px] shadow-2xl w-full max-w-lg overflow-hidden animate-scale-in flex flex-col">
            <div className="px-8 py-6 border-b border-slate-100 flex justify-between items-center bg-slate-50">
              <h3 className="text-xl font-black text-slate-900 flex items-center gap-2"><Settings size={20} className="text-blue-600"/> {t('schedule:modal.positionSettings')}</h3>
              <button onClick={() => setIsTemplateModalOpen(false)} className="p-2 text-slate-400 hover:text-slate-600 rounded-xl"><X size={20}/></button>
            </div>
            <div className="p-8 space-y-6">
              <div className="space-y-4">
                <p className="text-[11px] font-black text-slate-400 uppercase tracking-widest">{t('schedule:modal.addEditPosition')}</p>
                <div className="grid grid-cols-1 gap-5">
                  <div>
                    <label className="text-xs font-bold text-slate-500 mb-1.5 block">{t('schedule:modal.positionName')}</label>
                    <input
                      type="text"
                      className="w-full h-11 border border-slate-200 rounded-xl px-4 text-sm font-bold outline-none focus:ring-2 focus:ring-blue-500/10 focus:border-blue-500 transition-all bg-slate-50 focus:bg-white"
                      placeholder={t('schedule:modal.positionNamePlaceholder')}
                      value={newTemplate.name || ''}
                      onChange={e => setNewTemplate(prev => ({...prev, name: e.target.value}))}
                    />
                  </div>
                  <div>
                    <label className="text-xs font-bold text-slate-500 mb-1.5 block">{t('schedule:modal.positionTime')}</label>
                    <input
                      type="text"
                      className="w-full h-11 border border-slate-200 rounded-xl px-4 text-sm font-bold outline-none focus:ring-2 focus:ring-blue-500/10 focus:border-blue-500 transition-all bg-slate-50 focus:bg-white"
                      placeholder={t('schedule:modal.positionTimePlaceholder')}
                      value={newTemplate.timeRange || ''}
                      onChange={e => setNewTemplate(prev => ({...prev, timeRange: e.target.value}))}
                    />
                  </div>
                  <div>
                    <label className="text-xs font-bold text-slate-500 mb-1.5 block">{t('schedule:modal.category')}</label>
                    <div className="grid grid-cols-3 gap-2">
                      {(['RESPONSIBLE', 'SUPERVISOR', 'MANAGER'] as const).map(cat => (
                        <button
                          key={cat}
                          onClick={() => setNewTemplate(prev => ({...prev, category: cat}))}
                          className={`py-2.5 rounded-xl border text-[11px] font-black transition-all ${newTemplate.category === cat ? 'bg-blue-600 border-blue-600 text-white shadow-md' : 'border-slate-100 text-slate-400 hover:bg-slate-50'}`}
                        >
                          {cat === 'RESPONSIBLE' ? t('schedule:modal.categoryResponsible') : (cat === 'SUPERVISOR' ? t('schedule:modal.categorySupervisor') : t('schedule:modal.categoryManager'))}
                        </button>
                      ))}
                    </div>
                  </div>
                </div>
              </div>

              <div className="pt-6 border-t border-slate-100">
                <p className="text-[11px] font-black text-slate-400 uppercase tracking-widest mb-4">{t('schedule:modal.currentPositions')}</p>
                <div className="flex flex-wrap gap-2 max-h-40 overflow-y-auto no-scrollbar p-1">
                  {shiftTemplates.map(tmpl => (
                    <div key={tmpl.id} className="flex items-center gap-2 px-3 py-1.5 bg-slate-50 border border-slate-100 rounded-lg group">
                      <span className="text-[11px] font-black text-slate-700">{tmpl.name}</span>
                      <span className="text-[9px] font-bold text-slate-400">({tmpl.timeRange})</span>
                      <button onClick={() => setShiftTemplates(prev => prev.filter(x => x.id !== tmpl.id))} className="text-slate-300 hover:text-red-500 transition-colors opacity-0 group-hover:opacity-100"><X size={10}/></button>
                    </div>
                  ))}
                </div>
              </div>
            </div>
            <div className="px-8 py-6 bg-slate-50 border-t border-slate-100 flex justify-end gap-3">
              <button onClick={() => setIsTemplateModalOpen(false)} className="px-6 py-2.5 bg-white border border-slate-200 text-sm font-black text-slate-500 rounded-xl hover:bg-slate-100 transition-all">{t('schedule:action.cancel')}</button>
              <button onClick={handleAddTemplate} className="px-10 py-2.5 bg-blue-600 text-white text-sm font-black rounded-xl shadow-lg shadow-blue-200 hover:bg-blue-700 transition-all">{t('schedule:action.confirmAdd')}</button>
            </div>
          </div>
        </div>
      )}

      {/* 快速排班弹窗 */}
      {isQuickScheduleModalOpen && (
        <div className="fixed inset-0 z-[110] flex items-center justify-center bg-black/50 backdrop-blur-md animate-fade-in p-4">
          <div className="bg-white rounded-[40px] shadow-2xl w-full max-w-2xl overflow-hidden animate-scale-in flex flex-col ring-1 ring-black/5">
            <div className="px-10 py-8 bg-blue-600 border-b border-blue-700 flex justify-between items-center text-white">
              <div className="flex items-center gap-4">
                <div className="p-3 bg-white/20 rounded-2xl">
                  <Zap size={24}/>
                </div>
                <div>
                  <h3 className="text-2xl font-black">{t('schedule:quick.title')}</h3>
                  <p className="text-sm opacity-80 font-bold">{t('schedule:quick.subtitle')}</p>
                </div>
              </div>
              <button onClick={() => setIsQuickScheduleModalOpen(false)} className="p-2 hover:bg-white/10 rounded-xl transition-all"><X size={24}/></button>
            </div>

            <div className="p-10 space-y-10">
              {/* 步骤一：选择目标护士 */}
              <div className="space-y-4">
                <label className="text-[11px] font-black text-slate-400 uppercase tracking-widest ml-1 flex items-center gap-2">
                  <div className="w-5 h-5 rounded-full bg-slate-100 flex items-center justify-center text-[10px] text-slate-500">1</div>
                  {t('schedule:quick.targetNurse')}
                </label>
                <div className="relative">
                  <select
                    className="w-full h-14 pl-5 pr-10 border border-slate-200 rounded-2xl bg-slate-50 font-black text-slate-800 outline-none focus:ring-4 focus:ring-blue-500/10 focus:border-blue-500 transition-all appearance-none shadow-inner"
                    value={quickTargetStaffId}
                    onChange={(e) => setQuickTargetStaffId(e.target.value)}
                  >
                    {staff.map(s => (
                      <option key={s.id} value={s.id}>{s.name} ({s.level})</option>
                    ))}
                  </select>
                  <ChevronDown size={20} className="absolute right-5 top-1/2 -translate-y-1/2 text-slate-400 pointer-events-none"/>
                </div>
              </div>

              {/* 步骤二：选择数据源 */}
              <div className="space-y-6">
                <label className="text-[11px] font-black text-slate-400 uppercase tracking-widest ml-1 flex items-center gap-2">
                  <div className="w-5 h-5 rounded-full bg-slate-100 flex items-center justify-center text-[10px] text-slate-500">2</div>
                  {t('schedule:quick.selectSource')}
                </label>

                <div className="grid grid-cols-2 gap-4">
                  <button
                    onClick={() => setQuickSourceType('SELF')}
                    className={`p-6 rounded-[28px] border-2 flex flex-col items-center gap-3 transition-all ${quickSourceType === 'SELF' ? 'bg-blue-50 border-blue-500 shadow-lg shadow-blue-100' : 'bg-white border-slate-100 hover:border-slate-200'}`}
                  >
                    <div className={`p-4 rounded-2xl ${quickSourceType === 'SELF' ? 'bg-blue-600 text-white' : 'bg-slate-100 text-slate-400'}`}>
                      <Copy size={24}/>
                    </div>
                    <span className={`font-black text-sm ${quickSourceType === 'SELF' ? 'text-blue-700' : 'text-slate-500'}`}>{t('schedule:quick.copySelf')}</span>
                  </button>
                  <button
                    onClick={() => setQuickSourceType('OTHER')}
                    className={`p-6 rounded-[28px] border-2 flex flex-col items-center gap-3 transition-all ${quickSourceType === 'OTHER' ? 'bg-indigo-50 border-indigo-500 shadow-lg shadow-indigo-100' : 'bg-white border-slate-100 hover:border-slate-200'}`}
                  >
                    <div className={`p-4 rounded-2xl ${quickSourceType === 'OTHER' ? 'bg-indigo-600 text-white' : 'bg-slate-100 text-slate-400'}`}>
                      <ArrowRightLeft size={24}/>
                    </div>
                    <span className={`font-black text-sm ${quickSourceType === 'OTHER' ? 'text-indigo-700' : 'text-slate-500'}`}>{t('schedule:quick.copyOther')}</span>
                  </button>
                </div>

                {quickSourceType === 'OTHER' && (
                  <div className="animate-slide-up">
                    <p className="text-[10px] font-bold text-slate-400 mb-2 ml-1">{t('schedule:quick.selectCopyPerson')}</p>
                    <div className="relative">
                      <select
                        className="w-full h-14 pl-5 pr-10 border border-slate-200 rounded-2xl bg-slate-50 font-black text-slate-800 outline-none focus:ring-4 focus:ring-indigo-500/10 focus:border-indigo-500 transition-all appearance-none shadow-inner"
                        value={quickSourceStaffId}
                        onChange={(e) => setQuickSourceStaffId(e.target.value)}
                      >
                        {staff.map(s => (
                          <option key={s.id} value={s.id}>{s.name} ({s.level})</option>
                        ))}
                      </select>
                      <ChevronDown size={20} className="absolute right-5 top-1/2 -translate-y-1/2 text-slate-400 pointer-events-none"/>
                    </div>
                  </div>
                )}
              </div>
            </div>

            <div className="px-10 py-8 bg-slate-50 border-t border-slate-200 flex justify-end gap-4 shrink-0">
              <button onClick={() => setIsQuickScheduleModalOpen(false)} className="px-8 py-3.5 bg-white border border-slate-200 rounded-2xl text-sm font-black text-slate-500 hover:bg-slate-100 transition-all">{t('schedule:action.cancelAction')}</button>
              <button
                onClick={handleApplyQuickSchedule}
                className="px-12 py-3.5 bg-blue-600 text-white rounded-2xl text-sm font-black shadow-xl shadow-blue-200 hover:bg-blue-700 transition-all flex items-center gap-3"
              >
                <CheckCircle2 size={20} className="stroke-[2.5px]"/> {t('schedule:action.applySchedule')}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* 护士排班操作弹窗 */}
      {isStaffScheduling && schedulingData && (
        <div className="fixed inset-0 z-[110] flex items-center justify-center bg-black/50 backdrop-blur-md animate-fade-in p-4">
          <div className="bg-white rounded-[32px] shadow-2xl w-full max-w-xl overflow-hidden animate-scale-in flex flex-col">
            <div className="px-8 py-6 border-b border-slate-100 flex justify-between items-center bg-slate-50">
              <div>
                <h3 className="text-xl font-black text-slate-900">{t('schedule:staffModal.assignPosition')}</h3>
                <p className="text-xs text-slate-400 mt-1 font-bold">{t('schedule:staffModal.nurse')} <span className="text-blue-600">{staff.find(s => s.id === schedulingData.staffId)?.name}</span> · {t('schedule:staffModal.date')} {schedulingData.date}</p>
              </div>
              <button onClick={() => setIsStaffScheduling(false)} className="p-2 text-slate-400 hover:text-slate-600 rounded-xl"><X size={20}/></button>
            </div>
            <div className="p-8 space-y-6">
              <p className="text-[11px] font-black text-slate-400 uppercase tracking-widest mb-2">{t('schedule:staffModal.selectPreset')}</p>
              <div className="grid grid-cols-2 gap-3">
                {shiftTemplates.map(tmpl => (
                  <button
                    key={tmpl.id}
                    onClick={() => {
                      const newShift = {
                        date: schedulingData.date,
                        staffId: schedulingData.staffId!,
                        label: tmpl.name,
                        timeRange: tmpl.timeRange,
                        category: tmpl.category,
                        type: tmpl.category[0], // 兼容原有数据结构
                        area: 'A区'
                      }
                      setStaffSchedule(prev => {
                        const filtered = prev.filter(s => !(s.date === newShift.date && s.staffId === newShift.staffId))
                        return [...filtered, newShift]
                      })
                      setIsStaffScheduling(false)
                    }}
                    className="flex flex-col items-center justify-center p-4 border border-slate-100 rounded-2xl bg-slate-50/50 hover:bg-blue-50 hover:border-blue-200 transition-all group"
                  >
                    <span className="text-sm font-black text-slate-700 group-hover:text-blue-700 mb-1">{tmpl.name}</span>
                    <span className="text-[10px] font-bold text-slate-400 group-hover:text-blue-400">{tmpl.timeRange}</span>
                  </button>
                ))}
                <button
                  onClick={() => {
                    setStaffSchedule(prev => {
                      const filtered = prev.filter(s => !(s.date === schedulingData.date && s.staffId === schedulingData.staffId))
                      return [...filtered, { date: schedulingData.date, staffId: schedulingData.staffId!, type: 'OFF', area: '' }]
                    })
                    setIsStaffScheduling(false)
                  }}
                  className="flex flex-col items-center justify-center p-4 border border-dashed border-slate-200 rounded-2xl hover:bg-slate-50 transition-all"
                >
                  <span className="text-sm font-black text-slate-400 italic">{t('schedule:staffModal.dayOff')}</span>
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* 患者排班操作弹窗 */}
      {isScheduling && schedulingData && (
        <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/50 backdrop-blur-md animate-fade-in p-4">
          <div className="bg-white rounded-[40px] shadow-2xl w-full max-w-2xl overflow-hidden animate-scale-in flex flex-col ring-1 ring-black/5">
            <div className="px-10 py-8 bg-slate-50 border-b border-slate-200 flex justify-between items-center">
              <div>
                <h3 className="text-2xl font-black text-slate-900">{t('schedule:patientModal.title')}</h3>
                <p className="text-sm text-slate-400 mt-1 font-bold tracking-tight">{t('schedule:patientModal.bed')} <span className="text-blue-600">{schedulingData.bedNumber}</span> · {t('schedule:patientModal.date')} {schedulingData.date}</p>
              </div>
              <button onClick={() => setIsScheduling(false)} className="p-3 text-slate-400 hover:text-slate-600 hover:bg-white rounded-2xl transition-all shadow-sm border border-transparent hover:border-slate-200"><X size={24}/></button>
            </div>

            <div className="p-10 space-y-10">
              <div className="grid grid-cols-2 gap-10">
                <div className="space-y-3">
                  <label className="text-[11px] font-black text-slate-400 uppercase tracking-widest ml-1">{t('schedule:patientModal.selectPatient')}</label>
                  <div className="relative">
                    <Search size={16} className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-400"/>
                    <input
                      type="text"
                      placeholder={t('schedule:patientModal.searchPlaceholder')}
                      className="w-full pl-11 pr-4 py-4 bg-slate-50 border border-slate-200 rounded-2xl outline-none focus:ring-4 focus:ring-blue-500/10 focus:border-blue-500 focus:bg-white transition-all text-sm font-bold shadow-inner"
                      value={searchPatientQuery}
                      onChange={(e) => { setSearchPatientQuery(e.target.value); setShowSearchDropdown(true) }}
                      onFocus={() => setShowSearchDropdown(true)}
                    />
                    {showSearchDropdown && searchPatientQuery && (
                      <div className="absolute top-full left-0 right-0 mt-3 bg-white border border-slate-200 rounded-3xl shadow-2xl max-h-56 overflow-y-auto z-50 no-scrollbar ring-1 ring-black/5 p-2 animate-slide-up">
                        {searchResults.map(p => (
                          <button
                            key={p.id}
                            onClick={() => {
                              setSelectedPatientId(p.id!)
                              setSearchPatientQuery(p.name || '')
                              setShowSearchDropdown(false)
                              setSelectedTreatmentMode(p.defaultMode || 'HD')
                            }}
                            className="w-full text-left px-5 py-4 hover:bg-blue-50 flex items-center justify-between rounded-2xl transition-colors mb-1 last:mb-0"
                          >
                            <div className="flex items-center gap-4">
                              <div className="w-10 h-10 rounded-xl bg-blue-100 flex items-center justify-center text-blue-600 font-black">{(p.name || '?')[0]}</div>
                              <p className="text-sm font-black text-slate-800">{p.name}</p>
                            </div>
                            <span className="text-[10px] bg-slate-100 px-3 py-1 rounded-lg font-black text-slate-500 uppercase">{p.defaultMode || 'HD'}</span>
                          </button>
                        ))}
                        {searchResults.length === 0 && (
                          <div className="text-center py-4 text-slate-400 text-sm">未找到匹配患者</div>
                        )}
                      </div>
                    )}
                  </div>
                </div>
                <div className="space-y-3">
                  <label className="text-[11px] font-black text-slate-400 uppercase tracking-widest ml-1">{t('schedule:patientModal.treatmentMode')}</label>
                  <div className="relative">
                    <select
                      className="w-full px-5 py-4 bg-slate-50 border border-slate-200 rounded-2xl outline-none focus:ring-4 focus:ring-blue-500/10 focus:border-blue-500 focus:bg-white transition-all text-sm font-bold appearance-none shadow-inner"
                      value={selectedTreatmentMode}
                      onChange={(e) => setSelectedTreatmentMode(e.target.value)}
                    >
                      {['HD', 'HDF', 'HF', 'HP', 'HD+HP'].map(m => <option key={m} value={m}>{m}</option>)}
                    </select>
                    <ChevronDown size={18} className="absolute right-5 top-1/2 -translate-y-1/2 text-slate-400 pointer-events-none"/>
                  </div>
                </div>
              </div>
            </div>

            <div className="px-10 py-8 bg-slate-50 border-t border-slate-200 flex justify-end gap-4 shrink-0">
              <button onClick={() => setIsScheduling(false)} className="px-8 py-3.5 bg-white border border-slate-200 rounded-2xl text-sm font-black text-slate-500 hover:bg-slate-100 transition-all shadow-sm">{t('schedule:action.cancel')}</button>
              <button
                disabled={!selectedPatientId}
                onClick={async () => {
                  const patient = patients.find(p => p.id === selectedPatientId)
                  if (patient && schedulingData.bedNumber && schedulingData.shift) {
                    // 找到对应的后端 shiftId
                    const shiftTypeMap: Record<string, string[]> = {
                      'Morning': ['早', '上午', 'morning'],
                      'Afternoon': ['中', '下午', 'afternoon'],
                      'Evening': ['晚', '夜', 'evening']
                    }
                    const keywords = shiftTypeMap[schedulingData.shift] || []
                    const matchedShift = shifts.find(s => keywords.some(k => s.name.toLowerCase().includes(k)))
                    const shiftId = matchedShift?.id || shifts[0]?.id

                    if (shiftId) {
                      try {
                        await restApi.createPatientShift({
                          patientId: Number(patient.id),
                          scheduleDate: schedulingData.date,
                          shiftId,
                        })
                      } catch (err) {
                        console.error('创建排班失败:', err)
                      }
                    }

                    // 同时更新本地状态（乐观更新）
                    const newSession: PatientScheduleItem = {
                      id: `S-NEW-${Date.now()}`,
                      bedNumber: schedulingData.bedNumber,
                      date: schedulingData.date,
                      shift: schedulingData.shift,
                      patientName: patient.name || '',
                      mode: selectedTreatmentMode,
                      patientId: patient.id
                    }
                    setPatientSchedule(prev => [...prev, newSession])
                    setIsScheduling(false)
                  }
                }}
                className="px-12 py-3.5 bg-blue-600 text-white rounded-2xl text-sm font-black shadow-xl shadow-blue-200 hover:bg-blue-700 transition-all flex items-center gap-3 disabled:opacity-30 disabled:pointer-events-none"
              >
                <CheckCircle2 size={20} className="stroke-[2.5px]"/> {t('schedule:action.confirmSchedule')}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
