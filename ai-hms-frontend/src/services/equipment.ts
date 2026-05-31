import type {
  EquipmentInfo,
  EquipmentDisinfection,
  EquipmentMaintenanceRecord,
  EquipmentUsageLog,
  PaginatedResponse,
} from './types/api'
import { apiCache, cacheKey, CACHE_TTL } from '@/utils/cache'
import { apiClient, type ApiSuccessResponse } from './restClient'

interface DeviceApiItem {
  id: string
  tenantId: number
  name: string
  idNo?: string
  serialNo?: string
  brand?: string
  model?: string
  dialysisMethod?: string
  deviceType?: string
  manufacturer?: string
  bedNumber?: string
  bedId?: number | null
  wardId?: number | null
  wardName?: string
  status?: string
  lastDisinfectionTime?: string
  installDate?: string
  manufactureDate?: string
  lastMaintained?: string
  maintenance?: number | null
  maintenanceCycle?: string
  flux?: string
  notes?: string
  isDisabled?: boolean
  creatorId?: number
  createdAt?: string
  updatedAt?: string
}

interface DeviceListApiResponse {
  items: DeviceApiItem[]
  total: number
  page: number
  pageSize: number
  totalPage: number
}

interface DeviceDisinfectionApiItem {
  id: number
  tenantId?: number
  equipmentId: number
  disinfectUserId?: number
  disinfectWay?: string
  startTime?: string
  description?: string
  note?: string
}

interface DeviceDisinfectionApiResponse {
  items: DeviceDisinfectionApiItem[]
  total: number
  page: number
  pageSize: number
  totalPage: number
}

interface DeviceUsageLogApiItem {
  id: number
  tenantId?: number
  equipmentId: number
  useUserId?: number
  useStartTime?: string
  useDuration?: number
  note?: string
  creatorId?: number
  createTime?: string
  lastModifyTime?: string
}

interface DeviceUsageLogApiResponse {
  items: DeviceUsageLogApiItem[]
  total: number
  page: number
  pageSize: number
  totalPage: number
}

interface DeviceMaintenanceApiItem {
  id: number
  tenantId?: number
  equipmentId: number
  type?: string
  mode?: string
  operatorId?: number
  operateTime?: string
  description?: string
  note?: string
  creatorId?: number
  createTime?: string
  lastModifyTime?: string
}

interface DeviceMaintenanceApiResponse {
  items: DeviceMaintenanceApiItem[]
  total: number
  page: number
  pageSize: number
  totalPage: number
}

const toEquipmentInfo = (item: DeviceApiItem): EquipmentInfo => ({
  Id: Number(item.id),
  TenantId: item.tenantId,
  Name: item.name,
  IDNo: item.idNo || '',
  SerialNo: item.serialNo || '',
  Brand: item.brand || '',
  ModelNo: item.model || '',
  DialysisMethod: item.dialysisMethod || '',
  DeviceType: item.deviceType || '',
  Manufacturer: item.manufacturer || '',
  BedNumber: item.bedNumber || '',
  BedId: item.bedId ?? null,
  WardId: item.wardId ?? null,
  WardName: item.wardName || '',
  Status: item.status || 'normal',
  LastDisinfectionTime: item.lastDisinfectionTime,
  InstallDate: item.installDate,
  ManufactureDate: item.manufactureDate,
  LastMaintained: item.lastMaintained,
  Maintenance: item.maintenance ?? null,
  MaintenanceCycle: item.maintenanceCycle || '',
  Flux: item.flux || '',
  Notes: item.notes || '',
  IsDisabled: item.isDisabled ?? false,
  CreatorId: item.creatorId || 0,
  CreatedAt: item.createdAt,
  UpdatedAt: item.updatedAt,
})

const toEquipmentDisinfection = (item: DeviceDisinfectionApiItem): EquipmentDisinfection => ({
  Id: item.id,
  TenantId: item.tenantId,
  EquipmentId: item.equipmentId,
  DisinfectUserId: item.disinfectUserId,
  DisinfectWay: item.disinfectWay,
  StartTime: item.startTime,
  Description: item.description,
  Note: item.note,
})

const toEquipmentUsageLog = (item: DeviceUsageLogApiItem): EquipmentUsageLog => ({
  Id: item.id,
  TenantId: item.tenantId,
  EquipmentId: item.equipmentId,
  UseUserId: item.useUserId,
  UseStartTime: item.useStartTime,
  UseDuration: item.useDuration,
  Note: item.note,
  CreatorId: item.creatorId,
  CreateTime: item.createTime,
  LastModifyTime: item.lastModifyTime,
})

const toEquipmentMaintenanceRecord = (item: DeviceMaintenanceApiItem): EquipmentMaintenanceRecord => ({
  Id: item.id,
  TenantId: item.tenantId,
  EquipmentId: item.equipmentId,
  Type: item.type,
  Mode: item.mode,
  OperatorId: item.operatorId,
  OperateTime: item.operateTime,
  Description: item.description,
  Note: item.note,
  CreatorId: item.creatorId,
  CreateTime: item.createTime,
  LastModifyTime: item.lastModifyTime,
})

export async function getEquipmentList(
  page: number = 1,
  pageSize: number = 100,
  forceRefresh = false
): Promise<PaginatedResponse<EquipmentInfo>> {
  const key = cacheKey('equipment:list', page, pageSize)
  if (forceRefresh) {
    apiCache.invalidate('equipment')
  }
  return apiCache.withCache(key, async () => {
    const response = await apiClient.get<ApiSuccessResponse<DeviceListApiResponse>>('/api/v1/devices', {
      params: { page, pageSize }
    })
    const data = response.data.data
    return {
      data: data.items.map(toEquipmentInfo),
      total: data.total,
    }
  }, CACHE_TTL.DEFAULT)
}

export async function getAllEquipments(forceRefresh = false): Promise<EquipmentInfo[]> {
  const key = cacheKey('equipment:all')
  if (forceRefresh) {
    apiCache.invalidate('equipment')
  }
  return apiCache.withCache(key, async () => {
    const result = await getEquipmentList(1, 500, forceRefresh)
    return result.data
  }, CACHE_TTL.DEFAULT)
}

export async function getEquipmentById(id: number): Promise<EquipmentInfo | null> {
  const key = cacheKey('equipment:detail', id)
  return apiCache.withCache(key, async () => {
    const response = await apiClient.get<ApiSuccessResponse<DeviceApiItem>>(`/api/v1/devices/${id}`)
    return toEquipmentInfo(response.data.data)
  }, CACHE_TTL.EQUIPMENT_LIST)
}

export async function getEquipmentDisinfections(
  equipmentId: number,
  page: number = 1,
  pageSize: number = 20
): Promise<PaginatedResponse<EquipmentDisinfection>> {
  const response = await apiClient.get<ApiSuccessResponse<DeviceDisinfectionApiResponse>>(
    `/api/v1/devices/${equipmentId}/disinfections`,
    { params: { page, pageSize } }
  )
  const data = response.data.data
  return {
    data: data.items.map(toEquipmentDisinfection),
    total: data.total,
  }
}

export async function getRecentDisinfections(
  equipmentId: number,
  limit: number = 10
): Promise<EquipmentDisinfection[]> {
  const result = await getEquipmentDisinfections(equipmentId, 1, limit)
  return result.data
}

export async function getEquipmentUsageLogs(
  equipmentId: number,
  page: number = 1,
  pageSize: number = 20
): Promise<PaginatedResponse<EquipmentUsageLog>> {
  const response = await apiClient.get<ApiSuccessResponse<DeviceUsageLogApiResponse>>(
    `/api/v1/devices/${equipmentId}/usage-logs`,
    { params: { page, pageSize } }
  )
  const data = response.data.data
  return {
    data: data.items.map(toEquipmentUsageLog),
    total: data.total,
  }
}

export async function getEquipmentMaintenanceRecords(
  equipmentId: number,
  page: number = 1,
  pageSize: number = 20
): Promise<PaginatedResponse<EquipmentMaintenanceRecord>> {
  const response = await apiClient.get<ApiSuccessResponse<DeviceMaintenanceApiResponse>>(
    `/api/v1/devices/${equipmentId}/maintenance-records`,
    { params: { page, pageSize } }
  )
  const data = response.data.data
  return {
    data: data.items.map(toEquipmentMaintenanceRecord),
    total: data.total,
  }
}

export interface EquipmentStats {
  total: number
  byBrand: Record<string, number>
  byDialysisMethod: Record<string, number>
}

export async function getEquipmentStats(forceRefresh = false): Promise<EquipmentStats> {
  const result = await getEquipmentList(1, 500, forceRefresh)
  const equipments = result.data

  const stats: EquipmentStats = {
    total: equipments.length,
    byBrand: {},
    byDialysisMethod: {},
  }

  equipments.forEach((e) => {
    const brand = e.Brand || 'Unknown'
    stats.byBrand[brand] = (stats.byBrand[brand] || 0) + 1

    const method = e.DialysisMethod || 'Unknown'
    stats.byDialysisMethod[method] = (stats.byDialysisMethod[method] || 0) + 1
  })

  return stats
}

export interface EquipmentOverview {
  equipment: EquipmentInfo
  recentDisinfections: EquipmentDisinfection[]
  recentUsageLogs: EquipmentUsageLog[]
  recentMaintenanceRecords: EquipmentMaintenanceRecord[]
}

export async function getEquipmentOverview(
  equipmentId: number
): Promise<EquipmentOverview | null> {
  const equipment = await getEquipmentById(equipmentId)
  if (!equipment) return null

  const [disinfections, usageLogs, maintenanceRecords] = await Promise.all([
    getRecentDisinfections(equipmentId, 5),
    getEquipmentUsageLogs(equipmentId, 1, 5).then(r => r.data).catch(() => []),
    getEquipmentMaintenanceRecords(equipmentId, 1, 5).then(r => r.data).catch(() => []),
  ])
  return {
    equipment,
    recentDisinfections: disinfections,
    recentUsageLogs: usageLogs,
    recentMaintenanceRecords: maintenanceRecords,
  }
}

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
