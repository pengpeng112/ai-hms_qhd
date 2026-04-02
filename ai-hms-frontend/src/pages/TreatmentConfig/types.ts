// TreatmentConfig 类型定义

export type SubView = 'PLAN' | 'ORDER' | 'MATERIAL' | 'DRUG'

// Tab 配置
export interface TabConfig {
  key: SubView
  label: string
  icon: React.ElementType
  color: string
  bgColor: string
}

// 模态框中的材料
export interface ModalMaterial {
  id: string
  name: string
  category: string
  count: number
  code: string
  brand: string
  spec: string
  note: string
}

// 模态框中的医嘱
export interface ModalOrder {
  id: string
  name?: string
  content: string
  spec?: string
  minUnitDose?: string
  dosage: string
  unit: string
  route: string
  frequency: string
  timing: string
  status: string
  selected?: boolean
  groupId?: string
}

// 方案表单
export interface PlanFormData {
  id: string
  name: string
  method: string
  time: string
  bloodFlow: string
  note: string
  initialAnticoag: string
  initialDose: string
  maintenanceAnticoag: string
  infusionRate: string
  infusionTime: string
  maintenanceDose: string
  totalDose: string
  dialysateType: string
  dialysateGroup: string
  dialysateFlow: string
  na: string
  ca: string
  k: string
  hco3: string
  glucose: string
  conductivity: string
  temp: string
  volume: string
  category: string
  description: string
}

// 医嘱组表单
export interface OrderGroupFormData {
  name: string
  type: string
}
