import { apiClient, type ApiSuccessResponse } from './restClient'

export interface UploadResult {
  id: string
  fileName: string
  path: string
  size: number
  contentType: string
}

export const uploadApi = {
  async upload(file: File): Promise<UploadResult> {
    const formData = new FormData()
    formData.append('file', file)
    const response = await apiClient.post<ApiSuccessResponse<UploadResult>>('/api/v1/uploads', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    return response.data.data
  },
}

export interface WardItem {
  id: string
  name: string
  sort: number
  patientType: string
  patientTypeLabel: string
  infectionType: string
  responsibleUsers: string
  responsibleUserNames: string
  note: string
  isDisabled: boolean
  bedCount: number
  createdAt?: string
  updatedAt?: string
}

export interface WardPayload {
  name: string
  sort?: number
  patientType?: string
  infectionType?: string
  responsibleUsers?: string
  note?: string
  isDisabled?: boolean
}

export interface BedEquipmentItem {
  equipmentId: string
  equipmentName: string
  isDefault: boolean
  sort: number
}

export interface BedItem {
  id: string
  name: string
  wardId: string
  wardName: string
  sort: number
  note: string
  fepId?: number | null
  fepName?: string
  acquisiteConnectId?: number | null
  acquisiteConnectName?: string
  equipments: BedEquipmentItem[]
  defaultEquipmentName: string
  equipmentCount: number
  isDisabled: boolean
  createdAt?: string
  updatedAt?: string
}

export interface BedPayload {
  name: string
  wardId: number
  sort?: number
  note?: string
  isDisabled?: boolean
  fepId?: number | null
  acquisiteConnectId?: number | null
  equipments?: Array<{ equipmentId: string; isDefault?: boolean; sort?: number }>
}

export interface EducationItem {
  id: string
  name: string
  description: string
  sort: number
  attachmentIds: string
  type: string
  classify: string
  note: string
  isDisabled: boolean
  createdAt?: string
  updatedAt?: string
}

export interface EducationPayload {
  name: string
  description?: string
  sort?: number
  attachmentIds?: string
  type?: string
  classify?: string
  note?: string
  isDisabled?: boolean
}

interface ListResponse<T> {
  items: T[]
  total: number
  page: number
  pageSize: number
  totalPage: number
}

export const wardManagementApi = {
  async list(includeDisabled = true): Promise<WardItem[]> {
    const response = await apiClient.get<ApiSuccessResponse<ListResponse<WardItem>>>('/api/v1/wards', {
      params: { page: 1, pageSize: 200, includeDisabled },
    })
    return response.data.data.items
  },
  async create(payload: WardPayload): Promise<WardItem> {
    const response = await apiClient.post<ApiSuccessResponse<WardItem>>('/api/v1/wards', payload)
    return response.data.data
  },
  async update(id: string, payload: Partial<WardPayload>): Promise<WardItem> {
    const response = await apiClient.put<ApiSuccessResponse<WardItem>>(`/api/v1/wards/${id}`, payload)
    return response.data.data
  },
  async remove(id: string): Promise<void> {
    await apiClient.delete<ApiSuccessResponse<unknown>>(`/api/v1/wards/${id}`)
  },
}

export const bedManagementApi = {
  async list(includeDisabled = true): Promise<BedItem[]> {
    const response = await apiClient.get<ApiSuccessResponse<ListResponse<BedItem>>>('/api/v1/beds', {
      params: { page: 1, pageSize: 200, includeDisabled },
    })
    return response.data.data.items
  },
  async create(payload: BedPayload): Promise<BedItem> {
    const response = await apiClient.post<ApiSuccessResponse<BedItem>>('/api/v1/beds', payload)
    return response.data.data
  },
  async update(id: string, payload: Partial<BedPayload>): Promise<BedItem> {
    const response = await apiClient.put<ApiSuccessResponse<BedItem>>(`/api/v1/beds/${id}`, payload)
    return response.data.data
  },
  async remove(id: string): Promise<void> {
    await apiClient.delete<ApiSuccessResponse<unknown>>(`/api/v1/beds/${id}`)
  },
}

export const educationManagementApi = {
  async list(): Promise<EducationItem[]> {
    const response = await apiClient.get<ApiSuccessResponse<EducationItem[]>>('/api/v1/health-educations', {
      params: { includeDisabled: true },
    })
    return response.data.data
  },
  async create(payload: EducationPayload): Promise<EducationItem> {
    const response = await apiClient.post<ApiSuccessResponse<EducationItem>>('/api/v1/health-educations', payload)
    return response.data.data
  },
  async update(id: string, payload: Partial<EducationPayload>): Promise<EducationItem> {
    const response = await apiClient.put<ApiSuccessResponse<EducationItem>>(`/api/v1/health-educations/${id}`, payload)
    return response.data.data
  },
  async remove(id: string): Promise<void> {
    await apiClient.delete<ApiSuccessResponse<unknown>>(`/api/v1/health-educations/${id}`)
  },
}
