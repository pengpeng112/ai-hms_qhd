// orderApi.ts
// 医嘱 & 处方 API 服务层

import { apiClient } from './restClient'

// ===== 通用响应类型 =====
interface ApiSuccessResponse<T> {
  success: true
  data: T
  timestamp: string
}

interface ApiErrorResponse {
  success: false
  error: {
    code: string
    message: string
  }
  timestamp: string
}

// ===== 辅助函数（复用 treatmentConfigApi 模式）=====
type RequestParams = Record<string, string | number | boolean | undefined>

async function get<T, P extends RequestParams = RequestParams>(url: string, params?: P): Promise<T> {
  const response = await apiClient.get<ApiSuccessResponse<T> | ApiErrorResponse>(url, { params })
  if (!response.data.success) {
    const errorResp = response.data as ApiErrorResponse
    throw new Error(errorResp.error.message || 'API 请求失败')
  }
  return (response.data as ApiSuccessResponse<T>).data
}

async function post<T>(url: string, data?: unknown): Promise<T> {
  const response = await apiClient.post<ApiSuccessResponse<T> | ApiErrorResponse>(url, data)
  if (!response.data.success) {
    const errorResp = response.data as ApiErrorResponse
    throw new Error(errorResp.error.message || 'API 请求失败')
  }
  return (response.data as ApiSuccessResponse<T>).data
}

async function put<T>(url: string, data?: unknown): Promise<T> {
  const response = await apiClient.put<ApiSuccessResponse<T> | ApiErrorResponse>(url, data)
  if (!response.data.success) {
    const errorResp = response.data as ApiErrorResponse
    throw new Error(errorResp.error.message || 'API 请求失败')
  }
  return (response.data as ApiSuccessResponse<T>).data
}

// ===== 类型定义 =====

export interface Order {
  id: string
  patientId: string
  type: '长期' | '临时'
  category: string
  name: string
  content: string
  dose: string
  unit: string
  route: string
  timing: string
  execTiming: string
  drugId?: number
  spec: string
  groupId?: string
  doctorId: string
  doctorName: string
  status: '待执行' | '执行中' | '已执行' | '已停止'
  startTime: string
  endTime?: string
  frequency?: string
  priority: string
  notes: string
  executedAt?: string
  executedBy?: string
  stopReason?: string
  createdAt: string
  updatedAt: string
}

export interface OrderCreateRequest {
  type: '长期' | '临时'
  category?: string
  name?: string
  content?: string
  dose?: string
  unit?: string
  route?: string
  timing?: string
  execTiming?: string
  drugId?: number
  spec?: string
  groupId?: string
  frequency?: string
  priority?: string
  startTime?: string
  endTime?: string
  notes?: string
}

export interface OrderUpdateRequest {
  category?: string
  name?: string
  content?: string
  dose?: string
  unit?: string
  route?: string
  timing?: string
  execTiming?: string
  drugId?: number
  spec?: string
  groupId?: string
  frequency?: string
  priority?: string
  startTime?: string
  endTime?: string
  notes?: string
}

export interface OrderReviseRequest {
  category?: string
  name?: string
  content?: string
  dose?: string
  unit?: string
  route?: string
  timing?: string
  execTiming?: string
  drugId?: number
  spec?: string
  frequency?: string
  priority?: string
  startTime?: string
  stopDate?: string
  notes?: string
}

export interface CreateFromTemplateItemRequest {
  templateItemId: string
  name?: string
  content?: string
  dose?: string
  unit?: string
  route?: string
  frequency?: string
  timing?: string
  execTiming?: string
  spec?: string
}

export interface CreateFromTemplateRequest {
  templateId: string
  type: '长期' | '临时'
  items: CreateFromTemplateItemRequest[]
}

export interface OrderListParams {
  type?: string
  statuses?: string
  includeExpired?: boolean
  [key: string]: string | number | boolean | undefined
}

// ===== API 函数 =====

export const orderApi = {
  list: async (patientId: string, params?: OrderListParams): Promise<Order[]> => {
    return get<Order[]>(`/api/v1/patients/${patientId}/orders`, params)
  },

  create: async (patientId: string, data: OrderCreateRequest): Promise<Order> => {
    return post<Order>(`/api/v1/patients/${patientId}/orders`, data)
  },

  update: async (patientId: string, orderId: string, data: OrderUpdateRequest): Promise<Order> => {
    return put<Order>(`/api/v1/patients/${patientId}/orders/${orderId}`, data)
  },

  revise: async (patientId: string, orderId: string, data: OrderReviseRequest): Promise<Order> => {
    return post<Order>(`/api/v1/patients/${patientId}/orders/${orderId}/revise`, data)
  },

  stop: async (patientId: string, orderId: string, stopReason?: string, stopDate?: string): Promise<Order[]> => {
    return post<Order[]>(`/api/v1/patients/${patientId}/orders/${orderId}/stop`, { stopReason, stopDate })
  },

  copy: async (patientId: string, orderId: string): Promise<Order> => {
    return post<Order>(`/api/v1/patients/${patientId}/orders/${orderId}/copy`)
  },

  group: async (patientId: string, orderIds: string[]): Promise<Order[]> => {
    return post<Order[]>(`/api/v1/patients/${patientId}/orders/group`, { orderIds })
  },

  ungroup: async (patientId: string, orderIds: string[]): Promise<Order[]> => {
    return post<Order[]>(`/api/v1/patients/${patientId}/orders/ungroup`, { orderIds })
  },

  createFromTemplate: async (patientId: string, data: CreateFromTemplateRequest): Promise<Order[]> => {
    return post<Order[]>(`/api/v1/patients/${patientId}/orders/from-template`, data)
  },
}

// ===== 处方类型定义 =====

export interface PrescriptionOrderItem {
  orderId: string
  name: string
  category: string
  dose: string
  unit: string
  frequency: string
  route: string
  spec: string
}

export interface DialysisMode {
  mode: string
  bloodFlow: number
  substituteInputMode?: string
  substituteFlow?: number
  substituteVolume?: number
  bv: string
  frequencyDesc: string
  autoConfirm: boolean
  status: string
  notes: string
}

export interface Anticoagulant {
  initialDrug: string
  initialDose: string
  maintenanceDrug: string
  infusionRate: string
  infusionTime: string
  maintenanceDose: string
  totalDose: string
}

export interface DialysisParameters {
  dialysateType: string
  dialysateGroup: string
  flowRate: number
  na: number
  ca: number
  k: number
  hco3: number
  glucose: string
  conductivity: number
  temp: number
  volume: number
}

export interface PrescriptionMaterial {
  id: string
  name: string
  category: string
  count: number
  code: string
  brand: string
  spec: string
  note: string
}

export interface Prescription {
  id: string
  patientId: string
  treatmentPlanId: string
  prescriptionDate: string
  doctorId: string
  doctorName: string
  status: '待执行' | '执行中' | '已执行' | '已取消'
  duration: number
  dryWeight: number
  extraWeight: number
  dialysisMode: DialysisMode
  anticoagulant: Anticoagulant
  parameters: DialysisParameters
  materials: PrescriptionMaterial[]
  orderItems: PrescriptionOrderItem[]
  notes: string
  executedAt?: string
  executedBy?: string
  createdAt: string
  updatedAt: string
}

export interface PrescriptionCreateRequest {
  prescriptionDate: string
  duration?: number
  dryWeight?: number
  extraWeight?: number
  dialysisMode?: DialysisMode
  anticoagulant?: Anticoagulant
  parameters?: DialysisParameters
  materials?: PrescriptionMaterial[]
  orderItems?: PrescriptionOrderItem[]
  notes?: string
}

export interface PrescriptionUpdateRequest {
  duration?: number
  dryWeight?: number
  extraWeight?: number
  dialysisMode?: DialysisMode
  anticoagulant?: Anticoagulant
  parameters?: DialysisParameters
  materials?: PrescriptionMaterial[]
  orderItems?: PrescriptionOrderItem[]
  notes?: string
}

// ===== 处方 API =====

export const prescriptionApi = {
  list: async (patientId: string): Promise<Prescription[]> => {
    return get<Prescription[]>(`/api/v1/patients/${patientId}/prescriptions`)
  },

  get: async (patientId: string, prescriptionId: string): Promise<Prescription> => {
    return get<Prescription>(`/api/v1/patients/${patientId}/prescriptions/${prescriptionId}`)
  },

  create: async (patientId: string, data: PrescriptionCreateRequest): Promise<Prescription> => {
    return post<Prescription>(`/api/v1/patients/${patientId}/prescriptions`, data)
  },

  update: async (patientId: string, prescriptionId: string, data: PrescriptionUpdateRequest): Promise<Prescription> => {
    return put<Prescription>(`/api/v1/patients/${patientId}/prescriptions/${prescriptionId}`, data)
  },

  execute: async (patientId: string, prescriptionId: string): Promise<Prescription> => {
    return post<Prescription>(`/api/v1/patients/${patientId}/prescriptions/${prescriptionId}/execute`)
  },

  cancel: async (patientId: string, prescriptionId: string): Promise<Prescription> => {
    return post<Prescription>(`/api/v1/patients/${patientId}/prescriptions/${prescriptionId}/cancel`)
  },

  extract: async (patientId: string, date: string): Promise<Prescription> => {
    return post<Prescription>(`/api/v1/patients/${patientId}/prescriptions/extract`, { date })
  },
}
