/**
 * 治疗相关服务
 */

import type {
  Treatment,
  PatientPrescription,
  PatientPlan,
  PaginatedResponse,
} from './types/api'
import {
  fetchPaginatedData,
  fetchFilteredData,
  getTodayString,
} from './api'
import { apiCache, cacheKey, CACHE_TTL } from '@/utils/cache'

// ============ 字段定义 ============

const TREATMENT_FIELDS = [
  'Id', 'PatientId', 'TreatmentDate', 'ShiftId', 'BedNo',
  'StartTime', 'EndTime', 'Status', 'DoctorId', 'NurseId'
]

const PRESCRIPTION_FIELDS = [
  'Id', 'PatientId', 'TreatmentId', 'DialysisMode', 'DialyzerType',
  'BloodFlow', 'DialysateFlow', 'Duration', 'Anticoagulant', 'DryWeight'
]

const PLAN_FIELDS = [
  'Id', 'PatientId', 'PlanName', 'Frequency', 'Duration',
  'DialysisMode', 'Status', 'StartDate', 'EndDate'
]

// ============ 治疗记录服务 ============

export async function getTreatmentList(
  page: number = 1,
  pageSize: number = 50
): Promise<PaginatedResponse<Treatment>> {
  return fetchPaginatedData<Treatment>('Treatment', TREATMENT_FIELDS, { page, pageSize })
}

export async function getTodayTreatments(): Promise<Treatment[]> {
  const today = getTodayString()
  const key = cacheKey('treatment:today', today)
  return apiCache.withCache(key, async () => {
    const result = await fetchFilteredData<Treatment>(
      'Treatment',
      TREATMENT_FIELDS,
      { TreatmentDate: today },
      1,
      200
    )
    return result.data
  }, CACHE_TTL.TODAY_TREATMENTS)
}

export async function getPatientTreatments(
  patientId: string,
  page: number = 1,
  pageSize: number = 20
): Promise<PaginatedResponse<Treatment>> {
  const key = cacheKey('treatment:patient', patientId, page, pageSize)
  return apiCache.withCache(key, () =>
    fetchFilteredData<Treatment>(
      'Treatment',
      TREATMENT_FIELDS,
      { PatientId: patientId },
      page,
      pageSize
    ),
    CACHE_TTL.PATIENT_TREATMENTS
  )
}

export async function getOngoingTreatments(): Promise<Treatment[]> {
  const today = getTodayString()
  const result = await fetchFilteredData<Treatment>(
    'Treatment',
    TREATMENT_FIELDS,
    { TreatmentDate: today },
    1,
    200
  )
  return result.data.filter(t =>
    t.Status === 'ongoing' || t.Status === '1' || t.Status === '进行中'
  )
}

export async function getTreatmentsByDateRange(
  startDate: string,
  endDate: string,
  patientId?: string
): Promise<Treatment[]> {
  const filters: Record<string, string | number> = {}
  if (patientId) filters.PatientId = patientId

  const result = await fetchFilteredData<Treatment>(
    'Treatment',
    TREATMENT_FIELDS,
    filters,
    1,
    500
  )

  return result.data.filter(t => {
    const date = t.TreatmentDate
    return date && date >= startDate && date <= endDate
  })
}

// ============ 透析处方服务 ============

export async function getPatientPrescriptions(
  patientId: string,
  page: number = 1,
  pageSize: number = 20
): Promise<PaginatedResponse<PatientPrescription>> {
  return fetchFilteredData<PatientPrescription>(
    'PatientPrescription',
    PRESCRIPTION_FIELDS,
    { PatientId: patientId },
    page,
    pageSize
  )
}

export async function getTodayPrescriptions(): Promise<PatientPrescription[]> {
  const treatments = await getTodayTreatments()
  if (treatments.length === 0) return []

  const result = await fetchPaginatedData<PatientPrescription>(
    'PatientPrescription',
    PRESCRIPTION_FIELDS,
    { page: 1, pageSize: 500 }
  )

  const treatmentIds = new Set(treatments.map(t => t.Id))
  return result.data.filter(p => p.TreatmentId && treatmentIds.has(p.TreatmentId))
}

export async function getPrescriptionByTreatment(
  treatmentId: number
): Promise<PatientPrescription | null> {
  const result = await fetchFilteredData<PatientPrescription>(
    'PatientPrescription',
    PRESCRIPTION_FIELDS,
    { TreatmentId: treatmentId },
    1,
    1
  )
  return result.data[0] || null
}

// ============ 透析方案服务 ============

export async function getPatientPlans(
  patientId: string
): Promise<PatientPlan[]> {
  const result = await fetchFilteredData<PatientPlan>(
    'PatientPlan',
    PLAN_FIELDS,
    { PatientId: patientId },
    1,
    50
  )
  return result.data
}

export async function getCurrentPlan(
  patientId: string
): Promise<PatientPlan | null> {
  const plans = await getPatientPlans(patientId)
  return plans.find(p =>
    p.Status === '1' || p.Status === 'active' || p.Status === '生效中'
  ) || null
}

// ============ 统计 ============

export interface TreatmentStats {
  total: number
  ongoing: number
  completed: number
  byShift: Record<number, number>
}

export async function getTodayTreatmentStats(): Promise<TreatmentStats> {
  const treatments = await getTodayTreatments()

  const stats: TreatmentStats = {
    total: treatments.length,
    ongoing: 0,
    completed: 0,
    byShift: {},
  }

  treatments.forEach(t => {
    if (t.Status === 'ongoing' || t.Status === '1' || t.Status === '进行中') {
      stats.ongoing++
    } else if (t.Status === 'completed' || t.Status === '2' || t.Status === '已完成') {
      stats.completed++
    }

    if (t.ShiftId) {
      stats.byShift[t.ShiftId] = (stats.byShift[t.ShiftId] || 0) + 1
    }
  })

  return stats
}
