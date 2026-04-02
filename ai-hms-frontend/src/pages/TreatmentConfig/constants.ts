// TreatmentConfig 静态配置

import { FileText, Container, Pill, ClipboardList } from 'lucide-react'
import type { TabConfig, PlanFormData, OrderGroupFormData } from './types'

// Tab 配置
export const TAB_CONFIG: TabConfig[] = [
  { key: 'PLAN', label: '方案模版', icon: FileText, color: 'text-blue-600', bgColor: 'bg-blue-50' },
  { key: 'ORDER', label: '医嘱模版', icon: ClipboardList, color: 'text-purple-600', bgColor: 'bg-purple-50' },
  { key: 'MATERIAL', label: '材料目录', icon: Container, color: 'text-orange-600', bgColor: 'bg-orange-50' },
  { key: 'DRUG', label: '药品目录', icon: Pill, color: 'text-emerald-600', bgColor: 'bg-emerald-50' },
]

// 方案表单初始值
export const INITIAL_PLAN_FORM: PlanFormData = {
  id: '',
  name: '',
  method: 'HD',
  time: '4.0',
  bloodFlow: '250',
  note: '',
  initialAnticoag: '肝素钠',
  initialDose: '',
  maintenanceAnticoag: '',
  infusionRate: '',
  infusionTime: '',
  maintenanceDose: '',
  totalDose: '',
  dialysateType: 'A液+B液',
  dialysateGroup: 'A16-K2/Ca1.25',
  dialysateFlow: '500',
  na: '140',
  ca: '1.5',
  k: '2.0',
  hco3: '35',
  glucose: '无糖',
  conductivity: '',
  temp: '37.0',
  volume: '',
  category: '常规',
  description: ''
}

// 医嘱组表单初始值
export const INITIAL_ORDER_GROUP_FORM: OrderGroupFormData = {
  name: '',
  type: '基础类'
}

// 材料表单初始值
export const INITIAL_MATERIAL_FORM = {
  code: '',
  name: '',
  shortName: '',
  mnemonic: '',
  category: '',
  spec: '',
  standardType: '',
  brand: '',
  unit: '',
  packaging: '',
  manufacturer: '',
  sortOrder: 0,
  isEnabled: true,
  notes: ''
}

// 药品表单初始值
export const INITIAL_DRUG_FORM = {
  code: '',
  name: '',
  shortName: '',
  mnemonic: '',
  genericName: '',
  category: '',
  spec: '',
  concentration: '',
  specUnit: '',
  minUnitDose: '',
  baseUnit: '',
  brand: '',
  packaging: '',
  manufacturer: '',
  standardType: '',
  timing: '',
  tips: '',
  sortOrder: 0,
  isDisabled: false,
  note: ''
}

// 分页大小
export const PAGE_SIZE = 10
export const LOAD_ALL_PAGE_SIZE = 9999
