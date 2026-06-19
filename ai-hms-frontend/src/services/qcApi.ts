// 医疗质控赋分 API（⑤ v1b）。/api/v1/qc，走主系统 v1 信封。
import { apiClient } from './restClient'

export interface QCDoctorScore {
  doctorId: string
  doctorName?: string
  patientCount: number
  quantityScore: number
  qualityScore: number
  totalScore: number
  onTargetRate: Record<string, number>
}

export interface QCPatientScore {
  base: number
  items: Record<string, number>
  onTarget: Record<string, boolean>
  quality: number
  total: number
}

export interface QCPatientRow {
  patientId: string
  patientName: string
  score: QCPatientScore
}

export async function getQCDoctors(month: string): Promise<QCDoctorScore[]> {
  const res = await apiClient.get('/api/v1/qc/doctors', { params: { month } })
  return res.data?.data ?? []
}

export async function getQCDoctorDetail(doctorId: string, month: string): Promise<{ doctor: QCDoctorScore; patients: QCPatientRow[] }> {
  const res = await apiClient.get(`/api/v1/qc/doctor/${doctorId}`, { params: { month } })
  return res.data?.data ?? { doctor: null, patients: [] }
}

export const QC_ITEM_LABELS: Record<string, string> = {
  bloodPressure: '血压',
  heartRate: '心率',
  CTR: 'CTR',
  dialysisAdequacy: '充分性',
  fluidControl: '控水',
  anemia: '贫血Hb',
  nutrition: '营养Alb',
  calcium: '血钙',
  phosphorus: '血磷',
  PTH: 'PTH',
}
export const QC_ITEM_ORDER = ['bloodPressure', 'heartRate', 'CTR', 'dialysisAdequacy', 'fluidControl', 'anemia', 'nutrition', 'calcium', 'phosphorus', 'PTH']
// 数据源尚未接入的指标（不计入达标率/质量分，前端标注"待接"）。CTR 待接 ACTRS/胸片。
export const QC_NOT_CONNECTED = new Set<string>(['CTR'])
