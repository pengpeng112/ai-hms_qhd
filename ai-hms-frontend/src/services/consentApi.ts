// 知情同意 API（C2 模块）。/api/v1/consents，走主系统 v1 信封。
import { apiClient } from './restClient'

export type ConsentStatus = 'pending' | 'signed' | 'expired' | 'revoked'

export interface ConsentTemplate {
  consentType: string
  name: string
  version: string
  timing: string
  validMonths: number
  enabled: boolean
}

export interface ConsentRecord {
  id: string
  patientId: string
  consentType: string
  templateVersion?: string
  signedBy?: string
  signRecordId?: string
  issuedBy?: string
  signedAt?: string
  expiresAt?: string
  status: ConsentStatus
  docRef?: string
  note?: string
  createdAt?: string
}

export interface ConsentAlerts {
  pending: ConsentRecord[]
  expired: ConsentRecord[]
}

export interface IssueConsentRequest {
  patientId: string
  consentType: string
  issuedBy?: string
  note?: string
}

export interface SignConsentRequest {
  signedBy: string
  docRef?: string
}

const base = '/api/v1'

export async function getConsentTemplates(): Promise<ConsentTemplate[]> {
  const res = await apiClient.get(`${base}/consents/templates`)
  return res.data?.data ?? []
}

export async function issueConsent(body: IssueConsentRequest): Promise<ConsentRecord> {
  const res = await apiClient.post(`${base}/consents`, body)
  return res.data?.data
}

export async function getConsents(params: { patientId?: string; consentType?: string; status?: string } = {}): Promise<ConsentRecord[]> {
  const res = await apiClient.get(`${base}/consents`, { params })
  return res.data?.data ?? []
}

export async function getConsentAlerts(): Promise<ConsentAlerts> {
  const res = await apiClient.get(`${base}/consents/alerts`)
  return res.data?.data ?? { pending: [], expired: [] }
}

export async function signConsent(id: string, body: SignConsentRequest): Promise<ConsentRecord> {
  const res = await apiClient.post(`${base}/consents/${id}/sign`, body)
  return res.data?.data
}

export async function revokeConsent(id: string): Promise<ConsentRecord> {
  const res = await apiClient.post(`${base}/consents/${id}/revoke`, {})
  return res.data?.data
}
