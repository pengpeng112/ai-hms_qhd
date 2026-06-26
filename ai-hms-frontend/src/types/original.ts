// 前端 UI 聚合模型（camelCase），用于患者详情、列表、工作台等页面

// 角色定义 - 与原版 UI v1.3 保持一致
export const UserRole = {
  // 医生组
  DOCTOR_CHIEF: 'DOCTOR_CHIEF',           // 科室主任
  DOCTOR_SUPERVISOR: 'DOCTOR_SUPERVISOR', // 主管医生
  DOCTOR_DUTY: 'DOCTOR_DUTY',             // 值班医生
  // 护士组
  NURSE_HEAD: 'NURSE_HEAD',               // 护士长
  NURSE_SCHEDULER: 'NURSE_SCHEDULER',     // 主班护士（排班）
  NURSE_MANAGER: 'NURSE_MANAGER',         // 总管护士（库管）
  NURSE_RESPONSIBLE: 'NURSE_RESPONSIBLE', // 责任护士
  // 技术组
  ENGINEER: 'ENGINEER',                   // 设备工程师
} as const

export type UserRole = typeof UserRole[keyof typeof UserRole]

// 角色中文名称映射
export const UserRoleLabel: Record<UserRole, string> = {
  [UserRole.DOCTOR_CHIEF]: '科室主任',
  [UserRole.DOCTOR_SUPERVISOR]: '主管医生',
  [UserRole.DOCTOR_DUTY]: '值班医生',
  [UserRole.NURSE_HEAD]: '护士长',
  [UserRole.NURSE_SCHEDULER]: '主班护士',
  [UserRole.NURSE_MANAGER]: '总管护士',
  [UserRole.NURSE_RESPONSIBLE]: '责任护士',
  [UserRole.ENGINEER]: '设备工程师',
}

// 角色分组
export const RoleGroups = {
  DOCTOR: [UserRole.DOCTOR_CHIEF, UserRole.DOCTOR_SUPERVISOR, UserRole.DOCTOR_DUTY],
  NURSE: [UserRole.NURSE_HEAD, UserRole.NURSE_SCHEDULER, UserRole.NURSE_MANAGER, UserRole.NURSE_RESPONSIBLE],
  TECH: [UserRole.ENGINEER],
} as const

// 所有角色列表
export const ALL_ROLES = Object.values(UserRole)

export interface User {
  id: string;
  name: string;
  role: UserRole;
  avatar: string;
}

export type ViewState = 
  | 'DASHBOARD' 
  | 'WARD_OVERVIEW'
  | 'PATIENT_LIST'
  | 'PATIENT_FULL_VIEW'
  | 'PATIENT_DETAIL'
  | 'DIALYSIS_PROCESSING' 
  | 'MONITOR' 
  | 'SCHEDULE' 
  | 'STATISTICS' 
  | 'INVENTORY'
  | 'DEVICE_BINDING'
  | 'MASTER_DATA'
  | 'SETTINGS';

export interface LabResult {
  id: string;
  name: string;
  value: string;
  unit: string;
  date: string;
  isAbnormal: boolean;
  reference: string;
}

export interface MedicalOrder {
  id: string;
  content: string;
  type: '长期' | '临时';
  status: '已执行' | '待执行';
  doctor: string;
  startTime: string;
}

export interface EMRDocument {
  id: string;
  title: string;
  type: '病程记录' | '知情同意书' | '首诊记录' | '检查报告' | '评估单';
  author: string;
  date: string;
  content?: string;
}

export interface InfectionInfo {
  hbsag: '阴性' | '阳性';
  hcvab: '阴性' | '阳性';
  hivab: '阴性' | '阳性';
  tpab: '阴性' | '阳性';
  tb?: '阴性' | '阳性';  // 结核 (可选)
  updateDate: string;
}

export interface PlanAdjustmentRecord {
  id: string;
  date: string;
  content: string;
  operator: string;
  reason?: string;
}

export interface TreatmentPlan {
  weeklyFrequency: number;
  biweeklyFrequency: number;
  duration: number;
  dryWeight: number;
  extraWeight: number;
  vascularAccess: string;
  indicators: {
    mode: string;
    bloodFlow: number;
    bv: string;
    frequencyDesc: string;
    autoConfirm: boolean;
    status: '启用' | '禁用';
    notes: string;
  };
  anticoagulant: {
    initialDrug: string;
    initialDose: string;
    maintenanceDrug: string;
    infusionRate: string;
    infusionTime: string;
    maintenanceDose: string;
    totalDose: string;
  };
  parameters: {
    dialysateType: string;
    dialysateGroup: string;
    flowRate: number;
    na: number;
    ca: number;
    k: number;
    hco3: number;
    glucose: string;
    conductivity: number;
    temp: number;
    volume: number;
  };
  materials: {
    id: string;
    name: string;
    category: string;
    count: number;
    code: string;
    brand: string;
    spec: string;
    note: string;
  }[];
  adjustmentHistory: PlanAdjustmentRecord[];
}

export interface DialysisSession {
  id: string;
  startTime: string;
  endTime?: string;
  status: 'ongoing' | 'completed' | 'paused';
  mode: string;
  deviceId?: string;
}

export interface Patient {
  id: string;
  name: string;
  avatar?: string;
  age: number;
  gender: '男' | '女';
  bedNumber: string;
  diagnosis: string;
  riskLevel: '高危' | '中危' | '低危';
  status: string;
  patientType: string;
  insuranceType: string;
  dryWeight: number;
  defaultMode: string;
  doctorName: string;
  vitals: {
    bp: string;
    hr: number;
    spO2: number;
    weight: number;
  };
  dialysisParams: {
    timeRemaining: string;
    ufRate: number;
    targetUf: number;
    accumulatedUf: number;
    bloodFlow: number;
    dialysateFlow: number;
    mode: string;
  };
  vascularAccess: {
    type: string;
    site: string;
    status: string;
    firstTime?: string;
    history?: string;
  };
  recentLabs: LabResult[];
  orders: MedicalOrder[];
  infection?: InfectionInfo;
  documents: EMRDocument[];
  progressNotes: { date: string, author: string, content: string }[];
  medicalHistory: {
    allergies: string[];
    primaryDisease: string;
    pathology: string;
    tumorInfo: string;
    medicalHistory: string;
    complications: string[];
  };
  outcome: {
    status: string;
    date?: string;
    desc?: string;
  };
  currentSession?: DialysisSession;
  treatmentPlan?: TreatmentPlan;
}

export interface DashboardCardConfig {
  id: string;
  title: string;
  type: 'stat' | 'list' | 'chart' | 'action' | 'monitor' | 'inventory' | 'binding';
  size: 'small' | 'medium' | 'large';
  roles: UserRole[];
}

export interface MonitorAlert {
  metric: string;
  level: string;
  value: number;
}

export interface MonitorIdhRisk {
  available: boolean;
  probability: number;
  level: string;
}

export interface MonitorDevice {
  id: string;
  bedNumber: string;
  patientName: string;
  patientId?: string;
  status: 'normal' | 'warning' | 'alarm' | 'offline' | 'unknown';
  mode: string;
  timeRemaining: string;
  vitals: {
    sbp: number;
    dbp: number;
    hr: number;
    bf: number;
    tmp: number;
    ufGoal: number;
    ufVolume: number;
    conductivity: number;
    temp: number;
    vp?: number;
  };
  alarms: string[];
  // 实时监控 §9 扩展（来自 live-data，可选，不影响其它消费者）
  treatmentId?: number;
  age?: number;
  dialysisNo?: string;
  accessType?: string;
  wardId?: number;
  startTime?: string;
  estimatedDuration?: number;
  alarmLevel?: 'normal' | 'warning' | 'danger';
  alerts?: MonitorAlert[];
  idhRisk?: MonitorIdhRisk;
  isMine?: boolean;
  vitalsSeries?: VitalSample[];
  rnaCompletion?: RNaCompletionView;
}

export interface RNaCompletionView {
  available: boolean;
  percent: number;
  targetRNa: number;
  mTarget: number;
  mRealized: number;
  cPre: number;
  cPreAt: string;
}

export interface VitalSample {
  t: string;
  sbp: number;
  dbp: number;
  map: number;
  hr: number;
  kind: string;
}

// Added StaffMember interface
export interface StaffMember {
  id: string;
  name: string;
  role: string;
  level: string;
}

// Added Shift interface
export interface Shift {
  date: string;
  staffId: string;
  type: 'A' | 'P' | 'N' | 'OFF';
  area: string;
  isConflict?: boolean;
}

// Added PatientScheduleItem interface
export interface PatientScheduleItem {
  id: string;
  bedNumber: string;
  date: string;
  shift: 'Morning' | 'Afternoon' | 'Evening';
  patientName: string;
  mode: string;
  patientId?: string;  // 后端关联的患者ID
}

// ============ 医嘱相关类型 ============

/** 详细医嘱数据 - 用于 SchemeOrderTab */
export interface DetailedOrder {
  id: string
  name: string
  dose: string
  unit: string
  doctor: string
  startTime: string
  stopTime?: string
  type: '长期' | '临时'
  status: '在用' | '停用'
}

/** 药品分组数据 - 用于医嘱单 */
export interface MedGroup {
  category: string
  items: { drug: string; dose: string; freq: string }[]
}

/** 医嘱单日期项 */
export type OrderSheetDate = string
