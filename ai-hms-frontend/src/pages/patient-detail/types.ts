// 患者详情页共享类型

import type { Patient } from '@/types/original'

// Tab ID 类型（保留旧类型兼容）
export type TabID = 'overview' | 'basic_info' | 'treatment_plan' | 'medical_record' | 'scheme_order' | 'labs_exams' | 'vascular' | 'history' | 'monthly_summary'

// 主 Tab ID（U3 重构后）
export type MainTabID = 'overview' | 'treatment' | 'records' | 'history'
export type TreatmentSubTab = 'plan' | 'schemeOrder' | 'vascular'
export type RecordsSubTab = 'basicInfo' | 'labs' | 'monthly' | 'medicalRecord'

// 趋势数据项
export interface TrendDataItem {
  month: string
  hgb: number
  hgbColor: string
  ca: number
  caColor: string
  p: number
  pColor: string
}

// 治疗转归记录
export interface OutcomeRecord {
  id: string
  type: string          // 存储字典 code
  typeName?: string     // 显示名称（从字典获取）
  reason: string        // 存储字典 code
  reasonName?: string   // 显示名称（从字典获取）
  time: string
  remarks: string
  registrar: string
  registrationTime: string
  isDoorRule: boolean  // 是否门规
}

// Tab 组件通用 Props
export interface TabProps {
  patient: Patient
}

// 检验结果类型
export interface LabResultItem {
  id: string
  code: string
  name: string
  value: string
  unit: string
  reference: string
  date: string
  isAbnormal: boolean
  abnormalType?: 'high' | 'low'
  pendingSync?: boolean
}

// 血管通路记录
export interface VascularRecord {
  id: string
  accessType: string
  site: string
  artery: string[]
  vein: string[]
  side: string
  hospital: string
  surgeon: string
  surgeryDate: string
  firstUseDate: string
  accessNumber: number
  interventionCount: number
  interventionDate: string
  catheterMethod?: string
  catheterDepth?: string
  vPuncturePosition: string[]
  aPuncturePosition: string[]
  notes: string
  images: string[]
  isDefault: boolean
  isDisabled: boolean
  createdAt: string
}

// 治疗历史记录
export interface TreatmentHistoryItem {
  id: string
  date: string
  timeRange: string
  mode: string
  duration?: string
  weightLoss?: number
  startBP?: string
  endBP?: string
  complications?: string
  doctor?: string
  doctorSummary: string
  treatmentSummary: string
}

// 医嘱项
export interface OrderItem {
  id: string
  name: string
  dose?: string
  route?: string
  frequency?: string
  startTime: string
  endTime?: string
  doctor: string
  status: '执行中' | '已停止' | '已完成'
  isGrouped?: boolean
}
