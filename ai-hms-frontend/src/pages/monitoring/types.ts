import type { MonitorDevice } from '@/types/original'

export type ModalType = 'COMPREHENSIVE' | 'PRESCRIPTION' | 'ORDERS' | 'SUMMARY' | null

export type BedDisplayStatus = 'empty' | 'active' | 'warning' | 'danger' | 'offline'
export type StatusFilter = 'ALL' | 'alerts' | 'empty' | 'active' | 'warning' | 'danger' | 'offline'

export function classifyBedStatus(device: MonitorDevice): BedDisplayStatus {
  if (device.status === 'offline') return 'offline'
  if (!device.patientName && !device.patientId) return 'empty'
  if (device.alarmLevel === 'danger') return 'danger'
  if (device.alarmLevel === 'warning') return 'warning'
  if (device.status === 'alarm') return 'danger'
  if (device.status === 'warning') return 'warning'
  return 'active'
}

export const NA_CONDUCTIVITY_FACTOR = 9.9

export function computeMAP(device: MonitorDevice): number {
  const { sbp, dbp } = device.vitals
  return sbp > 0 && dbp > 0 ? Math.round((sbp + 2 * dbp) / 3) : 0
}

export function computeNa(device: MonitorDevice): number {
  const c = device.vitals.conductivity
  return c > 0 ? Math.round(c * NA_CONDUCTIVITY_FACTOR * 10) / 10 : 0
}

export function alertLevelFor(device: MonitorDevice, metric: string): 'warning' | 'danger' | undefined {
  const a = device.alerts?.find((x) => x.metric === metric)
  return (a?.level as 'warning' | 'danger') || undefined
}

export const MONITOR_METRIC_LABELS: Record<string, string> = {
  map: '平均压',
  heartRate: '心率',
  vp: '静脉压',
  dialysateNa: '透析液钠',
  ufr: '超滤率',
}

export function formatTimeProgress(device: MonitorDevice): { elapsed: string; planned: string } | null {
  if (!device.startTime || !device.estimatedDuration) return null
  const planMin = device.estimatedDuration
  const elapsedMin = Math.max(0, Math.min(planMin, (Date.now() - new Date(device.startTime).getTime()) / 60000))
  const fmt = (m: number) => `${Math.floor(m / 60)}:${String(Math.round(m % 60)).padStart(2, '0')}`
  return { elapsed: fmt(elapsedMin), planned: fmt(planMin) }
}

export interface MonitorSummary {
  total: number
  active: number
  empty: number
  warning: number
  danger: number
  offline: number
}

export function computeMonitorSummary(devices: MonitorDevice[]): MonitorSummary {
  const summary: MonitorSummary = { total: 0, active: 0, empty: 0, warning: 0, danger: 0, offline: 0 }
  for (const d of devices) {
    summary.total++
    switch (classifyBedStatus(d)) {
      case 'active': summary.active++; break
      case 'empty': summary.empty++; break
      case 'warning': summary.warning++; break
      case 'danger': summary.danger++; break
      case 'offline': summary.offline++; break
    }
  }
  return summary
}

export type TrendChartRow = { time: string; ts: number; kind: string } & Record<string, number | string | null>

const TREND_SERIES_TO_CHART: Record<string, string> = {
  sbp: 'sbp', dbp: 'dbp', map: 'map', heartRate: 'hr',
  ap: 'ap', vp: 'vp', tmp: 'tmp', bf: 'bf', ufVolume: 'uf', conductivity: 'na',
}

export function trendToChartRows(series: Record<string, { t: string; v: number; kind: string }[]>): TrendChartRow[] {
  const byTs = new Map<number, TrendChartRow>()
  for (const [seriesKey, chartKey] of Object.entries(TREND_SERIES_TO_CHART)) {
    for (const p of series[seriesKey] || []) {
      const ts = new Date(p.t).getTime()
      if (Number.isNaN(ts)) continue
      let row = byTs.get(ts)
      if (!row) {
        row = {
          ts, kind: p.kind,
          time: new Date(ts).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' }),
          sbp: null, dbp: null, map: null, hr: null, ap: null, vp: null, tmp: null, bf: null, uf: null, na: null,
        }
        byTs.set(ts, row)
      }
      row[chartKey] = chartKey === 'na' ? Math.round(p.v * NA_CONDUCTIVITY_FACTOR * 10) / 10 : p.v
    }
  }
  return Array.from(byTs.values()).sort((a, b) => a.ts - b.ts)
}

export type MiniGraphPoint = { sbp: number; hr: number }

export type HistoryPoint = {
  time: string
  sbp: number
  dbp: number
  hr: number
  ap: number
  vp: number
  tmp: number
  bf: number
  uf: number
}

export type DeviceAssignment = {
  patientId: string
  patientName: string
  mode: string
}

// 缓存设备图表数据（按需填充）
export const cachedGraphData = new Map<string, MiniGraphPoint[]>()
export const cachedHistoryData = new Map<string, HistoryPoint[]>()

export function ensureDeviceCache(device: MonitorDevice) {
  if (!cachedGraphData.has(device.id)) {
    cachedGraphData.set(device.id, [])
  }
  if (!cachedHistoryData.has(device.id)) {
    cachedHistoryData.set(device.id, [])
  }
}

export function formatPositive(value: number, suffix = '') {
  return value > 0 ? `${value}${suffix}` : '--'
}

export function formatBloodPressure(device: MonitorDevice) {
  return device.vitals.sbp > 0 && device.vitals.dbp > 0
    ? `${device.vitals.sbp}/${device.vitals.dbp}`
    : '--'
}

export function buildDeviceAssignments(items: { name?: string; bedNumber?: string; status?: string; id: string; defaultMode?: string }[]): Map<string, DeviceAssignment> {
  const assignments = new Map<string, DeviceAssignment>()

  items.forEach((item) => {
    const patientName = item.name?.trim()
    const bedNumber = item.bedNumber?.trim()
    if (!patientName || !bedNumber) {
      return
    }

    const normalizedStatus = item.status?.trim().toLowerCase()
    if (normalizedStatus === 'discharged') {
      return
    }

    assignments.set(bedNumber, {
      patientId: item.id,
      patientName,
      mode: item.defaultMode?.trim() || '',
    })
  })

  return assignments
}

export function toMonitorDevice(d: { id: string; bedNumber?: string; name: string; status: string }, assignment?: DeviceAssignment): MonitorDevice {
  const statusMap: Record<string, MonitorDevice['status']> = {
    normal: 'normal',
    warning: 'warning',
    alarm: 'alarm',
    offline: 'offline',
    maintenance: 'offline',
  }
  const status = statusMap[d.status] ?? 'unknown'
  return {
    id: d.id,
    bedNumber: d.bedNumber || d.name,
    patientName: assignment?.patientName || '',
    patientId: assignment?.patientId,
    status,
    mode: assignment?.mode || '',
    timeRemaining: '--',
    vitals: {
      sbp: 0,
      dbp: 0,
      hr: 0,
      bf: 0,
      tmp: 0,
      ufGoal: 0,
      ufVolume: 0,
      conductivity: 0,
      temp: 0,
    },
    alarms: [],
  }
}
