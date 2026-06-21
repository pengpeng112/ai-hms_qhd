import { apiClient } from './restClient'

export interface ActrStatus {
  enabled: boolean
  configured: boolean
}

export interface PatientACTR {
  id: string
  tenantId: number
  patientId: string
  dialysisNo: string
  actrsXrayId: number
  analysisDate?: string
  ctr?: number
  actr?: number
  actr1?: number
  actr2?: number
  actrNorm?: number
  heartWidth?: number
  lungWidth?: number
  tiltAngle?: number
  qcPass: number
  qcPaAp?: string
  qcWarnings?: string
  modelVersion?: string
  source?: string
  imagePath?: string
  overlayPath?: string
  maskPath?: string
  doctorCorrection?: number
  correctedBy?: string
  correctedAt?: string
  adoptedBy?: string
  adoptedAt?: string
  adoptedPrescriptionId?: string
  adoptedDryWeight?: number
  adoptedUfQuantity?: number
  notes?: string
  syncedAt?: string
  createdAt: string
  updatedAt: string
}

export interface AdoptActrRequest {
  prescriptionId: string
  actrRecordId: string
  dryWeight?: number
  ufQuantity?: number
}

export const actrApi = {
  status: (): Promise<ActrStatus> =>
    apiClient.get('/api/v1/actr/status').then((r) => r.data?.data as ActrStatus),

  history: (patientId: string | number): Promise<PatientACTR[]> =>
    apiClient.get(`/api/v1/patients/${patientId}/actr`).then((r) => r.data?.data ?? []),

  analyze: (patientId: string | number, file: File): Promise<PatientACTR> => {
    const fd = new FormData()
    fd.append('file', file)
    return apiClient.post(`/api/v1/patients/${patientId}/actr/analyze`, fd).then((r) => r.data?.data as PatientACTR)
  },

  adopt: (patientId: string | number, body: AdoptActrRequest): Promise<{ ok: boolean }> =>
    apiClient.post(`/api/v1/patients/${patientId}/actr/adopt`, body).then((r) => r.data?.data as { ok: boolean }),

  correct: (patientId: string | number, recordId: string, body: { value: number; notes?: string }): Promise<PatientACTR> =>
    apiClient.patch(`/api/v1/patients/${patientId}/actr/${recordId}/correction`, body).then((r) => r.data?.data as PatientACTR),
}
