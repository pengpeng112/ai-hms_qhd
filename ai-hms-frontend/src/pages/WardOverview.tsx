/**
 * 病区概览页面（科室运行全景）
 *
 * 提供科室运行驾驶舱视图，包括：
 * - 页头卡片（标题、更新时间、刷新、在科人数、报警提示）
 * - 核心 KPI 卡片（排班、工作量、设备、时长）
 * - 治疗进度统计（进度条 + 空状态）
 * - 设备与床位运行矩阵（自适应列）
 * - 底部辅助区（重点提醒、设备利用、快捷操作）
 *
 * 改造原则：不修改接口、字段、数据结构，不引入模拟数据。
 */

import { useState, useEffect, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { useNavigate } from 'react-router-dom'
import {
  Users, Monitor, Activity, CheckCircle2,
  AlertCircle, Clock, Battery, RefreshCw,
  Stethoscope, Calendar, ChevronRight,
} from 'lucide-react'
import { restApi } from '@/services'
import { getAllEquipments } from '@/services/equipment'
import { getMonitoringLiveData, type RestMonitoringLiveData } from '@/services/monitoringApi'

// 治疗状态类型
type TreatmentStatus = 'waiting' | 'dialysis' | 'disinfect' | 'completed'

interface ProcessDataItem {
  name: string
  value: number
  status: TreatmentStatus
}

interface BedStatus {
  id: string
  label: string
  status: 'active' | 'empty' | 'alarm' | 'maintenance'
}

// 进度条颜色映射
const PROCESS_COLORS: Record<TreatmentStatus, string> = {
  waiting: 'bg-blue-500',
  dialysis: 'bg-emerald-500',
  disinfect: 'bg-amber-500',
  completed: 'bg-violet-500',
}

// ─── KPI 卡片组件 ───
function KpiCard({
  title,
  value,
  suffix,
  description,
  icon,
  tone,
}: {
  title: string
  value: React.ReactNode
  suffix?: React.ReactNode
  description?: string
  icon: React.ReactNode
  tone: 'blue' | 'emerald' | 'purple' | 'orange'
}) {
  const toneMap = {
    blue: { bar: 'bg-blue-600', iconBg: 'bg-blue-50', iconText: 'text-blue-600' },
    emerald: { bar: 'bg-emerald-600', iconBg: 'bg-emerald-50', iconText: 'text-emerald-600' },
    purple: { bar: 'bg-violet-600', iconBg: 'bg-violet-50', iconText: 'text-violet-600' },
    orange: { bar: 'bg-orange-500', iconBg: 'bg-orange-50', iconText: 'text-orange-500' },
  }[tone]

  return (
    <div className="relative h-[100px] overflow-hidden rounded-[20px] border border-slate-200 bg-white p-5 shadow-sm">
      <div className={`absolute inset-x-0 top-0 h-1 ${toneMap.bar}`} />
      <div className="flex items-center gap-4">
        <div className={`flex h-[52px] w-[52px] shrink-0 items-center justify-center rounded-2xl ${toneMap.iconBg} ${toneMap.iconText}`}>
          {icon}
        </div>
        <div className="min-w-0">
          <div className="text-[13px] font-extrabold text-slate-500">{title}</div>
          <div className="mt-1 text-[30px] font-black leading-none text-slate-950">
            {value}
            {suffix && <span className="ml-1 text-sm font-bold text-slate-400">{suffix}</span>}
          </div>
          {description && (
            <div className="mt-1.5 text-xs font-bold text-slate-400">{description}</div>
          )}
        </div>
      </div>
    </div>
  )
}

// ─── 进度条行组件 ───
function ProgressRow({
  label,
  value,
  total,
  colorClass,
}: {
  label: string
  value: number
  total: number
  colorClass: string
}) {
  const percent = total > 0 ? Math.round((value / total) * 100) : 0

  return (
    <div>
      <div className="mb-2 flex items-center justify-between text-sm">
        <span className="flex items-center gap-2 font-extrabold text-slate-700">
          <span className={`h-2.5 w-2.5 rounded-full ${colorClass}`} />
          {label}
        </span>
        <span className="font-black text-slate-950">{value}人</span>
      </div>
      <div className="h-2.5 overflow-hidden rounded-full bg-slate-100">
        <div className={`h-2.5 rounded-full ${colorClass} transition-all duration-500`} style={{ width: `${percent}%` }} />
      </div>
    </div>
  )
}

// ─── 床位状态卡片组件 ───
function getBedStyle(status: BedStatus['status']) {
  switch (status) {
    case 'active':
      return 'border-blue-200 bg-blue-50 text-blue-600'
    case 'alarm':
      return 'border-rose-200 bg-rose-50 text-rose-600 animate-pulse'
    case 'maintenance':
      return 'border-orange-200 bg-orange-50 text-orange-600'
    default:
      return 'border-slate-200 bg-slate-50 text-slate-500'
  }
}

function BedStatusCard({ bed, statusLabel }: { bed: BedStatus; statusLabel: string }) {
  return (
    <div
      className={`h-[58px] rounded-[14px] border px-3 py-2 transition hover:-translate-y-0.5 hover:shadow-md ${getBedStyle(bed.status)}`}
      title={`${bed.label} - ${statusLabel}`}
    >
      <div className="flex items-center justify-between">
        <span className="text-[13px] font-black">{bed.label}</span>
        <span className="text-[11px] font-extrabold">{statusLabel}</span>
      </div>
      <Monitor size={16} className="mx-auto mt-0.5 opacity-80" />
    </div>
  )
}

// ─── 状态 Chip 组件 ───
function StatusChip({ label, count, tone }: { label: string; count: number; tone: string }) {
  const toneMap: Record<string, string> = {
    blue: 'bg-blue-50 text-blue-700 border-blue-200',
    slate: 'bg-slate-50 text-slate-600 border-slate-200',
    rose: 'bg-rose-50 text-rose-700 border-rose-200',
    orange: 'bg-orange-50 text-orange-700 border-orange-200',
  }

  return (
    <span className={`inline-flex items-center gap-1.5 rounded-lg border px-2.5 py-1 text-xs font-bold ${toneMap[tone]}`}>
      {label}
      <span className="font-black">{count}</span>
    </span>
  )
}

// ─── 快捷操作按钮 ───
function QuickActionButton({
  icon,
  label,
  onClick,
}: {
  icon: React.ReactNode
  label: string
  onClick: () => void
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="flex w-full items-center gap-3 rounded-xl border border-slate-200 bg-white px-4 py-3 text-left transition hover:border-blue-300 hover:bg-blue-50"
    >
      <span className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-blue-50 text-blue-600">
        {icon}
      </span>
      <span className="flex-1 text-sm font-bold text-slate-700">{label}</span>
      <ChevronRight size={16} className="text-slate-400" />
    </button>
  )
}

export default function WardOverview() {
  const { t, i18n } = useTranslation('wardOverview')
  const navigate = useNavigate()
  const [loading, setLoading] = useState(true)
  const [lastUpdate, setLastUpdate] = useState(new Date())

  // 统计数据
  const [stats, setStats] = useState({
    scheduledPatients: 0,
    totalCapacity: 50,
    workCompletion: 0,
    deviceUtilization: 0,
    avgDialysisHours: 0,
    onSiteCount: 0,
    alarmDevices: 0,
  })

  // 治疗进度数据
  const [processData, setProcessData] = useState<ProcessDataItem[]>([
    { name: t('process.waiting'), value: 0, status: 'waiting' },
    { name: t('process.dialysis'), value: 0, status: 'dialysis' },
    { name: t('process.disinfect'), value: 0, status: 'disinfect' },
    { name: t('process.completed'), value: 0, status: 'completed' },
  ])

  // 床位状态数据
  const [bedStatuses, setBedStatuses] = useState<BedStatus[]>([])

  // 加载数据 — 保持原有调用逻辑不变
  const loadData = useCallback(async () => {
    setLoading(true)
    try {
      const [patientsRes, equipmentsRes, dashboardRes, liveData] = await Promise.all([
        restApi.getPatientList({ page: 1, pageSize: 100 }).catch(() => null),
        getAllEquipments().catch(() => null),
        restApi.getDashboardStats().catch(() => null),
        getMonitoringLiveData().catch(() => [] as RestMonitoringLiveData[]),
      ])

      const scheduledPatients = dashboardRes?.todaySchedules || 0
      const totalEquipments = equipmentsRes?.length || dashboardRes?.equipmentCount || 0
      const totalPatients = patientsRes?.data?.pagination?.total || dashboardRes?.activePatients || 0

      const activeCount = dashboardRes?.todaySchedules || scheduledPatients
      const alarmCount = dashboardRes?.alertItems || 0
      const liveArray = Array.isArray(liveData) ? liveData : []

      setStats({
        scheduledPatients: activeCount,
        totalCapacity: totalEquipments,
        workCompletion: dashboardRes?.todayTreatments ? Math.round((dashboardRes.todayTreatments / Math.max(activeCount, 1)) * 100) : 0,
        deviceUtilization: totalEquipments > 0
          ? Math.round((activeCount / totalEquipments) * 1000) / 10
          : 0,
        avgDialysisHours: liveArray.length > 0
          ? liveArray.reduce((s: number, d: RestMonitoringLiveData) => s + (d.estimatedDuration || 0), 0) / liveArray.length / 60
          : 0,
        onSiteCount: totalPatients,
        alarmDevices: alarmCount,
      })

      const waitingCount = Math.max(0, activeCount - (dashboardRes?.todayTreatments || 0))
      const completedCount = dashboardRes?.todayTreatments || 0

      setProcessData([
        { name: t('process.waiting'), value: waitingCount, status: 'waiting' },
        { name: t('process.dialysis'), value: activeCount, status: 'dialysis' },
        { name: t('process.disinfect'), value: 0, status: 'disinfect' },
        { name: t('process.completed'), value: completedCount, status: 'completed' },
      ])

      const beds: BedStatus[] = liveArray.map((d, i) => ({
        id: `bed-${d.bedId || i}-${i}`,
        label: d.bedName || `A${String(i + 1).padStart(2, '0')}`,
        status: 'active' as BedStatus['status'],
      }))
      const totalBeds = totalEquipments > liveArray.length ? totalEquipments : liveArray.length
      for (let i = beds.length; i < totalBeds; i++) {
        beds.push({
          id: `bed-empty-${i}`,
          label: `A${String(i + 1).padStart(2, '0')}`,
          status: 'empty',
        })
      }
      setBedStatuses(beds)

      setLastUpdate(new Date())
    } catch (error) {
      console.error(t('loadError'), error)
    } finally {
      setLoading(false)
    }
  }, [t])

  useEffect(() => {
    loadData()
  }, [loadData])

  // 格式化时间
  const formatTime = (date: Date) => {
    const locale = i18n.language === 'en-US' ? 'en-US' : 'zh-CN'
    return date.toLocaleString(locale, {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  // 派生变量 — 基于真实数据计算
  const activeBeds = bedStatuses.filter(b => b.status === 'active').length
  const emptyBeds = bedStatuses.filter(b => b.status === 'empty').length
  const alarmBeds = bedStatuses.filter(b => b.status === 'alarm').length
  const maintenanceBeds = bedStatuses.filter(b => b.status === 'maintenance').length
  const totalProcessCount = processData.reduce((sum, item) => sum + item.value, 0)
  const dialysisInProgress = processData.find(p => p.status === 'dialysis')?.value || 0
  const utilizationPercent = bedStatuses.length > 0 ? Math.round((activeBeds / bedStatuses.length) * 100) : 0

  return (
    <div className="flex h-full flex-col gap-4 max-w-[1600px] mx-auto pb-8">
      {/* ═══ 页头卡片 ═══ */}
      <section className="rounded-[22px] border border-blue-100 bg-white px-6 py-4 shadow-sm">
        <div className="flex items-center justify-between gap-6">
          <div className="min-w-0">
            <h1 className="text-[28px] font-black tracking-tight text-slate-950">{t('title')}</h1>
            <p className="mt-1 text-sm font-semibold text-slate-500">
              {t('lastUpdate', { time: formatTime(lastUpdate) })} · 第一病区 · 今日运行总览
            </p>
          </div>

          <div className="flex shrink-0 items-center gap-3">
            <button
              type="button"
              onClick={loadData}
              disabled={loading}
              className="flex h-[42px] items-center rounded-[13px] border border-slate-200 bg-white px-4 text-sm font-black text-slate-700 shadow-sm transition hover:bg-slate-50 disabled:opacity-50"
            >
              <RefreshCw size={15} className={`mr-2 ${loading ? 'animate-spin' : ''}`} />
              {t('refresh')}
            </button>

            <div className="flex h-[42px] items-center rounded-[13px] border border-emerald-200 bg-emerald-50 px-4 text-sm font-black text-emerald-700 shadow-sm">
              <span className="mr-2 h-2 w-2 animate-pulse rounded-full bg-emerald-500" />
              {t('onSiteCount', { count: stats.onSiteCount })}
            </div>

            {stats.alarmDevices > 0 && (
              <div className="flex h-[42px] items-center rounded-[13px] border border-rose-200 bg-rose-50 px-4 text-sm font-black text-rose-600 shadow-sm">
                <AlertCircle size={15} className="mr-2" />
                {t('alarmDevices', { count: stats.alarmDevices })}
              </div>
            )}
          </div>
        </div>
      </section>

      {/* ═══ 核心 KPI 区 ═══ */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <KpiCard
          title={t('stats.scheduledPatients')}
          value={stats.scheduledPatients}
          suffix={` / ${stats.totalCapacity}`}
          icon={<Users size={24} />}
          tone="blue"
        />
        <KpiCard
          title={t('stats.workCompletion')}
          value={stats.workCompletion}
          suffix="%"
          icon={<CheckCircle2 size={24} />}
          tone="emerald"
        />
        <KpiCard
          title={t('stats.deviceUtilization')}
          value={stats.deviceUtilization}
          suffix="%"
          icon={<Battery size={24} />}
          tone="purple"
        />
        <KpiCard
          title={t('stats.avgDialysisHours')}
          value={stats.avgDialysisHours > 0 ? stats.avgDialysisHours.toFixed(1) : '0'}
          suffix="h"
          icon={<Clock size={24} />}
          tone="orange"
        />
      </div>

      {/* ═══ 主体运行区 ═══ */}
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-3">
        {/* 左侧：治疗进度统计 */}
        <div className="rounded-[20px] border border-slate-200 bg-white p-5 shadow-sm">
          <h3 className="mb-4 flex items-center gap-2 text-sm font-black text-slate-800">
            <Activity size={18} className="text-blue-600" />
            {t('chart.treatmentProgress')}
          </h3>

          {totalProcessCount === 0 ? (
            <div className="flex h-48 flex-col items-center justify-center gap-2 text-center">
              <div className="rounded-full bg-slate-100 p-4">
                <Activity size={24} className="text-slate-300" />
              </div>
              <p className="text-sm font-semibold text-slate-500">{t('process.emptyState')}</p>
            </div>
          ) : (
            <div className="space-y-4">
              {processData.map((item) => (
                <ProgressRow
                  key={item.status}
                  label={item.name}
                  value={item.value}
                  total={totalProcessCount}
                  colorClass={PROCESS_COLORS[item.status]}
                />
              ))}
            </div>
          )}
        </div>

        {/* 右侧：设备与床位运行矩阵 */}
        <div className="rounded-[20px] border border-slate-200 bg-white p-5 shadow-sm lg:col-span-2">
          <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
            <h3 className="flex items-center gap-2 text-sm font-black text-slate-800">
              <Monitor size={18} className="text-blue-600" />
              {t('chart.equipmentMatrix')}
            </h3>

            {/* 状态 Chip */}
            <div className="flex flex-wrap items-center gap-2">
              <StatusChip label={t('bedStatus.active')} count={activeBeds} tone="blue" />
              <StatusChip label={t('bedStatus.empty')} count={emptyBeds} tone="slate" />
              <StatusChip label={t('bedStatus.alarm')} count={alarmBeds} tone="rose" />
              <StatusChip label={t('bedStatus.maintenance')} count={maintenanceBeds} tone="orange" />
            </div>
          </div>

          {/* 床位矩阵 — auto-fill 自适应列 */}
          {bedStatuses.length === 0 ? (
            <div className="flex h-40 flex-col items-center justify-center gap-2 text-center">
              <div className="rounded-full bg-slate-100 p-4">
                <Monitor size={24} className="text-slate-300" />
              </div>
              <p className="text-sm font-semibold text-slate-400">暂无设备数据</p>
            </div>
          ) : (
            <div className="grid gap-3 [grid-template-columns:repeat(auto-fill,minmax(96px,1fr))]">
              {bedStatuses.map((bed) => (
                <BedStatusCard key={bed.id} bed={bed} statusLabel={t(`bedStatus.${bed.status}`)} />
              ))}
            </div>
          )}
        </div>
      </div>

      {/* ═══ 底部辅助区 ═══ */}
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-3">
        {/* 重点提醒 */}
        <div className="rounded-[20px] border border-slate-200 bg-white p-5 shadow-sm">
          <h3 className="mb-4 flex items-center gap-2 text-sm font-black text-slate-800">
            <AlertCircle size={18} className="text-amber-500" />
            {t('section.keyReminders')}
          </h3>
          <div className="space-y-3">
            {alarmBeds > 0 || stats.alarmDevices > 0 ? (
              <div className="flex items-start gap-2 rounded-lg border border-rose-200 bg-rose-50 px-3 py-2.5 text-sm font-semibold text-rose-700">
                <AlertCircle size={16} className="mt-0.5 shrink-0" />
                <span>{t('reminder.alarmWarning', { count: Math.max(alarmBeds, stats.alarmDevices) })}</span>
              </div>
            ) : (
              <div className="flex items-start gap-2 rounded-lg border border-emerald-200 bg-emerald-50 px-3 py-2.5 text-sm font-semibold text-emerald-700">
                <CheckCircle2 size={16} className="mt-0.5 shrink-0" />
                <span>{t('reminder.noAlarm')}</span>
              </div>
            )}

            {dialysisInProgress > 0 ? (
              <div className="flex items-start gap-2 rounded-lg border border-blue-200 bg-blue-50 px-3 py-2.5 text-sm font-semibold text-blue-700">
                <Activity size={16} className="mt-0.5 shrink-0" />
                <span>{t('reminder.dialysisInProgress', { count: dialysisInProgress })}</span>
              </div>
            ) : (
              <div className="flex items-start gap-2 rounded-lg border border-slate-200 bg-slate-50 px-3 py-2.5 text-sm font-semibold text-slate-500">
                <Clock size={16} className="mt-0.5 shrink-0" />
                <span>{t('reminder.noDialysis')}</span>
              </div>
            )}

            <div className="flex items-start gap-2 rounded-lg border border-slate-200 bg-slate-50 px-3 py-2.5 text-sm font-semibold text-slate-600">
              <Users size={16} className="mt-0.5 shrink-0" />
              <span>{t('reminder.scheduled', { scheduled: stats.scheduledPatients, total: stats.totalCapacity })}</span>
            </div>
          </div>
        </div>

        {/* 设备利用概览 */}
        <div className="rounded-[20px] border border-slate-200 bg-white p-5 shadow-sm">
          <h3 className="mb-4 flex items-center gap-2 text-sm font-black text-slate-800">
            <Monitor size={18} className="text-blue-600" />
            {t('section.deviceUsage')}
          </h3>
          <div className="space-y-3">
            <div className="flex items-center justify-between text-sm">
              <span className="font-bold text-slate-500">{t('usage.totalDevices', { count: bedStatuses.length })}</span>
              <span className="font-black text-slate-950">{utilizationPercent}%</span>
            </div>

            {/* 利用率进度条 */}
            <div className="h-3 overflow-hidden rounded-full bg-slate-100">
              <div
                className="h-3 rounded-full bg-blue-600 transition-all duration-500"
                style={{ width: `${utilizationPercent}%` }}
              />
            </div>

            {/* 明细 */}
            <div className="grid grid-cols-2 gap-2 pt-1 text-xs">
              <div className="flex items-center justify-between rounded-lg bg-blue-50 px-3 py-2">
                <span className="font-bold text-blue-600">{t('summary.active')}</span>
                <span className="font-black text-blue-700">{activeBeds}</span>
              </div>
              <div className="flex items-center justify-between rounded-lg bg-slate-50 px-3 py-2">
                <span className="font-bold text-slate-500">{t('summary.empty')}</span>
                <span className="font-black text-slate-600">{emptyBeds}</span>
              </div>
              <div className="flex items-center justify-between rounded-lg bg-rose-50 px-3 py-2">
                <span className="font-bold text-rose-600">{t('summary.alarm')}</span>
                <span className="font-black text-rose-700">{alarmBeds}</span>
              </div>
              <div className="flex items-center justify-between rounded-lg bg-orange-50 px-3 py-2">
                <span className="font-bold text-orange-600">{t('summary.maintenance')}</span>
                <span className="font-black text-orange-700">{maintenanceBeds}</span>
              </div>
            </div>
          </div>
        </div>

        {/* 快捷操作 */}
        <div className="rounded-[20px] border border-slate-200 bg-white p-5 shadow-sm">
          <h3 className="mb-4 flex items-center gap-2 text-sm font-black text-slate-800">
            <ChevronRight size={18} className="text-blue-600" />
            {t('section.quickActions')}
          </h3>
          <div className="space-y-2.5">
            <QuickActionButton
              icon={<Monitor size={18} />}
              label={t('quickAction.monitoring')}
              onClick={() => navigate('/monitoring')}
            />
            <QuickActionButton
              icon={<Calendar size={18} />}
              label={t('quickAction.schedule')}
              onClick={() => navigate('/schedule')}
            />
            <QuickActionButton
              icon={<Stethoscope size={18} />}
              label={t('quickAction.dialysis')}
              onClick={() => navigate('/dialysis-processing')}
            />
          </div>
        </div>
      </div>
    </div>
  )
}
