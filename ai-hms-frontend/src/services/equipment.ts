/**
 * 设备管理服务
 * 基于 HDIS API 文档 4.3.29 设备档案信息
 */

import type {
  EquipmentInfo,
  EquipmentDisinfection,
  PaginatedResponse,
} from './types/api'
import {
  fetchPaginatedData,
  fetchListData,
  fetchFilteredData,
} from './api'
import { apiCache, cacheKey, CACHE_TTL } from '@/utils/cache'

// ============ 字段定义 ============

// 根据 HDIS API 文档 4.3.29
const EQUIPMENT_FIELDS = [
  'Id', 'TenantId', 'Name', 'IDNo', 'SerialNo', 'Brand', 'ModelNo', 'DialysisMethod'
]

// 根据 HDIS API 文档 4.3.30
const DISINFECTION_FIELDS = [
  'Id', 'TenantId', 'EquipmentId', 'DisinfectUserId', 'DisinfectWay',
  'StartTime', 'Description', 'Note'
]

// ============ 透析机信息服务 ============

export async function getEquipmentList(
  page: number = 1,
  pageSize: number = 100
): Promise<PaginatedResponse<EquipmentInfo>> {
  const key = cacheKey('equipment:list', page, pageSize)
  return apiCache.withCache(key, () =>
    fetchPaginatedData<EquipmentInfo>('EquipmentInfomation', EQUIPMENT_FIELDS, { page, pageSize }),
    CACHE_TTL.EQUIPMENT_LIST
  )
}

export async function getAllEquipments(): Promise<EquipmentInfo[]> {
  const key = cacheKey('equipment:all')
  return apiCache.withCache(key, () =>
    fetchListData<EquipmentInfo>('EquipmentInfomation', EQUIPMENT_FIELDS),
    CACHE_TTL.EQUIPMENT_LIST
  )
}

export async function getEquipmentById(id: number): Promise<EquipmentInfo | null> {
  const key = cacheKey('equipment:detail', id)
  return apiCache.withCache(key, async () => {
    const result = await fetchFilteredData<EquipmentInfo>(
      'EquipmentInfomation',
      EQUIPMENT_FIELDS,
      { Id: id },
      1,
      1
    )
    return result.data[0] || null
  }, CACHE_TTL.EQUIPMENT_LIST)
}

// ============ 设备消毒记录服务 ============

export async function getEquipmentDisinfections(
  equipmentId: number,
  page: number = 1,
  pageSize: number = 20
): Promise<PaginatedResponse<EquipmentDisinfection>> {
  return fetchFilteredData<EquipmentDisinfection>(
    'EquipmentDisinfection',
    DISINFECTION_FIELDS,
    { EquipmentId: equipmentId },
    page,
    pageSize
  )
}

export async function getRecentDisinfections(
  equipmentId: number,
  limit: number = 10
): Promise<EquipmentDisinfection[]> {
  const result = await getEquipmentDisinfections(equipmentId, 1, limit)
  return result.data
}

// ============ 设备统计 ============

export interface EquipmentStats {
  total: number
  byBrand: Record<string, number>
  byDialysisMethod: Record<string, number>
}

export async function getEquipmentStats(): Promise<EquipmentStats> {
  const result = await getEquipmentList(1, 500)
  const equipments = result.data

  const stats: EquipmentStats = {
    total: equipments.length,
    byBrand: {},
    byDialysisMethod: {},
  }

  equipments.forEach(e => {
    const brand = e.Brand || 'Unknown'
    stats.byBrand[brand] = (stats.byBrand[brand] || 0) + 1

    const method = e.DialysisMethod || 'Unknown'
    stats.byDialysisMethod[method] = (stats.byDialysisMethod[method] || 0) + 1
  })

  return stats
}

// ============ 综合查询 ============

export interface EquipmentOverview {
  equipment: EquipmentInfo
  recentDisinfections: EquipmentDisinfection[]
}

export async function getEquipmentOverview(
  equipmentId: number
): Promise<EquipmentOverview | null> {
  const equipment = await getEquipmentById(equipmentId)
  if (!equipment) return null

  const disinfections = await getRecentDisinfections(equipmentId, 5)

  return {
    equipment,
    recentDisinfections: disinfections,
  }
}

// ============ Dashboard 数据 ============

export interface DashboardEquipmentData {
  stats: EquipmentStats
  equipments: EquipmentInfo[]
}

export async function getDashboardEquipmentData(): Promise<DashboardEquipmentData> {
  const [stats, equipments] = await Promise.all([
    getEquipmentStats(),
    getAllEquipments(),
  ])

  return { stats, equipments }
}
