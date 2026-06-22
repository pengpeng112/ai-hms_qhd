import { apiClient } from './restClient'

export interface MedicationAdmin {
  id: string
  tenantId: number
  patientId: number
  orderId: number
  treatmentId: number
  drugName: string
  category?: string
  dose?: string
  route?: string
  timing?: string
  administeredBy: string
  administeredName?: string
  administeredAt: string
  secondCheckBy?: string
  secondCheckName?: string
  secondCheckAt?: string
  status: 'recorded' | 'verified'
  note?: string
  createdAt: string
  updatedAt: string
}

export interface MaRecordBody {
  patientId: number
  orderId: number
  treatmentId: number
  drugName: string
  category?: string
  dose?: string
  route?: string
  timing?: string
  note?: string
}

export interface MaSecondCheckBody {
  checkerId?: string
  checkerName?: string
}

export interface MedSuggestion {
  indicator: string
  label: string
  value?: number
  unit: string
  status: 'low' | 'high' | 'normal' | 'no_data'
  drug: string
  drugLabel: string
  direction?: string
  advice?: string
}

export interface MedDefaultDose {
  drug: string
  name: string
  route: string
  defaultDose: string
  frequency: string
  note: string
}

export const medicationApi = {
  async record(body: MaRecordBody): Promise<MedicationAdmin> {
    const res = await apiClient.post('/api/v1/medication-admins', body)
    return res.data?.data
  },

  async list(params?: { treatmentId?: number; patientId?: number; orderId?: number }): Promise<MedicationAdmin[]> {
    const res = await apiClient.get('/api/v1/medication-admins', { params })
    return res.data?.data ?? []
  },

  async secondCheck(id: string, body?: MaSecondCheckBody): Promise<MedicationAdmin> {
    const res = await apiClient.post(`/api/v1/medication-admins/${id}/second-check`, body || {})
    return res.data?.data
  },

  async suggestions(patientId: number | string): Promise<MedSuggestion[]> {
    const res = await apiClient.get(`/api/v1/patients/${patientId}/medication-suggestions`)
    return res.data?.data ?? []
  },

  async defaultDoses(): Promise<MedDefaultDose[]> {
    const res = await apiClient.get('/api/v1/medication-default-doses')
    return res.data?.data ?? []
  },
}
