/**
 * 检验检查服务
 */

import type {
  Examination,
  ExaminationItem,
  PaginatedResponse,
} from './types/api'
import { fetchPaginatedData, fetchFilteredData, graphqlQuery, buildFilteredQuery } from './api'

// ============ 字段定义 ============

const EXAM_FIELDS = [
  'Id', 'PatientId', 'ExamType', 'ExamDate', 'ReportDate', 'Status', 'DoctorName'
]

const EXAM_ITEM_FIELDS = [
  'Id', 'ExaminationId', 'ItemName', 'ItemCode', 'Result', 'Unit', 'RefRange', 'AbnormalFlag'
]

// ============ 常用检验类型 ============

export const COMMON_EXAM_TYPES = {
  BLOOD_ROUTINE: '血常规',
  LIVER_FUNCTION: '肝功能',
  KIDNEY_FUNCTION: '肾功能',
  ELECTROLYTE: '电解质',
  BLOOD_GAS: '血气分析',
  COAGULATION: '凝血功能',
  INFECTION: '感染指标',
} as const

// ============ 检验报告服务 ============

export async function getExaminationList(
  page: number = 1,
  pageSize: number = 50
): Promise<PaginatedResponse<Examination>> {
  return fetchPaginatedData<Examination>('Examination', EXAM_FIELDS, { page, pageSize })
}

export async function getPatientExaminations(
  patientId: number,
  page: number = 1,
  pageSize: number = 20
): Promise<PaginatedResponse<Examination>> {
  return fetchFilteredData<Examination>(
    'Examination',
    EXAM_FIELDS,
    { PatientId: patientId },
    page,
    pageSize
  )
}

export async function getExaminationsByType(
  patientId: number,
  examType: string
): Promise<Examination[]> {
  const result = await fetchFilteredData<Examination>(
    'Examination',
    EXAM_FIELDS,
    { PatientId: patientId, ExamType: examType },
    1,
    50
  )
  return result.data
}

export async function getLatestExaminations(
  patientId: number,
  limit: number = 5
): Promise<Examination[]> {
  const result = await getPatientExaminations(patientId, 1, limit)
  return result.data.sort((a, b) => {
    const dateA = a.ExamDate || ''
    const dateB = b.ExamDate || ''
    return dateB.localeCompare(dateA)
  })
}

// ============ 检验项目服务 ============

export async function getExaminationItems(
  examinationId: number
): Promise<ExaminationItem[]> {
  const result = await fetchFilteredData<ExaminationItem>(
    'ExaminationItem',
    EXAM_ITEM_FIELDS,
    { ExaminationId: examinationId },
    1,
    100
  )
  return result.data
}

export async function getExaminationWithItems(
  examinationId: number
): Promise<{ examination: Examination | null; items: ExaminationItem[] }> {
  const query = buildFilteredQuery('Examination', EXAM_FIELDS, { Id: examinationId }, 1, 1)
  const result = await graphqlQuery<{ Examination: Examination[] }>(query)
  const examination = result.Examination?.[0] || null

  const items = examination ? await getExaminationItems(examinationId) : []

  return { examination, items }
}

// ============ 异常项目 ============

export interface AbnormalItem extends ExaminationItem {
  examDate?: string
  examType?: string
}

export async function getPatientAbnormalItems(
  patientId: number,
  limit: number = 20
): Promise<AbnormalItem[]> {
  const exams = await getLatestExaminations(patientId, 10)
  const abnormalItems: AbnormalItem[] = []

  for (const exam of exams) {
    const items = await getExaminationItems(exam.Id)
    const abnormal = items.filter(item =>
      item.AbnormalFlag === 'H' ||
      item.AbnormalFlag === 'L' ||
      item.AbnormalFlag === '1' ||
      item.AbnormalFlag === '异常'
    )

    abnormal.forEach(item => {
      abnormalItems.push({
        ...item,
        examDate: exam.ExamDate,
        examType: exam.ExamType,
      })
    })

    if (abnormalItems.length >= limit) break
  }

  return abnormalItems.slice(0, limit)
}

// ============ 透析相关检验 ============

export interface DialysisExamOverview {
  latestKidney: Examination | null
  latestElectrolyte: Examination | null
  latestBloodRoutine: Examination | null
  abnormalCount: number
}

export async function getDialysisExamOverview(
  patientId: number
): Promise<DialysisExamOverview> {
  const [kidney, electrolyte, bloodRoutine] = await Promise.all([
    getExaminationsByType(patientId, COMMON_EXAM_TYPES.KIDNEY_FUNCTION),
    getExaminationsByType(patientId, COMMON_EXAM_TYPES.ELECTROLYTE),
    getExaminationsByType(patientId, COMMON_EXAM_TYPES.BLOOD_ROUTINE),
  ])

  const latestKidney = kidney.sort((a, b) =>
    (b.ExamDate || '').localeCompare(a.ExamDate || '')
  )[0] || null

  const latestElectrolyte = electrolyte.sort((a, b) =>
    (b.ExamDate || '').localeCompare(a.ExamDate || '')
  )[0] || null

  const latestBloodRoutine = bloodRoutine.sort((a, b) =>
    (b.ExamDate || '').localeCompare(a.ExamDate || '')
  )[0] || null

  const abnormalItems = await getPatientAbnormalItems(patientId, 50)

  return {
    latestKidney,
    latestElectrolyte,
    latestBloodRoutine,
    abnormalCount: abnormalItems.length,
  }
}

// ============ 检验趋势 ============

export interface ExamItemTrend {
  itemName: string
  data: Array<{ date: string; value: number; refRange?: string }>
}

export async function getExamItemTrend(
  patientId: number,
  itemCode: string,
  limit: number = 10
): Promise<ExamItemTrend> {
  const exams = await getLatestExaminations(patientId, limit * 2)
  const trendData: ExamItemTrend = { itemName: itemCode, data: [] }

  for (const exam of exams) {
    const items = await getExaminationItems(exam.Id)
    const targetItem = items.find(i => i.ItemCode === itemCode || i.ItemName === itemCode)

    if (targetItem && targetItem.Result) {
      const numValue = parseFloat(targetItem.Result)
      if (!isNaN(numValue)) {
        trendData.itemName = targetItem.ItemName || itemCode
        trendData.data.push({
          date: exam.ExamDate || '',
          value: numValue,
          refRange: targetItem.RefRange,
        })
      }
    }

    if (trendData.data.length >= limit) break
  }

  return trendData
}
