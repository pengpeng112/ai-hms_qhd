import { apiClient, type ApiSuccessResponse } from './restClient'

export interface MonthlySummaryData {
  id: number
  patientId: number
  year: number
  month: number
  content: Record<string, unknown>
  createdAt: string
  updatedAt: string
}

export async function getMonthlySummary(patientId: string, year: number, month: number): Promise<MonthlySummaryData> {
  const response = await apiClient.get<ApiSuccessResponse<MonthlySummaryData>>(
    `/api/v1/patients/${patientId}/monthly-summaries`,
    { params: { year, month } }
  )
  if (!response.data.success) {
    throw new Error('Failed to get monthly summary')
  }
  return response.data.data
}

export async function saveMonthlySummary(patientId: string, year: number, month: number, content: Record<string, unknown>): Promise<MonthlySummaryData> {
  const response = await apiClient.put<ApiSuccessResponse<MonthlySummaryData>>(
    `/api/v1/patients/${patientId}/monthly-summaries`,
    { content },
    { params: { year, month } }
  )
  if (!response.data.success) {
    throw new Error('Failed to save monthly summary')
  }
  return response.data.data
}
