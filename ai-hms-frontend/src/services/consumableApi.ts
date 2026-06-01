import { apiClient, type ApiSuccessResponse } from './restClient'

export interface ConsumableRecord {
  id: number
  treatmentId: number
  materialId: number
  materialName: string
  num: number
  unit: string
  batch: string
  serialNo: string
  note: string
  createdAt: string
}

export interface CreateConsumableRequest {
  materialId: number
  num?: number
  unit?: string
  batch?: string
  serialNo?: string
  note?: string
}

export async function listConsumables(treatmentId: number): Promise<ConsumableRecord[]> {
  const response = await apiClient.get<ApiSuccessResponse<ConsumableRecord[]>>(`/api/v1/treatments/${treatmentId}/consumables`)
  if (!response.data.success) throw new Error('Failed to load consumables')
  return response.data.data
}

export async function createConsumable(treatmentId: number, data: CreateConsumableRequest): Promise<ConsumableRecord> {
  const response = await apiClient.post<ApiSuccessResponse<ConsumableRecord>>(`/api/v1/treatments/${treatmentId}/consumables`, data)
  if (!response.data.success) throw new Error('Failed to save consumable')
  return response.data.data
}

export async function deleteConsumable(treatmentId: number, id: number): Promise<void> {
  const response = await apiClient.delete<ApiSuccessResponse<unknown>>(`/api/v1/treatments/${treatmentId}/consumables/${id}`)
  if (!response.data.success) throw new Error('Failed to delete consumable')
}
