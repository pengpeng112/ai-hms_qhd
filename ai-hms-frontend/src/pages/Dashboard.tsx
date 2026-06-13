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
  Monitor,
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
  runningTreatments?: number
  completedTreatments?: number
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

const QUICK_ACTIONS: QuickAction[] = [
  {
    title: '进入透析执行',
    desc: '透前评估、医嘱、监测',
    route: '/dialysis-processing',
    icon: <ClipboardCheck size={20} />,
    tone: 'bg-blue-50 text-blue-700 border-blue-100',
  },
  {
    title: '查看实时监测',
    desc: '床旁状态、报警',
    route: '/monitoring',
    icon: <Monitor size={20} />,
    tone: 'bg-orange-50 text-orange-700 border-orange-100',
  },
  {
    title: '维护今日排班',
    desc: '班次、床位安排',
    route: '/schedule',
    icon: <CalendarClock size={20} />,
    tone: 'bg-sky-50 text-sky-700 border-sky-100',
  },
  {
    title: '患者资料中心',
    desc: '病史、通路、处方',
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
      restApi.getPatientList({ page: 1, pageSize: 50, onlyActive: true }).catch((error) => {
        console.warn('[Dashboard] 患者列表加载失败', error)
        return null
      }),
      getAllEquipments().catch((error) => {
        console.warn('[Dashboard] 设备列表加载失败', error)
        return []
      }),
      getTodayTreatments().catch((error) => {
        console.warn('[Dashboard] 今日治疗列表加载失败', error)
        return []
      }),
      restApi.getDashboardStats().catch((error) => {
        console.warn('[Dashboard] 看板统计加载失败', error)
        return null
      }),
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
  const runningTreatments = dashboardStats?.runningTreatments ?? treatments.filter((item) => isTreatmentRunning(item.Status)).length
  const attentionDevices = equipments.filter((item) => isDeviceAttention(item.Status)).length
  const alertCount = dashboardStats?.alertItems ?? attentionDevices
  const completedTreatments = dashboardStats?.completedTreatments ?? treatments.filter((item) =>
    ['completed', '2', '30', '已完成'].includes(item.Status ?? '')
  ).length
  const completionRate =
    todayTreatmentCount > 0
      ? Math.round((completedTreatments / todayTreatmentCount) * 100)
      : 0

  const hasScheduleWithoutTreatment = todayScheduleCount > 0 && todayTreatmentCount === 0
  const pendingScheduleCount = Math.max(todayScheduleCount - todayTreatmentCount, 0)

  const visibleTreatments = treatments.slice(0, 6)
  const visiblePatients = patients.slice(0, 6)
  const visibleDevices = equipments.slice(0, 8)

  const hourBars = (
    dashboardStats?.treatmentsByHour?.length
      ? dashboardStats.treatmentsByHour
      : todayTreatmentCount > 0
        ? [
            { name: '08:00', value: Math.round(todayTreatmentCount * 0.25) },
            { name: '10:00', value: Math.round(todayTreatmentCount * 0.35) },
            { name: '12:00', value: Math.round(todayTreatmentCount * 0.25) },
            { name: '14:00', value: Math.round(todayTreatmentCount * 0.15) },
          ]
        : [
            { name: '08:00', value: 0 },
            { name: '09:00', value: 0 },
            { name: '10:00', value: 0 },
            { name: '11:00', value: 0 },
            { name: '12:00', value: 0 },
            { name: '13:00', value: 0 },
          ]
  ).slice(0, 6)
  const maxHourValue = Math.max(...hourBars.map((item) => item.value), 1)

  const todoItems = [
    {
      index: '1',
      title: '创建/补录治疗记录',
      desc: `${pendingScheduleCount} 项排班待转入执行`,
      action: '透析执行',
      route: '/dialysis-processing',
      tone: 'blue' as const,
    },
    {
      index: '2',
      title: '确认未开始患者',
      desc: '关注接诊和候诊状态',
      action: '患者列表',
      route: '/patients',
      tone: 'teal' as const,
    },
    {
      index: '3',
      title: '设备与床位巡检',
      desc: attentionDevices > 0 ? `${attentionDevices} 台设备需关注` : '当前无异常设备',
      action: '实时监测',
      route: '/monitoring',
      tone: (attentionDevices > 0 ? 'orange' : 'green'),
    },
    {
      index: '4',
      title: '透后评估补齐',
      desc: `${Math.max(todayTreatmentCount - completedTreatments, 0)} 人待评估`,
      action: '透后评估',
      route: '/dialysis-processing',
      tone: 'indigo' as const,
    },
  ]

  const metricCards = [
    {
      label: '今日透析',
      value: todayTreatmentCount,
      unit: '人次',
      hint: '当前治疗记录数',
      icon: <Droplets size={20} />,
      barColor: 'bg-blue-500',
      tone: 'text-blue-600',
    },
    {
      label: '透析中',
      value: runningTreatments,
      unit: '人',
      hint: '需要持续监测',
      icon: <Activity size={20} />,
      barColor: 'bg-teal-500',
      tone: 'text-teal-600',
    },
    {
      label: '待关注',
      value: alertCount,
      unit: '项',
      hint: alertCount > 0 ? '需及时处理' : '无异常',
      icon: <AlertTriangle size={20} />,
      barColor: alertCount > 0 ? 'bg-orange-500' : 'bg-emerald-500',
      tone: alertCount > 0 ? 'text-orange-600' : 'text-emerald-600',
    },
    {
      label: '在档患者',
      value: activePatientCount,
      unit: '人',
      hint: '系统登记患者',
      icon: <Users size={20} />,
      barColor: 'bg-indigo-500',
      tone: 'text-indigo-600',
    },
    {
      label: '完成率',
      value: completionRate,
      unit: '%',
      hint: todayTreatmentCount > 0 ? `已完成 ${completedTreatments}/${todayTreatmentCount}` : '暂无治疗',
      icon: <CheckCircle2 size={20} />,
      barColor: completionRate >= 80 ? 'bg-emerald-500' : completionRate >= 40 ? 'bg-amber-500' : 'bg-slate-400',
      tone: completionRate >= 80 ? 'text-emerald-600' : completionRate >= 40 ? 'text-amber-600' : 'text-slate-500',
    },
  ]

  return (
    <div className="max-w-[1600px] mx-auto space-y-5 pb-10">
      {apiError && (
        <div className="rounded-lg border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-800">
          {apiError}
        </div>
      )}

      <section className="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
        <div className="flex flex-col gap-5 lg:flex-row lg:items-center lg:justify-between">
          <div>
            <p className="text-xs font-black uppercase tracking-[0.28em] text-teal-600">Dashboard</p>
            <h1 className="mt-1 text-2xl font-black tracking-tight text-slate-900">今日工作台</h1>
            <p className="mt-1.5 max-w-3xl text-sm text-slate-500">
              聚焦今日排班、透析执行、设备风险与待办处理。
            </p>
          </div>
          <div className="flex items-center gap-4">
            <div className="rounded-xl bg-slate-50 px-4 py-3 text-center">
              <div className="text-[11px] font-bold uppercase tracking-wide text-slate-400">今日排班</div>
              <div className="mt-0.5 text-lg font-black text-slate-800">{todayScheduleCount}</div>
              <div className="text-[11px] text-slate-400">项</div>
            </div>
            <div className="rounded-xl bg-slate-50 px-4 py-3 text-center">
              <div className="text-[11px] font-bold uppercase tracking-wide text-slate-400">系统状态</div>
              <div className="mt-0.5 flex items-center gap-1.5">
                <span className={`inline-block h-2 w-2 rounded-full ${attentionDevices > 0 ? 'bg-orange-500' : 'bg-emerald-500'}`} />
                <span className="text-sm font-bold text-slate-700">
                  {attentionDevices > 0 ? `${attentionDevices} 台需关注` : '正常'}
                </span>
              </div>
              <div className="text-[11px] text-slate-400">{equipmentCount} 台设备</div>
            </div>
          </div>
        </div>
      </section>

      {hasScheduleWithoutTreatment && (
        <section className="rounded-2xl border border-blue-200 bg-blue-50 px-5 py-4 text-sm">
          <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <span className="font-bold text-blue-800">提示：</span>
              <span className="text-blue-700">
                今日已有 {todayScheduleCount} 个排班安排，但暂无治疗记录。建议优先进入「透析执行」开始接诊或补录。
              </span>
            </div>
            <button
              type="button"
              onClick={() => navigate('/dialysis-processing')}
              className="shrink-0 inline-flex items-center gap-1.5 rounded-xl bg-blue-600 px-4 py-2 text-xs font-bold text-white transition hover:bg-blue-700"
            >
              进入透析执行 <ArrowRight size={14} />
            </button>
          </div>
        </section>
      )}

      <section className="grid gap-4 md:grid-cols-3 xl:grid-cols-5">
        {metricCards.map((card) => (
          <div key={card.label} className="rounded-2xl border border-slate-200 bg-white shadow-sm overflow-hidden">
            <div className={`h-1 w-full ${card.barColor}`} />
            <div className="p-5">
              <div className="flex items-center justify-between">
                <span className="text-xs font-semibold text-slate-500">{card.label}</span>
                <span className={card.tone}>{card.icon}</span>
              </div>
              <div className="mt-3 flex items-baseline gap-1">
                <span className="text-2xl font-black text-slate-900">{card.value}</span>
                <span className="text-xs text-slate-400">{card.unit}</span>
              </div>
              <div className="mt-1 text-[11px] text-slate-400">{card.hint}</div>
            </div>
          </div>
        ))}
      </section>

      <section className="grid gap-5 xl:grid-cols-[1fr_1fr]">
        <div className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
          <div className="mb-4 flex items-center gap-2">
            <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-blue-50 text-blue-600">
              <ClipboardCheck size={16} />
            </span>
            <h2 className="text-base font-bold text-slate-900">今日待办</h2>
          </div>
          <div className="space-y-3">
            {todoItems.map((item) => {
              const toneMap: Record<string, string> = {
                blue: 'border-l-blue-500',
                teal: 'border-l-teal-500',
                orange: 'border-l-orange-500',
                green: 'border-l-emerald-500',
                indigo: 'border-l-indigo-500',
              }
              return (
                <button
                  key={item.index}
                  type="button"
                  onClick={() => navigate(item.route)}
                  className={`w-full rounded-xl border border-slate-100 bg-slate-50 border-l-[3px] ${toneMap[item.tone]} px-4 py-3 text-left transition hover:bg-slate-100`}
                >
                  <div className="flex items-center justify-between gap-3">
                    <div className="min-w-0">
                      <div className="flex items-center gap-2">
                        <span className="flex h-5 w-5 items-center justify-center rounded-full bg-slate-200 text-[11px] font-bold text-slate-600">
                          {item.index}
                        </span>
                        <span className="text-sm font-bold text-slate-800">{item.title}</span>
                      </div>
                      <div className="mt-1 text-xs text-slate-500">{item.desc}</div>
                    </div>
                    <span className="inline-flex shrink-0 items-center gap-1 rounded-lg bg-white px-2.5 py-1 text-xs font-semibold text-slate-600 border border-slate-200">
                      {item.action} <ArrowRight size={12} />
                    </span>
                  </div>
                </button>
              )
            })}
          </div>
        </div>

        <div className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
          <div className="mb-4 flex items-center gap-2">
            <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-teal-50 text-teal-600">
              <Activity size={16} />
            </span>
            <h2 className="text-base font-bold text-slate-900">治疗节奏</h2>
            {todayTreatmentCount === 0 && (
              <span className="rounded-full bg-slate-100 px-2 py-0.5 text-[11px] text-slate-400">暂无治疗</span>
            )}
          </div>
          {todayTreatmentCount === 0 ? (
            <div className="flex min-h-[148px] flex-col items-center justify-center text-center">
              <p className="text-sm text-slate-400">今日暂无治疗记录，治疗节奏无数据。</p>
              <button
                type="button"
                onClick={() => navigate('/dialysis-processing')}
                className="mt-3 inline-flex items-center gap-1.5 rounded-xl bg-blue-600 px-4 py-2 text-xs font-bold text-white transition hover:bg-blue-700"
              >
                进入透析执行
              </button>
            </div>
          ) : (
            <div className="space-y-3">
              {hourBars.map((item) => (
                <div key={item.name} className="grid grid-cols-[52px_1fr_40px] items-center gap-3 text-sm">
                  <span className="text-slate-500">{item.name}</span>
                  <div className="h-3 overflow-hidden rounded-md bg-slate-100">
                    <div
                      className="h-full rounded-md bg-teal-500 transition-all duration-300"
                      style={{ width: `${maxHourValue > 0 ? Math.max(4, (item.value / maxHourValue) * 100) : 0}%` }}
                    />
                  </div>
                  <span className="text-right font-semibold text-slate-800">{item.value}</span>
                </div>
              ))}
              <div className="flex items-center gap-4 pt-2 text-xs text-slate-400">
                <span>治疗中 {runningTreatments}</span>
                <span>已完成 {completedTreatments}</span>
                <span>未开始 {Math.max(todayTreatmentCount - treatments.length, 0)}</span>
              </div>
            </div>
          )}
        </div>
      </section>

      <section className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        {QUICK_ACTIONS.map((item) => (
          <button
            key={item.title}
            type="button"
            onClick={() => navigate(item.route)}
            className={`group rounded-xl border p-4 text-left transition hover:-translate-y-0.5 hover:shadow-md active:scale-[0.98] ${item.tone}`}
          >
            <div className="mb-3 flex items-center justify-between">
              <span className="inline-flex h-9 w-9 items-center justify-center rounded-lg bg-white/70">{item.icon}</span>
              <ArrowRight size={16} className="opacity-40 transition group-hover:translate-x-1 group-hover:opacity-100" />
            </div>
            <div className="text-sm font-bold">{item.title}</div>
            <p className="mt-1 text-xs opacity-70">{item.desc}</p>
          </button>
        ))}
      </section>

      <section className="grid gap-5 xl:grid-cols-3">
        <Panel
          title="今日透析患者"
          action="患者列表"
          onAction={() => navigate('/patients')}
          className="xl:col-span-1"
          icon={<Users size={16} />}
        >
          {isLoading ? (
            <SkeletonRows />
          ) : visiblePatients.length > 0 ? (
            <div className="divide-y divide-slate-100">
              {visiblePatients.map((patient) => (
                <button
                  key={patient.id}
                  type="button"
                  onClick={() => patient.id && navigate(`/patients/${patient.id}`)}
                  className="flex w-full items-center gap-3 py-3 text-left transition hover:bg-slate-50 active:scale-[0.98]"
                >
                  <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-emerald-50 font-semibold text-emerald-700">
                    {patient.name?.slice(0, 1) || '患'}
                  </span>
                  <span className="min-w-0 flex-1">
                    <span className="block truncate font-semibold text-slate-900">
                      {patient.name || '未命名患者'}
                    </span>
                    <span className="mt-0.5 block text-xs text-slate-500">
                      床位 {patient.bedNumber || '--'} · {patient.diagnosis || patient.defaultMode || '待完善治疗信息'}
                    </span>
                  </span>
                  <span className="rounded-md bg-slate-100 px-2 py-1 text-xs text-slate-600">
                    {patient.riskLevel || '常规'}
                  </span>
                </button>
              ))}
            </div>
          ) : (
            <EmptyState text="暂无今日患者数据" />
          )}
        </Panel>

        <Panel
          title="透析执行动态"
          action="进入执行"
          onAction={() => navigate('/dialysis-processing')}
          className="xl:col-span-1"
          icon={<ClipboardCheck size={16} />}
        >
          {isLoading ? (
            <SkeletonRows />
          ) : visibleTreatments.length > 0 ? (
            <div className="space-y-2.5">
              {visibleTreatments.map((item) => (
                <button
                  key={item.Id}
                  type="button"
                  onClick={() => navigate('/dialysis-processing')}
                  className="w-full rounded-xl border border-slate-100 px-4 py-3 text-left transition hover:border-teal-200 hover:bg-teal-50 active:scale-[0.98]"
                >
                  <div className="flex items-center justify-between gap-3">
                    <span className="font-semibold text-slate-900">患者 ID：{item.PatientId}</span>
                    <span className="rounded-lg bg-teal-50 px-2 py-1 text-xs font-medium text-teal-700">
                      {treatmentStatusText[item.Status ?? ''] || item.Status || '待确认'}
                    </span>
                  </div>
                  <div className="mt-1.5 flex items-center gap-3 text-xs text-slate-500">
                    <span>班次 {item.ShiftId || '--'}</span>
                    <span>
                      {formatClock(item.StartTime)} - {formatClock(item.EndTime)}
                    </span>
                  </div>
                </button>
              ))}
            </div>
          ) : (
            <EmptyState
              text="暂无今日治疗记录"
              hint={hasScheduleWithoutTreatment ? `已有 ${todayScheduleCount} 个排班安排，建议创建治疗记录` : undefined}
              action="进入透析执行"
              onAction={() => navigate('/dialysis-processing')}
            />
          )}
        </Panel>

        <Panel
          title="设备与床位关注"
          action="设备管理"
          onAction={() => navigate('/device-binding')}
          className="xl:col-span-1"
          icon={<Bed size={16} />}
        >
          {isLoading ? (
            <SkeletonRows />
          ) : visibleDevices.length > 0 ? (
            <div className="grid grid-cols-2 gap-2.5">
              {visibleDevices.map((device) => {
                const needsAttention = isDeviceAttention(device.Status)
                return (
                  <button
                    key={device.Id}
                    type="button"
                    onClick={() => navigate('/monitoring')}
                    className={`rounded-xl border p-3 text-left transition active:scale-[0.98] ${
                      needsAttention
                        ? 'border-orange-200 bg-orange-50 hover:bg-orange-100'
                        : 'border-slate-100 bg-slate-50 hover:bg-emerald-50'
                    }`}
                  >
                    <div className="mb-2 flex items-center justify-between gap-2">
                      <Bed size={15} className={needsAttention ? 'text-orange-600' : 'text-emerald-600'} />
                      <span
                        className={`rounded-md px-1.5 py-0.5 text-[11px] font-medium ${
                          needsAttention ? 'bg-orange-100 text-orange-700' : 'bg-emerald-100 text-emerald-700'
                        }`}
                      >
                        {deviceStatusText[device.Status ?? ''] || device.Status || '未知'}
                      </span>
                    </div>
                    <div className="truncate text-sm font-semibold text-slate-900">
                      {device.Name || `设备 ${device.Id}`}
                    </div>
                    <div className="mt-1 truncate text-xs text-slate-500">编号 {device.IDNo || '--'}</div>
                  </button>
                )
              })}
            </div>
          ) : (
            <EmptyState text="暂无设备数据" />
          )}
        </Panel>
      </section>
    </div>
  )
}

function Panel({
  title,
  action,
  onAction,
  className = '',
  icon,
  children,
}: {
  title: string
  action: string
  onAction: () => void
  className?: string
  icon?: ReactNode
  children: ReactNode
}) {
  return (
    <section className={`rounded-2xl border border-slate-200 bg-white p-5 shadow-sm ${className}`}>
      <div className="mb-4 flex items-center justify-between gap-4">
        <div className="flex items-center gap-2">
          <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-teal-50 text-teal-600">
            {icon}
          </span>
          <h2 className="text-sm font-bold text-slate-900">{title}</h2>
        </div>
        <button
          type="button"
          onClick={onAction}
          className="inline-flex items-center gap-1 rounded-lg px-2 py-1 text-xs font-medium text-teal-700 transition hover:bg-teal-50 active:scale-[0.98]"
        >
          {action} <ArrowRight size={13} />
        </button>
      </div>
      {children}
    </section>
  )
}

function EmptyState({
  text,
  hint,
  action,
  onAction,
}: {
  text: string
  hint?: string
  action?: string
  onAction?: () => void
}) {
  return (
    <div className="flex min-h-32 flex-col items-center justify-center rounded-xl border border-dashed border-slate-200 bg-slate-50 px-4 text-center">
      <div className="text-sm font-semibold text-slate-500">{text}</div>
      {hint && <div className="mt-2 text-xs text-slate-400">{hint}</div>}
      {action && onAction && (
        <button
          type="button"
          onClick={onAction}
          className="mt-4 inline-flex items-center gap-1.5 rounded-xl bg-blue-600 px-4 py-2 text-xs font-bold text-white transition hover:bg-blue-700"
        >
          {action} <ArrowRight size={13} />
        </button>
      )}
    </div>
  )
}

function SkeletonRows() {
  return (
    <div className="space-y-3" aria-live="polite">
      {[0, 1, 2].map((item) => (
        <div key={item} className="h-14 animate-pulse rounded-lg bg-slate-100" />
      ))}
    </div>
  )
}
