// 血管通路全生命周期事件 API（B1 模块）
import { apiClient } from './restClient'

export interface VascTimelineEntry {
  accessId: number
  eventType: string
  eventDate?: string
  detail?: string
  note?: string
}

export interface VascReminder {
  accessId: number
  patientId: number
  kind: string
  message: string
}

const base = '/api/v1'

export const vascularEventApi = {
  recordEvent: (patientId: number, body: {
    accessId: number
    eventType: string
    eventDate: string
    detail?: string
    operatorId?: string
    note?: string
  }) =>
    apiClient.post(`${base}/patients/${patientId}/vascular-access-events`, body).then(r => r.data?.data),

  timeline: (patientId: number) =>
    apiClient.get(`${base}/patients/${patientId}/vascular-access-timeline`).then(r => (r.data?.data ?? []) as VascTimelineEntry[]),

  reminders: (patientId: number) =>
    apiClient.get(`${base}/patients/${patientId}/vascular-access-reminders`).then(r => (r.data?.data ?? []) as VascReminder[]),

  alerts: () =>
    apiClient.get(`${base}/vascular-access/alerts`).then(r => (r.data?.data ?? []) as VascReminder[]),
}
