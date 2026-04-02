/**
 * 患者 API 服务
 * 用于处理患者相关的 API 请求
 */

import { apiClient } from './restClient'
import type { ApiSuccessResponse } from './restClient'

// ===== 类型定义 =====

// 透析模式
interface DialysisMode {
  mode: string           // HD, HDF, HD+HP
  bloodFlow: number      // 血流量
  bv: string             // 抗凝剂标识
  frequencyDesc: string  // 频率描述
  autoConfirm: boolean   // 自动确认
  status: string         // 启用, 禁用
  notes: string
  substituteInputMode?: string  // 置换液输入方式
  substituteFlow?: number       // 置换液流速
  substituteVolume?: number     // 置换液总量
}

// 抗凝剂
interface Anticoagulant {
  initialDrug: string    // 首剂量药物
  initialDose: string    // 首剂量
  totalDose: string      // 总量
  maintenanceDrug: string // 维持量药物
  infusionRate: string   // 输注速度
  infusionTime: string   // 输注时间
  maintenanceDose: string // 维持量
}

// 透析参数
interface DialysisParameters {
  dialysateType: string  // 透析液类型
  dialysateGroup: string // 透析液组号
  flowRate: number       // 透析液流量
  na: number             // 钠
  ca: number             // 钙
  k: number              // 钾
  hco3: number           // 碳酸氢根
  glucose: string        // 葡萄糖
  conductivity: number   // 电导度
  temp: number           // 温度
  volume: number         // 透析液量
}

// 材料
interface Material {
  id: string
  name: string
  category: string
  count: number
  code: string
  brand: string
  spec: string
  note: string
}

// 治疗方案
export interface TreatmentPlan {
  id: string
  patientId: string
  weeklyFrequency: number
  biweeklyFrequency: number
  duration: number
  dryWeight: number
  extraWeight: number
  status: string
  doctorId?: string
  startDate?: string
  endDate?: string
  notes: string
  dialysisMode: DialysisMode
  anticoagulant: Anticoagulant
  parameters: DialysisParameters
  materials: Material[]
  createdAt: string
  updatedAt: string
}

// 创建治疗方案请求
interface CreateTreatmentPlanRequest {
  weeklyFrequency?: number
  biweeklyFrequency?: number
  duration?: number
  dryWeight?: number
  extraWeight?: number
  status?: string
  notes?: string
  dialysisMode: DialysisMode
  anticoagulant: Anticoagulant
  parameters: DialysisParameters
  materials: Material[]
}

// 更新治疗方案请求
interface UpdateTreatmentPlanRequest {
  weeklyFrequency?: number
  biweeklyFrequency?: number
  duration?: number
  dryWeight?: number
  extraWeight?: number
  status?: string
  notes?: string
  dialysisMode?: DialysisMode
  anticoagulant?: Anticoagulant
  parameters?: DialysisParameters
  materials?: Material[]
}

// 方案调整记录
export interface AdjustmentRecord {
  id: string
  patientId: string
  content: string       // 调整内容描述
  operator: string      // 调整人
  createdAt: string     // 调整时间
}

// 创建调整记录请求
interface CreateAdjustmentRecordRequest {
  content: string
  operator?: string
}

// ===== 辅助函数 =====

async function get<T>(url: string): Promise<T> {
  const response = await apiClient.get<ApiSuccessResponse<T>>(url)
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

async function del(url: string): Promise<void> {
  const response = await apiClient.delete<ApiSuccessResponse<void>>(url)
  if (response.status === 204) {
    return
  }
  if (!response.data?.success) {
    throw new Error('API 请求失败')
  }
}

// ===== 患者 API =====

export const patientApi = {
  // 获取患者治疗方案列表
  getTreatmentPlans: async (patientId: string): Promise<TreatmentPlan[]> => {
    try {
      return await get<TreatmentPlan[]>(`/api/v1/patients/${patientId}/treatment-plans`)
    } catch (error) {
      if (error && typeof error === 'object' && 'response' in error) {
        const err = error as { response?: { status?: number } }
        if (err.response?.status === 404) {
          return []
        }
      }
      throw error
    }
  },

  // 获取患者特定模式的治疗方案
  getTreatmentPlan: async (patientId: string, mode?: string): Promise<TreatmentPlan | null> => {
    try {
      const url = mode
        ? `/api/v1/patients/${patientId}/treatment-plan?mode=${mode}`
        : `/api/v1/patients/${patientId}/treatment-plan`
      return await get<TreatmentPlan>(url)
    } catch (error) {
      if (error && typeof error === 'object' && 'response' in error) {
        const err = error as { response?: { status?: number } }
        if (err.response?.status === 404) {
          return null
        }
      }
      throw error
    }
  },

  // 创建患者治疗方案
  createTreatmentPlan: async (patientId: string, data: CreateTreatmentPlanRequest): Promise<TreatmentPlan> => {
    return post<TreatmentPlan>(`/api/v1/patients/${patientId}/treatment-plan`, data)
  },

  // 更新患者治疗方案
  updateTreatmentPlan: async (patientId: string, data: UpdateTreatmentPlanRequest): Promise<TreatmentPlan> => {
    return put<TreatmentPlan>(`/api/v1/patients/${patientId}/treatment-plan`, data)
  },

  // 删除患者治疗方案
  deleteTreatmentPlan: async (patientId: string): Promise<void> => {
    return del(`/api/v1/patients/${patientId}/treatment-plan`)
  },

  // 获取方案调整记录列表
  getAdjustmentRecords: async (patientId: string): Promise<AdjustmentRecord[]> => {
    try {
      return await get<AdjustmentRecord[]>(`/api/v1/patients/${patientId}/adjustment-records`)
    } catch (error) {
      if (error && typeof error === 'object' && 'response' in error) {
        const err = error as { response?: { status?: number } }
        if (err.response?.status === 404) {
          return []
        }
      }
      throw error
    }
  },

  // 创建方案调整记录
  createAdjustmentRecord: async (patientId: string, data: CreateAdjustmentRecordRequest): Promise<AdjustmentRecord> => {
    return post<AdjustmentRecord>(`/api/v1/patients/${patientId}/adjustment-records`, data)
  },
}

export default patientApi
