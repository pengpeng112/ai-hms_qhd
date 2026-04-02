/**
 * 体征监测服务
 */

import type {
  BeforeSigns,
  DuringSigns,
  AfterSigns,
  PaginatedResponse,
} from './types/api'
import { fetchFilteredData } from './api'

// ============ 字段定义 ============

const BEFORE_SIGNS_FIELDS = [
  'Id', 'TreatmentId', 'PatientId', 'Weight', 'Temperature',
  'SystolicBP', 'DiastolicBP', 'Pulse', 'RecordTime'
]

const DURING_SIGNS_FIELDS = [
  'Id', 'TreatmentId', 'PatientId', 'SystolicBP', 'DiastolicBP',
  'Pulse', 'BloodFlow', 'VenousPressure', 'TMP', 'UFVolume', 'RecordTime'
]

const AFTER_SIGNS_FIELDS = [
  'Id', 'TreatmentId', 'PatientId', 'Weight', 'SystolicBP',
  'DiastolicBP', 'Pulse', 'ActualUF', 'RecordTime'
]

// ============ 透前体征服务 ============

export async function getPatientBeforeSigns(
  patientId: number,
  page: number = 1,
  pageSize: number = 20
): Promise<PaginatedResponse<BeforeSigns>> {
  return fetchFilteredData<BeforeSigns>(
    'BeforeSigns',
    BEFORE_SIGNS_FIELDS,
    { PatientId: patientId },
    page,
    pageSize
  )
}

export async function getBeforeSignsByTreatment(
  treatmentId: number
): Promise<BeforeSigns | null> {
  const result = await fetchFilteredData<BeforeSigns>(
    'BeforeSigns',
    BEFORE_SIGNS_FIELDS,
    { TreatmentId: treatmentId },
    1,
    1
  )
  return result.data[0] || null
}

// ============ 透中体征服务 ============

export async function getPatientDuringSigns(
  patientId: number,
  page: number = 1,
  pageSize: number = 50
): Promise<PaginatedResponse<DuringSigns>> {
  return fetchFilteredData<DuringSigns>(
    'DuringSigns',
    DURING_SIGNS_FIELDS,
    { PatientId: patientId },
    page,
    pageSize
  )
}

export async function getDuringSignsByTreatment(
  treatmentId: number
): Promise<DuringSigns[]> {
  const result = await fetchFilteredData<DuringSigns>(
    'DuringSigns',
    DURING_SIGNS_FIELDS,
    { TreatmentId: treatmentId },
    1,
    100
  )
  return result.data
}

export async function getLatestDuringSigns(
  treatmentId: number
): Promise<DuringSigns | null> {
  const signs = await getDuringSignsByTreatment(treatmentId)
  if (signs.length === 0) return null

  return signs.sort((a, b) => {
    const timeA = a.RecordTime || ''
    const timeB = b.RecordTime || ''
    return timeB.localeCompare(timeA)
  })[0]
}

// ============ 透后体征服务 ============

export async function getPatientAfterSigns(
  patientId: number,
  page: number = 1,
  pageSize: number = 20
): Promise<PaginatedResponse<AfterSigns>> {
  return fetchFilteredData<AfterSigns>(
    'AfterSigns',
    AFTER_SIGNS_FIELDS,
    { PatientId: patientId },
    page,
    pageSize
  )
}

export async function getAfterSignsByTreatment(
  treatmentId: number
): Promise<AfterSigns | null> {
  const result = await fetchFilteredData<AfterSigns>(
    'AfterSigns',
    AFTER_SIGNS_FIELDS,
    { TreatmentId: treatmentId },
    1,
    1
  )
  return result.data[0] || null
}

// ============ 综合体征查询 ============

export interface TreatmentVitals {
  before: BeforeSigns | null
  during: DuringSigns[]
  after: AfterSigns | null
}

export async function getTreatmentVitals(
  treatmentId: number
): Promise<TreatmentVitals> {
  const [before, during, after] = await Promise.all([
    getBeforeSignsByTreatment(treatmentId),
    getDuringSignsByTreatment(treatmentId),
    getAfterSignsByTreatment(treatmentId),
  ])

  return { before, during, after }
}

// ============ 趋势分析 ============

export interface VitalTrendPoint {
  date: string
  systolicBP?: number
  diastolicBP?: number
  pulse?: number
  weight?: number
}

export async function getPatientVitalTrends(
  patientId: number,
  limit: number = 10
): Promise<VitalTrendPoint[]> {
  const result = await getPatientBeforeSigns(patientId, 1, limit)

  return result.data.map(sign => ({
    date: sign.RecordTime || '',
    systolicBP: sign.SystolicBP,
    diastolicBP: sign.DiastolicBP,
    pulse: sign.Pulse,
    weight: sign.Weight,
  }))
}

// ============ 血压统计 ============

export interface BPStats {
  avgSystolic: number
  avgDiastolic: number
  maxSystolic: number
  minSystolic: number
  maxDiastolic: number
  minDiastolic: number
}

export function calculateBPStats(signs: Array<{ SystolicBP?: number; DiastolicBP?: number }>): BPStats {
  const systolicValues = signs.map(s => s.SystolicBP).filter((v): v is number => v !== undefined)
  const diastolicValues = signs.map(s => s.DiastolicBP).filter((v): v is number => v !== undefined)

  const avg = (arr: number[]) => arr.length > 0 ? arr.reduce((a, b) => a + b, 0) / arr.length : 0

  return {
    avgSystolic: Math.round(avg(systolicValues)),
    avgDiastolic: Math.round(avg(diastolicValues)),
    maxSystolic: systolicValues.length > 0 ? Math.max(...systolicValues) : 0,
    minSystolic: systolicValues.length > 0 ? Math.min(...systolicValues) : 0,
    maxDiastolic: diastolicValues.length > 0 ? Math.max(...diastolicValues) : 0,
    minDiastolic: diastolicValues.length > 0 ? Math.min(...diastolicValues) : 0,
  }
}
