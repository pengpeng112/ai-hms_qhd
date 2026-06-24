import { apiClient } from './restClient'

export interface ChargeLine {
  id: string
  tenantId: number
  chargeRecordId: string
  category: 'treatment' | 'material' | 'nursing' | 'injection' | 'drug'
  itemCode?: string
  itemName: string
  spec?: string
  unit?: string
  quantity?: number
  unitPrice?: number
  amount?: number
  billable: boolean
  source: 'auto' | 'manual'
  chargeItemId?: number
  hisPriceItemId?: string
  hisItemCode?: string
  hisItemClass?: string
  hisItemName?: string
  priceSource?: string
  matchedStatus?: string
  note?: string
  createdAt: string
}

export interface ChargeRecord {
  id: string
  tenantId: number
  patientId?: number
  treatmentId: number
  prescriptionId?: number
  chargeDate?: string
  shift?: string
  dialysisMode?: string
  accessType?: string
  crrtHours?: number
  totalAmount?: number
  status: 'draft' | 'confirmed' | 'checked' | 'pushed' | 'settled' | 'cancelled'
  recordedBy?: string
  recordedName?: string
  checkedBy?: string
  checkedName?: string
  checkedAt?: string
  exportedAt?: string
  pushedAt?: string
  note?: string
  lines?: ChargeLine[]
  createdAt: string
  updatedAt: string
}

export interface BuildDraftBody {
  treatmentId: number
  prescriptionId?: number
  dialysisMode?: string
  accessType?: string
  shift?: string
  crrtHours?: number
}

export interface ListChargesParams {
  patientId?: number
  status?: string
  dateFrom?: string
  dateTo?: string
  page?: number
  pageSize?: number
}

export interface ListChargesResponse {
  items: ChargeRecord[]
  total: number
  page: number
  pageSize: number
  totalPage: number
}

export interface PushResult {
  record: ChargeRecord
  push: { accepted: boolean; message: string }
}

export async function buildCharge(body: BuildDraftBody): Promise<ChargeRecord> {
  return apiClient.post('/api/v1/charges/build', body)
}

export async function listCharges(params: ListChargesParams = {}): Promise<ListChargesResponse> {
  return apiClient.get('/api/v1/charges', { params })
}

export async function getCharge(id: string): Promise<ChargeRecord> {
  return apiClient.get(`/api/v1/charges/${id}`)
}

export async function addChargeLine(id: string, line: Partial<ChargeLine>): Promise<ChargeLine> {
  return apiClient.post(`/api/v1/charges/${id}/lines`, line)
}

export async function updateChargeLine(lineId: string, patch: Partial<ChargeLine>): Promise<ChargeLine> {
  return apiClient.patch(`/api/v1/charges/lines/${lineId}`, patch)
}

export async function deleteChargeLine(lineId: string): Promise<void> {
  return apiClient.delete(`/api/v1/charges/lines/${lineId}`)
}

export async function confirmCharge(id: string): Promise<ChargeRecord> {
  return apiClient.post(`/api/v1/charges/${id}/confirm`)
}

export async function checkCharge(id: string): Promise<ChargeRecord> {
  return apiClient.post(`/api/v1/charges/${id}/check`)
}

export async function markExported(id: string): Promise<ChargeRecord> {
  return apiClient.post(`/api/v1/charges/${id}/exported`)
}

export async function cancelCharge(id: string, reason: string): Promise<ChargeRecord> {
  return apiClient.post(`/api/v1/charges/${id}/cancel`, { reason })
}

export async function pushCharge(id: string): Promise<PushResult> {
  return apiClient.post(`/api/v1/charges/${id}/push`)
}
