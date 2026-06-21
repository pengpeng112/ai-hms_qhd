// 机器消毒 API（A3 消毒模块）。/api/v1/disinfection，走主系统 v1 信封。
import { apiClient } from './restClient'

export interface MachineDisinfStatus {
  deviceId: number
  state: 'OK' | 'WARN' | 'BLOCKED_RESIDUAL'
  lastTerminal?: string
  lastDecalc?: string
  lastHeat?: string
  terminalToday: boolean
  decalcOverdue: boolean
  heatLag: boolean
  residualFail: boolean
  reasons: string[]
}

export interface DisinfAlerts {
  blocked: MachineDisinfStatus[]
  warn: MachineDisinfStatus[]
}

export interface DisinfRecordBody {
  deviceId: number
  disinfectType: string
  disinfectant?: string
  concentration?: string
  operatorId: number
  startTime: string
  endTime: string
  treatmentId?: number
  residualCheck?: string
  result?: string
  docRef?: string
  source?: string
}

export interface DisinfComplianceBody {
  concentration?: string
  residualCheck?: string
  result?: string
  docRef?: string
}

const base = '/api/v1'

export const disinfectionApi = {
  /** 提交消毒记录 */
  async record(body: DisinfRecordBody) {
    const res = await apiClient.post(`${base}/disinfection/record`, body)
    return res.data?.data as { disinfectionId: number; complianceId: number }
  },

  /** 补录合规明细 */
  async saveCompliance(id: number, body: DisinfComplianceBody) {
    const res = await apiClient.post(`${base}/disinfection/${id}/compliance`, body)
    return res.data?.data
  },

  /** 批量查询机器消毒状态 */
  async machines(deviceIds: number[]) {
    const res = await apiClient.get(`${base}/disinfection/machines`, {
      params: { deviceIds: deviceIds.join(',') },
    })
    return (res.data?.data ?? []) as MachineDisinfStatus[]
  },

  /** 消毒预警（blocked / warn） */
  async alerts(deviceIds: number[]) {
    const res = await apiClient.get(`${base}/disinfection/alerts`, {
      params: { deviceIds: deviceIds.join(',') },
    })
    return (res.data?.data ?? { blocked: [], warn: [] }) as DisinfAlerts
  },

  /** 消毒统计（热消毒/治疗数） */
  async stats(deviceIds: number[]) {
    const res = await apiClient.get(`${base}/disinfection/stats`, {
      params: { deviceIds: deviceIds.join(',') },
    })
    return (res.data?.data ?? { heatToday: 0, treatmentToday: 0 }) as {
      heatToday: number
      treatmentToday: number
    }
  },
}
