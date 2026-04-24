/**
 * 鏈嶅姟灞傜粺涓€瀵煎嚭
 *
 * 鎸夋ā鍧楃粍缁囷細
 * - api: 鍩虹 GraphQL 瀹㈡埛绔? * - patient: 鎮ｈ€呯浉鍏虫湇鍔? * - schedule: 鎺掔彮鐩稿叧鏈嶅姟
 * - treatment: 娌荤枟鐩稿叧鏈嶅姟
 * - order: 鍖诲槺鐩稿叧鏈嶅姟
 * - vitals: 浣撳緛鐩戞祴鏈嶅姟
 * - examination: 妫€楠屾鏌ユ湇鍔? * - equipment: 璁惧绠＄悊鏈嶅姟
 */

// ============ 绫诲瀷瀵煎嚭 ============
export type {
  // 閫氱敤绫诲瀷
  PaginatedItem,
  PaginatedResponse,
  QueryParams,
  EntityName,
  // 鎮ｈ€呯浉鍏?  Hospitalization,
  Infection,
  VascularAccess,
  CaseHistory,
  // 鎺掔彮鐩稿叧
  Shift,
  PatientShift,
  Bed,
  Ward,
  // 娌荤枟鐩稿叧
  Treatment,
  PatientPrescription,
  PatientPlan,
  // 鍖诲槺鐩稿叧
  OrderTPL,
  OrderTemplate,
  PatientOrder,
  PatientDayOrder,
  // 浣撳緛鐩稿叧
  BeforeSigns,
  DuringSigns,
  AfterSigns,
  // 妫€楠岀浉鍏?  Examination,
  ExaminationItem,
  // 璁惧鐩稿叧
  EquipmentInfo,
  EquipmentDisinfection,
  EquipmentMaintenanceRecord,
  EquipmentUsageLog,
  MachineInfo,
  MachineRunRecord,
} from './types/api'

// ============ API 鍩虹鍑芥暟 ============
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

// ============ REST API 瀹㈡埛绔?============
export { restApi, convertRestPatientToUI, convertRestPatientList, getErrorMessage } from './restClient'
export type { RestPatient, RestShift, RestPatientShift, RestTreatment, PaginationMeta } from './restClient'
export type { PaginatedResponse as RestPaginatedResponse } from './restClient'

// ============ 涓存椂 Mock 杈呭姪鍑芥暟 ============

// ============ 鎺掔彮鏈嶅姟 ============
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

// ============ 娌荤枟鏈嶅姟 ============
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

// ============ 鍖诲槺鏈嶅姟 ============
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

// ============ 浣撳緛鏈嶅姟 ============
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

// ============ 妫€楠屾湇鍔?============
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

// ============ 璁惧鏈嶅姟 ============
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

// ============ 瑙掕壊鏈嶅姟 ============
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
