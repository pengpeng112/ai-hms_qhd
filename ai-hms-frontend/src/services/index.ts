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
  // 排班相关
  Shift,
  PatientShift,
  Bed,
  Ward,
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
  RestShift,
  RestPatientShift,
  RestScheduleWard,
  RestScheduleBed,
  RestScheduleWeekShift,
  RestSchedulePendingPatient,
  RestScheduleWeekResponse,
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

// ============ 临时 Mock 辅助函数 ============

// ============ 排班服务 ============
export {
  getShiftList,
  getActiveShifts,
  getPatientShiftList,
  getTodaySchedule,
  getScheduleByDate,
  getScheduleByShift,
  getPatientSchedule,
  getBedList,
  getBedsByWard,
  getAvailableBeds,
  getWardList,
  getActiveWards,
  getTodayScheduleOverview,
} from './schedule'

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
