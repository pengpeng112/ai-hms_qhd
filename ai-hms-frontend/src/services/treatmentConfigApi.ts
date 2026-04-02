// treatmentConfigApi.ts
// 诊疗配置 API 服务层

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

interface ApiPaginatedResponse<T> {
  success: true
  data: {
    items: T[]
    pagination: {
      page: number
      pageSize: number
      total: number
      totalPages: number
    }
  }
  timestamp: string
}

// ===== 辅助函数 =====
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

async function del<T>(url: string): Promise<T> {
  const response = await apiClient.delete<ApiSuccessResponse<T> | ApiErrorResponse>(url)
  // 204 No Content 表示删除成功，响应体为空
  if (response.status === 204) {
    return undefined as T
  }
  if (!response.data?.success) {
    const errorResp = response.data as ApiErrorResponse
    throw new Error(errorResp.error.message || 'API 请求失败')
  }
  return (response.data as ApiSuccessResponse<T>).data
}

async function getPaginated<T, P extends RequestParams = RequestParams>(url: string, params?: P): Promise<PaginatedResponse<T>> {
  const response = await apiClient.get<ApiPaginatedResponse<T> | ApiErrorResponse>(url, { params })
  if (!response.data.success) {
    const errorResp = response.data as ApiErrorResponse
    throw new Error(errorResp.error.message || 'API 请求失败')
  }
  return (response.data as ApiPaginatedResponse<T>).data
}

// ===== 类型定义 =====

// 方案模板
export interface PlanTemplate {
  id: string
  name: string
  description: string
  mode: 'HD' | 'HFD' | 'HP' | 'HF' | 'HDF'
  category: string
  isDefault: boolean
  isEnabled: boolean
  createdAt: string
  updatedAt: string
  templateContent: PlanTemplateContent
}

export interface PlanTemplateContent {
  weeklyFrequency: number
  biweeklyFrequency: number
  duration: number
  dryWeight: number
  dialysisMode: {
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
  anticoagulant: {
    initialDrug: string
    initialDose: string
    totalDose: string
    maintenanceDrug: string
    infusionRate: string
    infusionTime: string
    maintenanceDose: string
  }
  parameters: {
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
  materials: PlanTemplateMaterial[]
}

export interface PlanTemplateMaterial {
  id: string
  name: string
  category: string
  count: number
  code: string
  brand: string
  spec: string
  note: string
}

export interface PlanTemplateCreateRequest {
  name: string
  description?: string
  mode: 'HD' | 'HFD' | 'HP' | 'HF' | 'HDF'
  category?: string
  isDefault?: boolean
  isEnabled?: boolean
  templateContent: PlanTemplateContent
}

export interface PlanTemplateUpdateRequest {
  name?: string
  description?: string
  mode?: 'HD' | 'HFD' | 'HP' | 'HF' | 'HDF'
  category?: string
  isDefault?: boolean
  isEnabled?: boolean
  templateContent?: PlanTemplateContent
}

export interface PlanTemplateListParams {
  page?: number
  pageSize?: number
  search?: string
  mode?: string
  category?: string
  isEnabled?: boolean
  [key: string]: string | number | boolean | undefined
}

// 材料目录
export interface MaterialCatalog {
  id: number
  code: string
  name: string
  shortName?: string
  mnemonic?: string
  category: string
  spec: string
  standardType?: string
  brand: string
  unit: string
  packaging?: string
  manufacturer?: string
  sortOrder: number
  isEnabled: boolean
  notes: string
  createdAt: string
  updatedAt: string
}

export interface MaterialCatalogCreateRequest {
  code?: string
  name: string
  shortName: string
  mnemonic?: string
  category: string
  spec?: string
  standardType?: string
  brand?: string
  unit: string
  packaging?: string
  manufacturer?: string
  sortOrder?: number
  isEnabled?: boolean
  notes?: string
}

export interface MaterialCatalogUpdateRequest {
  code?: string
  name?: string
  shortName?: string
  mnemonic?: string
  category?: string
  spec?: string
  standardType?: string
  brand?: string
  unit?: string
  packaging?: string
  manufacturer?: string
  sortOrder?: number
  isEnabled?: boolean
  notes?: string
}

export interface MaterialCatalogListParams {
  page?: number
  pageSize?: number
  search?: string
  category?: string
  isEnabled?: boolean
  [key: string]: string | number | boolean | undefined
}

// 药品目录
export interface DrugCatalog {
  id: number
  code: string
  name: string
  shortName: string
  mnemonic: string
  genericName: string
  category: string
  spec: string
  concentration?: string
  specUnit: string
  minUnitDose: string
  baseUnit: string
  brand: string
  packaging: string
  manufacturer: string
  standardType: string
  timing: string
  tips: string
  sortOrder: number
  isEnabled: boolean
  note: string
  createdAt: string
  updatedAt: string
}

export interface DrugCatalogCreateRequest {
  code: string
  name: string
  shortName?: string
  mnemonic?: string
  genericName?: string
  category: string
  spec?: string
  concentration?: string
  specUnit?: string
  minUnitDose?: string
  baseUnit?: string
  brand?: string
  packaging?: string
  manufacturer?: string
  standardType?: string
  timing?: string
  tips?: string
  sortOrder?: number
  isEnabled?: boolean
  note?: string
}

export interface DrugCatalogUpdateRequest {
  name?: string
  shortName?: string
  mnemonic?: string
  genericName?: string
  category?: string
  spec?: string
  concentration?: string
  specUnit?: string
  minUnitDose?: string
  baseUnit?: string
  brand?: string
  packaging?: string
  manufacturer?: string
  standardType?: string
  timing?: string
  tips?: string
  sortOrder?: number
  isEnabled?: boolean
  note?: string
}

export interface DrugCatalogListParams {
  page?: number
  pageSize?: number
  search?: string
  category?: string
  isEnabled?: boolean
  [key: string]: string | number | boolean | undefined
}

// 医嘱模板条目
export interface OrderTemplateItem {
  id: string
  drugId?: number
  drugName: string
  spec?: string
  minUnitDose?: string
  dosage?: string
  unit?: string
  route?: string
  frequency?: string
  timing?: string
  groupId?: string
  sortOrder: number
}

export interface OrderTemplateItemRequest {
  drugId?: number
  drugName: string
  spec?: string
  minUnitDose?: string
  dosage?: string
  unit?: string
  route?: string
  frequency?: string
  timing?: string
  groupId?: string
  sortOrder?: number
}

// 医嘱模板
export interface OrderTemplate {
  id: string
  name: string
  type: '长期' | '临时'
  category: string
  content: string
  frequency?: string
  priority: string
  isDefault: boolean
  isEnabled: boolean
  items?: OrderTemplateItem[]
  createdAt: string
  updatedAt: string
}

export interface OrderTemplateCreateRequest {
  name: string
  type: '长期' | '临时'
  category: string
  content?: string
  frequency?: string
  priority?: string
  isDefault?: boolean
  isEnabled?: boolean
  items?: OrderTemplateItemRequest[]
}

export interface OrderTemplateUpdateRequest {
  name?: string
  type?: '长期' | '临时'
  category?: string
  content?: string
  frequency?: string
  priority?: string
  isDefault?: boolean
  isEnabled?: boolean
  items?: OrderTemplateItemRequest[]
}

export interface OrderTemplateListParams {
  page?: number
  pageSize?: number
  search?: string
  type?: string
  category?: string
  isEnabled?: boolean
  [key: string]: string | number | boolean | undefined
}

// 分页响应类型
export interface PaginatedResponse<T> {
  items: T[]
  pagination: {
    page: number
    pageSize: number
    total: number
    totalPages: number
  }
}

// ===== API 函数 =====

// 方案模板 API
export const planTemplateApi = {
  // 获取方案模板列表
  list: async (params?: PlanTemplateListParams): Promise<PaginatedResponse<PlanTemplate>> => {
    return getPaginated<PlanTemplate>('/api/v1/treatment-templates', params)
  },

  // 获取方案模板详情
  get: async (id: string): Promise<PlanTemplate> => {
    return get<PlanTemplate>(`/api/v1/treatment-templates/${id}`)
  },

  // 创建方案模板
  create: async (data: PlanTemplateCreateRequest): Promise<PlanTemplate> => {
    return post<PlanTemplate>('/api/v1/treatment-templates', data)
  },

  // 更新方案模板
  update: async (id: string, data: PlanTemplateUpdateRequest): Promise<PlanTemplate> => {
    return put<PlanTemplate>(`/api/v1/treatment-templates/${id}`, data)
  },

  // 删除方案模板
  delete: async (id: string): Promise<void> => {
    return del<void>(`/api/v1/treatment-templates/${id}`)
  },

  // 切换启用状态
  toggleEnabled: async (id: string): Promise<{ id: string; isEnabled: boolean }> => {
    return post<{ id: string; isEnabled: boolean }>(`/api/v1/treatment-templates/${id}/toggle`)
  },

  // 设置默认模板
  setDefault: async (id: string): Promise<{ id: string; message: string }> => {
    return post<{ id: string; message: string }>(`/api/v1/treatment-templates/${id}/set-default`)
  },
}

// 材料目录 API
export const materialCatalogApi = {
  // 获取材料目录列表
  list: async (params?: MaterialCatalogListParams): Promise<PaginatedResponse<MaterialCatalog>> => {
    return getPaginated<MaterialCatalog>('/api/v1/materials/catalog', params)
  },

  // 获取材料目录详情
  get: async (id: number): Promise<MaterialCatalog> => {
    return get<MaterialCatalog>(`/api/v1/materials/catalog/${id}`)
  },

  // 创建材料目录
  create: async (data: MaterialCatalogCreateRequest): Promise<MaterialCatalog> => {
    return post<MaterialCatalog>('/api/v1/materials/catalog', data)
  },

  // 更新材料目录
  update: async (id: number, data: MaterialCatalogUpdateRequest): Promise<MaterialCatalog> => {
    return put<MaterialCatalog>(`/api/v1/materials/catalog/${id}`, data)
  },

  // 删除材料目录
  delete: async (id: number): Promise<void> => {
    return del<void>(`/api/v1/materials/catalog/${id}`)
  },

  // 切换启用状态
  toggleEnabled: async (id: number): Promise<{ id: number; isEnabled: boolean }> => {
    return post<{ id: number; isEnabled: boolean }>(`/api/v1/materials/catalog/${id}/toggle`)
  },

  // 获取材料分类列表
  getCategories: async (): Promise<string[]> => {
    return get<string[]>('/api/v1/materials/categories')
  },
}

// 药品目录 API
export const drugCatalogApi = {
  // 获取药品目录列表
  list: async (params?: DrugCatalogListParams): Promise<PaginatedResponse<DrugCatalog>> => {
    return getPaginated<DrugCatalog>('/api/v1/drugs/catalog', params)
  },

  // 获取药品目录详情
  get: async (id: number): Promise<DrugCatalog> => {
    return get<DrugCatalog>(`/api/v1/drugs/catalog/${id}`)
  },

  // 创建药品目录
  create: async (data: DrugCatalogCreateRequest): Promise<DrugCatalog> => {
    return post<DrugCatalog>('/api/v1/drugs/catalog', data)
  },

  // 更新药品目录
  update: async (id: number, data: DrugCatalogUpdateRequest): Promise<DrugCatalog> => {
    return put<DrugCatalog>(`/api/v1/drugs/catalog/${id}`, data)
  },

  // 删除药品目录
  delete: async (id: number): Promise<void> => {
    return del<void>(`/api/v1/drugs/catalog/${id}`)
  },

  // 切换启用状态
  toggleEnabled: async (id: number): Promise<{ id: number; isEnabled: boolean }> => {
    return post<{ id: number; isEnabled: boolean }>(`/api/v1/drugs/catalog/${id}/toggle`)
  },

  // 获取药品分类列表
  getCategories: async (): Promise<string[]> => {
    return get<string[]>('/api/v1/drugs/categories')
  },
}

// 医嘱模板 API
export const orderTemplateApi = {
  // 获取医嘱模板列表
  list: async (params?: OrderTemplateListParams): Promise<PaginatedResponse<OrderTemplate>> => {
    return getPaginated<OrderTemplate>('/api/v1/order-templates', params)
  },

  // 获取医嘱模板详情
  get: async (id: string): Promise<OrderTemplate> => {
    return get<OrderTemplate>(`/api/v1/order-templates/${id}`)
  },

  // 创建医嘱模板
  create: async (data: OrderTemplateCreateRequest): Promise<OrderTemplate> => {
    return post<OrderTemplate>('/api/v1/order-templates', data)
  },

  // 更新医嘱模板
  update: async (id: string, data: OrderTemplateUpdateRequest): Promise<OrderTemplate> => {
    return put<OrderTemplate>(`/api/v1/order-templates/${id}`, data)
  },

  // 删除医嘱模板
  delete: async (id: string): Promise<void> => {
    return del<void>(`/api/v1/order-templates/${id}`)
  },

  // 切换启用状态
  toggleEnabled: async (id: string): Promise<{ id: string; isEnabled: boolean }> => {
    return post<{ id: string; isEnabled: boolean }>(`/api/v1/order-templates/${id}/toggle`)
  },

  // 设置默认模板
  setDefault: async (id: string): Promise<{ id: string; message: string }> => {
    return post<{ id: string; message: string }>(`/api/v1/order-templates/${id}/set-default`)
  },
}

// 导出所有 API
export default {
  planTemplate: planTemplateApi,
  materialCatalog: materialCatalogApi,
  drugCatalog: drugCatalogApi,
  orderTemplate: orderTemplateApi,
}
