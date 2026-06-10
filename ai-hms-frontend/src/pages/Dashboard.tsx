import { useEffect, useState, type ReactNode } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  Activity,
  AlertTriangle,
  ArrowRight,
  Bed,
  CalendarClock,
  CheckCircle2,
  ClipboardCheck,
  Droplets,
  HeartPulse,
  Monitor,
  Stethoscope,
  Users,
} from 'lucide-react'
import type { Patient } from '@/types/original'
import {
  restApi,
  convertRestPatientList,
  getAllEquipments,
  getTodayTreatments,
  type EquipmentInfo,
  type Treatment,
} from '@/services'

type DashboardStats = {
  activePatients: number
  shiftCount: number
  equipmentCount: number
  todaySchedules: number
  todayTreatments: number
  alertItems: number
  treatmentsByHour: { name: string; value: number }[]
  qualityByHour: { name: string; value: number }[]
}

type QuickAction = {
  title: string
  desc: string
  route: string
  icon: ReactNode
  tone: string
}

const treatmentStatusText: Record<string, string> = {
  ongoing: '透析中',
  completed: '已完成',
  paused: '暂停',
  pending: '待开始',
  '1': '透析中',
  '2': '已完成',
  '3': '暂停',
  '10': '待接诊',
  '20': '透析中',
  '30': '已完成',
}

const deviceStatusText: Record<string, string> = {
  normal: '正常',
  warning: '预警',
  alarm: '报警',
  offline: '离线',
  maintenance: '维护',
  active: '正常',
  '1': '正常',
  '0': '停用',
}

const formatClock = (value?: string) => {
  if (!value) return '--:--'
  const time = value.includes('T') ? value.split('T')[1] : value
  return time.slice(0, 5)
}

const isTreatmentRunning = (status?: string) => {
  if (!status) return false
  return ['ongoing', '1', '20', '透析中', '进行中'].includes(status)
}

const isDeviceAttention = (status?: string) => {
  if (!status) return false
  return ['warning', 'alarm', 'offline', 'maintenance', '0', '预警', '报警', '离线', '维护'].includes(status)
}

const actionItems: QuickAction[] = [
  {
    title: '进入透析执行',
    desc: '核对、透前评估、医嘱执行、透中监测',
    route: '/dialysis-processing',
    icon: <ClipboardCheck size={20} />,
    tone: 'bg-teal-50 text-teal-700 border-teal-100',
  },
  {
    title: '查看实时监测',
    desc: '床旁状态、报警、透析机在线情况',
    route: '/monitoring',
    icon: <Monitor size={20} />,
    tone: 'bg-orange-50 text-orange-700 border-orange-100',
  },
  {
    title: '维护今日排班',
    desc: '班次、床位、治疗安排调整',
    route: '/schedule',
    icon: <CalendarClock size={20} />,
    tone: 'bg-sky-50 text-sky-700 border-sky-100',
  },
  {
    title: '患者资料中心',
    desc: '病史、通路、处方、检验检查',
    route: '/patients',
    icon: <Users size={20} />,
    tone: 'bg-emerald-50 text-emerald-700 border-emerald-100',
  },
]

export default function Dashboard() {
  const navigate = useNavigate()
  const [patients, setPatients] = useState<Partial<Patient>[]>([])
  const [patientTotal, setPatientTotal] = useState<number | null>(null)
  const [equipments, setEquipments] = useState<EquipmentInfo[]>([])
  const [treatments, setTreatments] = useState<Treatment[]>([])
  const [dashboardStats, setDashboardStats] = useState<DashboardStats | null>(null)
  const [apiError, setApiError] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    let alive = true

    Promise.all([
      restApi.getPatientList({ page: 1, pageSize: 50, onlyActive: true }).catch(() => null),
      getAllEquipments().catch(() => []),
      getTodayTreatments().catch(() => []),
      restApi.getDashboardStats().catch(() => null),
    ]).then(([patientResult, equipmentsData, treatmentsData, statsData]) => {
      if (!alive) return
      if (patientResult?.data?.items) {
        setPatients(convertRestPatientList(patientResult.data.items))
        setPatientTotal(patientResult.data.pagination?.total ?? patientResult.data.items.length)
      }
      setEquipments(equipmentsData)
      setTreatments(treatmentsData)
      setDashboardStats(statsData)
      setApiError(null)
      setIsLoading(false)
    }).catch(() => {
      if (!alive) return
      setApiError('工作台数据暂时不可用，请检查后端服务或接口配置。')
      setIsLoading(false)
    })

    return () => {
      alive = false
    }
  }, [])

  const activePatientCount = dashboardStats?.activePatients ?? patientTotal ?? patients.length
  const todayTreatmentCount = dashboardStats?.todayTreatments ?? treatments.length
  const todayScheduleCount = dashboardStats?.todaySchedules ?? 0
  const equipmentCount = dashboardStats?.equipmentCount ?? equipments.length
  const runningTreatments = treatments.filter(item => isTreatmentRunning(item.Status)).length
  const attentionDevices = equipments.filter(item => isDeviceAttention(item.Status)).length
  const alertCount = dashboardStats?.alertItems ?? attentionDevices
  const completionRate = todayTreatmentCount > 0
    ? Math.round((treatments.filter(item => ['completed', '2', '30', '已完成'].includes(item.Status ?? '')).length / todayTreatmentCount) * 100)
    : 0
  const visibleTreatments = treatments.slice(0, 6)
  const visiblePatients = patients.slice(0, 6)
  const visibleDevices = equipments.slice(0, 8)
  const hourBars = (dashboardStats?.treatmentsByHour?.length ? dashboardStats.treatmentsByHour : [
    { name: '07:00', value: Math.max(1, Math.round(todayTreatmentCount * 0.25)) },
    { name: '11:00', value: Math.max(1, Math.round(todayTreatmentCount * 0.35)) },
    { name: '15:00', value: Math.max(1, Math.round(todayTreatmentCount * 0.25)) },
    { name: '19:00', value: Math.max(0, Math.round(todayTreatmentCount * 0.15)) },
  ]).slice(0, 6)
  const maxHourValue = Math.max(...hourBars.map(item => item.value), 1)

  return (
    <div className="max-w-[1600px] mx-auto space-y-6 pb-10">
      {apiError && (
        <div className="rounded-lg border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-800">
          {apiError}
        </div>
      )}

      <section className="relative overflow-hidden rounded-lg border border-slate-200 bg-gradient-to-br from-slate-950 via-slate-900 to-teal-900 text-white shadow-sm">
        <div className="absolute inset-y-0 right-0 w-1/2 bg-[radial-gradient(circle_at_top_right,rgba(45,212,191,0.28),transparent_42%)]" />
        <div className="relative grid gap-6 p-6 lg:grid-cols-[1.35fr_0.65fr] lg:p-8">
          <div>
            <div className="mb-4 inline-flex items-center gap-2 rounded-md border border-white/15 bg-white/10 px-3 py-1 text-sm text-teal-50 backdrop-blur">
              <HeartPulse size={16} /> 血液透析工作台
            </div>
            <h1 className="text-3xl font-bold tracking-tight md:text-4xl">今日透析运行总览</h1>
            <p className="mt-3 max-w-[68ch] text-sm leading-6 text-slate-200">
              聚合今日排班、透析执行、设备状态和重点患者，帮助医护快速判断当前治疗压力，并一键进入对应业务菜单处理。
            </p>
            <div className="mt-6 grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
              <HeroMetric label="今日透析" value={todayTreatmentCount} suffix="人次" icon={<Droplets size={18} />} />
              <HeroMetric label="透析中" value={runningTreatments} suffix="人" icon={<Activity size={18} />} />
              <HeroMetric label="待关注" value={alertCount} suffix="项" icon={<AlertTriangle size={18} />} />
              <HeroMetric label="在档患者" value={activePatientCount} suffix="人" icon={<Users size={18} />} />
            </div>
          </div>
          <div className="rounded-lg border border-white/10 bg-white/10 p-5 backdrop-blur">
            <div className="flex items-center justify-between">
              <span className="text-sm text-slate-200">今日完成率</span>
              <span className="text-2xl font-bold">{completionRate}%</span>
            </div>
            <div className="mt-4 h-2 overflow-hidden rounded-md bg-white/15">
              <div className="h-full rounded-md bg-teal-300 transition-all duration-500" style={{ width: `${Math.min(completionRate, 100)}%` }} />
            </div>
            <div className="mt-6 space-y-3 text-sm text-slate-100">
              <StatusLine label="今日排班" value={`${todayScheduleCount} 个班次/安排`} />
              <StatusLine label="设备总数" value={`${equipmentCount} 台`} />
              <StatusLine label="设备需关注" value={`${attentionDevices} 台`} danger={attentionDevices > 0} />
            </div>
            <button
              type="button"
              onClick={() => navigate('/monitoring')}
              className="mt-6 inline-flex w-full items-center justify-center gap-2 rounded-md bg-white px-4 py-2.5 text-sm font-semibold text-slate-900 transition hover:bg-teal-50 active:scale-[0.98]"
            >
              打开实时监测 <ArrowRight size={16} />
            </button>
          </div>
        </div>
      </section>

      <section className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        {actionItems.map(item => (
          <button
            key={item.title}
            type="button"
            onClick={() => navigate(item.route)}
            className={`group rounded-lg border p-4 text-left transition hover:-translate-y-0.5 hover:shadow-md active:scale-[0.98] ${item.tone}`}
          >
            <div className="mb-4 flex items-center justify-between">
              <span className="inline-flex h-10 w-10 items-center justify-center rounded-md bg-white/70">{item.icon}</span>
              <ArrowRight size={18} className="transition group-hover:translate-x-1" />
            </div>
            <div className="font-semibold">{item.title}</div>
            <p className="mt-1 text-sm opacity-80">{item.desc}</p>
          </button>
        ))}
      </section>

      <section className="grid gap-5">
        <Panel title="今日治疗节奏" action="查看统计" onAction={() => navigate('/statistics')}>
          <div className="grid gap-5 lg:grid-cols-[0.9fr_1.1fr]">
            <div className="space-y-3">
              {hourBars.map(item => (
                <div key={item.name} className="grid grid-cols-[52px_1fr_40px] items-center gap-3 text-sm">
                  <span className="text-slate-500">{item.name}</span>
                  <div className="h-3 overflow-hidden rounded-md bg-slate-100">
                    <div className="h-full rounded-md bg-teal-500" style={{ width: `${Math.max(8, (item.value / maxHourValue) * 100)}%` }} />
                  </div>
                  <span className="text-right font-semibold text-slate-800">{item.value}</span>
                </div>
              ))}
            </div>
            <div className="grid gap-3 sm:grid-cols-3">
              <MiniMetric label="治疗中" value={runningTreatments} hint="需持续监测" />
              <MiniMetric label="已完成" value={treatments.filter(item => ['completed', '2', '30', '已完成'].includes(item.Status ?? '')).length} hint="可做透后评估" />
              <MiniMetric label="未开始" value={Math.max(todayTreatmentCount - treatments.length, 0)} hint="关注接诊排队" />
            </div>
          </div>
          </Panel>
      </section>

      <section className="grid gap-5 xl:grid-cols-3">
        <Panel title="今日透析患者" action="患者列表" onAction={() => navigate('/patients')} className="xl:col-span-1">
          {isLoading ? <SkeletonRows /> : visiblePatients.length > 0 ? (
            <div className="divide-y divide-slate-100">
              {visiblePatients.map(patient => (
                <button key={patient.id} type="button" onClick={() => patient.id && navigate(`/patients/${patient.id}`)} className="flex w-full items-center gap-3 py-3 text-left transition hover:bg-slate-50 active:scale-[0.98]">
                  <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-md bg-emerald-50 font-semibold text-emerald-700">{patient.name?.slice(0, 1) || '患'}</span>
                  <span className="min-w-0 flex-1">
                    <span className="block truncate font-semibold text-slate-900">{patient.name || '未命名患者'}</span>
                    <span className="mt-0.5 block text-sm text-slate-500">床位 {patient.bedNumber || '--'} · {patient.diagnosis || patient.defaultMode || '待完善治疗信息'}</span>
                  </span>
                  <span className="rounded-md bg-slate-100 px-2 py-1 text-xs text-slate-600">{patient.riskLevel || '常规'}</span>
                </button>
              ))}
            </div>
          ) : <EmptyState text="暂无今日患者数据" />}
        </Panel>

        <Panel title="透析执行动态" action="进入执行" onAction={() => navigate('/dialysis-processing')} className="xl:col-span-1">
          {isLoading ? <SkeletonRows /> : visibleTreatments.length > 0 ? (
            <div className="space-y-3">
              {visibleTreatments.map(item => (
                <button key={item.Id} type="button" onClick={() => navigate('/dialysis-processing')} className="w-full rounded-lg border border-slate-100 px-4 py-3 text-left transition hover:border-teal-200 hover:bg-teal-50 active:scale-[0.98]">
                  <div className="flex items-center justify-between gap-3">
                    <span className="font-semibold text-slate-900">患者 ID：{item.PatientId}</span>
                    <span className="rounded-md bg-teal-50 px-2 py-1 text-xs font-medium text-teal-700">{treatmentStatusText[item.Status ?? ''] || item.Status || '待确认'}</span>
                  </div>
                  <div className="mt-2 flex items-center gap-3 text-sm text-slate-500">
                    <span>班次 {item.ShiftId || '--'}</span>
                    <span>{formatClock(item.StartTime)} - {formatClock(item.EndTime)}</span>
                  </div>
                </button>
              ))}
            </div>
          ) : <EmptyState text="暂无今日治疗记录" />}
        </Panel>

        <Panel title="设备与床位关注" action="设备管理" onAction={() => navigate('/device-binding')} className="xl:col-span-1">
          {isLoading ? <SkeletonRows /> : visibleDevices.length > 0 ? (
            <div className="grid grid-cols-2 gap-3">
              {visibleDevices.map(device => {
                const needsAttention = isDeviceAttention(device.Status)
                return (
                  <button key={device.Id} type="button" onClick={() => navigate('/monitoring')} className={`rounded-lg border p-3 text-left transition active:scale-[0.98] ${needsAttention ? 'border-orange-200 bg-orange-50 hover:bg-orange-100' : 'border-slate-100 bg-slate-50 hover:bg-emerald-50'}`}>
                    <div className="mb-2 flex items-center justify-between gap-2">
                      <Bed size={16} className={needsAttention ? 'text-orange-600' : 'text-emerald-600'} />
                      <span className={`rounded-md px-2 py-0.5 text-xs ${needsAttention ? 'bg-orange-100 text-orange-700' : 'bg-emerald-100 text-emerald-700'}`}>{deviceStatusText[device.Status ?? ''] || device.Status || '未知'}</span>
                    </div>
                    <div className="truncate font-semibold text-slate-900">{device.Name || `设备 ${device.Id}`}</div>
                    <div className="mt-1 truncate text-xs text-slate-500">编号 {device.IDNo || '--'}</div>
                  </button>
                )
              })}
            </div>
          ) : <EmptyState text="暂无设备数据" />}
        </Panel>
      </section>
    </div>
  )
}

function HeroMetric({ label, value, suffix, icon }: { label: string; value: number; suffix: string; icon: ReactNode }) {
  return (
    <div className="rounded-lg border border-white/10 bg-white/10 p-4 backdrop-blur">
      <div className="mb-3 flex items-center justify-between text-teal-50">
        <span className="text-sm">{label}</span>
        {icon}
      </div>
      <div className="flex items-end gap-1">
        <span className="text-3xl font-bold leading-none">{value}</span>
        <span className="text-sm text-slate-200">{suffix}</span>
      </div>
    </div>
  )
}

function StatusLine({ label, value, danger }: { label: string; value: string; danger?: boolean }) {
  return (
    <div className="flex items-center justify-between gap-3">
      <span className="text-slate-300">{label}</span>
      <span className={danger ? 'font-semibold text-orange-200' : 'font-semibold text-white'}>{value}</span>
    </div>
  )
}

function Panel({ title, action, onAction, className = '', children }: { title: string; action: string; onAction: () => void; className?: string; children: ReactNode }) {
  return (
    <section className={`rounded-lg border border-slate-200 bg-white p-5 shadow-sm ${className}`}>
      <div className="mb-5 flex items-center justify-between gap-4">
        <div className="flex items-center gap-2">
          <span className="flex h-9 w-9 items-center justify-center rounded-md bg-teal-50 text-teal-700"><Stethoscope size={18} /></span>
          <h2 className="text-base font-bold text-slate-900">{title}</h2>
        </div>
        <button type="button" onClick={onAction} className="inline-flex items-center gap-1 rounded-md px-2 py-1 text-sm font-medium text-teal-700 transition hover:bg-teal-50 active:scale-[0.98]">
          {action} <ArrowRight size={15} />
        </button>
      </div>
      {children}
    </section>
  )
}

function MiniMetric({ label, value, hint }: { label: string; value: number; hint: string }) {
  return (
    <div className="rounded-lg border border-slate-100 bg-slate-50 p-4">
      <div className="flex items-center justify-between">
        <span className="text-sm text-slate-500">{label}</span>
        <CheckCircle2 size={16} className="text-teal-600" />
      </div>
      <div className="mt-3 text-2xl font-bold text-slate-900">{value}</div>
      <div className="mt-1 text-xs text-slate-500">{hint}</div>
    </div>
  )
}

function EmptyState({ text }: { text: string }) {
  return (
    <div className="flex min-h-32 items-center justify-center rounded-lg border border-dashed border-slate-200 bg-slate-50 text-sm text-slate-500">
      {text}
    </div>
  )
}

function SkeletonRows() {
  return (
    <div className="space-y-3" aria-live="polite">
      {[0, 1, 2].map(item => (
        <div key={item} className="h-14 animate-pulse rounded-lg bg-slate-100" />
      ))}
    </div>
  )
}
