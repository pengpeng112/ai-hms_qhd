import { apiClient, type ApiSuccessResponse } from './restClient'

export interface RestMonitoringAlert {
  metric: string
  level: string
  value: number
}

export interface RestMonitoringIdhRisk {
  available: boolean
  probability: number
  level: string
}

export interface RestMonitoringLiveData {
  treatmentId: number
  patientId: number
  patientName: string
  age: number
  dialysisNo: string
  bedId: number
  bedName: string
  wardId: number
  wardName: string
  status: string
  startTime: string
  estimatedDuration: number
  dryWeight: number
  dialysisMode: string
  sbp: number
  dbp: number
  heartRate: number
  respiration: number
  spO2: number
  bf: number
  tmp: number
  ufVolume: number
  ufGoal: number
  conductivity: number
  machineTmp: number
  arterialPressure: number
  venousPressure: number
  // 阈值引擎 + AI③ 产出
  accessType: string
  alarmLevel: string
  alerts: RestMonitoringAlert[] | null
  idhRisk: RestMonitoringIdhRisk
  // 卡面整场曲线（决①）：整场体征序列，每点带 SBP/DBP/MAP/HR，kind: actual | predicted。
  vitalsSeries: RestVitalSample[] | null
  // 实时钠清除比完成率（RNa 联动）。available=false 时显示"待数据"。
  rnaCompletion: RestRNaCompletion
  // 上机前双人核对状态（软门禁提醒）。
  firstChecked: boolean
  secondChecked: boolean
  doubleChecked: boolean
}

export interface RestRNaCompletion {
  available: boolean
  percent: number
  targetRNa: number
  mTarget: number
  mRealized: number
  cPre: number
  cPreAt: string
}

export interface RestVitalSample {
  t: string
  sbp: number
  dbp: number
  map: number
  hr: number
  kind: string
}

export async function getMonitoringLiveData(): Promise<RestMonitoringLiveData[]> {
  const response = await apiClient.get<ApiSuccessResponse<RestMonitoringLiveData[]>>('/api/v1/monitoring/live-data')
  if (!response.data.success) {
    throw new Error('Failed to get monitoring live data')
  }
  return response.data.data
}

// 整场趋势（决①）。每点带 kind: actual（实线）| predicted（虚线，本期后端留空）。
export interface RestTrendPoint {
  t: string
  v: number
  kind: string
}

export interface RestTreatmentTrend {
  treatmentId: number
  start: string
  now: string
  plannedEnd: string
  series: Record<string, RestTrendPoint[]>
}

export async function getTreatmentTrend(treatmentId: number): Promise<RestTreatmentTrend> {
  const response = await apiClient.get<ApiSuccessResponse<RestTreatmentTrend>>(`/api/v1/monitoring/treatments/${treatmentId}/trend`)
  if (!response.data.success) {
    throw new Error('Failed to get treatment trend')
  }
  return response.data.data
}
