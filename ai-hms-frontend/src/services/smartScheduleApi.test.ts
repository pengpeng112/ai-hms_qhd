import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('axios', () => {
  const mockInstance = {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
    interceptors: {
      request: { use: vi.fn(), eject: vi.fn(), clear: vi.fn() },
      response: { use: vi.fn(), eject: vi.fn(), clear: vi.fn() },
    },
  }
  return {
    default: {
      create: vi.fn(() => mockInstance),
    },
  }
})

import axios from 'axios'

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const mockInstance = (axios.create as any)() as {
  get: ReturnType<typeof vi.fn>
  post: ReturnType<typeof vi.fn>
  put: ReturnType<typeof vi.fn>
  delete: ReturnType<typeof vi.fn>
}

beforeEach(() => {
  vi.clearAllMocks()
})

function mockGetSuccess<T>(data: T) {
  mockInstance.get.mockResolvedValue({ data: { success: true, data } })
}

function mockPostSuccess<T>(data: T) {
  mockInstance.post.mockResolvedValue({ data: { success: true, data } })
}

import {
  getBoard, generateSchedule,
  confirmPlan, confirmDay,
  cancelShift, absentShift, moveShift,
  startTreatment, completeTreatment,
  insertTemporary, insertCrrt, listCrrt,
  machineOutage, setHoliday, planChange,
  makeup,
  listConflicts, resolveConflict, getDiffs, getQuality,
  listPatients, rebuildTemplate, listIncompleteProfiles,
  dischargePatient, placePatient, setInfectionStatus, waiveInfection,
  seedDemo,
  type GenerateRequest,
  type MoveShiftRequest,
  type TemporaryRequest,
  type CrrtRequest,
} from '@services/smartScheduleApi'

describe('smartScheduleApi — 看板 & 生成', () => {
  it('getBoard 应返回 WeekBoard', async () => {
    mockGetSuccess({ weekStart: '2026-06-08', weekEnd: '2026-06-14', dates: [], shifts: [], wards: [] })
    const result = await getBoard('2026-06-10')
    expect(result).toHaveProperty('weekStart')
  })

  it('generateSchedule 应发送 GenerateRequest', async () => {
    mockPostSuccess({ startDate: '', weeks: 2, dialysisDays: 3, drafts: 5, conflicts: 0, parityAssigned: 0 })
    const payload: GenerateRequest = { startDate: '2026-06-08', weeks: 2 }
    const result = await generateSchedule(payload)
    expect(result).toHaveProperty('drafts')
  })
})

describe('smartScheduleApi — 确认', () => {
  it('confirmPlan 应发送确认方案请求', async () => {
    mockPostSuccess({ confirmed: 10 })
    const result = await confirmPlan({ weekStart: '2026-06-08', weeks: 2 })
    expect(result.confirmed).toBe(10)
  })

  it('confirmDay 应发送日确认请求', async () => {
    mockPostSuccess({ confirmed: 5 })
    const result = await confirmDay({ date: '2026-06-10', level: 2 })
    expect(result.confirmed).toBe(5)
  })
})

describe('smartScheduleApi — 操作', () => {
  it('cancelShift 应发送取消请求', async () => {
    mockPostSuccess({ ok: true })
    const result = await cancelShift(42, '请假')
    expect(result.ok).toBe(true)
  })

  it('absentShift 应发送缺席请求', async () => {
    mockPostSuccess({ ok: true })
    await absentShift(42, '未到')
  })

  it('moveShift 应发送移床请求', async () => {
    mockPostSuccess({ ok: true })
    const payload: MoveShiftRequest = { machineId: 5, date: '2026-06-10', shiftId: 2 }
    const result = await moveShift(42, payload)
    expect(result.ok).toBe(true)
  })
})

describe('smartScheduleApi — 治疗执行', () => {
  it('startTreatment 应发送上机请求', async () => {
    mockPostSuccess({ message: '已上机' })
    const result = await startTreatment(42)
    expect(result.message).toBe('已上机')
  })

  it('completeTreatment 应发送下机请求', async () => {
    mockPostSuccess({ message: '已下机' })
    const result = await completeTreatment(42)
    expect(result.message).toBe('已下机')
  })
})

describe('smartScheduleApi — 临时透析 & CRRT', () => {
  it('insertTemporary 应发送临时透析请求', async () => {
    mockPostSuccess({ ok: true, shift: {} })
    const payload: TemporaryRequest = { patientId: 100, wardId: 3 }
    const result = await insertTemporary(payload)
    expect(result.ok).toBe(true)
  })

  it('insertCrrt 应发送 CRRT 请求', async () => {
    mockPostSuccess({ ok: true, shift: {} })
    const payload: CrrtRequest = { patientId: 100, wardId: 3, machineId: 10 }
    await insertCrrt(payload)
  })

  it('listCrrt 应返回 CRRT 列表', async () => {
    mockGetSuccess({ count: 2, items: [] })
    const result = await listCrrt('2026-06-10')
    expect(result.count).toBe(2)
  })
})

describe('smartScheduleApi — 停机 & 假日 & 方案变更', () => {
  it('machineOutage 应发送停机请求', async () => {
    mockPostSuccess({})
    await machineOutage(5, { startDate: '2026-06-10', endDate: '2026-06-12' })
  })

  it('setHoliday 应发送假日请求', async () => {
    mockPostSuccess({})
    await setHoliday({ date: '2026-06-10', mode: 10 })
  })

  it('planChange 应发送方案变更请求', async () => {
    mockPostSuccess({})
    await planChange(100, { changeType: 'frequency', newValue: '3' })
  })
})

describe('smartScheduleApi — 补透', () => {
  it('makeup 应发送补排请求', async () => {
    mockPostSuccess({})
    await makeup(100, { weekStart: '2026-06-08', weeks: 2 })
  })
})

describe('smartScheduleApi — 冲突 & 质量', () => {
  it('listConflicts 应返回冲突列表', async () => {
    mockGetSuccess({ total: 3, count: 3, conflicts: [] })
    const result = await listConflicts(0)
    expect(result.total).toBe(3)
  })

  it('resolveConflict 应发送解决请求', async () => {
    mockPostSuccess({ ok: true })
    await resolveConflict(1, 'accept')
  })

  it('getDiffs 应返回差异列表', async () => {
    mockGetSuccess({ weekStart: '', weeks: 2, items: [] })
    const result = await getDiffs('2026-06-10', 2)
    expect(result.weeks).toBe(2)
  })

  it('getQuality 应返回质量评分', async () => {
    mockGetSuccess({ weekStart: '', weeks: 2, patientsTotal: 100, patientsOnTarget: 95, onTargetRate: 0.95, capacitySlots: 200, usedSlots: 180, utilization: 0.9, patientsScheduled: 95, singleMachine: 10, stabilityRate: 0.92, openConflicts: 5, score: 85 })
    const result = await getQuality('2026-06-10')
    expect(result.score).toBe(85)
  })
})

describe('smartScheduleApi — 管理 API', () => {
  it('listPatients 应返回患者列表', async () => {
    mockGetSuccess({ items: [] })
    const result = await listPatients()
    expect(result).toHaveProperty('items')
  })

  it('rebuildTemplate 应发送重建请求', async () => {
    mockPostSuccess({})
    await rebuildTemplate('新模板')
  })

  it('listIncompleteProfiles 应返回不完整档案', async () => {
    mockGetSuccess({ items: [] })
    const result = await listIncompleteProfiles()
    expect(result).toHaveProperty('items')
  })
})

describe('smartScheduleApi — 生命周期', () => {
  it('dischargePatient 应发送出院请求', async () => {
    mockPostSuccess({})
    await dischargePatient(100, { reason: '康复' })
  })

  it('placePatient 应发送安置请求', async () => {
    mockPostSuccess({})
    await placePatient(100, { start: '2026-06-10', weeks: 2 })
  })

  it('setInfectionStatus 应设置感染状态', async () => {
    mockPostSuccess({})
    await setInfectionStatus(100, { status: 'positive' })
  })

  it('waiveInfection 应发送豁免请求', async () => {
    mockPostSuccess({})
    await waiveInfection(100)
  })
})

describe('smartScheduleApi — 演示', () => {
  it('seedDemo 应发送演示数据请求', async () => {
    mockPostSuccess({ seeded: '5' })
    const result = await seedDemo()
    expect(result).toHaveProperty('seeded')
  })
})
