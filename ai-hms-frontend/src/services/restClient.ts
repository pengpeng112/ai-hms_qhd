/**
 * REST API 客户端
 * 用于对接后端 REST 接口
 */

import axios from 'axios'

// API 配置
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL?.trim()

if (!API_BASE_URL) {
  throw new Error('缺少必填环境变量 VITE_API_BASE_URL')
}

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

// 患者 REST 数据格式（后端返回）
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
}

// ============ /core 接口类型定义 ============

/**
 * 患者 Core 接口响应
 * GET /api/v1/patients/{id}/core
 */
export interface PatientCoreResponse {
  header: PatientCoreHeader
  overview: PatientCoreOverview
  clinicalFocus: PatientCoreClinical
  navigation?: PatientCoreNavigation
}

/**
 * 患者头部信息
 */
export interface PatientCoreHeader {
  id: string
  name: string
  age: number
  gender: 'M' | 'F'
  bedNumber: string
  status: string  // "治疗中", "待诊", "已结束"
  patientType: string
  insuranceType: string
  doctorName: string
  riskLevel: string  // "高危", "中危", "低危"
  dialysisAge?: string  // 如 "3年2个月"
}

/**
 * 传染病标志
 */
export interface PatientCoreInfection {
  hbsag: string  // "阳性" | "阴性"
  hcvab: string  // "阳性" | "阴性"
  hivab: string  // "阳性" | "阴性"
  tpab: string   // "阳性" | "阴性"
  tb?: string    // "阳性" | "阴性" (可选)
  updateDate: string
}

/**
 * 当前治疗方案摘要
 */
export interface PatientCoreCurrentPlan {
  dialysisMode: string   // HD/HDF/CRRT
  frequency: string      // "3次/周"
  duration: number       // 时长(小时)
  dryWeight: number      // 干体重
  bloodFlow: number      // 血流量
  anticoagulant: string  // 抗凝剂方案
  lastTreatmentNote?: string  // 上次治疗动态
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
 * 关键检验指标趋势
 */
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
  infection?: PatientCoreInfection    // 可选，后端无数据时返回 null/omit
  currentPlan?: PatientCoreCurrentPlan  // 可选，后端无数据时返回 null/omit
  activeOrders: PatientCoreOrder[]
  labTrends: PatientCoreLabTrend[]
}

/**
 * 危急值提醒
 */
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
 * 文书状态
 */
export interface PatientCoreDoc {
  id: string
  documentName: string
  status: string  // '待签署' | '已完成'
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
 * 患者导航信息
 */
export interface PatientCoreNavigation {
  previous?: PatientCoreNavPatient
  next?: PatientCoreNavPatient
  total: number
  currentIndex: number
}

/**
 * 患者导航中的患者信息
 */
export interface PatientCoreNavPatient {
  id: string
  name: string
  bedNumber: string
}

// ============ /basic-info 接口类型定义 ============

/**
 * 患者基本信息档案响应
 * GET /api/v1/patients/{id}/basic-info
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
 * 生命体征与社会信息
 */
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
  type?: string        // 分类（字典值）
  checkTime?: string   // 检查时间
  checkDoctor?: string // 检查医生
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
 * 血管通路 API 响应
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
 * 创建/更新血管通路请求
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
 * 血管通路干预记录 API 响应
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
 * 创建血管通路干预记录请求
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

  // 请求拦截器 - 添加 token
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

  // 响应拦截器 - 统一错误处理
  instance.interceptors.response.use(
    (response) => {
      return response
    },
    (error) => {
      // Token 过期或无效
      if (error.response?.status === 401) {
        // 清除本地存储的认证信息
        localStorage.removeItem('hdis_access_token')
        localStorage.removeItem('hdis_user_info')
        localStorage.removeItem('hdis_token_expiry')
        // 触发重新登录
        window.location.href = '/login'
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
   * 获取患者列表
   */
  async getPatientList(params?: {
    page?: number
    pageSize?: number
    status?: string
    bedNumber?: string
    name?: string
    riskLevel?: string
  }): Promise<ApiSuccessResponse<PaginatedResponse<RestPatient>>> {
    const response = await apiClient.get<ApiSuccessResponse<PaginatedResponse<RestPatient>>>(
      '/api/v1/patients',
      { params }
    )

    if (!response.data.success) {
      throw new Error('获取患者列表失败')
    }

    return response.data
  }

  /**
   * 创建患者
   */
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
      throw new Error('创建患者失败')
    }

    return response.data.data
  }

  /**
   * 获取患者详情
   */
  async getPatient(id: string) {
    const response = await apiClient.get<ApiSuccessResponse<unknown>>(`/api/v1/patients/${id}`)

    if (!response.data.success) {
      throw new Error('获取患者详情失败')
    }

    return response.data.data
  }

  /**
   * 获取患者核心信息（首屏聚合）
   */
  async getPatientCore(id: string): Promise<PatientCoreResponse> {
    const response = await apiClient.get<ApiSuccessResponse<PatientCoreResponse>>(
      `/api/v1/patients/${id}/core`
    )

    if (!response.data.success) {
      throw new Error('获取患者核心信息失败')
    }

    return response.data.data
  }

  /**
   * 获取患者基本信息档案
   */
  async getPatientBasicInfo(id: string): Promise<PatientBasicInfoResponse> {
    const response = await apiClient.get<ApiSuccessResponse<PatientBasicInfoResponse>>(
      `/api/v1/patients/${id}/basic-info`
    )

    if (!response.data.success) {
      throw new Error('获取患者基本信息失败')
    }

    return response.data.data
  }

  /**
   * 更新患者基本信息档案
   */
  async updatePatientBasicInfo(id: string, data: unknown) {
    const response = await apiClient.put<ApiSuccessResponse<unknown> | ApiErrorResponse>(
      `/api/v1/patients/${id}/basic-info`,
      data
    )

    if ('success' in response.data && !response.data.success) {
      throw new Error(response.data.error?.message || '更新患者基本信息失败')
    }

    return (response.data as ApiSuccessResponse<unknown>).data
  }

  /**
   * 获取患者详情（带转换）
   */
  async getPatientById(id: string | number): Promise<Partial<Patient>> {
    const patientData = await this.getPatient(String(id))
    return convertRestPatientToUI(patientData as RestPatient)
  }

  /**
   * 删除患者
   */
  async deletePatient(id: string) {
    const response = await apiClient.delete<ApiSuccessResponse<unknown>>(
      `/api/v1/patients/${id}`
    )

    // 204 No Content 响应没有 body，response.data 可能是空字符串
    // 只要状态码是 2xx 就认为成功
    if (response.status === 204) {
      return
    }

    if (!response.data?.success) {
      throw new Error('删除患者失败')
    }

    return response.data.data
  }

  // ============ 临床病史 API ============

  /**
   * 获取患者临床病史
   */
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
   * 保存患者临床病史
   */
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

  // ============ 检验报告 API ============

  /**
   * 获取患者检验报告列表
   */
  async getLabReports(patientId: string, params?: LabReportListParams): Promise<PaginatedResponse<LabReportApi>> {
    const response = await apiClient.get<ApiSuccessResponse<PaginatedResponse<LabReportApi>>>(
      `/api/v1/patients/${patientId}/lab-reports`,
      { params }
    )

    if (!response.data.success) {
      throw new Error('获取检验报告列表失败')
    }

    return response.data.data
  }

  /**
   * 触发患者检验报告同步
   */
  async syncLabReports(patientId: string): Promise<LabReportSyncResult> {
    const response = await apiClient.post<ApiSuccessResponse<LabReportSyncResult>>(
      `/api/v1/patients/${patientId}/lab-reports/sync`
    )

    if (!response.data.success) {
      throw new Error('同步检验报告失败')
    }

    return response.data.data
  }

  /**
   * 触发患者检查报告同步
   */
  async syncExamReports(patientId: string): Promise<LabReportSyncResult> {
    const response = await apiClient.post<ApiSuccessResponse<LabReportSyncResult>>(
      `/api/v1/patients/${patientId}/exam-reports/sync`
    )

    if (!response.data.success) {
      throw new Error('同步检查报告失败')
    }

    return response.data.data
  }

  /**
   * 获取患者检查报告列表
   */
  async getExamReports(patientId: string, params?: LabReportListParams): Promise<PaginatedResponse<ExamReportApi>> {
    const response = await apiClient.get<ApiSuccessResponse<PaginatedResponse<ExamReportApi>>>(
      `/api/v1/patients/${patientId}/exam-reports`,
      { params }
    )

    if (!response.data.success) {
      throw new Error('获取检查报告列表失败')
    }

    return response.data.data
  }

  /**
   * 触发患者关键指标同步
   */
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
   * 获取患者关键指标列表
   */
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

  // ============ 血管通路 API ============

  /**
   * 获取患者血管通路列表
   */
  async getVascularAccesses(patientId: string): Promise<VascularAccessApi[]> {
    const response = await apiClient.get<ApiSuccessResponse<VascularAccessApi[]>>(
      `/api/v1/patients/${patientId}/vascular-accesses`
    )

    if (!response.data.success) {
      throw new Error('获取血管通路列表失败')
    }

    return response.data.data
  }

  /**
   * 创建血管通路
   */
  async createVascularAccess(patientId: string, data: VascularAccessCreateRequest): Promise<VascularAccessApi> {
    const response = await apiClient.post<ApiSuccessResponse<VascularAccessApi>>(
      `/api/v1/patients/${patientId}/vascular-accesses`,
      data
    )

    if (!response.data.success) {
      throw new Error('创建血管通路失败')
    }

    return response.data.data
  }

  /**
   * 更新血管通路
   */
  async updateVascularAccess(patientId: string, accessId: string, data: VascularAccessCreateRequest): Promise<VascularAccessApi> {
    const response = await apiClient.put<ApiSuccessResponse<VascularAccessApi>>(
      `/api/v1/patients/${patientId}/vascular-accesses/${accessId}`,
      data
    )

    if (!response.data.success) {
      throw new Error('更新血管通路失败')
    }

    return response.data.data
  }

  /**
   * 删除血管通路
   */
  async deleteVascularAccess(patientId: string, accessId: string): Promise<void> {
    const response = await apiClient.delete<ApiSuccessResponse<unknown>>(
      `/api/v1/patients/${patientId}/vascular-accesses/${accessId}`
    )

    if (response.status !== 204 && !response.data?.success) {
      throw new Error('删除血管通路失败')
    }
  }

  // ============ 血管通路干预记录接口 ============

  /**
   * 获取患者的血管通路干预记录列表
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
   * 创建血管通路干预记录
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
   * 删除血管通路干预记录
   */
  async deleteVascularAccessIntervention(patientId: string, interventionId: string): Promise<void> {
    const response = await apiClient.delete<ApiSuccessResponse<unknown>>(
      `/api/v1/patients/${patientId}/vascular-access-interventions/${interventionId}`
    )

    if (response.status !== 204 && !response.data?.success) {
      throw new Error('删除干预记录失败')
    }
  }
}

// 导出单例
export const restApi = new RestApiService()

// 错误处理辅助函数
export function getErrorMessage(error: unknown): string {
  if (axios.isAxiosError(error)) {
    const data = error.response?.data as ApiErrorResponse | undefined
    if (data?.error) {
      return data.error.message
    }
    return error.message
  }
  if (error instanceof Error) {
    return error.message
  }
  return '未知错误'
}

// ============ 数据转换函数 ============

import type {
  Patient,
  LabResult,
  TreatmentPlan,
  MedicalOrder,
  InfectionInfo,
} from '@/types/original'

/**
 * 后端患者数据转换为前端 UI 格式
 * 只转换接口实际返回的字段，返回 Partial<Patient> 类型
 */
export function convertRestPatientToUI(restPatient: RestPatient): Partial<Patient> {
  const statusMap: Record<string, string> = {
    active: '透析中',
    inactive: '候诊',
    discharged: '已结束',
  }
  const status = statusMap[restPatient.status] || '居家'

  // 后端返回 ISO 5218 标准的 M/F，转换为中文
  const gender: '男' | '女' = restPatient.gender === 'M' ? '男' : '女'

  return {
    id: restPatient.id,
    name: restPatient.name,
    gender,
    age: restPatient.age,
    bedNumber: restPatient.bedNumber || '',
    status,
    patientType: restPatient.patientType || '门诊',
    insuranceType: restPatient.insuranceType || '自费',
    dryWeight: restPatient.dryWeight ?? 65,
    defaultMode: restPatient.defaultMode || 'HD',
    doctorName: restPatient.doctorName || '王医生',
  }
}

/**
 * 批量转换患者数据
 */
export function convertRestPatientList(patients: unknown[]): Partial<Patient>[] {
  return patients.map((p) => convertRestPatientToUI(p as RestPatient))
}

// ============ /core 接口数据转换函数 ============

/**
 * 将 /core 接口返回的数据转换为前端 Patient 类型
 *
 * 映射规则：
 * - header → 基础字段（id, name, age, gender 等）
 * - overview.infection → Patient.infection
 * - overview.currentPlan → Patient.treatmentPlan（基础字段）
 * - overview.activeOrders → Patient.orders
 * - overview.labTrends[] → Patient.recentLabs（取最新值）
 * - clinicalFocus.criticalAlerts → 用于填充临床焦点
 *
 * 对于 /core 接口不包含的字段，设置合理的默认值
 */
export function convertCoreResponseToPatient(
  coreData: PatientCoreResponse
): Partial<Patient> {
  const { header, overview, clinicalFocus } = coreData

  // Gender 转换: M → '男', F → '女'
  const gender: '男' | '女' = header.gender === 'M' ? '男' : '女'

  // LabTrends[] → recentLabs（取最新的实验室数据）
  const recentLabs: LabResult[] = []
  if (overview.labTrends && overview.labTrends.length > 0) {
    // 遍历所有趋势，取每个指标的最新值
    overview.labTrends.forEach(trend => {
      if (trend.data && trend.data.length > 0) {
        const latest = trend.data[trend.data.length - 1]
        recentLabs.push({
          id: `${trend.code}_${latest.date}`,
          name: trend.name,
          value: String(latest.value),
          unit: trend.unit,
          date: latest.date,
          isAbnormal: latest.isAbnormal,
          reference: trend.normalRange,
        })
      }
    })
  }

  // activeOrders → orders（转换类型）
  const orders: MedicalOrder[] = (overview.activeOrders || []).map(order => ({
    id: order.id,
    content: order.content,
    type: (order.type === '长期' || order.type === '临时') ? order.type : '长期',
    status: '待执行',  // 默认为待执行
    doctor: order.doctor,
    startTime: order.startTime,
  }))

  // infection → 转换为 InfectionInfo 类型（仅当有数据时）
  const infection: InfectionInfo | undefined = overview.infection ? {
    hbsag: (overview.infection.hbsag === '阳性' ? '阳性' : '阴性'),
    hcvab: (overview.infection.hcvab === '阳性' ? '阳性' : '阴性'),
    hivab: (overview.infection.hivab === '阳性' ? '阳性' : '阴性'),
    tpab: (overview.infection.tpab === '阳性' ? '阳性' : '阴性'),
    tb: overview.infection.tb === '阳性' ? '阳性' : '阴性',
    updateDate: overview.infection.updateDate,
  } : undefined

  // currentPlan → treatmentPlan（基础映射）
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

  // riskLevel 转换
  const riskLevel: '高危' | '中危' | '低危' =
    (header.riskLevel === '高危' || header.riskLevel === '中危' || header.riskLevel === '低危')
      ? header.riskLevel
      : '低危'

  // 合并基础数据
  const basePatient: Partial<Patient> = {
    id: header.id,
    name: header.name,
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
    diagnosis: '慢性肾脏病5期',  // 默认诊断
    ...(infection && { infection }),  // 只在有感染数据时才添加
    orders,
    recentLabs,
    treatmentPlan,
  }

  // 添加默认值（/core 接口不包含的字段）
  return {
    ...basePatient,
    // 生命体征（后续从独立 API 获取）
    vitals: {
      bp: '120/80',
      hr: 75,
      spO2: 98,
      weight: overview.currentPlan?.dryWeight || 65,
    },
    // 透析参数（后续从独立 API 获取）
    dialysisParams: {
      timeRemaining: '00:00',
      ufRate: 0,
      targetUf: 2.5,
      accumulatedUf: 0,
      bloodFlow: overview.currentPlan?.bloodFlow || 250,
      dialysateFlow: 500,
      mode: overview.currentPlan?.dialysisMode || 'HD',
    },
    // 血管通路（后续从独立 API 获取）
    vascularAccess: {
      type: '-',
      site: '-',
      status: '未知',
    },
    // EMR 文档（从 clinicalFocus.documentStatus 转换）
    documents: (clinicalFocus.documentStatus || []).map(doc => ({
      id: doc.id,
      title: doc.documentName,
      type: '知情同意书' as const,
      author: '系统',
      date: doc.dueDate || new Date().toISOString(),
    })),
    // 病程记录
    progressNotes: [],
    // 医疗史
    medicalHistory: {
      allergies: [],
      primaryDisease: '慢性肾脏病5期',
      pathology: '',
      tumorInfo: '',
      medicalHistory: '',
      complications: [],
    },
    // 治疗结局
    outcome: {
      status: '治疗中',
    },
  }
}

/**
 * 解析频次字符串为每周次数
 */
function parseFrequency(frequency: string): number {
  if (frequency.includes('3次/周')) return 3
  if (frequency.includes('2次/周')) return 2
  if (frequency.includes('4次/周')) return 4
  if (frequency.includes('1次/周')) return 1
  return 3  // 默认值
}
