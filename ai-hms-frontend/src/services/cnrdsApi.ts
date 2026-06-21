import { apiClient } from './restClient'

export interface CnrdsContentRow {
  patientId: string
  name: string
  gender: string
  birthDate: string
  primaryDiagnosis: string
  comorbidity: string
  firstDialysisDate: string
  dialysisMode: string
  frequency: string
  vascularAccess: string
  hb: number | null
  ca: number | null
  p: number | null
  pth: number | null
  albumin: number | null
  ktv: number | null
  infMarkers: string
  outcomeType: string
  outcomeDate: string
  deathReason: string
}

export interface CnrdsReport {
  id: string
  tenantId: number
  period: string
  reportType: string
  eventType: string
  patientId: string
  content: string
  patientCount: number
  status: string
  exportRef: string
  reviewedBy: string
  submittedAt: string | null
  createdAt: string
  updatedAt: string
}

export const cnrdsApi = {
  monthly: (period: string): Promise<CnrdsReport> =>
    apiClient.post(`/api/v1/cnrds/monthly?period=${encodeURIComponent(period)}`).then((r) => r.data?.data as CnrdsReport),

  event: (patientId: string, eventType: string): Promise<CnrdsReport> =>
    apiClient.post('/api/v1/cnrds/event', { patientId, eventType }).then((r) => r.data?.data as CnrdsReport),

  list: (params?: { period?: string; reportType?: string; status?: string }): Promise<CnrdsReport[]> =>
    apiClient.get('/api/v1/cnrds', { params }).then((r) => r.data?.data ?? []),

  get: (id: string): Promise<CnrdsReport> =>
    apiClient.get(`/api/v1/cnrds/${id}`).then((r) => r.data?.data as CnrdsReport),

  exportCsv: (id: string): Promise<{ blob: Blob; filename: string; contentType: string }> =>
    apiClient.get(`/api/v1/cnrds/${id}/export`, { responseType: 'blob' }).then((r) => {
      const contentDisposition = r.headers['content-disposition'] as string | undefined
      const contentType = String(r.headers['content-type'] || '')
      let filename = 'cnrds_export.csv'
      if (contentDisposition) {
        const match = contentDisposition.match(/filename="?([^"]+)"?/)
        if (match) filename = match[1]
      }
      return { blob: r.data as Blob, filename, contentType }
    }),

  submit: (id: string, reviewedBy: string): Promise<{ ok: boolean }> =>
    apiClient.post(`/api/v1/cnrds/${id}/submit`, { reviewedBy }).then((r) => r.data?.data as { ok: boolean }),
}
