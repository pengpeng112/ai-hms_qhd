/**
 * 智能排班 v1.3 REST API 模块
 *
 * 封装 /api/v2 下全部排班接口，使用 restRequest 统一 helper。
 */
import { restGet, restPost } from './restRequest'

// ============ 类型定义 ============

export interface WeekBoard {
  weekStart: string
  weekEnd: string
  dates: string[]
  shifts: ShiftDTO[]
  wards: WardDTO[]
}

export interface ShiftDTO {
  id: number
  name: string
  code: string
  sort: number
}

export interface WardDTO {
  id: number
  name: string
  zoneType: string
  machines: MachineDTO[]
}

export interface MachineDTO {
  id: number
  code: string
  machineType: string
  positionIndex: number
  cells: Record<string, CellDTO>
}

export interface CellDTO {
  id: number
  shiftId: number
  patientId: number
  patientName: string
  dialysisMode: string
  status: number
  sourceType: number
  confirms: number
}

export interface GenerateRequest {
  startDate?: string
  weeks?: number
}

export interface GenerateResult {
  startDate: string
  weeks: number
  dialysisDays: number
  drafts: number
  conflicts: number
  parityAssigned: number
}

export interface ConfirmPlanRequest {
  weekStart?: string
  weeks?: number
}

export interface ConfirmDayRequest {
  date: string
  level: number
}

export interface MoveShiftRequest {
  machineId: number
  date?: string
  shiftId?: number
}

export interface TemporaryRequest {
  patientId: number
  wardId: number
  date?: string
  mode?: string
}

export interface CrrtRequest {
  patientId: number
  wardId: number
  machineId?: number
  startAt?: string
  endAt?: string
}

export interface OutageRequest {
  startDate: string
  endDate: string
  type?: number
  reason?: string
}

export interface HolidayRequest {
  date: string
  mode?: number
  openWardIds?: string
}

export interface PlanChangeRequest {
  changeType: string
  newValue: string
  effectiveDate?: string
}

export interface QualityResult {
  weekStart: string
  weeks: number
  patientsTotal: number
  patientsOnTarget: number
  onTargetRate: number
  capacitySlots: number
  usedSlots: number
  utilization: number
  patientsScheduled: number
  singleMachine: number
  stabilityRate: number
  openConflicts: number
  score: number
}

export interface ConflictItem {
  id: number
  patientId: number
  scheduleDate: string
  shiftId: number
  wardId: number
  conflictType: string
  severity: number
  detail: string
  status: number
}

export interface DiffItem {
  patientId: number
  patientName: string
  freqPattern: number
  expected: number
  scheduled: number
  diff: number
}

export interface CrrtItem {
  id: number
  patientId: number
  patientName: string
  machineId: number
  machineCode: string
  wardId: number
  startAt: string
  endAt: string
  status: number
}

export interface DischargeRequest {
  reason: string
}

export interface PlaceRequest {
  start?: string
  weeks?: number
}

export interface InfectionRequest {
  status: string
}

export interface IncompleteItem {
  patientId: number
  patientName: string
  missing: string[]
}

// ============ 排班看板 & 生成 ============

export function getBoard(date: string) {
  return restGet<WeekBoard>(`/api/v2/schedule/board?date=${encodeURIComponent(date)}`)
}

export function getWeek(date: string) {
  return restGet<{ weekStart: string; weekEnd: string; count: number; shifts: unknown[] }>(`/api/v2/schedule/week?date=${encodeURIComponent(date)}`)
}

export function generateSchedule(payload: GenerateRequest) {
  return restPost<GenerateResult>('/api/v2/schedule/generate', payload)
}

// ============ 确认 ============

export function confirmPlan(payload: ConfirmPlanRequest) {
  return restPost<{ confirmed: number }>('/api/v2/schedule/confirm-plan', payload)
}

export function confirmDay(payload: ConfirmDayRequest) {
  return restPost<{ confirmed: number }>('/api/v2/schedule/confirm-day', payload)
}

// ============ 操作 ============

export function cancelShift(id: number, reason?: string) {
  return restPost<{ ok: boolean }>(`/api/v2/shifts/${id}/cancel`, { reason })
}

export function absentShift(id: number, reason?: string) {
  return restPost<{ ok: boolean }>(`/api/v2/shifts/${id}/absent`, { reason })
}

export function moveShift(id: number, payload: MoveShiftRequest) {
  return restPost<{ ok: boolean }>(`/api/v2/shifts/${id}/move`, payload)
}

// ============ 治疗执行(上机/下机) ============

export function startTreatment(id: number) {
  return restPost<{ message: string }>(`/api/v2/shifts/${id}/start`, {})
}

export function completeTreatment(id: number) {
  return restPost<{ message: string }>(`/api/v2/shifts/${id}/complete`, {})
}

// ============ 临时透析 & CRRT ============

export function insertTemporary(payload: TemporaryRequest) {
  return restPost<{ ok: boolean; shift: unknown }>('/api/v2/schedule/temporary', payload)
}

export function insertCrrt(payload: CrrtRequest) {
  return restPost<{ ok: boolean; shift: unknown }>('/api/v2/schedule/crrt', payload)
}

export function listCrrt(date: string) {
  return restGet<{ count: number; items: CrrtItem[] }>(`/api/v2/schedule/crrt?date=${encodeURIComponent(date)}`)
}

// ============ 停机 & 假日 & 方案变更 ============

export function machineOutage(machineId: number, payload: OutageRequest) {
  return restPost<unknown>(`/api/v2/machines/${machineId}/outage`, payload)
}

export function setHoliday(payload: HolidayRequest) {
  return restPost<unknown>('/api/v2/schedule/holiday', payload)
}

export function planChange(patientId: number, payload: PlanChangeRequest) {
  return restPost<unknown>(`/api/v2/patients/${patientId}/plan-change`, payload)
}

// ============ 补透 ============

export function makeup(patientId: number, payload: { weekStart?: string; weeks?: number }) {
  return restPost<unknown>(`/api/v2/patients/${patientId}/makeup`, payload)
}

// ============ 冲突 & 差异 & 质量 ============

export function listConflicts(status?: number) {
  const qs = status !== undefined ? `?status=${status}` : ''
  return restGet<{ total: number; count: number; conflicts: ConflictItem[] }>(`/api/v2/conflicts${qs}`)
}

export function resolveConflict(id: number, action: string) {
  return restPost<{ ok: boolean }>(`/api/v2/conflicts/${id}/resolve`, { action })
}

export function getDiffs(date: string, weeks?: number) {
  const w = weeks ? `&weeks=${weeks}` : ''
  return restGet<{ weekStart: string; weeks: number; items: DiffItem[] }>(`/api/v2/schedule/diffs?date=${encodeURIComponent(date)}${w}`)
}

export function getQuality(date: string, weeks?: number) {
  const w = weeks ? `&weeks=${weeks}` : ''
  return restGet<QualityResult>(`/api/v2/schedule/quality?date=${encodeURIComponent(date)}${w}`)
}

// ============ 管理 ============

export function listPatients() {
  return restGet<{ items: unknown[] }>('/api/v2/admin/patients')
}

export function upsertPatient(payload: { id: number; name: string; gender?: string }) {
  return restPost<unknown>('/api/v2/admin/patients', payload)
}

export function listProfiles() {
  return restGet<{ items: unknown[] }>('/api/v2/admin/profiles')
}

export function getProfile(patientId: number) {
  return restGet<unknown>(`/api/v2/admin/profiles/${patientId}`)
}

export function upsertProfile(payload: unknown) {
  return restPost<unknown>('/api/v2/admin/profiles', payload)
}

export function listTemplates() {
  return restGet<{ items: unknown[] }>('/api/v2/admin/templates')
}

export function rebuildTemplate(name?: string) {
  return restPost<unknown>('/api/v2/admin/templates/rebuild', { name })
}

export function listIncompleteProfiles() {
  return restGet<{ items: IncompleteItem[] }>('/api/v2/admin/incomplete-profiles')
}

// ============ 生命周期 ============

export function dischargePatient(patientId: number, payload: DischargeRequest) {
  return restPost<unknown>(`/api/v2/patients/${patientId}/discharge`, payload)
}

export function placePatient(patientId: number, payload: PlaceRequest) {
  return restPost<unknown>(`/api/v2/patients/${patientId}/place`, payload)
}

export function setInfectionStatus(patientId: number, payload: InfectionRequest) {
  return restPost<unknown>(`/api/v2/patients/${patientId}/infection`, payload)
}

export function waiveInfection(patientId: number) {
  return restPost<unknown>(`/api/v2/patients/${patientId}/infection-waive`, {})
}

// ============ 演示 ============

export function seedDemo() {
  return restPost<{ seeded: string }>('/api/v2/admin/seed')
}
