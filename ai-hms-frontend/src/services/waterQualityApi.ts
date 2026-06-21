// 水质监测 API（A2 模块）。/api/v1/water-quality，走主系统 v1 信封。
import { apiClient } from './restClient'

export interface WaterQualityRecord {
  id: string
  testDate?: string
  testType: string
  samplePoint: string
  deviceId?: string
  value: number
  unit: string
  standardLimit?: string
  result: 'pass' | 'fail' | 'pending'
  nextDueDate?: string
  handledAt?: string
  action?: string
}

export interface ConductivityPoint {
  day: string
  value: number
  inRange: boolean
}

export interface WaterQualityAlerts {
  exceed: WaterQualityRecord[]
  due: WaterQualityRecord[]
}

export const waterQualityApi = {
  /** 查询水质监测记录列表 */
  async list(q?: { testType?: string; samplePoint?: string }): Promise<WaterQualityRecord[]> {
    const res = await apiClient.get('/api/v1/water-quality', { params: q })
    return res.data?.data ?? []
  },

  /** 新增一条水质检测记录 */
  async record(body: {
    testDate: string
    testType: string
    samplePoint: string
    deviceId?: string
    value: number
    unit: string
  }): Promise<WaterQualityRecord> {
    const res = await apiClient.post('/api/v1/water-quality/record', body)
    return res.data?.data ?? null
  },

  /** 获取电导率趋势数据 */
  async conductivity(days?: number): Promise<ConductivityPoint[]> {
    const res = await apiClient.get('/api/v1/water-quality/conductivity', { params: { days: days ?? 7 } })
    return res.data?.data ?? []
  },

  /** 获取水质告警（超标 + 到期） */
  async alerts(): Promise<WaterQualityAlerts> {
    const res = await apiClient.get('/api/v1/water-quality/alerts')
    return res.data?.data ?? { exceed: [], due: [] }
  },

  /** 处理一条水质告警记录 */
  async handle(
    id: string,
    body: {
      role: 'engineer' | 'head_nurse'
      signerId: string
      signerName: string
      action: string
    },
  ): Promise<WaterQualityRecord> {
    const res = await apiClient.post(`/api/v1/water-quality/${id}/handle`, body)
    return res.data?.data ?? null
  },
}
