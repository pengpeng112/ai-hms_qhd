import { apiClient, type ApiSuccessResponse } from './restClient'

export interface RestMonitoringLiveData {
  treatmentId: number
  patientId: number
  patientName: string
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
}

export async function getMonitoringLiveData(): Promise<RestMonitoringLiveData[]> {
  const response = await apiClient.get<ApiSuccessResponse<RestMonitoringLiveData[]>>('/api/v1/monitoring/live-data')
  if (!response.data.success) {
    throw new Error('Failed to get monitoring live data')
  }
  return response.data.data
}
