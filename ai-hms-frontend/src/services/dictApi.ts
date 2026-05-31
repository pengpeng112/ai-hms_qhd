// dictApi.ts - 字典管理 API 服务层

import { apiClient } from './restClient'

// ===== 类型定义 =====

interface ApiSuccessResponse<T> {
  success: true
  data: T
  timestamp: string
}

// 级联选择选项类型
export type CascaderOption = {
  value: string
  label: string
  children?: CascaderOption[]
}

// 字典类型
export interface DictType {
  id: string
  code: string
  name: string
  description: string
  source: string
  icon?: string  // 图标（emoji）
  sortOrder: number
  isEnabled: boolean
  createdAt: string
  updatedAt: string
}

// 字典项
export interface DictItem {
  id: string
  typeCode: string
  code: string
  name: string
  description: string
  source: string
  sortOrder: number
  isEnabled: boolean
  extra?: string
  parentCode?: string
  createdAt: string
  updatedAt: string
  children?: DictItem[]  // 子项（用于树形结构）
}

// 字典类型创建请求
export interface DictTypeCreateRequest {
  code: string
  name: string
  description?: string
  icon?: string
  sortOrder?: number
  isEnabled?: boolean
}

// 字典项创建请求
export interface DictItemCreateRequest {
  typeCode: string
  code: string
  name: string
  description?: string
  sortOrder?: number
  isEnabled?: boolean
  extra?: string
  parentCode?: string
}

// ===== 辅助函数 =====

type RequestParams = Record<string, string | number | boolean | undefined>

async function get<T, P extends RequestParams = RequestParams>(url: string, params?: P): Promise<T> {
  const response = await apiClient.get<ApiSuccessResponse<T>>(url, { params })
  if (!response.data.success) {
    throw new Error('API 请求失败')
  }
  return response.data.data
}

async function post<T>(url: string, data?: unknown): Promise<T> {
  const response = await apiClient.post<ApiSuccessResponse<T>>(url, data)
  if (!response.data.success) {
    throw new Error('API 请求失败')
  }
  return response.data.data
}

async function put<T>(url: string, data?: unknown): Promise<T> {
  const response = await apiClient.put<ApiSuccessResponse<T>>(url, data)
  if (!response.data.success) {
    throw new Error('API 请求失败')
  }
  return response.data.data
}

async function del<T>(url: string): Promise<T> {
  const response = await apiClient.delete<ApiSuccessResponse<T>>(url)
  if (response.status === 204) {
    return undefined as T
  }
  if (!response.data?.success) {
    throw new Error('API 请求失败')
  }
  return response.data.data
}

// ===== 字典 API =====

export const dictApi = {
  // 获取字典类型列表
  listTypes: async (): Promise<DictType[]> => {
    return get<DictType[]>('/api/v1/dict/types')
  },

  // 获取字典类型详情
  getType: async (code: string): Promise<DictType> => {
    return get<DictType>(`/api/v1/dict/types/${code}`)
  },

  // 创建字典类型
  createType: async (data: DictTypeCreateRequest): Promise<DictType> => {
    return post<DictType>('/api/v1/dict/types', data)
  },

  // 更新字典类型
  updateType: async (id: string, data: Partial<DictType>): Promise<void> => {
    return put<void>(`/api/v1/dict/types/${id}`, data)
  },

  // 删除字典类型
  deleteType: async (id: string): Promise<void> => {
    return del<void>(`/api/v1/dict/types/${id}`)
  },

  // 获取字典项列表
  getItems: async (typeCode: string, isEnabledOnly = true): Promise<DictItem[]> => {
    return get<DictItem[]>(`/api/v1/dict/items/${typeCode}`, { isEnabled: isEnabledOnly })
  },

  // 获取字典项树形结构（用于级联选择）
  getItemsTree: async (typeCode: string, isEnabledOnly = true): Promise<DictItem[]> => {
    return get<DictItem[]>(`/api/v1/dict/items/${typeCode}/tree`, { isEnabled: isEnabledOnly })
  },

  // 创建字典项
  createItem: async (data: DictItemCreateRequest): Promise<DictItem> => {
    return post<DictItem>('/api/v1/dict/items', data)
  },

  // 更新字典项（后端接收 snake_case key）
  updateItem: async (id: string, data: Partial<DictItem> & { parent_code?: string | null }): Promise<void> => {
    return put<void>(`/api/v1/dict/items/${id}`, data)
  },

  // 删除字典项
  deleteItem: async (id: string): Promise<void> => {
    return del<void>(`/api/v1/dict/items/${id}`)
  },

  // 切换字典项启用状态
  toggleItemEnabled: async (id: string): Promise<void> => {
    return post<void>(`/api/v1/dict/items/${id}/toggle`)
  },

  // 初始化字典数据
  initData: async (): Promise<{ message: string }> => {
    return post<{ message: string }>('/api/v1/dict/items/init')
  },

  // 批量导入字典数据
  importData: async (data: { types: DictType[]; items: DictItem[] }): Promise<{
    typesCreated: number
    typesUpdated: number
    itemsCreated: number
    itemsUpdated: number
  }> => {
    return post('/api/v1/dict/import', data)
  },
}

// ===== 字典类型代码常量 =====

export const DICT_TYPES = {
  DIALYSIS_MODE: 'DIALYSIS_MODE',        // 透析方式
  ANTICOAGULANT: 'ANTICOAGULANT',        // 抗凝剂类型
  DIALYSATE_TYPE: 'DIALYSATE_TYPE',     // 透析液类型
  DIALYSATE_GROUP: 'DIALYSATE_GROUP',   // 透析液组
  DIALYSATE_FLOW: 'DIALYSATE_FLOW',     // 透析液流量
  DIALYZER_AREA: 'DIALYZER_AREA',       // 透析器面积
  GLUCOSE: 'GLUCOSE',                   // 葡萄糖类型
  MATERIAL_CATEGORY: 'MATERIAL_CATEGORY', // 材料分类
  DRUG_CATEGORY: 'DRUG_CATEGORY',       // 药品分类
  ORDER_TYPE: 'ORDER_TYPE',             // 医嘱类型
  ORDER_CATEGORY: 'ORDER_CATEGORY',     // 医嘱分类
  PATIENT_STATUS: 'PATIENT_STATUS',     // 患者状态
  VASCULAR_ACCESS: 'VASCULAR_ACCESS',   // 血管通路类型
  VASCULAR_SITE: 'VASCULAR_SITE',       // 血管通路部位
  VEIN_TYPE: 'VEIN_TYPE',               // 静脉类型
  ARTERY_TYPE: 'ARTERY_TYPE',           // 动脉类型
  INSURANCE_TYPE: 'INSURANCE_TYPE',     // 医保类型
  PATIENT_TYPE: 'PATIENT_TYPE',         // 患者类型
  ID_TYPE: 'ID_TYPE',                   // 证件类型
  VISIT_CATEGORY: 'VISIT_CATEGORY',     // 就诊类别
  BLOOD_TYPE_ABO: 'BLOOD_TYPE_ABO',     // ABO血型
  BLOOD_TYPE_RH: 'BLOOD_TYPE_RH',       // Rh血型
  EDUCATION_LEVEL: 'EDUCATION_LEVEL',   // 文化程度
  MARITAL_STATUS: 'MARITAL_STATUS',     // 婚姻状况
  RELATIONSHIP_OPTIONS: 'RelationshipOptions', // 患者关系
  DOCTOR: 'DOCTOR',                     // 医生列表
  NURSE: 'NURSE',                       // 护士列表
  HOSPITAL: 'HOSPITAL',                 // 手术医院
  SURGERY_TYPE: 'SURGERY_TYPE',         // 手术类型（血管通路干预）
  // 临床诊疗分类
  PRIMARY_DISEASE: 'PRIMARY_DISEASE',   // 原发病分类
  COMPLICATION: 'COMPLICATION',         // 并发症类型
  PATHOLOGY: 'PATHOLOGY',               // 病理诊断分类
  TUMOR: 'TUMOR',                       // 肿瘤分类
  ALLERGEN: 'ALLERGEN',                 // 过敏原类型
  OUTCOME: 'OUTCOME',                   // 患者转归（一级：在科/转出，二级：具体原因）
  HEALTH_EDUCATION_TYPE: 'HealthEducationType', // 宣教内容类型
  HEALTH_CLASSIFY: 'HealthClassify',     // 宣教内容分类
  // 医嘱用药扩展
  ORDER_ROUTE: 'ORDER_ROUTE',             // 医嘱用法
  ORDER_FREQUENCY: 'ORDER_FREQUENCY',     // 医嘱频次
  ORDER_TIMING: 'ORDER_TIMING',           // 医嘱使用时机
} as const

// ===== 字典数据缓存 =====

class DictCache {
  private cache: Map<string, { data: DictItem[]; timestamp: number }> = new Map()
  private ttl = 5 * 60 * 1000 // 5分钟缓存

  async getItems(typeCode: string, forceRefresh = false): Promise<DictItem[]> {
    const cached = this.cache.get(typeCode)
    const now = Date.now()

    // 如果有缓存且未过期，直接返回
    if (!forceRefresh && cached && (now - cached.timestamp) < this.ttl) {
      return cached.data
    }

    // 从 API 获取数据
    const items = await dictApi.getItems(typeCode, true)

    // 更新缓存
    this.cache.set(typeCode, { data: items, timestamp: now })

    return items
  }

  clear(typeCode?: string) {
    if (typeCode) {
      this.cache.delete(typeCode)
    } else {
      this.cache.clear()
    }
  }

  // 获取下拉选项（仅返回 code 和 name）
  async getOptions(typeCode: string, forceRefresh = false): Promise<Array<{ value: string; label: string }>> {
    const items = await this.getItems(typeCode, forceRefresh)
    return items.map(item => ({
      value: item.code,
      label: item.name,
    }))
  }

  // 根据代码获取名称
  async getNameByCode(typeCode: string, code: string): Promise<string> {
    if (!code) return ''
    const items = await this.getItems(typeCode)
    const item = items.find(i => i.code === code)
    return item?.name || code
  }

  // 批量获取名称（用于列表显示）
  async getNameMap(typeCode: string): Promise<Map<string, string>> {
    const items = await this.getItems(typeCode)
    const map = new Map<string, string>()
    items.forEach(item => {
      map.set(item.code, item.name)
    })
    return map
  }

  // 获取级联选择选项（用于 Cascader 组件）
  async getCascaderOptions(typeCode: string): Promise<CascaderOption[]> {
    // 树形数据从 API 获取，不使用本地缓存
    const items = await dictApi.getItemsTree(typeCode, true)
    return this.buildCascaderOptions(Array.isArray(items) ? items : [])
  }

  // 递归构建级联选项
  private buildCascaderOptions(items?: DictItem[] | null): CascaderOption[] {
    if (!Array.isArray(items)) {
      return []
    }

    return items.filter((item): item is DictItem => !!item).map(item => {
      const option: CascaderOption = {
        value: item.code,
        label: item.name,
      }
      // 如果有子项，递归构建
      if (Array.isArray(item.children) && item.children.length > 0) {
        option.children = this.buildCascaderOptions(item.children)
      }
      return option
    })
  }
}

export const dictCache = new DictCache()

export default {
  dict: dictApi,
  cache: dictCache,
  types: DICT_TYPES,
}
