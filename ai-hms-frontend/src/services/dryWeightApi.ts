import { apiClient } from './restClient'

export interface DryWeightAssessment {
  id: string
  tenantId: number
  patientId: number
  assessType: 'daily' | 'cycle'
  phase: 'induction' | 'maintenance'
  sbp?: number
  dbp?: number
  heartRate?: number
  edema: boolean
  palpitation: boolean
  heartFailure: boolean
  cramp: boolean
  ctr?: number
  actr?: number
  biaOh?: number
  biaTbw?: number
  biaEcw?: number
  postWeight?: number
  targetWeight?: number
  decision?: string
  adjustKg?: number
  rnaSetting?: number
  mainMet: boolean
  failedReasons?: string
  assessorId?: string
  assessorName?: string
  createdAt: string
}

export interface DwAssessBody {
  assessType: string
  phase: string
  sbp?: number
  dbp?: number
  heartRate?: number
  edema?: boolean
  palpitation?: boolean
  heartFailure?: boolean
  cramp?: boolean
  ctr?: number
  actr?: number
  biaOh?: number
  biaTbw?: number
  biaEcw?: number
  postWeight?: number
  targetWeight?: number
  decision?: string
  adjustKg?: number
  rnaSetting?: number
}

export interface DwConfirmBody {
  dryWeight: number
  phase: string
  actr?: number
  ctr?: number
}

export interface DwConfirmResult {
  id: string
  tenantId: number
  patientId: number
  dryWeight: number
  standardActr?: number
  standardCtr?: number
  phase: string
  confirmedBy?: string
  confirmedName?: string
  confirmedAt: string
  legacyPlanUpdated: boolean
}

export interface DwCurrentData {
  dryWeight?: number
  standardActr?: number
  standardCtr?: number
  phase: string
  suggestedRNa: number
  confirmedAt?: string
}

export const dryWeightApi = {
  async assess(patientId: number | string, body: DwAssessBody): Promise<DryWeightAssessment> {
    const res = await apiClient.post(`/api/v1/patients/${patientId}/dry-weight-assessments`, body)
    return res.data?.data
  },

  async listAssessments(patientId: number | string): Promise<DryWeightAssessment[]> {
    const res = await apiClient.get(`/api/v1/patients/${patientId}/dry-weight-assessments`)
    return res.data?.data ?? []
  },

  async confirm(patientId: number | string, body: DwConfirmBody): Promise<DwConfirmResult> {
    const res = await apiClient.post(`/api/v1/patients/${patientId}/dry-weight/confirm`, body)
    return res.data?.data
  },

  async current(patientId: number | string): Promise<DwCurrentData> {
    const res = await apiClient.get(`/api/v1/patients/${patientId}/dry-weight`)
    return res.data?.data ?? { phase: 'induction', suggestedRNa: 1.05 }
  },
}
