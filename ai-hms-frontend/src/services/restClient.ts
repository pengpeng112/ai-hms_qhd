/**
 * REST API 瀹㈡埛绔? * 鐢ㄤ簬瀵规帴鍚庣 REST 鎺ュ彛
 */

import axios from 'axios'

// API 閰嶇疆
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL?.trim()

if (!API_BASE_URL) {
  throw new Error('缂哄皯蹇呭～鐜鍙橀噺 VITE_API_BASE_URL')
}

const API_CONFIG = {
  baseURL: API_BASE_URL,
  timeout: 10000,
}

// 鏍囧噯鍝嶅簲鏍煎紡
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

// 鐧诲綍璇锋眰
export interface LoginRequest {
  username: string
  password: string
}

// 鐧诲綍鍝嶅簲
export interface LoginResponse {
  token: string
  userId: string
  username: string
  realName: string
  role: string
}

// 鍒嗛〉鍝嶅簲鍏冩暟鎹?export interface PaginationMeta {
  page: number
  pageSize: number
  total: number
  totalPages: number
}

// 鍒嗛〉鍝嶅簲
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
}

// ============ 鐝 & 鎺掔彮绫诲瀷 ============

export interface RestUser {
  id: string
  username: string
  realName: string
  role: string
  status: string
  departmentId: number | null
}

export interface RestInventoryItem {
  id: string
  tenantId: number
  name: string
  spec: string
  category: string
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

export interface RestShift {
  id: number
  tenantId: number
  name: string        // 鐝鍚嶇О锛氭棭鐝?涓彮/鏅氱彮
  startTime: string   // HH:MM
  endTime: string     // HH:MM
  type: string
  isDisabled: boolean
  sort: number
  notes: string
}

export interface RestPatientShift {
  id: number
  tenantId: number
  patientId: number
  scheduleDate: string   // ISO 鏃ユ湡
  shiftId: number
  bedId?: number
  wardId?: number
  status: number
  isDisabled: boolean
  notes: string
  creatorId: number
  createTime: string
  lastModifyTime: string
  // Preloaded 鍏宠仈
  patient?: RestPatient
  shift?: RestShift
  bed?: { id: number; name: string; wardId: number; sort: number; status: string }
  ward?: { id: number; name: string; sort: number; status: string }
}

// ============ 娌荤枟璁板綍绫诲瀷 ============

export interface RestDuringParam {
  id: number
  tenantId: number
  treatmentId: number
  recordTime: string
  code: string
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

export interface RestTreatment {
  id: number
  tenantId: number
  patientId: number
  treatmentDate: string
  shiftId?: number
  treatmentType: string
  status: number       // 0-寰呭紑濮?1-杩涜涓?2-宸插畬鎴?3-宸插彇娑?  startTime?: string
  endTime?: string
  notes?: string
  creatorId: number
  createTime: string
  lastModifyTime: string
  patient?: RestPatient
  shift?: RestShift
  duringParams?: RestDuringParam[]
}

export interface CreateTreatmentRequest {
  patientId: number
  treatmentDate: string
  type: number
  status?: number
  notes?: string
}

export interface UpdateTreatmentRequest {
  status?: number
  notes?: string
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

// ============ /core 鎺ュ彛绫诲瀷瀹氫箟 ============

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
  age: number
  gender: 'M' | 'F'
  bedNumber: string
  status: string  // "娌荤枟涓?, "寰呰瘖", "宸茬粨鏉?
  patientType: string
  insuranceType: string
  doctorName: string
  riskLevel: string  // "楂樺嵄", "涓嵄", "浣庡嵄"
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
 * 褰撳墠娌荤枟鏂规鎽樿
 */
export interface PatientCoreCurrentPlan {
  dialysisMode: string   // HD/HDF/CRRT
  frequency: string      // "3娆?鍛?
  duration: number       // 鏃堕暱(灏忔椂)
  dryWeight: number      // 骞蹭綋閲?  bloodFlow: number      // 琛€娴侀噺
  anticoagulant: string  // 鎶楀嚌鍓傛柟妗?  lastTreatmentNote?: string  // 涓婃娌荤枟鍔ㄦ€?}

/**
 * 鍖诲槺鎽樿
 */
export interface PatientCoreOrder {
  id: string
  content: string
  type: string  // "闀挎湡", "涓存椂"
  startTime: string
  doctor: string
}

/**
 * 瀹為獙瀹よ秼鍔挎暟鎹偣
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
 * Overview Tab 鏁版嵁
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
 * 涓村簥鐒︾偣鏁版嵁
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

// ============ /basic-info 鎺ュ彛绫诲瀷瀹氫箟 ============

/**
 * 鎮ｈ€呭熀鏈俊鎭。妗堝搷搴? * GET /api/v1/patients/{id}/basic-info
 */
export interface PatientBasicInfoResponse {
  personalInfo: PatientBasicPersonal
  medicalInfo: PatientBasicMedical
  vitalSocialInfo: PatientBasicVitalSocial
  contactInfo: PatientBasicContact
  // TODO: 鍚庣画娣诲姞浠ヤ笅瀛楁
  // familyContacts: FamilyContact[]
  // electronicDocuments: ElectronicDocument[]
}

/**
 * 韬唤鏍稿績淇℃伅
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
 * 鍖荤枟鐧昏淇℃伅
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
 * 鑱旂郴淇℃伅
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
 * 瀹跺睘涓庣揣鎬ヨ仈绯讳汉锛堝悗缁疄鐜帮級
 */
export interface FamilyContact {
  id: string
  name: string
  phone: string
  type: 'primary' | 'family' | 'emergency'
  relation: string
}

/**
 * 鐢靛瓙鏂囦功锛堝悗缁疄鐜帮級
 */
export interface ElectronicDocument {
  id: string
  name: string
  type: string
  status: 'signed' | 'pending'
  date: string
}

// ============ /medical-history 鎺ュ彛绫诲瀷瀹氫箟 ============

/**
 * 鐥呭彶鍐呭
 */
export interface HistoryContent {
  content: string
}

/**
 * 甯﹀悕绉扮殑鐥呭彶鍐呭锛堜笓绉戣瘖鐤楄褰曪級
 */
export interface HistoryNamedContent {
  name: string
  content: string
  type?: string        // 鍒嗙被锛堝瓧鍏稿€硷級
  checkTime?: string   // 妫€鏌ユ椂闂?  checkDoctor?: string // 妫€鏌ュ尰鐢?}

/**
 * 涓村簥鐥呭彶鍝嶅簲
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
 * 杞綊璁板綍
 */
export interface OutcomeRecordApi {
  id: string
  type: string
  reason: string
  time: string
  remarks: string
  registrar: string
  registrationTime: string
  isDoorRule: boolean  // 鏄惁闂ㄨ
}

// ============ /vascular-accesses 鎺ュ彛绫诲瀷瀹氫箟 ============

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

// ============ /lab-reports 鎺ュ彛绫诲瀷瀹氫箟 ============

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

// 鍒涘缓 axios 瀹炰緥
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
      // Token 杩囨湡鎴栨棤鏁?      if (error.response?.status === 401) {
        // 娓呴櫎鏈湴瀛樺偍鐨勮璇佷俊鎭?        localStorage.removeItem('hdis_access_token')
        localStorage.removeItem('hdis_user_info')
        localStorage.removeItem('hdis_token_expiry')
        // 瑙﹀彂閲嶆柊鐧诲綍
        window.location.href = '/login'
      }
      return Promise.reject(error)
    }
  )

  return instance
}

// 瀵煎嚭 axios 瀹炰緥
export const apiClient = createAxiosInstance()

// API 鏈嶅姟绫?class RestApiService {
  /**
   * 鐢ㄦ埛鐧诲綍
   */
  async login(credentials: LoginRequest): Promise<LoginResponse> {
    const response = await apiClient.post<ApiSuccessResponse<LoginResponse>>(
      '/api/v1/auth/login',
      credentials
    )

    if (!response.data.success) {
      throw new Error('鐧诲綍澶辫触')
    }

    return response.data.data
  }

  /**
   * 鑾峰彇褰撳墠鐢ㄦ埛淇℃伅
   */
  async getCurrentUser() {
    const response = await apiClient.get<ApiSuccessResponse<unknown>>('/api/v1/me')

    if (!response.data.success) {
      throw new Error('鑾峰彇鐢ㄦ埛淇℃伅澶辫触')
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
  }): Promise<ApiSuccessResponse<PaginatedResponse<RestPatient>>> {
    const response = await apiClient.get<ApiSuccessResponse<PaginatedResponse<RestPatient>>>(
      '/api/v1/patients',
      { params }
    )

    if (!response.data.success) {
      throw new Error('鑾峰彇鎮ｈ€呭垪琛ㄥけ璐?)
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
    // 鍩烘湰淇℃伅妗ｆ
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
      throw new Error('鍒涘缓鎮ｈ€呭け璐?)
    }

    return response.data.data
  }

  /**
   * 鑾峰彇鎮ｈ€呰鎯?   */
  async getPatient(id: string) {
    const response = await apiClient.get<ApiSuccessResponse<unknown>>(`/api/v1/patients/${id}`)

    if (!response.data.success) {
      throw new Error('鑾峰彇鎮ｈ€呰鎯呭け璐?)
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
      throw new Error('鑾峰彇鎮ｈ€呮牳蹇冧俊鎭け璐?)
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
      throw new Error('鑾峰彇鎮ｈ€呭熀鏈俊鎭け璐?)
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
      throw new Error(response.data.error?.message || '鏇存柊鎮ｈ€呭熀鏈俊鎭け璐?)
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

  /**
   * 鍒犻櫎鎮ｈ€?   */
  async deletePatient(id: string) {
    const response = await apiClient.delete<ApiSuccessResponse<unknown>>(
      `/api/v1/patients/${id}`
    )

    // 204 No Content 鍝嶅簲娌℃湁 body锛宺esponse.data 鍙兘鏄┖瀛楃涓?    // 鍙鐘舵€佺爜鏄?2xx 灏辫涓烘垚鍔?    if (response.status === 204) {
      return
    }

    if (!response.data?.success) {
      throw new Error('鍒犻櫎鎮ｈ€呭け璐?)
    }

    return response.data.data
  }

  // ============ 涓村簥鐥呭彶 API ============

  /**
   * 鑾峰彇鎮ｈ€呬复搴婄梾鍙?   */
  async getMedicalHistory(patientId: string): Promise<MedicalHistoryApiResponse> {
    const response = await apiClient.get<ApiSuccessResponse<MedicalHistoryApiResponse>>(
      `/api/v1/patients/${patientId}/medical-history`
    )

    if (!response.data.success) {
      throw new Error('鑾峰彇涓村簥鐥呭彶澶辫触')
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
      throw new Error('淇濆瓨涓村簥鐥呭彶澶辫触')
    }

    return response.data.data
  }

  /**
   * 鑾峰彇杞綊璁板綍鍒楄〃
   */
  async getOutcomeRecords(patientId: string): Promise<OutcomeRecordApi[]> {
    const response = await apiClient.get<ApiSuccessResponse<OutcomeRecordApi[]>>(
      `/api/v1/patients/${patientId}/outcome-records`
    )

    if (!response.data.success) {
      throw new Error('鑾峰彇杞綊璁板綍澶辫触')
    }

    return response.data.data
  }

  /**
   * 鍒涘缓杞綊璁板綍
   */
  async createOutcomeRecord(patientId: string, data: Omit<OutcomeRecordApi, 'id'>): Promise<OutcomeRecordApi> {
    const response = await apiClient.post<ApiSuccessResponse<OutcomeRecordApi>>(
      `/api/v1/patients/${patientId}/outcome-records`,
      data
    )

    if (!response.data.success) {
      throw new Error('鍒涘缓杞綊璁板綍澶辫触')
    }

    return response.data.data
  }

  /**
   * 鏇存柊杞綊璁板綍
   */
  async updateOutcomeRecord(patientId: string, recordId: string, data: Omit<OutcomeRecordApi, 'id'>): Promise<OutcomeRecordApi> {
    const response = await apiClient.put<ApiSuccessResponse<OutcomeRecordApi>>(
      `/api/v1/patients/${patientId}/outcome-records/${recordId}`,
      data
    )

    if (!response.data.success) {
      throw new Error('鏇存柊杞綊璁板綍澶辫触')
    }

    return response.data.data
  }

  /**
   * 鍒犻櫎杞綊璁板綍
   */
  async deleteOutcomeRecord(patientId: string, recordId: string): Promise<void> {
    const response = await apiClient.delete<ApiSuccessResponse<unknown>>(
      `/api/v1/patients/${patientId}/outcome-records/${recordId}`
    )

    if (response.status !== 204 && !response.data?.success) {
      throw new Error('鍒犻櫎杞綊璁板綍澶辫触')
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
      throw new Error('鑾峰彇妫€楠屾姤鍛婂垪琛ㄥけ璐?)
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
      throw new Error('鍚屾妫€楠屾姤鍛婂け璐?)
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
      throw new Error('鍚屾妫€鏌ユ姤鍛婂け璐?)
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
      throw new Error('鑾峰彇妫€鏌ユ姤鍛婂垪琛ㄥけ璐?)
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
      throw new Error('鍚屾鍏抽敭鎸囨爣澶辫触')
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
      throw new Error('鑾峰彇鍏抽敭鎸囨爣澶辫触')
    }

    return response.data.data
  }

  // ============ Settings / HDIS Integration API ============

  async getHdisIntegrationSettings(): Promise<HdisIntegrationSettings> {
    const response = await apiClient.get<ApiSuccessResponse<HdisIntegrationSettings>>(
      '/api/v1/settings/integrations/hdis'
    )

    if (!response.data.success) {
      throw new Error('鑾峰彇 HDIS 閰嶇疆澶辫触')
    }

    return response.data.data
  }

  async updateHdisIntegrationSettings(payload: HdisIntegrationSettingsUpdatePayload): Promise<HdisIntegrationSettings> {
    const response = await apiClient.put<ApiSuccessResponse<HdisIntegrationSettings>>(
      '/api/v1/settings/integrations/hdis',
      payload
    )

    if (!response.data.success) {
      throw new Error('淇濆瓨 HDIS 閰嶇疆澶辫触')
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
      throw new Error('鍒锋柊 HDIS Token 澶辫触')
    }

    return response.data.data
  }

  async getSystemLogs(params?: SystemLogsQuery): Promise<SystemLogsResponse> {
    const response = await apiClient.get<ApiSuccessResponse<SystemLogsResponse>>(
      '/api/v1/settings/logs',
      { params }
    )

    if (!response.data.success) {
      throw new Error('鑾峰彇绯荤粺鏃ュ織澶辫触')
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
      throw new Error('鑾峰彇骞查璁板綍澶辫触')
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
      throw new Error('鍒涘缓骞查璁板綍澶辫触')
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
      throw new Error('鍒犻櫎骞查璁板綍澶辫触')
    }
  }

  // ============ 鐢ㄦ埛绠＄悊 ============

  /**
   * 鑾峰彇鐢ㄦ埛鍒楄〃锛堢敤浜庤鑹查€夋嫨椤甸潰锛?   */
  async getUserList(params?: { role?: string; status?: string }): Promise<RestUser[]> {
    const response = await apiClient.get<ApiSuccessResponse<RestUser[]>>('/api/v1/users', { params })
    if (!response.data.success) {
      throw new Error('鑾峰彇鐢ㄦ埛鍒楄〃澶辫触')
    }
    return response.data.data
  }

  async getRolePermissions(role: string): Promise<ApiSuccessResponse<{ role: string; permissionCodes: string[] }>> {
    const response = await apiClient.get<ApiSuccessResponse<{ role: string; permissionCodes: string[] }>>(`/api/v1/role-permissions/${role}`)
    if (!response.data.success) {
      throw new Error('获取角色权限失败')
    }
    return response.data
  }

  async getPermissions(): Promise<ApiSuccessResponse<{ items: RestPermission[]; total: number }>> {
    const response = await apiClient.get<ApiSuccessResponse<{ items: RestPermission[]; total: number }>>('/api/v1/permissions')
    if (!response.data.success) {
      throw new Error('获取权限列表失败')
    }
    return response.data
  }

  async setRolePermissions(role: string, permissionCodes: string[]): Promise<ApiSuccessResponse<{ role: string; permissionCodes: string[] }>> {
    const response = await apiClient.put<ApiSuccessResponse<{ role: string; permissionCodes: string[] }>>(
      `/api/v1/role-permissions/${role}`,
      { permissionCodes }
    )
    if (!response.data.success) {
      throw new Error('更新角色权限失败')
    }
    return response.data
  }

  // ============ 鐪嬫澘缁熻 ============

  /**
   * 鑾峰彇鐪嬫澘缁熻姹囨€?   */
  async getDashboardStats(): Promise<{
    activePatients: number
    todayTreatments: number
    alertItems: number
    treatmentsByHour: { name: string; value: number }[]
    qualityByHour: { name: string; value: number }[]
  }> {
    const response = await apiClient.get<ApiSuccessResponse<{
      activePatients: number
      todayTreatments: number
      alertItems: number
      treatmentsByHour: { name: string; value: number }[]
      qualityByHour: { name: string; value: number }[]
    }>>('/api/v1/dashboard/stats')
    if (!response.data.success) {
      throw new Error('鑾峰彇鐪嬫澘缁熻澶辫触')
    }
    return response.data.data
  }

  // ============ 璁惧绠＄悊 ============

  /**
   * 鑾峰彇璁惧鍒楄〃
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
      throw new Error('鑾峰彇璁惧鍒楄〃澶辫触')
    }
    return response.data.data.items
  }

  // ============ 搴撳瓨绠＄悊 ============

  /**
   * 鑾峰彇搴撳瓨鍝佺洰鍒楄〃
   */
  async getInventoryItems(params?: {
    page?: number
    pageSize?: number
    category?: string
    keyword?: string
  }): Promise<{ items: RestInventoryItem[]; total: number; page: number; pageSize: number; totalPage: number }> {
    const response = await apiClient.get<ApiSuccessResponse<{ items: RestInventoryItem[]; total: number; page: number; pageSize: number; totalPage: number }>>('/api/v1/inventory/items', { params })
    if (!response.data.success) {
      throw new Error('鑾峰彇搴撳瓨鍒楄〃澶辫触')
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
      throw new Error('鑾峰彇鍑哄叆搴撹褰曞け璐?)
    }
    return response.data.data
  }

  /**
   * 鑾峰彇鏍囩鎵撳嵃浠诲姟
   */
  async getLabelTasks(params?: {
    page?: number
    pageSize?: number
    status?: string
  }): Promise<{ items: RestLabelTask[]; total: number; page: number; pageSize: number; totalPage: number }> {
    const response = await apiClient.get<ApiSuccessResponse<{ items: RestLabelTask[]; total: number; page: number; pageSize: number; totalPage: number }>>('/api/v1/inventory/labels', { params })
    if (!response.data.success) {
      throw new Error('鑾峰彇鏍囩浠诲姟澶辫触')
    }
    return response.data.data
  }

  // ============ 鐝绠＄悊 ============

  /**
   * 鑾峰彇鐝鍒楄〃
   */
  async getShifts(): Promise<ApiSuccessResponse<RestShift[]>> {
    const response = await apiClient.get<ApiSuccessResponse<RestShift[]>>('/api/v1/shifts')
    if (!response.data.success) {
      throw new Error('鑾峰彇鐝鍒楄〃澶辫触')
    }
    return response.data
  }

  // ============ 鎮ｈ€呮帓鐝鐞?============

  /**
   * 鑾峰彇鎮ｈ€呮帓鐝垪琛?   */
  async getPatientShifts(params?: {
    page?: number
    pageSize?: number
    patientId?: number
    shiftId?: number
    startDate?: string
    endDate?: string
    status?: number
  }): Promise<ApiSuccessResponse<PaginatedResponse<RestPatientShift>>> {
    const response = await apiClient.get<ApiSuccessResponse<PaginatedResponse<RestPatientShift>>>(
      '/api/v1/patient-shifts',
      { params }
    )
    if (!response.data.success) {
      throw new Error('鑾峰彇鎮ｈ€呮帓鐝垪琛ㄥけ璐?)
    }
    return response.data
  }

  /**
   * 鍒涘缓鎮ｈ€呮帓鐝?   */
  async createPatientShift(data: {
    patientId: number
    scheduleDate: string
    shiftId: number
    bedId?: number
    wardId?: number
    notes?: string
  }): Promise<unknown> {
    const response = await apiClient.post<unknown>('/api/v1/patient-shifts', data)
    return response.data
  }

  // ============ 娌荤枟璁板綍绠＄悊 ============

  /**
   * 鑾峰彇娌荤枟璁板綍鍒楄〃
   */
  async getTreatments(params?: {
    page?: number
    pageSize?: number
    patientId?: number
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
      throw new Error('鑾峰彇娌荤枟璁板綍鍒楄〃澶辫触')
    }
    return response.data
  }

  /**
   * 鑾峰彇娌荤枟璁板綍璇︽儏
   */
  async getTreatment(id: number): Promise<ApiSuccessResponse<RestTreatment>> {
    const response = await apiClient.get<ApiSuccessResponse<RestTreatment>>(`/api/v1/treatments/${id}`)
    if (!response.data.success) {
      throw new Error('鑾峰彇娌荤枟璁板綍璇︽儏澶辫触')
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
      throw new Error('鑾峰彇娌荤枟璁板綍澶辫触')
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

// 瀵煎嚭鍗曚緥
export const restApi = new RestApiService()

// 閿欒澶勭悊杈呭姪鍑芥暟
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
  return '鏈煡閿欒'
}

// ============ 鏁版嵁杞崲鍑芥暟 ============

import type {
  Patient,
  LabResult,
  TreatmentPlan,
  MedicalOrder,
  InfectionInfo,
} from '@/types/original'

/**
 * 鍚庣鎮ｈ€呮暟鎹浆鎹负鍓嶇 UI 鏍煎紡
 * 鍙浆鎹㈡帴鍙ｅ疄闄呰繑鍥炵殑瀛楁锛岃繑鍥?Partial<Patient> 绫诲瀷
 */
export function convertRestPatientToUI(restPatient: RestPatient): Partial<Patient> {
  const statusMap: Record<string, string> = {
    active: '閫忔瀽涓?,
    inactive: '鍊欒瘖',
    discharged: '宸茬粨鏉?,
  }
  const status = statusMap[restPatient.status] || '灞呭'

  // 鍚庣杩斿洖 ISO 5218 鏍囧噯鐨?M/F锛岃浆鎹负涓枃
  const gender: '鐢? | '濂? = restPatient.gender === 'M' ? '鐢? : '濂?

  return {
    id: restPatient.id,
    name: restPatient.name,
    gender,
    age: restPatient.age,
    bedNumber: restPatient.bedNumber || '',
    status,
    patientType: restPatient.patientType || '闂ㄨ瘖',
    insuranceType: restPatient.insuranceType || '鑷垂',
    dryWeight: restPatient.dryWeight ?? 65,
    defaultMode: restPatient.defaultMode || 'HD',
    doctorName: restPatient.doctorName || '鐜嬪尰鐢?,
  }
}

/**
 * 鎵归噺杞崲鎮ｈ€呮暟鎹? */
export function convertRestPatientList(patients: unknown[]): Partial<Patient>[] {
  return patients.map((p) => convertRestPatientToUI(p as RestPatient))
}

// ============ /core 鎺ュ彛鏁版嵁杞崲鍑芥暟 ============

/**
 * 灏?/core 鎺ュ彛杩斿洖鐨勬暟鎹浆鎹负鍓嶇 Patient 绫诲瀷
 *
 * 鏄犲皠瑙勫垯锛? * - header 鈫?鍩虹瀛楁锛坕d, name, age, gender 绛夛級
 * - overview.infection 鈫?Patient.infection
 * - overview.currentPlan 鈫?Patient.treatmentPlan锛堝熀纭€瀛楁锛? * - overview.activeOrders 鈫?Patient.orders
 * - overview.labTrends[] 鈫?Patient.recentLabs锛堝彇鏈€鏂板€硷級
 * - clinicalFocus.criticalAlerts 鈫?鐢ㄤ簬濉厖涓村簥鐒︾偣
 *
 * 瀵逛簬 /core 鎺ュ彛涓嶅寘鍚殑瀛楁锛岃缃悎鐞嗙殑榛樿鍊? */
export function convertCoreResponseToPatient(
  coreData: PatientCoreResponse
): Partial<Patient> {
  const { header, overview, clinicalFocus } = coreData

  // Gender 杞崲: M 鈫?'鐢?, F 鈫?'濂?
  const gender: '鐢? | '濂? = header.gender === 'M' ? '鐢? : '濂?

  // LabTrends[] 鈫?recentLabs锛堝彇鏈€鏂扮殑瀹為獙瀹ゆ暟鎹級
  const recentLabs: LabResult[] = []
  if (overview.labTrends && overview.labTrends.length > 0) {
    // 閬嶅巻鎵€鏈夎秼鍔匡紝鍙栨瘡涓寚鏍囩殑鏈€鏂板€?    overview.labTrends.forEach(trend => {
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

  // activeOrders 鈫?orders锛堣浆鎹㈢被鍨嬶級
  const orders: MedicalOrder[] = (overview.activeOrders || []).map(order => ({
    id: order.id,
    content: order.content,
    type: (order.type === '闀挎湡' || order.type === '涓存椂') ? order.type : '闀挎湡',
    status: '寰呮墽琛?,  // 榛樿涓哄緟鎵ц
    doctor: order.doctor,
    startTime: order.startTime,
  }))

  // infection 鈫?杞崲涓?InfectionInfo 绫诲瀷锛堜粎褰撴湁鏁版嵁鏃讹級
  const infection: InfectionInfo | undefined = overview.infection ? {
    hbsag: (overview.infection.hbsag === '闃虫€? ? '闃虫€? : '闃存€?),
    hcvab: (overview.infection.hcvab === '闃虫€? ? '闃虫€? : '闃存€?),
    hivab: (overview.infection.hivab === '闃虫€? ? '闃虫€? : '闃存€?),
    tpab: (overview.infection.tpab === '闃虫€? ? '闃虫€? : '闃存€?),
    tb: overview.infection.tb === '闃虫€? ? '闃虫€? : '闃存€?,
    updateDate: overview.infection.updateDate,
  } : undefined

  // currentPlan 鈫?treatmentPlan锛堝熀纭€鏄犲皠锛?  const treatmentPlan: TreatmentPlan | undefined = overview.currentPlan
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
          bv: '鏈缃?,
          frequencyDesc: overview.currentPlan.frequency,
          autoConfirm: false,
          status: '鍚敤',
          notes: overview.currentPlan.lastTreatmentNote || '',
        },
        anticoagulant: {
          initialDrug: overview.currentPlan.anticoagulant || '浣庡垎瀛愯倽绱?,
          initialDose: '',
          maintenanceDrug: overview.currentPlan.anticoagulant || '浣庡垎瀛愯倽绱?,
          infusionRate: '',
          infusionTime: '',
          maintenanceDose: '',
          totalDose: '',
        },
        parameters: {
          dialysateType: '鏍囧噯',
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

  // riskLevel 杞崲
  const riskLevel: '楂樺嵄' | '涓嵄' | '浣庡嵄' =
    (header.riskLevel === '楂樺嵄' || header.riskLevel === '涓嵄' || header.riskLevel === '浣庡嵄')
      ? header.riskLevel
      : '浣庡嵄'

  // 鍚堝苟鍩虹鏁版嵁
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
    diagnosis: '鎱㈡€ц偩鑴忕梾5鏈?,  // 榛樿璇婃柇
    ...(infection && { infection }),  // 鍙湪鏈夋劅鏌撴暟鎹椂鎵嶆坊鍔?    orders,
    recentLabs,
    treatmentPlan,
  }

  // 娣诲姞榛樿鍊硷紙/core 鎺ュ彛涓嶅寘鍚殑瀛楁锛?  return {
    ...basePatient,
    // 鐢熷懡浣撳緛锛堝悗缁粠鐙珛 API 鑾峰彇锛?    vitals: {
      bp: '120/80',
      hr: 75,
      spO2: 98,
      weight: overview.currentPlan?.dryWeight || 65,
    },
    // 閫忔瀽鍙傛暟锛堝悗缁粠鐙珛 API 鑾峰彇锛?    dialysisParams: {
      timeRemaining: '00:00',
      ufRate: 0,
      targetUf: 2.5,
      accumulatedUf: 0,
      bloodFlow: overview.currentPlan?.bloodFlow || 250,
      dialysateFlow: 500,
      mode: overview.currentPlan?.dialysisMode || 'HD',
    },
    // 琛€绠￠€氳矾锛堝悗缁粠鐙珛 API 鑾峰彇锛?    vascularAccess: {
      type: '-',
      site: '-',
      status: '鏈煡',
    },
    // EMR 鏂囨。锛堜粠 clinicalFocus.documentStatus 杞崲锛?    documents: (clinicalFocus.documentStatus || []).map(doc => ({
      id: doc.id,
      title: doc.documentName,
      type: '鐭ユ儏鍚屾剰涔? as const,
      author: '绯荤粺',
      date: doc.dueDate || new Date().toISOString(),
    })),
    // 鐥呯▼璁板綍
    progressNotes: [],
    // 鍖荤枟鍙?    medicalHistory: {
      allergies: [],
      primaryDisease: '鎱㈡€ц偩鑴忕梾5鏈?,
      pathology: '',
      tumorInfo: '',
      medicalHistory: '',
      complications: [],
    },
    // 娌荤枟缁撳眬
    outcome: {
      status: '娌荤枟涓?,
    },
  }
}

/**
 * 瑙ｆ瀽棰戞瀛楃涓蹭负姣忓懆娆℃暟
 */
function parseFrequency(frequency: string): number {
  if (frequency.includes('3娆?鍛?)) return 3
  if (frequency.includes('2娆?鍛?)) return 2
  if (frequency.includes('4娆?鍛?)) return 4
  if (frequency.includes('1娆?鍛?)) return 1
  return 3  // 榛樿鍊?}

