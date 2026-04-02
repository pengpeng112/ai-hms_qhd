// 治疗方案表单共享类型

// 透析模式数据
export interface DialysisModeValues {
  method: string
  duration: string
  bloodFlow: string
  bv: string
  substituteMode: string
  substituteFlow: string
  substituteVolume: string
  notes: string
}

// 抗凝剂数据
export interface AnticoagulantValues {
  initialDrug: string
  initialDose: string
  maintenanceDrug: string
  infusionRate: string
  infusionTime: string
  maintenanceDose: string
  totalDose: string
}

// 透析参数数据
export interface DialysisParamsValues {
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
  dialysateVolume: string
}

// 材料项
export interface MaterialItem {
  id: string
  name: string
  category: string
  count: number
  code: string
  brand: string
  spec: string
  note: string
}

// 字典选项
export type DictOptions = Record<string, Array<{ value: string; label: string }>>

// 药品选项
export interface DrugOption {
  value: string
  label: string
  concentration?: string
}

// NumericPad 配置
export type OpenNumericPad = (config: {
  label: string
  value: string
  suffix?: string
  disabled?: boolean
  onConfirm: (value: string) => void
}) => void
