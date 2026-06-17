/**
 * REST API 客户端
 * 用于对接后端 REST 接口
 */

import axios from 'axios'
import { userApi, type CreateUserRequest, type UpdateUserRequest, type UserListParams, type RestUser } from './userApi'
import { roleManagementApi, type AppRoleApi, type PermissionNodeApi } from './roleManagementApi'

// API 配置
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL?.trim() || ''

const API_CONFIG = {
  baseURL: API_BASE_URL,
  timeout: 10000,
}

// 标准响应格式
export interface ApiSuccessResponse<T> {
  success: true
  data: T
  timestamp: string
}

export interface ApiErrorResponse {
  success: false
  error: {
    code: string
    message: string
    details?: unknown
  }
  timestamp: string
}

export type ApiResponse<T> = ApiSuccessResponse<T> | ApiErrorResponse

// 登录请求
export interface LoginRequest {
  username: string
  password: string
}

// 登录响应
export interface LoginResponse {
  token: string
  userId: string
  username: string
  realName: string
  role: string
  roles?: string[]
}

// 分页响应元数据
export interface PaginationMeta {
  page: number
  pageSize: number
  total: number
  totalPages: number
}

// 分页响应
export interface PaginatedResponse<T> {
  items: T[]
  pagination: PaginationMeta
}

// 鎮ｈ€?REST 鏁版嵁鏍煎紡锛堝悗绔繑鍥烇級
export interface RestPatient {
  id: string
  name: string
  age: number
  gender: string
  bedNumber?: string
  status: string
  patientType?: string
  insuranceType?: string
  dryWeight?: number
  defaultMode?: string
  doctorName?: string
  imageBase64String?: string
}

// ============ 班次 & 排班类型 ============

export interface RestInventoryItem {
  id: string
  tenantId: number
  name: string
  spec: string
  category: string
  categoryLabel?: string
  price: number
  stock: number
  unit: string
  minStock: number
  maxStock: number
  location: string
  supplier: string
  isDisabled: boolean
  creatorId: number
  createdAt: string
  updatedAt: string
  alert: boolean        // computed: stock < minStock
  lastUpdated: string   // formatted updatedAt
}

export interface RestStockLog {
  id: string
  tenantId: number
  itemId: string
  itemName: string
  type: 'in' | 'out'
  quantity: number
  unit: string
  operator: string
  note: string
  createdAt: string
}

export interface RestLabelTask {
  id: string
  tenantId: number
  itemId: string
  itemName: string
  spec: string
  quantity: number
  status: 'pending' | 'printing' | 'completed'
  creatorId: number
  createdAt: string
  updatedAt: string
}

export interface RestDevice {
  id: string
  tenantId: number
  name: string
  serialNo: string
  model: string
  manufacturer: string
  bedNumber: string
  wardId: number | null
  status: string   // normal | warning | alarm | offline | maintenance
  notes: string
  isDisabled: boolean
  creatorId: number
  createdAt: string
  updatedAt: string
}

// ============ 治疗记录类型 ============

export interface RestDuringParam {
  id: number
  tenantId: number
  treatmentId: number
  recordTime: string
  code: string
  sbp?: number
  dbp?: number
  heartRate?: number
  respiration?: number
  spO2?: number
  bloodFlow?: number
  dialysateFlow?: number
  ufVolume?: number
  venousPressure?: number
  arterialPressure?: number
  tmp?: number
  temperature?: number
  conductivity?: number
  ufRate?: number
  notes?: string
  creatorId: number
  createTime: string
  lastModifyTime: string
}

export interface RestPrescriptionDayStatus {
  patientId: string
  hasPrescription: boolean
  signed: boolean // 已签 = 后端 ConfirmTime 非空
  prescriptionId?: string
}

// 统一电子签名留痕（处方/方案/小结共用）
export interface RestSignRecord {
  id: string
  targetType: string // prescription / plan / summary
  targetId: string
  signerId: string
  signerName?: string
  signTime: string
}

export interface RestTreatment {
  id: number
  tenantId: number
  patientId: string
  treatmentDate: string
  shiftId?: number
  wardId?: number | null
  treatmentType: string
  status: number       // 0-寰呭紑濮?1-杩涜涓?2-宸插畬鎴?3-宸插彇娑?
  legacyStatus?: string  // 老库原始状态码(10/20/30/40/50/60)
  startTime?: string
  endTime?: string
  notes?: string
  doctorSummary?: string
  treatmentSummary?: string
  timeRange?: string
  durationMinutes?: number
  weightLossKg?: number
  shiftName?: string
  queueNo?: string
  caseStatus?: string
  tmrPath?: string
  tmrTime?: string
  tmrPages?: number
  doctorName?: string
  startBp?: string
  endBp?: string
  complications?: string
  beforeSigns?: {
    weight?: number
    extraWeight?: number
    sbp?: number
    dbp?: number
    heartRate?: number
    respiration?: number
    temperature?: number
    pressurePoint?: string
    symptoms?: string
    notes?: string
    operateTime?: string
  }
  afterSigns?: {
    realUfVolume?: number
    realSubstituteVolume?: number
    weight?: number
    extraWeight?: number
    lossWeight?: number
    sbp?: number
    dbp?: number
    heartRate?: number
    respiration?: number
    temperature?: number
    realIntake?: number
    pressurePoint?: string
    complication?: string
    symptoms?: string
    notes?: string
    operateTime?: string
  }
  firstCheck?: {
    id: number
    treatmentId: number
    beforeSignsId?: number
    beforeSymptomId?: number
    operatorId?: number
    operateTime?: string
    materialsResult?: boolean
    materialsMistake?: string
    paramResult?: boolean
    paramMistake?: string
    vascularAccessResult?: boolean
    vascularAccessMistake?: string
    pipelineResult?: boolean
    pipelineMistake?: string
    creatorId: number
    createTime?: string
    lastModifyTime?: string
  }
  secondCheck?: {
    actionId: number
    treatmentId: number
    operatorId?: number
    recheckNurseId?: number
    qcNurseId?: number
    operateTime?: string
    paramResult?: boolean
    paramMistake?: string
    vascularAccessResult?: boolean
    vascularAccessMistake?: string
    pipelineResult?: boolean
    pipelineMistake?: string
    dialysisModeResult?: boolean
    dialysisModeMistake?: string
    prescriptionResult?: boolean
    prescriptionMistake?: string
    anticoagulantResult?: boolean
    anticoagulantMistake?: string
    lineConnectionResult?: boolean
    lineConnectionMistake?: string
    createTime?: string
    lastModifyTime?: string
  }
  beforeSymptomItems?: Array<{ code: string; value: string }>
  afterSymptomItems?: Array<{ code: string; value: string }>
  actions?: Array<{
    id: number
    name: string
    operatorId: number
    operateTime: string
    code?: string
    operator?: string
  }>
  creatorId: number
  createTime: string
  lastModifyTime: string
  patient?: RestPatient
  shift?: { id: number; name: string; startTime: string; endTime: string }
  duringParams?: RestDuringParam[]
}

export interface CreateTreatmentRequest {
  patientId: string
  treatmentDate: string
  type: number
  status?: number
  notes?: string
}

export interface UpdateTreatmentRequest {
  status?: number
  notes?: string
}

export interface UpdateTreatmentSummaryRequest {
  doctorSummary: string
  treatmentSummary: string
}

export interface TreatmentDuringParamRequest {
  recordTime?: string
  code?: string
  sbp?: number
  dbp?: number
  heartRate?: number
  respiration?: number
  spO2?: number
  bloodFlow?: number
  dialysateFlow?: number
  ufVolume?: number
  venousPressure?: number
  arterialPressure?: number
  tmp?: number
  temperature?: number
  conductivity?: number
  ufRate?: number
  notes?: string
}

export interface TreatmentBeforeSignsRequest {
  weight?: number
  extraWeight?: number
  sbp?: number
  dbp?: number
  heartRate?: number
  respiration?: number
  temperature?: number
  pressurePoint?: string
  notes?: string
  symptomItems?: Array<{ code: string; value: string }>
}

export interface TreatmentAfterSignsRequest {
  startTime?: string
  endTime?: string
  realUfVolume?: number
  realSubstituteVolume?: number
  weight?: number
  extraWeight?: number
  lossWeight?: number
  sbp?: number
  dbp?: number
  heartRate?: number
  respiration?: number
  temperature?: number
  realIntake?: number
  pressurePoint?: string
  complication?: string
  symptoms?: string
  notes?: string
  symptomItems?: Array<{ code: string; value: string }>
}

export interface TreatmentFirstCheckRequest {
  beforeSignsId?: number
  beforeSymptomId?: number
  operatorId?: number
  operateTime?: string
  materialsResult?: boolean
  materialsMistake?: string
  paramResult?: boolean
  paramMistake?: string
  vascularAccessResult?: boolean
  vascularAccessMistake?: string
  pipelineResult?: boolean
  pipelineMistake?: string
}

export interface TreatmentSecondCheckRequest {
  operatorId?: number
  recheckNurseId?: number
  qcNurseId?: number
  operateTime?: string
  paramResult?: boolean
  paramMistake?: string
  vascularAccessResult?: boolean
  vascularAccessMistake?: string
  pipelineResult?: boolean
  pipelineMistake?: string
  dialysisModeResult?: boolean
  dialysisModeMistake?: string
  prescriptionResult?: boolean
  prescriptionMistake?: string
  anticoagulantResult?: boolean
  anticoagulantMistake?: string
  lineConnectionResult?: boolean
  lineConnectionMistake?: string
}

export interface TreatmentDisinfectionRequest {
  equipmentId?: number
  disinfectUserId?: number
  disinfectWay?: string
  type?: string
  disinfectant?: string
  startTime?: string
  endTime?: string
  description?: string
  note?: string
}

export interface RestClinicalTask {
  id: number
  type: string
  title: string
  description: string
  patientId?: number
  patientName: string
  bedNumber: string
  severity: string
  status: string
  createdAt: string
}

export interface RestPatientOrder {
  id: string
  type: string
  content: string
  frequency?: string | null
  doctorName?: string
  status: string
  startTime?: string
  createdAt?: string
}

export interface RestQualityStatItem {
  month: number
  ktv: number
  hb: number
  alb: number
}

export interface RestInfectionStatItem {
  month: number
  hbsAg: number
  hcv: number
  hiv: number
  tp: number
}

export interface RestVascularStatItem {
  month: number
  avf: number
  avg: number
  tcc: number
}

export interface RestWorkloadStatItem {
  userId: number
  name: string
  treatments: number
  punctures: number
}

export interface RestPermission {
  id: number
  code: string
  name: string
  description: string
  module: string
  action: string
  status: string
  createdAt: string
  updatedAt: string
}

// ============ /core 接口类型定义 ============

/**
 * 鎮ｈ€?Core 鎺ュ彛鍝嶅簲
 * GET /api/v1/patients/{id}/core
 */
export interface PatientCoreResponse {
  header: PatientCoreHeader
  overview: PatientCoreOverview
  clinicalFocus: PatientCoreClinical
  navigation?: PatientCoreNavigation
}

/**
 * 鎮ｈ€呭ご閮ㄤ俊鎭? */
export interface PatientCoreHeader {
  id: string
  name: string
  avatar?: string
  age: number
  gender: 'M' | 'F'
  bedNumber: string
  status: string  // "娌荤枟涓?, "寰呰瘖", "宸茬粨鏉?
  patientType: string
  insuranceType: string
  doctorName: string
  riskLevel: string  // "高危", "中危", "低危"
  dialysisAge?: string  // 濡?"3骞?涓湀"
}

/**
 * 浼犳煋鐥呮爣蹇? */
export interface PatientCoreInfection {
  hbsag: string  // "闃虫€? | "闃存€?
  hcvab: string  // "闃虫€? | "闃存€?
  hivab: string  // "闃虫€? | "闃存€?
  tpab: string   // "闃虫€? | "闃存€?
  tb?: string    // "闃虫€? | "闃存€? (鍙€?
  updateDate: string
}

/**
 * 当前治疗方案摘要
 */
export interface PatientCoreCurrentPlan {
  dialysisMode: string   // HD/HDF/CRRT
  frequency: string      // "3娆?鍛?
  duration: number       // 时长(小时)
  dryWeight: number      // 骞蹭綋閲?
  bloodFlow: number      // 琛€娴侀噺
  anticoagulant: string  // 鎶楀嚌鍓傛柟妗?
  lastTreatmentNote?: string  // 涓婃娌荤枟鍔ㄦ€?
}

/**
 * 医嘱摘要
 */
export interface PatientCoreOrder {
  id: string
  content: string
  type: string  // "长期", "临时"
  startTime: string
  doctor: string
}

/**
 * 实验室趋势数据点
 */
export interface PatientCoreLabData {
  date: string
  value: number
  isAbnormal: boolean
}

/**
 * 鍏抽敭妫€楠屾寚鏍囪秼鍔? */
export interface PatientCoreLabTrend {
  code: string         // 'HGB' | 'Ca' | 'P'
  name: string
  unit: string
  normalRange: string
  data: PatientCoreLabData[]
}

/**
 * Overview Tab 数据
 */
export interface PatientCoreOverview {
  infection?: PatientCoreInfection    // 鍙€夛紝鍚庣鏃犳暟鎹椂杩斿洖 null/omit
  currentPlan?: PatientCoreCurrentPlan  // 鍙€夛紝鍚庣鏃犳暟鎹椂杩斿洖 null/omit
  activeOrders: PatientCoreOrder[]
  labTrends: PatientCoreLabTrend[]
}

/**
 * 鍗辨€ュ€兼彁閱? */
export interface PatientCoreAlert {
  id: string
  type: string  // 'lab' | 'vital' | 'medication'
  name: string
  value: string
  unit: string
  severity: string  // 'critical' | 'warning' | 'info'
  referenceRange: string
  aiSuggestion?: string
  measuredAt: string
}

/**
 * 鏂囦功鐘舵€? */
export interface PatientCoreDoc {
  id: string
  documentName: string
  status: string  // '寰呯缃? | '宸插畬鎴?
  dueDate?: string
  priority: string  // 'high' | 'medium' | 'low'
}

/**
 * 临床焦点数据
 */
export interface PatientCoreClinical {
  criticalAlerts: PatientCoreAlert[]
  documentStatus: PatientCoreDoc[]
  lastSyncAt: string
}

/**
 * 鎮ｈ€呭鑸俊鎭? */
export interface PatientCoreNavigation {
  previous?: PatientCoreNavPatient
  next?: PatientCoreNavPatient
  total: number
  currentIndex: number
}

/**
 * 鎮ｈ€呭鑸腑鐨勬偅鑰呬俊鎭? */
export interface PatientCoreNavPatient {
  id: string
  name: string
  bedNumber: string
}

// ============ /basic-info 接口类型定义 ============

/**
 * 鎮ｈ€呭熀鏈俊鎭。妗堝搷搴? * GET /api/v1/patients/{id}/basic-info
 */
export interface PatientBasicInfoResponse {
  personalInfo: PatientBasicPersonal
  medicalInfo: PatientBasicMedical
  vitalSocialInfo: PatientBasicVitalSocial
  contactInfo: PatientBasicContact
  // TODO: 后续添加以下字段
  // familyContacts: FamilyContact[]
  // electronicDocuments: ElectronicDocument[]
}

/**
 * 身份核心信息
 */
export interface PatientBasicPersonal {
  name: string
  pinyin?: string
  birthday?: string
  age: number
  gender: string
  ethnicity?: string
  idType: string
  idNumber: string
  patientType: string
}

/**
 * 医疗登记信息
 */
export interface PatientBasicMedical {
  visitCategory?: string
  admissionNo?: string
  visitNo?: string
  medicalRecordNo?: string
  insuranceNo?: string
  hdisPatientId?: number
  insuranceType: string
  dialysisNo?: string
  doctorName: string
  nurseName?: string
  firstDialysisDate?: string
  firstHospitalDate?: string
  firstDialysisHospital?: string
  currentDialysisAge?: string
}

/**
 * 鐢熷懡浣撳緛涓庣ぞ浼氫俊鎭? */
export interface PatientBasicVitalSocial {
  height?: string
  dryWeight: number
  aboBloodType?: string
  rhBloodType?: string
  educationLevel?: string
  occupation?: string
  maritalStatus?: string
  workplace?: string
}

/**
 * 联系信息
 */
export interface PatientBasicContact {
  phone?: string
  wechat?: string
  landline?: string
  address?: string
  district?: string
  contactName?: string
  contactPhone?: string
}

/**
 * 家属与紧急联系人（后续实现）
 */
export interface FamilyContact {
  id: string
  name: string
  phone: string
  type: 'primary' | 'family' | 'emergency'
  relation: string
}

/**
 * 电子文书（后续实现）
 */
export interface ElectronicDocument {
  id: string
  name: string
  type: string
  status: 'signed' | 'pending'
  date: string
}

// ============ /medical-history 接口类型定义 ============

/**
 * 病史内容
 */
export interface HistoryContent {
  content: string
}

/**
 * 带名称的病史内容（专科诊疗记录）
 */
export interface HistoryNamedContent {
  name: string
  content: string
  type?: string        // 鍒嗙被锛堝瓧鍏稿€硷級
  checkTime?: string   // 妫€鏌ユ椂闂?
  checkDoctor?: string // 妫€鏌ュ尰鐢?
}

/**
 * 临床病史响应
 * GET /api/v1/patients/{id}/medical-history
 */
export interface MedicalHistoryApiResponse {
  current: HistoryContent
  past: HistoryContent
  transfusion: HistoryContent
  marital: HistoryContent
  family: HistoryContent
  diagnosis: HistoryContent
  primary: HistoryNamedContent
  pathology: HistoryNamedContent
  allergen: HistoryNamedContent
  tumor: HistoryNamedContent
  complication: HistoryNamedContent
}

/**
 * 转归记录
 */
export interface OutcomeRecordApi {
  id: string
  type: string
  reason: string
  time: string
  remarks: string
  registrar: string
  registrationTime: string
  isDoorRule: boolean  // 是否门规
}

// ============ /vascular-accesses 接口类型定义 ============

/**
 * 琛€绠￠€氳矾 API 鍝嶅簲
 */
export interface VascularAccessApi {
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

/**
 * 鍒涘缓/鏇存柊琛€绠￠€氳矾璇锋眰
 */
export interface VascularAccessCreateRequest {
  accessType: string
  site?: string
  artery?: string[]
  vein?: string[]
  side?: string
  hospital?: string
  surgeon?: string
  surgeryDate?: string
  firstUseDate?: string
  accessNumber?: number
  interventionCount?: number
  interventionDate?: string
  catheterMethod?: string
  catheterDepth?: string
  vPuncturePosition?: string[]
  aPuncturePosition?: string[]
  notes?: string
  images?: string[]
  isDefault?: boolean
  isDisabled?: boolean
}

/**
 * 琛€绠￠€氳矾骞查璁板綍 API 鍝嶅簲
 */
export interface VascularAccessInterventionApi {
  id: string
  vascularAccessId: string
  patientId: string
  accessType: string
  avgBloodFlow: number
  usageDays: number
  surgeryType: string
  interventionReason: string
  doctor: string
  interventionDate: string
  description: string
  createdAt: string
}

/**
 * 鍒涘缓琛€绠￠€氳矾骞查璁板綍璇锋眰
 */
export interface VascularAccessInterventionCreateRequest {
  vascularAccessId: string
  accessType?: string
  avgBloodFlow?: number
  usageDays?: number
  surgeryType: string
  interventionReason: string
  doctor?: string
  interventionDate: string
  description?: string
}

export interface HealthEducationContentApi {
  id: string
  name: string
  description: string
  sort: number
  attachmentIds: string
  type: string
  classify: string
}

export interface PatientHealthEducationApi {
  id: string
  patientId: string
  healthEducationId: string
  healthEducationName: string
  operatorId: string
  operatorName: string
  educationTime: string
  educationType: string
  educationResult: string
  nurseSign: string
  patientSign: string
  finishTime: string | null
  note: string
  createdAt: string
}

export interface CreatePatientHealthEducationRequest {
  healthEducationId: string
  operatorId?: string
  educationTime: string
  educationType?: string
  educationResult?: string
  nurseSign?: string
  patientSign?: string
  finishTime?: string
  note?: string
}

// ============ /lab-reports 接口类型定义 ============

export interface LabReportItemApi {
  id: string
  labReportId: string
  itemCode: string
  itemName: string
  resultValue: string
  unit: string
  referenceRange: string
  abnormalFlag: string
  testedAt: string | null
  createdAt: string
  updatedAt: string
}

export interface LabReportApi {
  id: string
  patientId: string
  reportNo: string
  itemCode: string
  itemName: string
  clinicalDiagnosis: string
  specimenType: string
  urgency: string
  requestDoctor: string
  requestedAt: string | null
  sampledAt: string | null
  receivedAt: string | null
  reportedAt: string | null
  status: string
  externalReportId: string | null
  sourceSystem: string
  syncedAt: string | null
  createdAt: string
  updatedAt: string
  items: LabReportItemApi[]
}

export interface LabReportListParams {
  page?: number
  pageSize?: number
  startDate?: string
  endDate?: string
}

export interface LabReportSyncResult {
  created: number
  updated: number
  skipped: number
  errors: number
}

export interface ExamReportApi {
  id: string
  patientId: string
  examDate: string | null
  title: string
  conclusion: string
  department: string
  externalReportId: string | null
  sourceSystem: string
  syncedAt: string | null
  createdAt: string
  updatedAt: string
}

export interface KeyIndicatorApi {
  id: string
  patientId: string
  externalRecordId: string
  sourceSystem: string
  indexName: string
  indexCode: string
  result: string
  unit: string
  reference: string
  resultSign: string
  testTime: string | null
  evaluationResult: string
  syncedAt: string | null
  createdAt: string
  updatedAt: string
}

export interface HdisIntegrationSettings {
  webcmdUrl: string
  graphqlUrl: string
  authUrl: string
  clientId: string
  serviceUsername: string
  servicePasswordConfigured: boolean
  autoRefreshEnabled: boolean
  refreshLeadSeconds: number
  tokenConfigured: boolean
  tokenExpiresAt: string | null
  tokenStatus: 'MISSING' | 'UNKNOWN' | 'VALID' | 'EXPIRING' | 'EXPIRED'
  lastError: string
}

export interface HdisIntegrationSettingsUpdatePayload {
  webcmdUrl: string
  graphqlUrl: string
  authUrl: string
  clientId: string
  serviceUsername: string
  servicePassword?: string
  autoRefreshEnabled: boolean
  refreshLeadSeconds: number
}

export interface HdisRefreshTokenResult {
  tokenExpiresAt: string | null
  tokenStatus: 'MISSING' | 'UNKNOWN' | 'VALID' | 'EXPIRING' | 'EXPIRED'
}

export type SystemLogSource = 'app' | 'error' | 'all'
export type SystemLogLevel = 'INFO' | 'WARN' | 'ERROR'

export interface SystemLogEntry {
  raw: string
  timestamp?: string
  source: 'app' | 'error'
  level?: SystemLogLevel
}

export interface SystemLogMeta {
  source: SystemLogSource
  lines: number
  keyword?: string
  level?: SystemLogLevel
  redacted: boolean
  levelApplied?: boolean
  levelAppliedOnApp?: boolean
  levelAppliedOnError?: boolean
  fetchedAt: string
}

export interface SystemLogsSingleResponse {
  entries: SystemLogEntry[]
  meta: SystemLogMeta
}

export interface SystemLogsMergedResponse {
  merged: SystemLogEntry[]
  meta: SystemLogMeta
}

export type SystemLogsResponse = SystemLogsSingleResponse | SystemLogsMergedResponse

export interface SystemLogsQuery {
  source?: SystemLogSource
  lines?: number
  keyword?: string
  level?: SystemLogLevel
}

// 创建 axios 实例
const createAxiosInstance = () => {
  const instance = axios.create({
    baseURL: API_CONFIG.baseURL,
    timeout: API_CONFIG.timeout,
    headers: {
      'Content-Type': 'application/json',
    },
  })

  // 璇锋眰鎷︽埅鍣?- 娣诲姞 token
  instance.interceptors.request.use(
    (config) => {
      const token = localStorage.getItem('hdis_access_token')
      if (token) {
        config.headers.Authorization = `Bearer ${token}`
      }
      return config
    },
    (error) => {
      return Promise.reject(error)
    }
  )

  // 鍝嶅簲鎷︽埅鍣?- 缁熶竴閿欒澶勭悊
  instance.interceptors.response.use(
    (response) => {
      return response
    },
    (error) => {
      // 仅在用户已登录且非登录页时，对401做token清理（不自动跳转，交由页面自行处理）
      if (error.response?.status === 401) {
        const isLoginPage = window.location.pathname === '/login'
        const hasToken = !!localStorage.getItem('hdis_access_token')
        if (!isLoginPage && hasToken) {
          localStorage.removeItem('hdis_access_token')
          localStorage.removeItem('hdis_user_info')
          localStorage.removeItem('hdis_token_expiry')
        }
      }
      return Promise.reject(error)
    }
  )

  return instance
}

// 导出 axios 实例
export const apiClient = createAxiosInstance()

// API 服务类
class RestApiService {
  /**
   * 用户登录
   */
  async login(credentials: LoginRequest): Promise<LoginResponse> {
    const response = await apiClient.post<ApiSuccessResponse<LoginResponse>>(
      '/api/v1/auth/login',
      credentials
    )

    if (!response.data.success) {
      throw new Error('登录失败')
    }

    return response.data.data
  }

  /**
   * 获取当前用户信息
   */
  async getCurrentUser() {
    const response = await apiClient.get<ApiSuccessResponse<unknown>>('/api/v1/me')

    if (!response.data.success) {
      throw new Error('获取用户信息失败')
    }

    return response.data.data
  }

  /**
   * 鑾峰彇鎮ｈ€呭垪琛?   */
  async getPatientList(params?: {
    page?: number
    pageSize?: number
    status?: string
    bedNumber?: string
    name?: string
    riskLevel?: string
    onlyActive?: boolean
    onlyTransferred?: boolean
  }): Promise<ApiSuccessResponse<PaginatedResponse<RestPatient>>> {
    const response = await apiClient.get<ApiSuccessResponse<PaginatedResponse<RestPatient>>>(
      '/api/v1/patients',
      { params }
    )

    if (!response.data.success) {
      throw new Error('Failed to get patient list')
    }

    return response.data
  }

  /**
   * 鍒涘缓鎮ｈ€?   */
  async createPatient(data: {
    name: string
    age: number
    gender: string
    bedNumber?: string
    diagnosis?: string
    patientType?: string
    insuranceType?: string
    dryWeight?: number
    defaultMode?: string
    doctorName?: string
    // 基本信息档案
    birthday?: string
    height?: string
    idType?: string
    idNumber?: string
    visitCategory?: string
    visitNo?: string
    insuranceNo?: string
    phone?: string
    address?: string
  }) {
    const response = await apiClient.post<ApiSuccessResponse<RestPatient>>(
      '/api/v1/patients',
      data
    )

    if (!response.data.success) {
      throw new Error('Failed to create patient')
    }

    return response.data.data
  }

  /**
   * 鑾峰彇鎮ｈ€呰鎯?   */
  async getPatient(id: string) {
    const response = await apiClient.get<ApiSuccessResponse<unknown>>(`/api/v1/patients/${id}`)

    if (!response.data.success) {
      throw new Error('Failed to get patient detail')
    }

    return response.data.data
  }

  /**
   * 鑾峰彇鎮ｈ€呮牳蹇冧俊鎭紙棣栧睆鑱氬悎锛?   */
  async getPatientCore(id: string): Promise<PatientCoreResponse> {
    const response = await apiClient.get<ApiSuccessResponse<PatientCoreResponse>>(
      `/api/v1/patients/${id}/core`
    )

    if (!response.data.success) {
      throw new Error('Failed to get patient core')
    }

    return response.data.data
  }

  /**
   * 鑾峰彇鎮ｈ€呭熀鏈俊鎭。妗?   */
  async getPatientBasicInfo(id: string): Promise<PatientBasicInfoResponse> {
    const response = await apiClient.get<ApiSuccessResponse<PatientBasicInfoResponse>>(
      `/api/v1/patients/${id}/basic-info`
    )

    if (!response.data.success) {
      throw new Error('Failed to get patient basic info')
    }

    return response.data.data
  }

  /**
   * 鏇存柊鎮ｈ€呭熀鏈俊鎭。妗?   */
  async updatePatientBasicInfo(id: string, data: unknown) {
    const response = await apiClient.put<ApiSuccessResponse<unknown> | ApiErrorResponse>(
      `/api/v1/patients/${id}/basic-info`,
      data
    )

    if ('success' in response.data && !response.data.success) {
      throw new Error(response.data.error?.message || 'Failed to update patient basic info')
    }

    return (response.data as ApiSuccessResponse<unknown>).data
  }

  /**
   * 鑾峰彇鎮ｈ€呰鎯咃紙甯﹁浆鎹級
   */
  async getPatientById(id: string | number): Promise<Partial<Patient>> {
    const patientData = await this.getPatient(String(id))
    return convertRestPatientToUI(patientData as RestPatient)
  }

  async getPatientOrders(patientId: string, params?: { type?: 'LONG' | 'TEMP'; status?: string }): Promise<ApiSuccessResponse<RestPatientOrder[]>> {
    const mappedParams = {
      type: params?.type === 'LONG' ? '长期' : params?.type === 'TEMP' ? '临时' : undefined,
      statuses: params?.status,
    }
    const response = await apiClient.get<ApiSuccessResponse<RestPatientOrder[]>>(
      `/api/v1/patients/${patientId}/orders`,
      { params: mappedParams }
    )

    if (!response.data.success) {
      throw new Error('获取患者医嘱失败')
    }

    return response.data
  }

  /**
   * 鍒犻櫎鎮ｈ€?   */
  async deletePatient(id: string) {
    const response = await apiClient.delete<ApiSuccessResponse<unknown>>(
      `/api/v1/patients/${id}`
    )

    // 204 No Content: no response body, still treated as success.
    if (response.status === 204) {
      return
    }

    if (!response.data?.success) {
      throw new Error('Failed to delete patient')
    }

    return response.data.data
  }

  // ============ 临床病史 API ============

  /**
   * 鑾峰彇鎮ｈ€呬复搴婄梾鍙?   */
  async getMedicalHistory(patientId: string): Promise<MedicalHistoryApiResponse> {
    const response = await apiClient.get<ApiSuccessResponse<MedicalHistoryApiResponse>>(
      `/api/v1/patients/${patientId}/medical-history`
    )

    if (!response.data.success) {
      throw new Error('获取临床病史失败')
    }

    return response.data.data
  }

  /**
   * 淇濆瓨鎮ｈ€呬复搴婄梾鍙?   */
  async saveMedicalHistory(patientId: string, data: Partial<MedicalHistoryApiResponse>): Promise<MedicalHistoryApiResponse> {
    const response = await apiClient.put<ApiSuccessResponse<MedicalHistoryApiResponse>>(
      `/api/v1/patients/${patientId}/medical-history`,
      data
    )

    if (!response.data.success) {
      throw new Error('保存临床病史失败')
    }

    return response.data.data
  }

  /**
   * 获取转归记录列表
   */
  async getOutcomeRecords(patientId: string): Promise<OutcomeRecordApi[]> {
    const response = await apiClient.get<ApiSuccessResponse<OutcomeRecordApi[]>>(
      `/api/v1/patients/${patientId}/outcome-records`
    )

    if (!response.data.success) {
      throw new Error('获取转归记录失败')
    }

    return response.data.data
  }

  /**
   * 创建转归记录
   */
  async createOutcomeRecord(patientId: string, data: Omit<OutcomeRecordApi, 'id'>): Promise<OutcomeRecordApi> {
    const response = await apiClient.post<ApiSuccessResponse<OutcomeRecordApi>>(
      `/api/v1/patients/${patientId}/outcome-records`,
      data
    )

    if (!response.data.success) {
      throw new Error('创建转归记录失败')
    }

    return response.data.data
  }

  /**
   * 更新转归记录
   */
  async updateOutcomeRecord(patientId: string, recordId: string, data: Omit<OutcomeRecordApi, 'id'>): Promise<OutcomeRecordApi> {
    const response = await apiClient.put<ApiSuccessResponse<OutcomeRecordApi>>(
      `/api/v1/patients/${patientId}/outcome-records/${recordId}`,
      data
    )

    if (!response.data.success) {
      throw new Error('更新转归记录失败')
    }

    return response.data.data
  }

  /**
   * 删除转归记录
   */
  async deleteOutcomeRecord(patientId: string, recordId: string): Promise<void> {
    const response = await apiClient.delete<ApiSuccessResponse<unknown>>(
      `/api/v1/patients/${patientId}/outcome-records/${recordId}`
    )

    if (response.status !== 204 && !response.data?.success) {
      throw new Error('删除转归记录失败')
    }
  }

  // ============ 妫€楠屾姤鍛?API ============

  /**
   * 鑾峰彇鎮ｈ€呮楠屾姤鍛婂垪琛?   */
  async getLabReports(patientId: string, params?: LabReportListParams): Promise<PaginatedResponse<LabReportApi>> {
    const response = await apiClient.get<ApiSuccessResponse<PaginatedResponse<LabReportApi>>>(
      `/api/v1/patients/${patientId}/lab-reports`,
      { params }
    )

    if (!response.data.success) {
      throw new Error('Failed to get lab reports')
    }

    return response.data.data
  }

  /**
   * 瑙﹀彂鎮ｈ€呮楠屾姤鍛婂悓姝?   */
  async syncLabReports(patientId: string): Promise<LabReportSyncResult> {
    const response = await apiClient.post<ApiSuccessResponse<LabReportSyncResult>>(
      `/api/v1/patients/${patientId}/lab-reports/sync`
    )

    if (!response.data.success) {
      throw new Error('Failed to sync lab reports')
    }

    return response.data.data
  }

  /**
   * 瑙﹀彂鎮ｈ€呮鏌ユ姤鍛婂悓姝?   */
  async syncExamReports(patientId: string): Promise<LabReportSyncResult> {
    const response = await apiClient.post<ApiSuccessResponse<LabReportSyncResult>>(
      `/api/v1/patients/${patientId}/exam-reports/sync`
    )

    if (!response.data.success) {
      throw new Error('Failed to sync exam reports')
    }

    return response.data.data
  }

  /**
   * 鑾峰彇鎮ｈ€呮鏌ユ姤鍛婂垪琛?   */
  async getExamReports(patientId: string, params?: LabReportListParams): Promise<PaginatedResponse<ExamReportApi>> {
    const response = await apiClient.get<ApiSuccessResponse<PaginatedResponse<ExamReportApi>>>(
      `/api/v1/patients/${patientId}/exam-reports`,
      { params }
    )

    if (!response.data.success) {
      throw new Error('Failed to get exam reports')
    }

    return response.data.data
  }

  /**
   * 瑙﹀彂鎮ｈ€呭叧閿寚鏍囧悓姝?   */
  async syncKeyIndicators(patientId: string): Promise<LabReportSyncResult> {
    const response = await apiClient.post<ApiSuccessResponse<LabReportSyncResult>>(
      `/api/v1/patients/${patientId}/key-indicators/sync`
    )

    if (!response.data.success) {
      throw new Error('同步关键指标失败')
    }

    return response.data.data
  }

  /**
   * 鑾峰彇鎮ｈ€呭叧閿寚鏍囧垪琛?   */
  async getKeyIndicators(patientId: string, params?: LabReportListParams): Promise<PaginatedResponse<KeyIndicatorApi>> {
    const response = await apiClient.get<ApiSuccessResponse<PaginatedResponse<KeyIndicatorApi>>>(
      `/api/v1/patients/${patientId}/key-indicators`,
      { params }
    )

    if (!response.data.success) {
      throw new Error('获取关键指标失败')
    }

    return response.data.data
  }

  // ============ Settings / HDIS Integration API ============

  async getHdisIntegrationSettings(): Promise<HdisIntegrationSettings> {
    const response = await apiClient.get<ApiSuccessResponse<HdisIntegrationSettings>>(
      '/api/v1/settings/integrations/hdis'
    )

    if (!response.data.success) {
      throw new Error('获取 HDIS 配置失败')
    }

    return response.data.data
  }

  async updateHdisIntegrationSettings(payload: HdisIntegrationSettingsUpdatePayload): Promise<HdisIntegrationSettings> {
    const response = await apiClient.put<ApiSuccessResponse<HdisIntegrationSettings>>(
      '/api/v1/settings/integrations/hdis',
      payload
    )

    if (!response.data.success) {
      throw new Error('保存 HDIS 配置失败')
    }

    return response.data.data
  }

  async refreshHdisToken(): Promise<HdisRefreshTokenResult> {
    const response = await apiClient.post<ApiSuccessResponse<HdisRefreshTokenResult>>(
      '/api/v1/settings/integrations/hdis/refresh-token',
      null,
      { timeout: 120_000 }
    )

    if (!response.data.success) {
      throw new Error('刷新 HDIS Token 失败')
    }

    return response.data.data
  }

  async getSystemLogs(params?: SystemLogsQuery): Promise<SystemLogsResponse> {
    const response = await apiClient.get<ApiSuccessResponse<SystemLogsResponse>>(
      '/api/v1/settings/logs',
      { params }
    )

    if (!response.data.success) {
      throw new Error('获取系统日志失败')
    }

    return response.data.data
  }

  // ============ 琛€绠￠€氳矾 API ============

  /**
   * 鑾峰彇鎮ｈ€呰绠￠€氳矾鍒楄〃
   */
  async getVascularAccesses(patientId: string): Promise<VascularAccessApi[]> {
    const response = await apiClient.get<ApiSuccessResponse<VascularAccessApi[]>>(
      `/api/v1/patients/${patientId}/vascular-accesses`
    )

    if (!response.data.success) {
      throw new Error('鑾峰彇琛€绠￠€氳矾鍒楄〃澶辫触')
    }

    return response.data.data
  }

  /**
   * 鍒涘缓琛€绠￠€氳矾
   */
  async createVascularAccess(patientId: string, data: VascularAccessCreateRequest): Promise<VascularAccessApi> {
    const response = await apiClient.post<ApiSuccessResponse<VascularAccessApi>>(
      `/api/v1/patients/${patientId}/vascular-accesses`,
      data
    )

    if (!response.data.success) {
      throw new Error('鍒涘缓琛€绠￠€氳矾澶辫触')
    }

    return response.data.data
  }

  /**
   * 鏇存柊琛€绠￠€氳矾
   */
  async updateVascularAccess(patientId: string, accessId: string, data: VascularAccessCreateRequest): Promise<VascularAccessApi> {
    const response = await apiClient.put<ApiSuccessResponse<VascularAccessApi>>(
      `/api/v1/patients/${patientId}/vascular-accesses/${accessId}`,
      data
    )

    if (!response.data.success) {
      throw new Error('鏇存柊琛€绠￠€氳矾澶辫触')
    }

    return response.data.data
  }

  /**
   * 鍒犻櫎琛€绠￠€氳矾
   */
  async deleteVascularAccess(patientId: string, accessId: string): Promise<void> {
    const response = await apiClient.delete<ApiSuccessResponse<unknown>>(
      `/api/v1/patients/${patientId}/vascular-accesses/${accessId}`
    )

    if (response.status !== 204 && !response.data?.success) {
      throw new Error('鍒犻櫎琛€绠￠€氳矾澶辫触')
    }
  }

  // ============ 琛€绠￠€氳矾骞查璁板綍鎺ュ彛 ============

  /**
   * 鑾峰彇鎮ｈ€呯殑琛€绠￠€氳矾骞查璁板綍鍒楄〃
   */
  async getVascularAccessInterventions(patientId: string, vascularAccessId?: string): Promise<VascularAccessInterventionApi[]> {
    const params = vascularAccessId ? { vascularAccessId } : {}
    const response = await apiClient.get<ApiSuccessResponse<VascularAccessInterventionApi[]>>(
      `/api/v1/patients/${patientId}/vascular-access-interventions`,
      { params }
    )

    if (!response.data?.success) {
      throw new Error('获取干预记录失败')
    }
    return response.data.data
  }

  /**
   * 鍒涘缓琛€绠￠€氳矾骞查璁板綍
   */
  async createVascularAccessIntervention(patientId: string, data: VascularAccessInterventionCreateRequest): Promise<VascularAccessInterventionApi> {
    const response = await apiClient.post<ApiSuccessResponse<VascularAccessInterventionApi>>(
      `/api/v1/patients/${patientId}/vascular-access-interventions`,
      data
    )

    if (!response.data?.success) {
      throw new Error('创建干预记录失败')
    }
    return response.data.data
  }

  /**
   * 鍒犻櫎琛€绠￠€氳矾骞查璁板綍
   */
  async deleteVascularAccessIntervention(patientId: string, interventionId: string): Promise<void> {
    const response = await apiClient.delete<ApiSuccessResponse<unknown>>(
      `/api/v1/patients/${patientId}/vascular-access-interventions/${interventionId}`
    )

    if (response.status !== 204 && !response.data?.success) {
      throw new Error('删除干预记录失败')
    }
  }

  // ============ 健康宣教 ============

  async getHealthEducationContents(): Promise<HealthEducationContentApi[]> {
    const response = await apiClient.get<ApiSuccessResponse<HealthEducationContentApi[]>>('/api/v1/health-educations')
    if (!response.data.success) {
      throw new Error('获取健康宣教内容失败')
    }
    return response.data.data
  }

  async getPatientHealthEducations(patientId: string): Promise<PatientHealthEducationApi[]> {
    const response = await apiClient.get<ApiSuccessResponse<PatientHealthEducationApi[]>>(
      `/api/v1/patients/${patientId}/health-educations`
    )
    if (!response.data.success) {
      throw new Error('获取患者健康宣教记录失败')
    }
    return response.data.data
  }

  async createPatientHealthEducation(
    patientId: string,
    data: CreatePatientHealthEducationRequest
  ): Promise<PatientHealthEducationApi> {
    const response = await apiClient.post<ApiSuccessResponse<PatientHealthEducationApi>>(
      `/api/v1/patients/${patientId}/health-educations`,
      data
    )
    if (!response.data.success) {
      throw new Error('保存患者健康宣教记录失败')
    }
    return response.data.data
  }

  // ============ 用户管理 ============

  // ============ 用户管理（代理调用 userApi）============

  async getUserList(params?: UserListParams): Promise<{ items: RestUser[]; total: number }> {
    return userApi.getList(params)
  }

  async getUserById(id: string): Promise<unknown> {
    return userApi.getById(id)
  }

  async createUser(data: CreateUserRequest): Promise<unknown> {
    return userApi.create(data)
  }

  async updateUser(id: string, data: UpdateUserRequest): Promise<unknown> {
    return userApi.update(id, data)
  }

  async updateUserStatus(id: string, status: string): Promise<void> {
    return userApi.updateStatus(id, status)
  }

  async deleteUser(id: string): Promise<void> {
    return userApi.remove(id)
  }

  async resetPassword(id: string, newPassword: string): Promise<void> {
    return userApi.resetPassword(id, newPassword)
  }

  async getUserRoles(id: string): Promise<string[]> {
    return userApi.getRoles(id)
  }

  async setUserRoles(id: string, roleCodes: string[]): Promise<void> {
    return userApi.setRoles(id, roleCodes)
  }

  async getMyRoles(): Promise<ApiSuccessResponse<{ userId: string; username: string; realName: string; roles: string[] }>> {
    const data = await userApi.getMyRoles()
    return { success: true, data, timestamp: new Date().toISOString() }
  }

  async getRolePermissions(role: string): Promise<ApiSuccessResponse<{ role: string; permissionCodes: string[] }>> {
    const data = await roleManagementApi.getRolePermissions(role)
    return { success: true, data, timestamp: new Date().toISOString() }
  }

  async setRolePermissions(role: string, permissionCodes: string[]): Promise<ApiSuccessResponse<{ role: string; permissionCodes: string[] }>> {
    const data = await roleManagementApi.setRolePermissions(role, permissionCodes)
    return { success: true, data, timestamp: new Date().toISOString() }
  }
  async getRoleList(): Promise<AppRoleApi[]> {
    return roleManagementApi.getRoleList()
  }

  async createRole(data: Partial<AppRoleApi>): Promise<AppRoleApi> {
    return roleManagementApi.createRole(data)
  }

  async updateRole(code: string, data: Partial<AppRoleApi>): Promise<AppRoleApi> {
    return roleManagementApi.updateRole(code, data)
  }

  async deleteRole(code: string): Promise<void> {
    return roleManagementApi.deleteRole(code)
  }

  async getPermissionTree(): Promise<PermissionNodeApi[]> {
    return roleManagementApi.getPermissionTree()
  }


  async getPermissions(): Promise<ApiSuccessResponse<{ items: RestPermission[]; total: number }>> {
    const response = await apiClient.get<ApiSuccessResponse<{ items: RestPermission[]; total: number }>>('/api/v1/permissions')
    if (!response.data.success) {
      throw new Error('获取权限列表失败')
    }
    return response.data
  }



  // ============ 看板统计 ============

  /**
   * 鑾峰彇鐪嬫澘缁熻姹囨€?   */
  async getDashboardStats(): Promise<{
    activePatients: number
    shiftCount: number
    equipmentCount: number
    todaySchedules: number
    todayTreatments: number
    runningTreatments?: number
    completedTreatments?: number
    alertItems: number
    treatmentsByHour: { name: string; value: number }[]
    qualityByHour: { name: string; value: number }[]
  }> {
    const response = await apiClient.get<ApiSuccessResponse<{
      activePatients: number
      shiftCount: number
      equipmentCount: number
      todaySchedules: number
      todayTreatments: number
      runningTreatments?: number
      completedTreatments?: number
      alertItems: number
      treatmentsByHour: { name: string; value: number }[]
      qualityByHour: { name: string; value: number }[]
    }>>('/api/v1/dashboard/stats')
    if (!response.data.success) {
      throw new Error('获取看板统计失败')
    }
    return response.data.data
  }

  // ============ 设备管理 ============

  /**
   * 获取设备列表
   */
  async getDeviceList(params?: {
    page?: number
    pageSize?: number
    status?: string
    bedNumber?: string
    keyword?: string
  }): Promise<RestDevice[]> {
    const response = await apiClient.get<ApiSuccessResponse<{ items: RestDevice[]; total: number }>>('/api/v1/devices', { params })
    if (!response.data.success) {
      throw new Error('获取设备列表失败')
    }
    return response.data.data.items
  }

  // ============ 库存管理 ============

  /**
   * 获取库存品目列表
   */
  async getInventoryItems(params?: {
    page?: number
    pageSize?: number
    category?: string
    keyword?: string
  }): Promise<{ items: RestInventoryItem[]; total: number; page: number; pageSize: number; totalPage: number }> {
    const response = await apiClient.get<ApiSuccessResponse<{ items: RestInventoryItem[]; total: number; page: number; pageSize: number; totalPage: number }>>('/api/v1/inventory/items', { params })
    if (!response.data.success) {
      throw new Error('获取库存列表失败')
    }
    return response.data.data
  }

  /**
   * 鑾峰彇鍑哄叆搴撹褰?   */
  async getStockLogs(params?: {
    page?: number
    pageSize?: number
    itemId?: string
    type?: string
  }): Promise<{ items: RestStockLog[]; total: number; page: number; pageSize: number; totalPage: number }> {
    const response = await apiClient.get<ApiSuccessResponse<{ items: RestStockLog[]; total: number; page: number; pageSize: number; totalPage: number }>>('/api/v1/inventory/logs', { params })
    if (!response.data.success) {
      throw new Error('Failed to get stock logs')
    }
    return response.data.data
  }

  async adjustStock(data: {
    itemId: string
    type: 'in' | 'out'
    quantity: number
    operator?: string
    note?: string
  }): Promise<RestStockLog> {
    const response = await apiClient.post<ApiSuccessResponse<RestStockLog>>('/api/v1/inventory/adjust', data)
    if (!response.data.success) {
      throw new Error('Failed to adjust stock')
    }
    return response.data.data
  }

  /**
   * 获取标签打印任务
   */
  async getLabelTasks(params?: {
    page?: number
    pageSize?: number
    status?: string
  }): Promise<{ items: RestLabelTask[]; total: number; page: number; pageSize: number; totalPage: number }> {
    const response = await apiClient.get<ApiSuccessResponse<{ items: RestLabelTask[]; total: number; page: number; pageSize: number; totalPage: number }>>('/api/v1/inventory/labels', { params })
    if (!response.data.success) {
      throw new Error('获取标签任务失败')
    }
    return response.data.data
  }

  async createLabelTask(data: { itemId: string; quantity: number }): Promise<RestLabelTask> {
    const response = await apiClient.post<ApiSuccessResponse<RestLabelTask>>('/api/v1/inventory/labels', data)
    if (!response.data.success) {
      throw new Error('Failed to create label task')
    }
return response.data.data
  }

  // ============ 治疗记录管理 ============

  /**
   * 获取治疗记录列表
   */
  async getTreatments(params?: {
    page?: number
    pageSize?: number
    patientId?: string
    status?: number
    treatmentDate?: string
    treatmentDateStart?: string
    treatmentDateEnd?: string
  }): Promise<ApiSuccessResponse<PaginatedResponse<RestTreatment>>> {
    const response = await apiClient.get<ApiSuccessResponse<PaginatedResponse<RestTreatment>>>(
      '/api/v1/treatments',
      { params }
    )
    if (!response.data.success) {
      throw new Error('获取治疗记录列表失败')
    }
    return response.data
  }

  async getPatientStats(): Promise<{ totalCount: number; activeCount: number; outpatientCount: number; inpatientCount: number }> {
    const response = await apiClient.get<ApiSuccessResponse<{ totalCount: number; activeCount: number; outpatientCount: number; inpatientCount: number }>>('/api/v1/patients/stats')
    if (!response.data.success) {
      throw new Error('Failed to get patient stats')
    }
    return response.data.data
  }

  // 驾驶舱医生墙：当日每患者 是否开方/是否已签（批量，签发=ConfirmTime 非空）
  async getPrescriptionDayStatus(date: string): Promise<RestPrescriptionDayStatus[]> {
    const response = await apiClient.get<ApiSuccessResponse<RestPrescriptionDayStatus[]>>(
      '/api/v1/prescriptions/day-status',
      { params: { date } }
    )
    if (!response.data.success) {
      throw new Error('Failed to get prescription day status')
    }
    return response.data.data || []
  }

  // 签发处方（待签→已签，写统一签名留痕；不改执行态）
  async signPrescription(patientId: string, prescriptionId: string): Promise<void> {
    const response = await apiClient.post<ApiSuccessResponse<unknown>>(
      `/api/v1/patients/${patientId}/prescriptions/${prescriptionId}/sign`
    )
    if (!response.data.success) {
      throw new Error('Failed to sign prescription')
    }
  }

  /** 通用签发（方案/小结）：写一条统一签名留痕 */
  async signTarget(targetType: 'plan' | 'summary', targetId: string): Promise<void> {
    const response = await apiClient.post<ApiSuccessResponse<unknown>>('/api/v1/sign-records', { targetType, targetId })
    if (!response.data.success) {
      throw new Error('Failed to sign target')
    }
  }

  /** 查询某对象的签名留痕（审计/展示） */
  async getSignRecords(targetType: string, targetId: string): Promise<RestSignRecord[]> {
    const response = await apiClient.get<ApiSuccessResponse<RestSignRecord[]>>('/api/v1/sign-records', {
      params: { targetType, targetId },
    })
    if (!response.data.success) {
      throw new Error('Failed to get sign records')
    }
    return response.data.data || []
  }

  /**
   * 获取排班列表
   */
  async getTreatment(id: number): Promise<ApiSuccessResponse<RestTreatment>> {
    const response = await apiClient.get<ApiSuccessResponse<RestTreatment>>(`/api/v1/treatments/${id}`)
    if (!response.data.success) {
      throw new Error('获取治疗记录详情失败')
    }
    return response.data
  }

  async createTreatment(data: CreateTreatmentRequest): Promise<ApiSuccessResponse<RestTreatment>> {
    const response = await apiClient.post<ApiSuccessResponse<RestTreatment>>('/api/v1/treatments', data)
    if (!response.data.success) {
      throw new Error('创建治疗记录失败')
    }
    return response.data
  }

  async updateTreatment(id: number, data: UpdateTreatmentRequest): Promise<ApiSuccessResponse<RestTreatment>> {
    const response = await apiClient.put<ApiSuccessResponse<RestTreatment>>(`/api/v1/treatments/${id}`, data)
    if (!response.data.success) {
      throw new Error('更新治疗记录失败')
    }
    return response.data
  }

  async updateTreatmentSummary(id: number, data: UpdateTreatmentSummaryRequest): Promise<ApiSuccessResponse<RestTreatment>> {
    const response = await apiClient.put<ApiSuccessResponse<RestTreatment>>(`/api/v1/treatments/${id}/summary`, data)
    if (!response.data.success) {
      throw new Error('保存透析小结失败')
    }
    return response.data
  }

  async updateTreatmentStatus(id: number, status: number): Promise<ApiSuccessResponse<unknown>> {
    const response = await apiClient.put<ApiSuccessResponse<unknown>>(`/api/v1/treatments/${id}/status`, { status })
    if (!response.data.success) {
      throw new Error('更新治疗状态失败')
    }
    return response.data
  }

  /**
   * 鑾峰彇鎮ｈ€呮寚瀹氭棩鏈熺殑娌荤枟璁板綍锛堝惈 duringParams锛?   */
  async getPatientTreatmentByDate(patientId: string | number, date: string): Promise<ApiSuccessResponse<RestTreatment | null>> {
    const response = await apiClient.get<ApiSuccessResponse<RestTreatment | null>>(
      `/api/v1/patients/${patientId}/treatment`,
      { params: { date } }
    )
    if (!response.data.success) {
      throw new Error('获取治疗记录失败')
    }
    return response.data
  }

  async createTreatmentDuringParam(treatmentId: number, data: TreatmentDuringParamRequest): Promise<ApiSuccessResponse<RestDuringParam>> {
    const response = await apiClient.post<ApiSuccessResponse<RestDuringParam>>(
      `/api/v1/treatments/${treatmentId}/during-params`,
      data
    )
    if (!response.data.success) {
      throw new Error('创建透中监测记录失败')
    }
    return response.data
  }

  async updateTreatmentDuringParam(treatmentId: number, paramId: number, data: TreatmentDuringParamRequest): Promise<ApiSuccessResponse<RestDuringParam>> {
    const response = await apiClient.put<ApiSuccessResponse<RestDuringParam>>(
      `/api/v1/treatments/${treatmentId}/during-params/${paramId}`,
      data
    )
    if (!response.data.success) {
      throw new Error('更新透中监测记录失败')
    }
    return response.data
  }

  async deleteTreatmentDuringParam(treatmentId: number, paramId: number): Promise<ApiSuccessResponse<unknown>> {
    const response = await apiClient.delete<ApiSuccessResponse<unknown>>(
      `/api/v1/treatments/${treatmentId}/during-params/${paramId}`
    )
    if (!response.data.success) {
      throw new Error('删除透中监测记录失败')
    }
    return response.data
  }

  async saveTreatmentBeforeSigns(treatmentId: number, data: TreatmentBeforeSignsRequest): Promise<ApiSuccessResponse<unknown>> {
    const response = await apiClient.put<ApiSuccessResponse<unknown>>(`/api/v1/treatments/${treatmentId}/before-signs`, data)
    if (!response.data.success) {
      throw new Error('保存透前评估失败')
    }
    return response.data
  }

  async saveTreatmentAfterSigns(treatmentId: number, data: TreatmentAfterSignsRequest): Promise<ApiSuccessResponse<unknown>> {
    const response = await apiClient.put<ApiSuccessResponse<unknown>>(`/api/v1/treatments/${treatmentId}/after-signs`, data)
    if (!response.data.success) {
      throw new Error('保存透后评估失败')
    }
    return response.data
  }

  async submitTreatmentPostAssessment(treatmentId: number, data: TreatmentAfterSignsRequest): Promise<ApiSuccessResponse<RestTreatment>> {
    const response = await apiClient.put<ApiSuccessResponse<RestTreatment>>(`/api/v1/treatments/${treatmentId}/post-assessment-submit`, data)
    if (!response.data.success) {
      throw new Error('提交透后评估失败')
    }
    return response.data
  }

  async saveTreatmentFirstCheck(treatmentId: number, data: TreatmentFirstCheckRequest): Promise<ApiSuccessResponse<unknown>> {
    const response = await apiClient.put<ApiSuccessResponse<unknown>>(`/api/v1/treatments/${treatmentId}/first-check`, data)
    if (!response.data.success) {
      throw new Error('保存首次核对失败')
    }
    return response.data
  }

  async saveTreatmentSecondCheck(treatmentId: number, data: TreatmentSecondCheckRequest): Promise<ApiSuccessResponse<unknown>> {
    const response = await apiClient.put<ApiSuccessResponse<unknown>>(`/api/v1/treatments/${treatmentId}/second-check`, data)
    if (!response.data.success) {
      throw new Error('保存二次核对失败')
    }
    return response.data
  }

  async saveTreatmentDisinfection(treatmentId: number, data: TreatmentDisinfectionRequest): Promise<ApiSuccessResponse<unknown>> {
    const response = await apiClient.put<ApiSuccessResponse<unknown>>(`/api/v1/treatments/${treatmentId}/disinfection`, data)
    if (!response.data.success) {
      throw new Error('保存消毒登记失败')
    }
    return response.data
  }

  async getClinicalTasks(params?: { status?: string }): Promise<ApiSuccessResponse<{ items: RestClinicalTask[]; total: number }>> {
    const response = await apiClient.get<ApiSuccessResponse<{ items: RestClinicalTask[]; total: number }>>('/api/v1/clinical-tasks', { params })
    if (!response.data.success) {
      throw new Error('获取临床任务失败')
    }
    return response.data
  }

  async updateClinicalTaskStatus(id: number, status: string): Promise<ApiSuccessResponse<Record<string, never>>> {
    const response = await apiClient.put<ApiSuccessResponse<Record<string, never>>>(`/api/v1/clinical-tasks/${id}/status`, { status })
    if (!response.data.success) {
      throw new Error('更新临床任务状态失败')
    }
    return response.data
  }

  async getQualityStatistics(params: { year: number }): Promise<ApiSuccessResponse<{ items: RestQualityStatItem[] }>> {
    const response = await apiClient.get<ApiSuccessResponse<{ items: RestQualityStatItem[] }>>('/api/v1/statistics/quality', { params })
    if (!response.data.success) {
      throw new Error('获取质量统计失败')
    }
    return response.data
  }

  async getInfectionStatistics(params: { year: number }): Promise<ApiSuccessResponse<{ items: RestInfectionStatItem[] }>> {
    const response = await apiClient.get<ApiSuccessResponse<{ items: RestInfectionStatItem[] }>>('/api/v1/statistics/infection', { params })
    if (!response.data.success) {
      throw new Error('获取感染统计失败')
    }
    return response.data
  }

  async getVascularStatistics(params: { year: number }): Promise<ApiSuccessResponse<{ items: RestVascularStatItem[] }>> {
    const response = await apiClient.get<ApiSuccessResponse<{ items: RestVascularStatItem[] }>>('/api/v1/statistics/vascular', { params })
    if (!response.data.success) {
      throw new Error('获取通路统计失败')
    }
    return response.data
  }

  async getWorkloadStatistics(params: { yearMonth: string }): Promise<ApiSuccessResponse<{ items: RestWorkloadStatItem[] }>> {
    const response = await apiClient.get<ApiSuccessResponse<{ items: RestWorkloadStatItem[] }>>('/api/v1/statistics/workload', { params })
    if (!response.data.success) {
      throw new Error('获取工作量统计失败')
    }
    return response.data
  }
}

// 导出单例
export const restApi = new RestApiService()

export type RequestErrorKind = 'auth' | 'forbidden' | 'not_found' | 'server' | 'network' | 'unknown'

export function getRequestErrorKind(error: unknown): RequestErrorKind {
  if (axios.isAxiosError(error)) {
    const status = error.response?.status
    if (status === 401) return 'auth'
    if (status === 403) return 'forbidden'
    if (status === 404) return 'not_found'
    if (typeof status === 'number' && status >= 500) return 'server'
    if (!error.response) return 'network'
  }

  if (error instanceof TypeError) {
    return 'network'
  }

  return 'unknown'
}

export function getTreatmentLoadErrorMessage(error: unknown): string {
  switch (getRequestErrorKind(error)) {
    case 'not_found':
      return '暂无治疗记录'
    case 'auth':
      return '登录已失效，请重新登录'
    case 'forbidden':
      return '无权限访问，请重新登录'
    case 'server':
      return '治疗记录加载失败，请重试'
    case 'network':
      return '网络异常，请检查连接'
    default:
      return getErrorMessage(error)
  }
}

// 错误处理辅助函数
export function getErrorMessage(error: unknown): string {
  const data = axios.isAxiosError(error) ? (error.response?.data as ApiErrorResponse | undefined) : undefined
  const apiMessage = data?.error?.message?.trim() || ''
  const status = axios.isAxiosError(error) ? error.response?.status : undefined
  const kind = getRequestErrorKind(error)

  if (status === 401) {
    return '登录已失效，请重新登录'
  }
  if (status === 403) {
    return '无权限访问，请联系管理员'
  }
  if (status === 404) {
    return apiMessage || '资源不存在，请刷新后重试'
  }
  if (status === 400) {
    return apiMessage || '请求参数有误，请检查后重试'
  }
  if (status === 409) {
    return apiMessage || '数据已发生冲突，请刷新后重试'
  }
  if (typeof status === 'number' && status >= 500) {
    return apiMessage || '服务器异常，请稍后重试'
  }

  if (kind === 'network') {
    return '网络异常，请检查连接'
  }

  if (apiMessage) {
    return apiMessage
  }

  if (axios.isAxiosError(error)) {
    return error.message || '请求失败，请重试'
  }
  if (error instanceof Error) {
    return error.message
  }
  return '未知错误'
}

// ============ Data Converters ============

import type {
  Patient,
  LabResult,
  TreatmentPlan,
  MedicalOrder,
  InfectionInfo,
} from '@/types/original'

export function convertRestPatientToUI(restPatient: RestPatient): Partial<Patient> {
  const statusMap: Record<string, string> = {
    active: '透析中',
    inactive: '候诊',
    discharged: '已结束',
  }
  const status = statusMap[restPatient.status] || '居家'
  // 兼容后端返回 "M"/"F" 或 "男"/"女" 两种格式
  const gender: '男' | '女' = (restPatient.gender === 'M' || restPatient.gender === '男') ? '男' : '女'

  return {
    id: restPatient.id,
    name: restPatient.name,
    gender,
    age: restPatient.age || 0,
    bedNumber: restPatient.bedNumber || '',
    status,
    patientType: restPatient.patientType || '门诊',
    insuranceType: restPatient.insuranceType || '自费',
    dryWeight: restPatient.dryWeight ?? 65,
    defaultMode: restPatient.defaultMode || 'HD',
    doctorName: restPatient.doctorName || '',
    avatar: restPatient.imageBase64String || undefined,
  }
}

export function convertRestPatientList(patients: unknown[]): Partial<Patient>[] {
  return patients.map((p) => convertRestPatientToUI(p as RestPatient))
}

export function convertCoreResponseToPatient(coreData: PatientCoreResponse): Partial<Patient> {
  const { header, overview, clinicalFocus } = coreData
  // 兼容后端返回 "M"/"F" 或 "男"/"女" 两种格式
  const rawGender = String(header.gender || '')
  const gender: '男' | '女' = (rawGender === 'M' || rawGender === '男') ? '男' : '女'

  const recentLabs: LabResult[] = (overview.labTrends || [])
    .filter(trend => trend.data && trend.data.length > 0)
    .map(trend => {
      const latest = trend.data[trend.data.length - 1]
      return {
        id: `${trend.code}_${latest.date}`,
        name: trend.name,
        value: String(latest.value),
        unit: trend.unit,
        date: latest.date,
        isAbnormal: latest.isAbnormal,
        reference: trend.normalRange,
      }
    })

  const orders: MedicalOrder[] = (overview.activeOrders || []).map(order => ({
    id: order.id,
    content: order.content,
    type: (order.type === '长期' || order.type === '临时') ? order.type : '长期',
    status: '待执行',
    doctor: order.doctor,
    startTime: order.startTime,
  }))

  const infection: InfectionInfo | undefined = overview.infection
    ? {
      hbsag: overview.infection.hbsag === '阳性' ? '阳性' : '阴性',
      hcvab: overview.infection.hcvab === '阳性' ? '阳性' : '阴性',
      hivab: overview.infection.hivab === '阳性' ? '阳性' : '阴性',
      tpab: overview.infection.tpab === '阳性' ? '阳性' : '阴性',
      tb: overview.infection.tb === '阳性' ? '阳性' : '阴性',
      updateDate: overview.infection.updateDate,
    }
    : undefined

  const treatmentPlan: TreatmentPlan | undefined = overview.currentPlan
    ? {
      weeklyFrequency: parseFrequency(overview.currentPlan.frequency),
      biweeklyFrequency: 0,
      duration: overview.currentPlan.duration,
      dryWeight: overview.currentPlan.dryWeight,
      extraWeight: 0,
      vascularAccess: overview.currentPlan.dialysisMode,
      indicators: {
        mode: overview.currentPlan.dialysisMode,
        bloodFlow: overview.currentPlan.bloodFlow,
        bv: '未设置',
        frequencyDesc: overview.currentPlan.frequency,
        autoConfirm: false,
        status: '启用',
        notes: overview.currentPlan.lastTreatmentNote || '',
      },
      anticoagulant: {
        initialDrug: overview.currentPlan.anticoagulant || '低分子肝素',
        initialDose: '',
        maintenanceDrug: overview.currentPlan.anticoagulant || '低分子肝素',
        infusionRate: '',
        infusionTime: '',
        maintenanceDose: '',
        totalDose: '',
      },
      parameters: {
        dialysateType: '标准',
        dialysateGroup: 'A',
        flowRate: 500,
        na: 140,
        ca: 1.5,
        k: 2.0,
        hco3: 35,
        glucose: '5.5',
        conductivity: 14.0,
        temp: 36.5,
        volume: 0,
      },
      materials: [],
      adjustmentHistory: [],
    }
    : undefined

  const riskLevel: '高危' | '中危' | '低危' =
    (header.riskLevel === '高危' || header.riskLevel === '中危' || header.riskLevel === '低危')
      ? header.riskLevel
      : '低危'

  return {
    id: header.id,
    name: header.name,
    avatar: header.avatar,
    age: header.age,
    gender,
    bedNumber: header.bedNumber,
    status: header.status,
    patientType: header.patientType,
    insuranceType: header.insuranceType,
    dryWeight: overview.currentPlan?.dryWeight || 65,
    defaultMode: overview.currentPlan?.dialysisMode || 'HD',
    doctorName: header.doctorName,
    riskLevel,
    diagnosis: '慢性肾脏病5期',
    ...(infection && { infection }),
    orders,
    recentLabs,
    treatmentPlan,
    vitals: {
      bp: '120/80',
      hr: 75,
      spO2: 98,
      weight: overview.currentPlan?.dryWeight || 65,
    },
    dialysisParams: {
      timeRemaining: '00:00',
      ufRate: 0,
      targetUf: 2.5,
      accumulatedUf: 0,
      bloodFlow: overview.currentPlan?.bloodFlow || 250,
      dialysateFlow: 500,
      mode: overview.currentPlan?.dialysisMode || 'HD',
    },
    vascularAccess: {
      type: '-',
      site: '-',
      status: '未知',
    },
    documents: (clinicalFocus.documentStatus || []).map(doc => ({
      id: doc.id,
      title: doc.documentName,
      type: '知情同意书' as const,
      author: '系统',
      date: doc.dueDate || new Date().toISOString(),
    })),
    progressNotes: [],
    medicalHistory: {
      allergies: [],
      primaryDisease: '慢性肾脏病5期',
      pathology: '',
      tumorInfo: '',
      medicalHistory: '',
      complications: [],
    },
    outcome: {
      status: '治疗中',
    },
  }
}

function parseFrequency(frequency: string): number {
  if (frequency.includes('3次/周')) return 3
  if (frequency.includes('2次/周')) return 2
  if (frequency.includes('4次/周')) return 4
  if (frequency.includes('1次/周')) return 1
  return 3
}

// Dashboard Stats 类型定义
export interface DashboardStats {
  activePatients: number
  shiftCount: number
  equipmentCount: number
  todaySchedules: number
  todayTreatments: number
  alertItems: number
  treatmentsByHour: { name: string; value: number }[]
  qualityByHour: { name: string; value: number }[]
}
