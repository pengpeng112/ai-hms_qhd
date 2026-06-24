import { apiClient } from './restClient'

export interface HisPriceItem {
  id: string
  sourceSystem: string
  itemClass?: string
  itemCode: string
  itemName?: string
  itemSpec?: string
  units?: string
  price?: number
  preferPrice?: number
  foreignerPrice?: number
  isActive: boolean
  syncedAt: string
}

export interface HisPriceSearchParams {
  keyword?: string
  itemClass?: string
  activeOnly?: boolean
  page?: number
  pageSize?: number
}

export interface HisPriceSearchResponse {
  items: HisPriceItem[]
  total: number
  page: number
  pageSize: number
  totalPage: number
}

export async function searchHisPrices(params: HisPriceSearchParams = {}): Promise<HisPriceSearchResponse> {
  return apiClient.get('/api/v1/his-price-items', { params })
}

export async function syncHisPrices(): Promise<{ message: string; runId: string }> {
  return apiClient.post('/api/v1/his-price-items/sync')
}

export async function getHisPriceByCode(itemCode: string): Promise<HisPriceItem> {
  return apiClient.get(`/api/v1/his-price-items/${itemCode}`)
}
