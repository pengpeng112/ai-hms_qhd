import { apiClient } from '@services/restClient'
import type { PrescriptionContextData } from './PatientContextPanel'

/** 拉取处方开单聚合数据（体重/检验/上次治疗/钠清除比/血压心率）。 */
export async function fetchPatientContext(patientId: string): Promise<PrescriptionContextData> {
  const res = await apiClient.get<{ data: PrescriptionContextData }>(
    `/api/v1/patients/${patientId}/prescriptions/context`
  )
  return res.data.data
}

/** 从聚合数据中取出 LIS 最近一次血钠（SERUM_NA）的数值，无则 undefined。 */
export function extractLatestSerumNa(data: PrescriptionContextData | null | undefined): number | undefined {
  const na = data?.labs.find((l) => l.conceptId === 'SERUM_NA')
  return na?.value ?? undefined
}
