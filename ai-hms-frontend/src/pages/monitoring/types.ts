import type { MonitorDevice } from '@/types/original'

export type ModalType = 'COMPREHENSIVE' | 'PRESCRIPTION' | 'ORDERS' | 'SUMMARY' | null

export type BedDisplayStatus = 'empty' | 'active' | 'warning' | 'danger' | 'offline'
export type StatusFilter = 'ALL' | 'empty' | 'active' | 'warning' | 'danger' | 'offline'

export function classifyBedStatus(device: MonitorDevice): BedDisplayStatus {
  if (device.status === 'offline') return 'offline'
  if (device.status === 'alarm') return 'danger'
  if (device.status === 'warning') return 'warning'
  if (!device.patientName && !device.patientId) return 'empty'
  return 'active'
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
