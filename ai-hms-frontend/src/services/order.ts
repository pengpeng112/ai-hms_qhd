/**
 * 医嘱相关服务
 * 基于 HDIS API 文档 4.3.16-4.3.19
 */

import type {
  OrderTPL,
  PatientOrder,
  PatientDayOrder,
  PaginatedResponse,
} from './types/api'
import {
  fetchPaginatedData,
  fetchFilteredData,
} from './api'

// ============ 字段定义 ============

// 根据 HDIS API 文档 4.3.16
const ORDER_TPL_FIELDS = [
  'Id', 'TenantId', 'Name', 'OrderGroup', 'IsDisabled', 'Classification', 'DrugId', 'Content'
]

// 根据 HDIS API 文档 4.3.17
const PATIENT_ORDER_FIELDS = [
  'Id', 'TenantId', 'PatientId', 'OrderTPLId', 'OrderGroup', 'Type', 'DrugId', 'Classification'
]

// 根据 HDIS API 文档 4.3.18
const PATIENT_DAY_ORDER_FIELDS = [
  'Id', 'TenantId', 'PatientId', 'TreatmentTime', 'PatientOrderId', 'OrderGroup', 'Status', 'CaseStatus'
]

// ============ 医嘱模板服务 ============

export async function getOrderTemplates(
  page: number = 1,
  pageSize: number = 100
): Promise<PaginatedResponse<OrderTPL>> {
  return fetchPaginatedData<OrderTPL>('OrderTPL', ORDER_TPL_FIELDS, { page, pageSize })
}

export async function getOrderTemplatesByGroup(
  orderGroup: string
): Promise<OrderTPL[]> {
  const result = await fetchFilteredData<OrderTPL>(
    'OrderTPL',
    ORDER_TPL_FIELDS,
    { OrderGroup: orderGroup },
    1,
    100
  )
  return result.data
}

export async function getActiveOrderTemplates(): Promise<OrderTPL[]> {
  const result = await getOrderTemplates(1, 200)
  return result.data.filter(t => t.IsDisabled !== '1' && t.IsDisabled !== 'true')
}

// ============ 患者医嘱服务 ============

export async function getPatientOrders(
  patientId: number,
  page: number = 1,
  pageSize: number = 50
): Promise<PaginatedResponse<PatientOrder>> {
  return fetchFilteredData<PatientOrder>(
    'PatientOrder',
    PATIENT_ORDER_FIELDS,
    { PatientId: patientId },
    page,
    pageSize
  )
}

export async function getPatientOrdersByType(
  patientId: number,
  orderType: string
): Promise<PatientOrder[]> {
  const result = await fetchFilteredData<PatientOrder>(
    'PatientOrder',
    PATIENT_ORDER_FIELDS,
    { PatientId: patientId, Type: orderType },
    1,
    100
  )
  return result.data
}

export async function getPatientOrdersByGroup(
  patientId: number,
  orderGroup: string
): Promise<PatientOrder[]> {
  const result = await fetchFilteredData<PatientOrder>(
    'PatientOrder',
    PATIENT_ORDER_FIELDS,
    { PatientId: patientId, OrderGroup: orderGroup },
    1,
    100
  )
  return result.data
}

// ============ 患者日医嘱服务 ============

export async function getPatientDayOrders(
  patientId: number,
  treatmentTime?: string
): Promise<PatientDayOrder[]> {
  const filters: Record<string, string | number> = { PatientId: patientId }
  if (treatmentTime) {
    filters.TreatmentTime = treatmentTime
  }
  const result = await fetchFilteredData<PatientDayOrder>(
    'PatientDayOrder',
    PATIENT_DAY_ORDER_FIELDS,
    filters,
    1,
    100
  )
  return result.data
}

export async function getDayOrdersByTreatmentTime(
  treatmentTime: string
): Promise<PatientDayOrder[]> {
  const result = await fetchFilteredData<PatientDayOrder>(
    'PatientDayOrder',
    PATIENT_DAY_ORDER_FIELDS,
    { TreatmentTime: treatmentTime },
    1,
    200
  )
  return result.data
}

export async function getPendingDayOrders(
  patientId: number
): Promise<PatientDayOrder[]> {
  const orders = await getPatientDayOrders(patientId)
  return orders.filter(o =>
    o.Status === '0' || o.Status === 'pending' || o.Status === '待执行'
  )
}

export async function getExecutedDayOrders(
  patientId: number
): Promise<PatientDayOrder[]> {
  const orders = await getPatientDayOrders(patientId)
  return orders.filter(o =>
    o.Status === '1' || o.Status === 'executed' || o.Status === '已执行'
  )
}

// ============ 统计 ============

export interface OrderStats {
  totalTemplates: number
  activeTemplates: number
}

export async function getOrderStats(): Promise<OrderStats> {
  const templates = await getOrderTemplates(1, 1)
  const active = await getActiveOrderTemplates()

  return {
    totalTemplates: templates.total,
    activeTemplates: active.length,
  }
}

export interface PatientOrderOverview {
  orders: PatientOrder[]
  byGroup: Record<string, PatientOrder[]>
}

export async function getPatientOrderOverview(
  patientId: number
): Promise<PatientOrderOverview> {
  const result = await getPatientOrders(patientId, 1, 100)
  const orders = result.data

  const byGroup: Record<string, PatientOrder[]> = {}
  orders.forEach(order => {
    const group = order.OrderGroup || 'Other'
    if (!byGroup[group]) byGroup[group] = []
    byGroup[group].push(order)
  })

  return { orders, byGroup }
}
