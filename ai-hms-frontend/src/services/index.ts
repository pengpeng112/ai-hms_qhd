/**
 * 服务层统一导出
 *
 * 按模块组织：
 * - api: 基础 GraphQL 客户端（HDIS 兼容层）
 * - schedule/treatment/order/vitals/examination/equipment: HDIS GraphQL 服务
 * - restClient: 历史兼容 REST 客户端（禁止新增业务方法）
 * - *Api.ts: 新增 REST 接口优先在此实现
 */

// ============ 类型导出 ============
export type {
  // 通用类型
  PaginatedItem,
  PaginatedResponse,
  QueryParams,
  EntityName,
  // 患者相关
  Hospitalization,
  Infection,
  VascularAccess,
  CaseHistory,
  // 治疗相关
  Treatment,
  PatientPrescription,
  PatientPlan,
  // 医嘱相关
  OrderTPL,
  OrderTemplate,
  PatientOrder,
  PatientDayOrder,
  // 体征相关
  BeforeSigns,
  DuringSigns,
  AfterSigns,
  // 检验相关
  Examination,
  ExaminationItem,
  // 设备相关
  EquipmentInfo,
  EquipmentDisinfection,
  EquipmentMaintenanceRecord,
  EquipmentUsageLog,
  MachineInfo,
  MachineRunRecord,
} from './types/api'

// ============ API 基础函数 ============
export {
  graphqlQuery,
  isApiConfigured,
  buildPaginatedQuery,
  buildFilteredQuery,
  buildSimpleQuery,
  fetchPaginatedData,
  fetchListData,
  fetchFilteredData,
  getTodayString,
  formatDateForApi,
} from './api'

// ============ REST API 客户端 ============
// REST API 客户端（历史兼容 facade，新代码请使用独立 *Api.ts 文件）
export { restApi, convertRestPatientToUI, convertRestPatientList, getErrorMessage } from './restClient'
export type {
  HealthEducationContentApi,
  PatientHealthEducationApi,
  RestPatient,
  RestTreatment,
  PaginationMeta,
  DashboardStats,
} from './restClient'
export type { PaginatedResponse as RestPaginatedResponse } from './restClient'

// ============ 独立 REST API 模块 ============
export { userApi } from './userApi'
export type { RestUser, CreateUserRequest, UpdateUserRequest, UserListParams } from './userApi'
export { roleManagementApi } from './roleManagementApi'
export type { AppRoleApi, PermissionNodeApi } from './roleManagementApi'
export { vascularEventApi } from './vascularEventApi'
export type { VascTimelineEntry, VascReminder } from './vascularEventApi'
export { adverseEventApi } from './adverseEventApi'
export type { AdverseEvent, AeRegisterBody, AeReportTarget, AeReportBody, AeStatusBody, AeAlertsResponse } from './adverseEventApi'
export { medicationApi } from './medicationApi'
export type { MedicationAdmin, MaRecordBody, MaSecondCheckBody, MedSuggestion, MedDefaultDose } from './medicationApi'
export { dryWeightApi } from './dryWeightApi'
export type { DryWeightAssessment, DwAssessBody, DwConfirmBody, DwConfirmResult, DwCurrentData } from './dryWeightApi'

// ============ 临时 Mock 辅助函数 ============

// ============ 治疗服务 ============
export {
  getTreatmentList,
  getTodayTreatments,
  getPatientTreatments,
  getOngoingTreatments,
  getTreatmentsByDateRange,
  getPatientPrescriptions,
  getTodayPrescriptions,
  getPrescriptionByTreatment,
  getPatientPlans,
  getCurrentPlan,
  getTodayTreatmentStats,
} from './treatment'

// ============ 医嘱服务 ============
export {
  getOrderTemplates,
  getOrderTemplatesByGroup,
  getActiveOrderTemplates,
  getPatientOrders,
  getPatientOrdersByType,
  getPatientOrdersByGroup,
  getPatientDayOrders,
  getDayOrdersByTreatmentTime,
  getPendingDayOrders,
  getExecutedDayOrders,
  getOrderStats,
  getPatientOrderOverview,
} from './order'

// ============ 体征服务 ============
export {
  getPatientBeforeSigns,
  getBeforeSignsByTreatment,
  getPatientDuringSigns,
  getDuringSignsByTreatment,
  getLatestDuringSigns,
  getPatientAfterSigns,
  getAfterSignsByTreatment,
  getTreatmentVitals,
  getPatientVitalTrends,
  calculateBPStats,
} from './vitals'
export type { TreatmentVitals, VitalTrendPoint, BPStats } from './vitals'

// ============ 检验服务 ============
export {
  getExaminationList,
  getPatientExaminations,
  getExaminationsByType,
  getLatestExaminations,
  getExaminationItems,
  getExaminationWithItems,
  getPatientAbnormalItems,
  COMMON_EXAM_TYPES,
  getDialysisExamOverview,
  getExamItemTrend,
} from './examination'
export type { AbnormalItem, ExamItemTrend } from './examination'

// ============ 设备服务 ============
export {
  getEquipmentList,
  getAllEquipments,
  getEquipmentById,
  getEquipmentDisinfections,
  getEquipmentMaintenanceRecords,
  getEquipmentUsageLogs,
  getRecentDisinfections,
  getEquipmentStats,
  getEquipmentOverview,
  getDashboardEquipmentData,
} from './equipment'
export type { EquipmentStats, EquipmentOverview, DashboardEquipmentData } from './equipment'

// ============ 角色服务 ============
export {
  getRoleUsers,
  getRoleUsersByGroup,
  getDefaultRouteByRole,
  getMenusByRole,
  saveSelectedRoleUser,
  getSelectedRoleUser,
  getSelectedRole,
  clearSelectedRole,
  hasSelectedRole,
  UserRole,
  UserRoleLabel,
  RoleGroups,
} from './role'
export type { RoleUser, RoleGroup } from './role'

// ============ 智能排班 v1.3 ============
export {
  getBoard,
  getWeek,
  generateSchedule,
  confirmPlan,
  confirmDay,
  cancelShift,
  absentShift,
  moveShift,
  insertTemporary,
  insertCrrt,
  listCrrt,
  machineOutage,
  setHoliday,
  planChange,
  makeup,
  listConflicts,
  resolveConflict,
  getDiffs,
  getQuality,
  listPatients,
  upsertPatient,
  listProfiles,
  getProfile,
  upsertProfile,
  listTemplates,
  rebuildTemplate,
  listIncompleteProfiles,
  dischargePatient,
  placePatient,
  setInfectionStatus,
  waiveInfection,
  seedDemo,
  upsertStaffDuty,
  listStaffDuty,
  deleteStaffDuty,
  resolveDuty,
  createOverride,
  getMyDuties,
  getCheckInStatus,
  checkIn,
} from './smartScheduleApi'
export type {
  WeekBoard,
  GenerateResult,
  QualityResult,
  ConflictItem,
  DiffItem,
  CrrtItem,
  IncompleteItem,
  StaffDuty,
  StaffDutyInput,
  ResolvedDuty,
} from './smartScheduleApi'

export {
  getQCDoctors,
  getQCDoctorDetail,
  QC_ITEM_LABELS,
  QC_ITEM_ORDER,
} from './qcApi'
export type {
  QCDoctorScore,
  QCPatientScore,
  QCPatientRow,
} from './qcApi'

// ============ 传染病筛查 A1 ============
export { infectiousApi } from './infectiousApi'
export type {
  InfectiousRecord,
  GateResult,
  InfectiousScreenBody,
  InfectiousDisposeBody,
  InfectiousHistoryResponse,
  InfectiousAlertsResponse,
} from './infectiousApi'

// ============ ACTRS 胸片分析 ============
export { actrApi } from './actrApi'
export type { ActrStatus, PatientACTR, AdoptActrRequest } from './actrApi'

// ============ 水质监测 A2 ============
export { waterQualityApi } from './waterQualityApi'
export type { WaterQualityRecord, ConductivityPoint, WaterQualityAlerts } from './waterQualityApi'

// ============ CNRDS 上报 A4 ==========
export { cnrdsApi } from './cnrdsApi'
export type { CnrdsReport, CnrdsContentRow } from './cnrdsApi'

// ============ 消毒监管 A3 ============
export { disinfectionApi } from './disinfectionApi'
export type { MachineDisinfStatus, DisinfAlerts, DisinfRecordBody, DisinfComplianceBody } from './disinfectionApi'

// ============ 护理文书 C1 ============
export {
  getNursingScales,
  recordNursingScale,
  recordNursingDoc,
  getNursingDocs,
  getNursingAlerts,
} from './nursingApi'
export type {
  NursingRiskLevel,
  NursingScaleOption,
  NursingScaleItem,
  NursingScaleBand,
  NursingScale,
  NursingDoc,
  RecordScaleRequest,
  RecordDocRequest,
} from './nursingApi'

// ============ 知情同意 C2 ============
export {
  getConsentTemplates,
  issueConsent,
  getConsents,
  getConsentAlerts,
  signConsent,
  revokeConsent,
} from './consentApi'
export type {
  ConsentStatus,
  ConsentTemplate,
  ConsentRecord,
  ConsentAlerts,
  IssueConsentRequest,
  SignConsentRequest,
} from './consentApi'

// ============ C4 收费归集 ============
export {
  buildCharge,
  listCharges,
  getCharge,
  addChargeLine,
  updateChargeLine,
  deleteChargeLine,
  confirmCharge,
  checkCharge,
  markExported,
  cancelCharge,
  pushCharge,
} from './billingApi'
export type {
  ChargeLine,
  ChargeRecord,
  BuildDraftBody,
  ListChargesParams,
  ListChargesResponse,
} from './billingApi'

// ============ C4 HIS 价表 ============
export {
  searchHisPrices,
  syncHisPrices,
  getHisPriceByCode,
} from './hisPriceApi'
export type {
  HisPriceItem,
  HisPriceSearchParams,
  HisPriceSearchResponse,
} from './hisPriceApi'
