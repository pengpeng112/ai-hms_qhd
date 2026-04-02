/**
 * 临时 Mock 辅助函数
 * 等待页面对接 API 后删除
 */

import { MOCK_PATIENTS } from '@/constants'
import type { Patient } from '@/types/original'

// 临时的 getPatientList 函数，返回 Mock 数据
// eslint-disable-next-line @typescript-eslint/no-unused-vars
export async function getPatientList(_page: number, _pageSize: number) {
  return {
    data: MOCK_PATIENTS,
    total: MOCK_PATIENTS.length
  }
}

// 临时的 getPatientById 函数，返回 Mock 数据
// eslint-disable-next-line @typescript-eslint/no-unused-vars
export async function getPatientById(_id: number) {
  return MOCK_PATIENTS[0]
}

// 临时的 convertAPIPatientList 函数
export function convertAPIPatientList(patients: Patient[]) {
  return patients
}

// 临时的 convertAPIPatientToFullUI 函数
export function convertAPIPatientToFullUI(patient: Patient) {
  return patient
}
