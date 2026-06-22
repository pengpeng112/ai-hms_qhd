// 护理文书 API（C1 模块）。/api/v1/nursing，走主系统 v1 信封。
import { apiClient } from './restClient'

export type NursingRiskLevel = 'high' | 'moderate' | 'low' | 'none'

export interface NursingScaleOption {
  label: string
  value: number
}

export interface NursingScaleItem {
  key: string
  label: string
  options: NursingScaleOption[]
}

export interface NursingScaleBand {
  min: number
  max: number
  level: NursingRiskLevel
  label: string
}

export interface NursingScale {
  scaleType: string
  name: string
  enabled: boolean
  direction: 'higher_worse' | 'lower_worse'
  items: NursingScaleItem[]
  bands: NursingScaleBand[]
}

export interface NursingDoc {
  id: string
  patientId: string
  treatmentId?: string
  docType: 'scale' | 'record' | 'plan'
  scaleType?: string
  score?: number
  riskLevel?: NursingRiskLevel
  content?: string
  nurseId?: string
  nurseName?: string
  recordedAt?: string
}

export interface RecordScaleRequest {
  patientId: string
  treatmentId?: string
  scaleType: string
  items: Record<string, number>
  nurseId?: string
  nurseName?: string
}

export interface RecordDocRequest {
  patientId: string
  treatmentId?: string
  docType: 'record' | 'plan'
  content: string
  nurseId?: string
  nurseName?: string
}

const base = '/api/v1'

export async function getNursingScales(): Promise<NursingScale[]> {
  const res = await apiClient.get(`${base}/nursing/scales`)
  return res.data?.data ?? []
}

export async function recordNursingScale(body: RecordScaleRequest): Promise<NursingDoc> {
  const res = await apiClient.post(`${base}/nursing/scales`, body)
  return res.data?.data
}

export async function recordNursingDoc(body: RecordDocRequest): Promise<NursingDoc> {
  const res = await apiClient.post(`${base}/nursing/docs`, body)
  return res.data?.data
}

export async function getNursingDocs(params: {
  patientId?: string
  treatmentId?: string
  docType?: string
  scaleType?: string
} = {}): Promise<NursingDoc[]> {
  const res = await apiClient.get(`${base}/nursing/docs`, { params })
  return res.data?.data ?? []
}

export async function getNursingAlerts(): Promise<NursingDoc[]> {
  const res = await apiClient.get(`${base}/nursing/alerts`)
  return res.data?.data ?? []
}
