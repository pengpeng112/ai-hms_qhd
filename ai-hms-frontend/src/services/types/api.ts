/**
 * API 类型定义
 * 基于 HDIS GraphQL API 接口说明书
 */

// ============ 通用类型 ============

/** 带 RowCount 的分页项 */
export interface PaginatedItem {
  RowCount?: number
}

/** 分页响应 */
export interface PaginatedResponse<T> {
  data: T[]
  total: number
}

/** 查询参数 */
export interface QueryParams {
  page?: number
  pageSize?: number
  filter?: Record<string, unknown>
}

// ============ 患者相关 (4.3.1-4.3.8) ============

/** 4.3.1 患者基本信息 PatientInfomation */
export interface Patient extends PaginatedItem {
  Id: number
  TenantId: number
  Name: string
  Spell?: string
  Type?: string
  TreatmentStatus?: string
  OutComeStatus?: string
  Gender?: string
  BirthDate?: string
  PhoneNo?: string
  ExpenseType?: string
  DialysisNo?: string
}

/** 4.3.3 住院信息 Hospitalization */
export interface Hospitalization extends PaginatedItem {
  Id: number
  TenantId?: number
  PatientId: number
  CaseNo?: string
  HospNo?: string
  BarCode?: string
  HospPatientType?: string
  HospReceiveDept?: string
  Status?: string
}

/** 4.3.4 感染记录 Infection */
export interface Infection extends PaginatedItem {
  Id: number
  TenantId?: number
  PatientId: number
  InfectionDesc?: string
  InfectionType?: string
  OtherDesc?: string
  Note?: string
  CreatorId?: number
  CreateTime?: string
  TestDate?: string
  Result?: string
}

/** 4.3.5 血管通路 VascularAccess */
export interface VascularAccess extends PaginatedItem {
  Id: number
  TenantId?: number
  PatientId: number
  AccessType?: string
  AccessPosition?: string
  Artery?: string
  Venous?: string
  LeftAndRight?: string
  Status?: string
}

/** 病历记录 (未在 API 文档中找到) */
export interface CaseHistory extends PaginatedItem {
  Id: number
  PatientId: number
  RecordDate?: string
  RecordType?: string
  Content?: string
  DoctorName?: string
  CreateTime?: string
}

// ============ 排班相关 (4.3.20-4.3.24) ============

/** 4.3.20 班次 Shift */
export interface Shift extends PaginatedItem {
  Id: number
  TenantId: number
  Name?: string
  Sort?: number
  StartTime?: string
  EndTime?: string
  Type?: string
  Note?: string
  Status?: string
}

/** 4.3.21 患者排班 PatientShift */
export interface PatientShift extends PaginatedItem {
  Id: number
  TenantId?: number
  PatientId: number
  TreatmentTime?: string
  ShiftId: number
  WardId?: number
  BedId?: number
  PatientPlanId?: number
}

/** 4.3.22 床位 Bed */
export interface Bed extends PaginatedItem {
  Id: number
  TenantId: number
  Name?: string
  WardId?: number
  Sort?: number
  Status?: string
  Note?: string
}

/** 4.3.24 病区 Ward */
export interface Ward extends PaginatedItem {
  Id: number
  TenantId: number
  Name?: string
  Sort?: number
  Status?: string
  Note?: string
}

// ============ 治疗相关 (4.3.9-4.3.15, 4.3.32) ============

/** 4.3.32 治疗记录 Treatment */
export interface Treatment extends PaginatedItem {
  Id: number
  TenantId?: number
  PatientId: number
  ScheduleId?: number
  ReceptionDrId?: number
  SignInTime?: string
  QueueNo?: string
  ReceptionTime?: string
  TreatmentDate?: string
  ShiftId?: number
  Status?: string
  StartTime?: string
  EndTime?: string
}

/** 4.3.13 透析处方 PatientPrescription */
export interface PatientPrescription extends PaginatedItem {
  Id: number
  TenantId?: number
  PatientId: number
  TreatmentTime?: string
  TreatmentId?: number
  PatientPlanId?: number
  DialyzerType?: string
  DialysisDuration?: number
  BloodFlow?: number
}

/** 4.3.11 透析方案 PatientPlan */
export interface PatientPlan extends PaginatedItem {
  Id: number
  TenantId?: number
  PatientId: number
  PlanTPLId?: number
  Status?: string
  CreatorId?: number
  CreateTime?: string
}

// ============ 医嘱相关 (4.3.16-4.3.19) ============

/** 4.3.16 医嘱模板 OrderTPL */
export interface OrderTPL extends PaginatedItem {
  Id: number
  TenantId: number
  Name?: string
  OrderGroup?: string
  IsDisabled?: string
  Classification?: string
  DrugId?: number
  Content?: string
}

/** 4.3.17 患者医嘱 PatientOrder */
export interface PatientOrder extends PaginatedItem {
  Id: number
  TenantId?: number
  PatientId: number
  OrderTPLId?: number
  OrderGroup?: string
  Type?: string
  DrugId?: number
  Classification?: string
}

/** 4.3.18 患者日医嘱 PatientDayOrder */
export interface PatientDayOrder extends PaginatedItem {
  Id: number
  TenantId?: number
  PatientId: number
  TreatmentTime?: string
  PatientOrderId?: number
  OrderGroup?: string
  Status?: string
  CaseStatus?: string
}

// ============ 体征相关 (4.3.31, 4.3.36, 4.3.38) ============

/** 4.3.31 透前体征 BeforeSigns */
export interface BeforeSigns extends PaginatedItem {
  Id: number
  TenantId?: number
  TreatmentId: number
  PatientId?: number
  OperatorId?: number
  OperateTime?: string
  RecordTime?: string
  Weight?: number
  ExtraWeight?: number
  DryWeight?: number
  SystolicBP?: number
  DiastolicBP?: number
  Pulse?: number
  Temperature?: number
}

/** 4.3.38 透中体征 DuringSigns */
export interface DuringSigns extends PaginatedItem {
  Id: number
  TenantId?: number
  TreatmentId: number
  PatientId?: number
  OperatorId?: number
  OperateTime?: string
  RecordTime?: string
  SBP?: number
  DBP?: number
  SystolicBP?: number
  DiastolicBP?: number
  Pulse?: number
  BloodFlow?: number
  VenousPressure?: number
  TMP?: number
  UFVolume?: number
}

/** 4.3.36 透后体征 AfterSigns */
export interface AfterSigns extends PaginatedItem {
  Id: number
  TenantId?: number
  TreatmentId: number
  OperatorId?: number
  OperateTime?: string
  Weight?: number
  SBP?: number
  DBP?: number
}

// ============ 检验相关 (4.3.25-4.3.26) ============

/** 4.3.25 检验报告 Examination */
export interface Examination extends PaginatedItem {
  Id: number
  TenantId?: number
  PatientId: number
  ExamType?: string
  ExamDate?: string
  ReportDate?: string
  Status?: string
}

/** 4.3.26 检验项目 ExaminationItem */
export interface ExaminationItem extends PaginatedItem {
  Id: number
  TenantId?: number
  ExaminationId: number
  ItemName?: string
  ItemCode?: string
  Result?: string
  Unit?: string
  RefRange?: string
  AbnormalFlag?: string
}

// ============ 设备相关 (4.3.29-4.3.30) ============

/** 4.3.29 设备档案信息 EquipmentInfomation */
export interface EquipmentInfo extends PaginatedItem {
  Id: number
  TenantId: number
  Name?: string
  IDNo?: string
  SerialNo?: string
  Brand?: string
  ModelNo?: string
  DialysisMethod?: string
}

/** 4.3.30 设备消毒记录 EquipmentDisinfection */
export interface EquipmentDisinfection extends PaginatedItem {
  Id: number
  TenantId?: number
  EquipmentId: number
  DisinfectUserId?: number
  DisinfectWay?: string
  StartTime?: string
  Description?: string
  Note?: string
}

// ============ 兼容旧代码的别名 ============
export type OrderTemplate = OrderTPL
export type MachineInfo = EquipmentInfo
export type MachineRunRecord = EquipmentDisinfection

// ============ GraphQL 响应映射 ============

/** 实体名称类型 */
export type EntityName =
  | 'PatientInfomation'
  | 'IDInfomation'
  | 'Hospitalization'
  | 'Infection'
  | 'VascularAccess'
  | 'Shift'
  | 'PatientShift'
  | 'Bed'
  | 'Ward'
  | 'Treatment'
  | 'PatientPrescription'
  | 'PatientPlan'
  | 'OrderTPL'
  | 'PatientOrder'
  | 'PatientDayOrder'
  | 'BeforeSigns'
  | 'DuringSigns'
  | 'AfterSigns'
  | 'Examination'
  | 'ExaminationItem'
  | 'EquipmentInfomation'
  | 'EquipmentDisinfection'
