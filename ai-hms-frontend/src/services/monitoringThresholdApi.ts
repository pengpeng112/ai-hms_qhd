import { apiClient, type ApiSuccessResponse } from './restClient'

export interface FixedThreshold {
  metricKey: string
  label: string
  unit: string
  dangerLow: number | null
  warnLow: number | null
  warnHigh: number | null
  dangerHigh: number | null
  basis: string
  enabled: boolean
  sortOrder: number
}

export interface VPStratum {
  access: 'AVF' | 'AVG' | 'TCC' | 'NCC'
  bfMin: number
  bfMax: number
  normalLow: number
  warnHigh: number
  dangerHigh: number
  basis: string
  enabled: boolean
}

export interface ThresholdPayload {
  fixed: FixedThreshold[]
  vpReference: VPStratum[]
  naFactor: number
}

export async function getThresholds(): Promise<ThresholdPayload> {
  const res = await apiClient.get<ApiSuccessResponse<ThresholdPayload> | { success: false; error: { message: string } }>('/api/v1/monitoring/thresholds')
  if (!res.data.success) {
    const err = res.data as { success: false; error: { message: string } }
    throw new Error(err.error.message || '加载阈值表失败')
  }
  return (res.data as ApiSuccessResponse<ThresholdPayload>).data
}

export async function saveThresholds(payload: ThresholdPayload): Promise<void> {
  const res = await apiClient.put<ApiSuccessResponse<unknown> | { success: false; error: { message: string } }>('/api/v1/monitoring/thresholds', payload)
  if (!res.data.success) {
    const err = res.data as { success: false; error: { message: string } }
    throw new Error(err.error.message || '保存阈值表失败')
  }
}

export async function resetThresholds(): Promise<void> {
  const res = await apiClient.post<ApiSuccessResponse<unknown> | { success: false; error: { message: string } }>('/api/v1/monitoring/thresholds/reset')
  if (!res.data.success) {
    const err = res.data as { success: false; error: { message: string } }
    throw new Error(err.error.message || '恢复默认失败')
  }
}
