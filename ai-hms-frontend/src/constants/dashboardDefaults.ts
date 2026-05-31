import { UserRole } from '@/types/original'

const ADMIN = 'ADMIN' as UserRole

// 角色 → 默认卡片 key 列表（按位置 1-4）
export const DASHBOARD_DEFAULTS_BY_ROLE: Record<string, string[]> = {
  DOCTOR_SUPERVISOR: ['myPatientsToday', 'pendingPrescriptions', 'recent7dTreatments', 'abnormalLabs'],
  DOCTOR_DUTY:       ['myPatientsToday', 'pendingPrescriptions', 'recent7dTreatments', 'abnormalLabs'],
  DOCTOR_CHIEF:      ['myPatientsToday', 'pendingPrescriptions', 'recent7dTreatments', 'abnormalLabs'],
  NURSE_HEAD:        ['todayShiftMatrix', 'pendingOrders', 'pendingPreAssessment', 'onlineDevices'],
  NURSE_RESPONSIBLE: ['todayShiftMatrix', 'pendingOrders', 'pendingPreAssessment', 'onlineDevices'],
  NURSE_MANAGER:     ['todayShiftMatrix', 'pendingOrders', 'pendingPreAssessment', 'onlineDevices'],
  NURSE_SCHEDULER:   ['todayShiftMatrix', 'pendingScheduleQueue', 'bedUtilization', 'contractAnomaly'],
  ADMIN:             ['operationKpis', 'deviceUtilization', 'patientGrowth', 'treatmentTrend'],
  DEFAULT:           ['myPatientsToday', 'pendingPrescriptions', 'recent7dTreatments', 'abnormalLabs'],
}

export function getDefaultCardKeys(role: string): string[] {
  return DASHBOARD_DEFAULTS_BY_ROLE[role] || DASHBOARD_DEFAULTS_BY_ROLE.DEFAULT
}

// v2 卡片注册表（所有可用卡片定义）
export interface V2CardDef {
  id: string
  titleKey: string
  title: string
  type: 'stat' | 'list' | 'chart' | 'action' | 'monitor' | 'inventory' | 'binding'
  size: 'small' | 'medium' | 'large'
  roles: UserRole[]
}

export const V2_CARD_REGISTRY: V2CardDef[] = [
  { id: 'myPatientsToday',      titleKey: 'dashboard:card.myPatientsToday',      title: 'My Patients Today',      type: 'list',   size: 'large',  roles: [ADMIN, UserRole.DOCTOR_SUPERVISOR, UserRole.DOCTOR_DUTY] },
  { id: 'pendingPrescriptions', titleKey: 'dashboard:card.pendingPrescriptions', title: 'Pending Prescriptions',   type: 'action', size: 'medium', roles: [ADMIN, UserRole.DOCTOR_SUPERVISOR, UserRole.DOCTOR_DUTY, UserRole.DOCTOR_CHIEF] },
  { id: 'recent7dTreatments',   titleKey: 'dashboard:card.recent7dTreatments',   title: 'Recent 7d Treatments',    type: 'chart',  size: 'medium', roles: [ADMIN, UserRole.DOCTOR_SUPERVISOR, UserRole.DOCTOR_DUTY] },
  { id: 'abnormalLabs',         titleKey: 'dashboard:card.abnormalLabs',         title: 'Abnormal Labs',           type: 'list',   size: 'medium', roles: [ADMIN, UserRole.DOCTOR_SUPERVISOR, UserRole.DOCTOR_DUTY] },
  { id: 'todayShiftMatrix',     titleKey: 'dashboard:card.todayShiftMatrix',     title: 'Today Shift Matrix',      type: 'monitor',size: 'large',  roles: [ADMIN, UserRole.NURSE_HEAD, UserRole.NURSE_RESPONSIBLE, UserRole.NURSE_MANAGER, UserRole.NURSE_SCHEDULER] },
  { id: 'pendingOrders',        titleKey: 'dashboard:card.pendingOrders',        title: 'Pending Orders',          type: 'action', size: 'medium', roles: [ADMIN, UserRole.NURSE_HEAD, UserRole.NURSE_RESPONSIBLE, UserRole.NURSE_MANAGER] },
  { id: 'pendingPreAssessment', titleKey: 'dashboard:card.pendingPreAssessment', title: 'Pending Pre-Assessment',  type: 'action', size: 'medium', roles: [ADMIN, UserRole.NURSE_HEAD, UserRole.NURSE_RESPONSIBLE, UserRole.NURSE_MANAGER] },
  { id: 'onlineDevices',        titleKey: 'dashboard:card.onlineDevices',        title: 'Online Devices',          type: 'monitor',size: 'medium', roles: [ADMIN, UserRole.NURSE_HEAD, UserRole.NURSE_RESPONSIBLE, UserRole.NURSE_MANAGER] },
  { id: 'pendingScheduleQueue', titleKey: 'dashboard:card.pendingScheduleQueue', title: 'Pending Schedule Queue',  type: 'list',   size: 'medium', roles: [ADMIN, UserRole.NURSE_SCHEDULER] },
  { id: 'bedUtilization',       titleKey: 'dashboard:card.bedUtilization',       title: 'Bed Utilization',         type: 'stat',   size: 'medium', roles: [ADMIN, UserRole.NURSE_SCHEDULER] },
  { id: 'contractAnomaly',      titleKey: 'dashboard:card.contractAnomaly',      title: 'Contract Anomaly',        type: 'action', size: 'medium', roles: [ADMIN, UserRole.NURSE_SCHEDULER] },
  { id: 'operationKpis',        titleKey: 'dashboard:card.operationKpis',        title: 'Operation KPIs',          type: 'stat',   size: 'large',  roles: [ADMIN] },
  { id: 'deviceUtilization',    titleKey: 'dashboard:card.deviceUtilization',    title: 'Device Utilization',      type: 'chart',  size: 'medium', roles: [ADMIN] },
  { id: 'patientGrowth',        titleKey: 'dashboard:card.patientGrowth',        title: 'Patient Growth',          type: 'chart',  size: 'medium', roles: [ADMIN] },
  { id: 'treatmentTrend',       titleKey: 'dashboard:card.treatmentTrend',       title: 'Treatment Trend',         type: 'chart',  size: 'medium', roles: [ADMIN] },
]

// localStorage v2 key
export const LAYOUT_STORAGE_KEY_V2 = 'dashboard_layout_config_v2'
