import { apiClient } from './restClient'

export interface AdverseEvent {
  id: string
  tenantId: number
  patientId: number
  treatmentId?: number
  eventType: string
  severity: 'mild' | 'moderate' | 'severe'
  occurredAt: string
  description?: string
  handling?: string
  outcome?: string
  reporterId?: string
  reportedTo?: string
  reportedAt?: string
  within6h?: boolean
  status: 'registered' | 'reported' | 'acknowledged' | 'processing' | 'closed'
  cqiLinked: boolean
  createdAt: string
  updatedAt: string
}

export interface AeRegisterBody {
  patientId: number
  treatmentId?: number
  eventType: string
  severity: string
  occurredAt: string
  description?: string
  handling?: string
  outcome?: string
  reporterId?: string
}

export interface AeReportTarget {
  role: string
  userId: string
}

export interface AeReportBody {
  reportedTo: AeReportTarget[]
}

export interface AeStatusBody {
  status: string
  cqiLinked?: boolean
}

export interface AeAlertsResponse {
  severeUnreported: AdverseEvent[]
  severeOverdue: AdverseEvent[]
  pending: AdverseEvent[]
}

export const adverseEventApi = {
  async register(body: AeRegisterBody): Promise<AdverseEvent> {
    const res = await apiClient.post('/api/v1/adverse-events', body)
    return res.data?.data
  },

  async list(params?: { severity?: string; status?: string; patientId?: string }): Promise<AdverseEvent[]> {
    const res = await apiClient.get('/api/v1/adverse-events', { params })
    return res.data?.data ?? []
  },

  async alerts(): Promise<AeAlertsResponse> {
    const res = await apiClient.get('/api/v1/adverse-events/alerts')
    return res.data?.data ?? { severeUnreported: [], severeOverdue: [], pending: [] }
  },

  async get(id: string): Promise<AdverseEvent> {
    const res = await apiClient.get(`/api/v1/adverse-events/${id}`)
    return res.data?.data
  },

  async report(id: string, body: AeReportBody): Promise<AdverseEvent> {
    const res = await apiClient.post(`/api/v1/adverse-events/${id}/report`, body)
    return res.data?.data
  },

  async updateStatus(id: string, body: AeStatusBody): Promise<AdverseEvent> {
    const res = await apiClient.post(`/api/v1/adverse-events/${id}/status`, body)
    return res.data?.data
  },
}
