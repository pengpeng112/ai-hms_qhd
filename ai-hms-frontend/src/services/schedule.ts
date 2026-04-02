/**
 * 排班相关服务
 */

import type {
  Shift,
  PatientShift,
  Bed,
  Ward,
  PaginatedResponse,
} from './types/api'
import {
  fetchPaginatedData,
  fetchListData,
  fetchFilteredData,
  getTodayString,
} from './api'
import { apiCache, cacheKey, CACHE_TTL } from '@/utils/cache'

// ============ 字段定义 ============

// 根据 HDIS API 文档 4.3.20
const SHIFT_FIELDS = ['Id', 'TenantId', 'Name', 'Sort', 'StartTime', 'EndTime', 'Type', 'Note']

// 根据 HDIS API 文档 4.3.21
const PATIENT_SHIFT_FIELDS = [
  'Id', 'TenantId', 'PatientId', 'TreatmentTime', 'ShiftId', 'WardId', 'BedId', 'PatientPlanId'
]

// 根据 HDIS API 文档 4.3.22
const BED_FIELDS = ['Id', 'TenantId', 'Name', 'WardId', 'Sort', 'Status', 'Note']

// 根据 HDIS API 文档 4.3.24
const WARD_FIELDS = ['Id', 'TenantId', 'Name', 'Sort', 'Status', 'Note']

// ============ 班次服务 ============

export async function getShiftList(
  page: number = 1,
  pageSize: number = 50
): Promise<PaginatedResponse<Shift>> {
  const key = cacheKey('shift:list', page, pageSize)
  return apiCache.withCache(key, () =>
    fetchPaginatedData<Shift>('Shift', SHIFT_FIELDS, { page, pageSize }),
    CACHE_TTL.SHIFT_LIST
  )
}

export async function getActiveShifts(): Promise<Shift[]> {
  const result = await getShiftList(1, 100)
  return result.data.filter(s => s.Status === '1' || s.Status === 'active')
}

// ============ 患者排班服务 ============

export async function getPatientShiftList(
  page: number = 1,
  pageSize: number = 50
): Promise<PaginatedResponse<PatientShift>> {
  return fetchPaginatedData<PatientShift>('PatientShift', PATIENT_SHIFT_FIELDS, { page, pageSize })
}

export async function getTodaySchedule(): Promise<PatientShift[]> {
  const today = getTodayString()
  const key = cacheKey('schedule:today', today)
  return apiCache.withCache(key, async () => {
    const result = await fetchFilteredData<PatientShift>(
      'PatientShift',
      PATIENT_SHIFT_FIELDS,
      { ScheduleDate: today },
      1,
      200
    )
    return result.data
  }, CACHE_TTL.TODAY_SCHEDULE)
}

export async function getScheduleByDate(date: string): Promise<PatientShift[]> {
  const result = await fetchFilteredData<PatientShift>(
    'PatientShift',
    PATIENT_SHIFT_FIELDS,
    { ScheduleDate: date },
    1,
    200
  )
  return result.data
}

export async function getScheduleByShift(
  shiftId: number,
  date?: string
): Promise<PatientShift[]> {
  const scheduleDate = date || getTodayString()
  const result = await fetchFilteredData<PatientShift>(
    'PatientShift',
    PATIENT_SHIFT_FIELDS,
    { ShiftId: shiftId, ScheduleDate: scheduleDate },
    1,
    100
  )
  return result.data
}

export async function getPatientSchedule(
  patientId: number,
  page: number = 1,
  pageSize: number = 20
): Promise<PaginatedResponse<PatientShift>> {
  return fetchFilteredData<PatientShift>(
    'PatientShift',
    PATIENT_SHIFT_FIELDS,
    { PatientId: patientId },
    page,
    pageSize
  )
}

// ============ 床位服务 ============

export async function getBedList(
  page: number = 1,
  pageSize: number = 100
): Promise<PaginatedResponse<Bed>> {
  const key = cacheKey('bed:list', page, pageSize)
  return apiCache.withCache(key, () =>
    fetchPaginatedData<Bed>('Bed', BED_FIELDS, { page, pageSize }),
    CACHE_TTL.BED_LIST
  )
}

export async function getBedsByWard(wardId: number): Promise<Bed[]> {
  const result = await fetchFilteredData<Bed>(
    'Bed',
    BED_FIELDS,
    { WardId: wardId },
    1,
    100
  )
  return result.data
}

export async function getAvailableBeds(): Promise<Bed[]> {
  const result = await getBedList(1, 200)
  return result.data.filter(b => b.Status === '1' || b.Status === 'available')
}

// ============ 病区服务 ============

export async function getWardList(): Promise<Ward[]> {
  const key = cacheKey('ward:list')
  return apiCache.withCache(key, () =>
    fetchListData<Ward>('Ward', WARD_FIELDS),
    CACHE_TTL.WARD_LIST
  )
}

export async function getActiveWards(): Promise<Ward[]> {
  const wards = await getWardList()
  return wards.filter(w => w.Status === '1' || w.Status === 'active')
}

// ============ 综合查询 ============

export interface ScheduleOverview {
  shifts: Shift[]
  todaySchedule: PatientShift[]
  totalPatients: number
  byShift: Record<number, number>
}

export async function getTodayScheduleOverview(): Promise<ScheduleOverview> {
  const [shifts, todaySchedule] = await Promise.all([
    getActiveShifts(),
    getTodaySchedule(),
  ])

  const byShift: Record<number, number> = {}
  todaySchedule.forEach(ps => {
    const shiftId = ps.ShiftId
    byShift[shiftId] = (byShift[shiftId] || 0) + 1
  })

  return {
    shifts,
    todaySchedule,
    totalPatients: todaySchedule.length,
    byShift,
  }
}
