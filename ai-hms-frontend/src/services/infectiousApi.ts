// 传染病筛查 API（A1 模块）。/api/v1 前缀，走主系统 v1 信封。
import { apiClient } from './restClient'

// ============ 类型定义 ============

export interface InfectiousRecord {
  id: string
  patientId: string
  screenDate?: string
  resultOverall: 'negative' | 'positive' | 'pending'
  positiveMarkers?: string
  nextDueDate?: string
  disposition?: string
  handledAt?: string
  zoneTag?: string
}

export interface GateResult {
  state: 'ALLOW_NORMAL' | 'REQUIRE_C_ZONE' | 'FROZEN' | 'C_ZONE_CRRT'
  reason?: string
}

export interface InfectiousScreenItem {
  item: string
  result: string
}

export interface InfectiousScreenBody {
  screenDate: string
  source: string
  items: InfectiousScreenItem[]
  note?: string
}

export interface InfectiousDisposeBody {
  disposition: 'c_zone_crrt' | 'transfer_out'
  role: 'doctor' | 'head_nurse'
  signerId: string
  signerName: string
}

export interface InfectiousHistoryResponse {
  records: InfectiousRecord[]
  gate: GateResult
}

export interface InfectiousAlertsResponse {
  positives: InfectiousRecord[]
  due: InfectiousRecord[]
}

// ============ API 方法 ============

export const infectiousApi = {
  /** 获取患者传染病筛查历史及分区门控 */
  async history(pid: string): Promise<InfectiousHistoryResponse> {
    const res = await apiClient.get(`/api/v1/patients/${pid}/infectious`)
    return res.data?.data ?? { records: [], gate: { state: 'ALLOW_NORMAL' } }
  },

  /** 提交筛查结果 */
  async screen(pid: string, body: InfectiousScreenBody): Promise<InfectiousRecord> {
    const res = await apiClient.post(`/api/v1/patients/${pid}/infectious/screen`, body)
    return res.data?.data
  },

  /** 处置阳性记录（分区/转出） */
  async dispose(pid: string, recordId: string, body: InfectiousDisposeBody): Promise<InfectiousRecord> {
    const res = await apiClient.post(`/api/v1/patients/${pid}/infectious/${recordId}/dispose`, body)
    return res.data?.data
  },

  /** 获取全局传染病预警（阳性 + 到期复查） */
  async alerts(): Promise<InfectiousAlertsResponse> {
    const res = await apiClient.get('/api/v1/infectious/alerts')
    return res.data?.data ?? { positives: [], due: [] }
  },
}
