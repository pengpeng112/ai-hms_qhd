/**
 * 排班模板类型定义
 */

export interface ScheduleTemplate {
  id: number
  name: string
  cycleWeeks: 2 | 4
  wardScope?: string
  isDefault: boolean
  isEnabled: boolean
  status: number
  note?: string
  creatorId?: number
  createdAt?: string
  updatedAt?: string
}

export interface ScheduleTemplateItem {
  id: number
  templateId: number
  weekIndex: number
  weekday: number
  shiftId: number
  wardId: number
  bedId: number
  patientId: number
  patientName?: string
  patientPlanId?: number
  dialysisMode?: string
  shiftTiming: 10 | 20
  status: number
  note?: string
}

export interface CreateTemplateRequest {
  name: string
  cycleWeeks: 2 | 4
  wardScope?: string
  note?: string
}

export interface ApplyTemplateRequest {
  startDate: string
  endDate: string
  applyMode: 'only_empty' | 'skip_existing' | 'overwrite_unconfirmed'
  wardIds?: number[]
}

export interface ScheduleValidationIssue {
  level: 'error' | 'warning' | 'info'
  type:
    | 'bed_conflict'
    | 'patient_conflict'
    | 'frequency_missing'
    | 'infection_mismatch'
    | 'device_mismatch'
    | 'disabled_ward'
    | 'disabled_bed'
    | 'locked_date'
  message: string
  patientId?: number
  patientName?: string
  date?: string
  shiftId?: number
  bedId?: number
  wardId?: number
}

export interface ApplyTemplatePreview {
  createdCount: number
  skippedCount: number
  conflictCount: number
  issues: ScheduleValidationIssue[]
}

export interface TemplateApplyRecord {
  id: number
  templateId: number
  templateName?: string
  startDate: string
  endDate: string
  applyMode: string
  createdCount: number
  skippedCount: number
  conflictCount: number
  operatorId?: number
  operatorName?: string
  operateTime: string
}
