import { useState, useEffect, useMemo, useCallback } from 'react'
import { message, Select, Modal, DatePicker } from 'antd'
import dayjs, { type Dayjs } from 'dayjs'
import { UserPlus, X, CalendarDays, AlertTriangle, CheckCircle2 } from 'lucide-react'
import { restApi } from '@/services'
import {
  getBoard, listStaffDuty, upsertStaffDuty, deleteStaffDuty,
  getStaffShifts, getNurseRatio,
  DUTY_ROLES, type StaffDuty, type WardDTO, type ShiftDef, type NurseRatioResult,
} from '@/services/smartScheduleApi'

interface StaffOption { value: number; label: string }

export default function StaffSchedulePage() {
  const [wards, setWards] = useState<WardDTO[]>([])
  const [wardId, setWardId] = useState<number>(0)
  const [month, setMonth] = useState<Dayjs>(dayjs())
  const [duties, setDuties] = useState<StaffDuty[]>([])
  const [staff, setStaff] = useState<StaffOption[]>([])
  const [loading, setLoading] = useState(false)

  const [modalOpen, setModalOpen] = useState(false)
  const [editCell, setEditCell] = useState<{ date: string; dutyRole: string } | null>(null)
  const [pickStaff, setPickStaff] = useState<number>(0)
  const [pickShift, setPickShift] = useState<string>('')
  const [shifts, setShifts] = useState<ShiftDef[]>([])

  // 护患比自检
  const [ratioDate, setRatioDate] = useState<Dayjs>(dayjs())
  const [ratioShift, setRatioShift] = useState<string>('')
  const [ratio, setRatio] = useState<NurseRatioResult | null>(null)

  const monthStr = month.format('YYYY-MM')
  const shiftName = useCallback((code: string) => shifts.find((s) => s.code === code)?.name || code, [shifts])

  useEffect(() => {
    getBoard(dayjs().format('YYYY-MM-DD'))
      .then((b) => {
        setWards(b.wards || [])
        setWardId((prev) => prev || b.wards?.[0]?.id || 0)
      })
      .catch(() => message.error('加载病区失败'))
    restApi.getUserList({ page: 1, pageSize: 500 })
      .then((res) => setStaff((res.items || []).map((u: { id: string; realName?: string; username: string }) => ({ value: Number(u.id), label: u.realName || u.username }))))
      .catch(() => {})
    getStaffShifts()
      .then((s) => {
        setShifts(s)
        setPickShift((p) => p || s[0]?.code || '')
        setRatioShift((p) => p || s[0]?.code || '')
      })
      .catch(() => {})
  }, [])

  const checkRatio = useCallback(() => {
    if (!wardId) { message.warning('请先选择病区'); return }
    getNurseRatio(wardId, ratioDate.format('YYYY-MM-DD'), ratioShift)
      .then(setRatio)
      .catch(() => message.error('护患比校验失败'))
  }, [wardId, ratioDate, ratioShift])

  const loadDuties = useCallback(() => {
    if (!wardId) return
    setLoading(true)
    listStaffDuty(wardId, monthStr)
      .then(setDuties)
      .catch(() => message.error('加载排班失败'))
      .finally(() => setLoading(false))
  }, [wardId, monthStr])

  useEffect(() => { loadDuties() }, [loadDuties])

  const cellMap = useMemo(() => {
    const m = new Map<string, StaffDuty[]>()
    for (const d of duties) {
      const key = `${dayjs(d.dutyDate).format('YYYY-MM-DD')}|${d.dutyRole}`
      const arr = m.get(key) || []
      arr.push(d)
      m.set(key, arr)
    }
    return m
  }, [duties])

  const days = useMemo(() => {
    const n = month.daysInMonth()
    return Array.from({ length: n }, (_, i) => month.date(i + 1))
  }, [month])

  const openAssign = (date: string, dutyRole: string) => {
    setEditCell({ date, dutyRole })
    setPickStaff(staff[0]?.value || 0)
    setPickShift((p) => p || shifts[0]?.code || '')
    setModalOpen(true)
  }

  const submitAssign = async () => {
    if (!editCell || !pickStaff) { message.warning('请选择人员'); return }
    if (!pickShift) { message.warning('请选择班次'); return }
    const s = staff.find((x) => x.value === pickStaff)
    try {
      await upsertStaffDuty({
        staffId: pickStaff, staffName: s?.label || '', dutyRole: editCell.dutyRole,
        wardId, dutyDate: editCell.date, shift: pickShift,
      })
      message.success('已指派')
      setModalOpen(false)
      loadDuties()
      if (ratio) checkRatio()
    } catch {
      message.error('指派失败')
    }
  }

  const removeDuty = async (id: number) => {
    try { await deleteStaffDuty(id); loadDuties() } catch { message.error('删除失败') }
  }

  const roleColor: Record<string, string> = {
    '当班医生': 'bg-blue-50 text-blue-700 border-blue-200',
    '主班护士': 'bg-violet-50 text-violet-700 border-violet-200',
    '当班护士': 'bg-emerald-50 text-emerald-700 border-emerald-200',
  }

  return (
    <div className="h-full overflow-y-auto bg-slate-50 p-6">
      <div className="flex items-center justify-between mb-5">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-blue-600 flex items-center justify-center text-white"><CalendarDays size={20} /></div>
          <div>
            <h1 className="text-lg font-black text-slate-800">医护人力排班 · 月基线</h1>
            <p className="text-xs text-slate-400">主任排医生 · 护士长排护士；点格指派当班人。当班解析/接班门禁为 v2。</p>
          </div>
        </div>
        <div className="flex items-center gap-3">
          <Select value={wardId || undefined} onChange={setWardId} placeholder="选择病区" style={{ width: 180 }}
            options={wards.map((w) => ({ value: w.id, label: `${w.name}${w.zoneType ? `（${w.zoneType}区）` : ''}` }))} />
          <DatePicker picker="month" value={month} onChange={(v) => v && setMonth(v)} allowClear={false} />
        </div>
      </div>

      {/* 护患比自检（C3：1 护士 : N 台机） */}
      <div className="mb-4 flex items-center gap-3 flex-wrap bg-white rounded-2xl border border-slate-200 px-4 py-3">
        <span className="text-sm font-black text-slate-600">护患比校验</span>
        <DatePicker value={ratioDate} onChange={(v) => v && setRatioDate(v)} allowClear={false} size="small" />
        <Select value={ratioShift || undefined} onChange={setRatioShift} size="small" style={{ width: 160 }} placeholder="班次"
          options={shifts.map((s) => ({ value: s.code, label: s.name }))} />
        <button type="button" onClick={checkRatio} className="px-3 py-1 rounded-lg bg-blue-600 text-white text-xs font-black hover:bg-blue-700">校验</button>
        {ratio && (
          <div className={`inline-flex items-center gap-2 px-3 py-1 rounded-lg text-xs font-black border ${
            ratio.status === 'ok' ? 'bg-green-50 text-green-600 border-green-100'
            : ratio.status === 'understaffed' ? 'bg-red-50 text-red-600 border-red-100'
            : 'bg-amber-50 text-amber-600 border-amber-100'}`}>
            {ratio.status === 'ok' ? <CheckCircle2 size={14} /> : <AlertTriangle size={14} />}
            {ratio.status === 'understaffed' ? '缺岗' : ratio.status === 'overstaffed' ? '超配' : '达标'}
            <span className="font-bold text-slate-500">
              机台 {ratio.machineCount} · 护士 {ratio.nurseCount} · 需 {ratio.requiredNurses}（1:{ratio.ratio}）
            </span>
          </div>
        )}
      </div>

      <div className="bg-white rounded-2xl border border-slate-200 overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="bg-slate-50 border-b border-slate-200">
              <th className="py-2.5 px-3 text-left font-black text-slate-500 w-28">日期</th>
              {DUTY_ROLES.map((r) => (
                <th key={r} className="py-2.5 px-3 text-left font-black text-slate-500">{r}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {days.map((d) => {
              const dateStr = d.format('YYYY-MM-DD')
              const isWeekend = d.day() === 0 || d.day() === 6
              return (
                <tr key={dateStr} className={`border-b border-slate-50 hover:bg-sky-50/40 ${isWeekend ? 'bg-slate-50/40' : ''}`}>
                  <td className="py-2 px-3 font-mono font-bold text-slate-600 whitespace-nowrap">
                    {d.format('MM-DD')} <span className="text-slate-400 text-xs">周{'日一二三四五六'[d.day()]}</span>
                  </td>
                  {DUTY_ROLES.map((role) => {
                    const items = cellMap.get(`${dateStr}|${role}`) || []
                    return (
                      <td key={role} className="py-2 px-3 align-top">
                        <div className="flex flex-wrap items-center gap-1.5">
                          {items.map((it) => (
                            <span key={it.id} className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-lg border text-xs font-bold ${roleColor[role] || 'bg-slate-50 border-slate-200'}`}>
                              {it.staffName || `#${it.staffId}`}{it.shift ? `·${shiftName(it.shift)}` : ''}
                              <button type="button" onClick={() => removeDuty(it.id)} className="text-slate-300 hover:text-rose-500"><X size={12} /></button>
                            </span>
                          ))}
                          <button type="button" onClick={() => openAssign(dateStr, role)}
                            className="inline-flex items-center gap-1 px-2 py-0.5 rounded-lg border border-dashed border-slate-300 text-xs font-bold text-slate-400 hover:text-blue-600 hover:border-blue-400">
                            <UserPlus size={12} /> 指派
                          </button>
                        </div>
                      </td>
                    )
                  })}
                </tr>
              )
            })}
          </tbody>
        </table>
        {loading && <div className="py-6 text-center text-slate-400 text-sm">加载中…</div>}
        {!wardId && <div className="py-10 text-center text-slate-400">请先选择病区</div>}
      </div>

      <Modal open={modalOpen} title={editCell ? `指派 ${editCell.dutyRole} · ${editCell.date}` : ''}
        onCancel={() => setModalOpen(false)} onOk={submitAssign} okText="指派">
        <div className="flex flex-col gap-3 py-2">
          <label className="text-sm font-bold text-slate-600">人员
            <Select showSearch optionFilterProp="label" value={pickStaff || undefined} onChange={setPickStaff}
              placeholder="选择人员" style={{ width: '100%' }} options={staff} className="mt-1" />
          </label>
          <label className="text-sm font-bold text-slate-600">班次
            <Select value={pickShift || undefined} onChange={setPickShift} style={{ width: '100%' }} className="mt-1"
              placeholder="选择班次"
              options={shifts.map((s) => ({ value: s.code, label: `${s.name}（${s.start}-${s.end}）` }))} />
          </label>
        </div>
      </Modal>
    </div>
  )
}
