import { useState, useEffect, useCallback, useRef } from 'react'
import { DatePicker, Button, Modal, Spin, Tabs, Table, Tag, Select, Input, InputNumber } from 'antd'
import dayjs, { type Dayjs } from 'dayjs'
import {
  getBoard, generateSchedule, confirmPlan, confirmDay, cancelShift, absentShift, moveShift,
  insertTemporary, insertCrrt, listCrrt, machineOutage, setHoliday, planChange, makeup,
  listConflicts, resolveConflict, getDiffs, getQuality,
  upsertPatient, upsertProfile, rebuildTemplate,
  dischargePatient, placePatient, setInfectionStatus,
  seedDemo,
  startTreatment, completeTreatment,
  type WeekBoard, type CellDTO, type MachineDTO,
  type ConflictItem, type QualityResult, type DiffItem, type CrrtItem,
} from '@/services/smartScheduleApi'

const WD = ['周一', '周二', '周三', '周四', '周五', '周六']
const STATUS_LABEL: Record<number, string> = { 0: '待排', 10: '草稿', 20: '已确认', 50: '透析中', 60: '完成', 70: '取消', 80: '缺席' }
const MODE_COLORS: Record<string, string> = {
  HD: '#0284c7', HDF: '#059669', HFD: '#0891b2', HP: '#d97706', HF: '#0d9488', CRRT: '#e11d48',
}
const FREQS: [number, string][] = [[10, '一三五'], [20, '二四六'], [30, '二四'], [40, '周四'], [90, '临时']]

interface AdminWard { id: number; name: string; zoneType: string; isDisabled: boolean }
interface AdminMachine { id: number; wardId: number; code: string; machineType: string; positionIndex: number }
interface AdminPatient { id: number; name: string; gender: string; infectionStatus: string }
interface AdminProfile { patientId: number; zoneTag: string; defaultMode: string; freqPattern: number; hdfEnabled: boolean; hdfWeekday: number; patientStatus: number }
interface AdminTemplate { id: number; name: string; isActive: boolean }
interface AdminShift { id: number; name: string }
interface AdminState { wards: AdminWard[]; machines: AdminMachine[]; patients: AdminPatient[]; profiles: AdminProfile[]; templates: AdminTemplate[]; shifts: AdminShift[] }

export default function SmartSchedulePage() {
  const [currentDate, setCurrentDate] = useState<Dayjs>(dayjs())
  const [weeks, setWeeks] = useState(2)
  const [board, setBoard] = useState<WeekBoard | null>(null)
  const [loading, setLoading] = useState(false)
  const [msg, setMsg] = useState('')
  const [msgErr, setMsgErr] = useState(false)
  const [conflicts, setConflicts] = useState<ConflictItem[]>([])
  const [quality, setQuality] = useState<QualityResult | null>(null)
  const [diffs, setDiffs] = useState<DiffItem[]>([])
  const [crrts, setCrrts] = useState<CrrtItem[]>([])

  const [moveSrc, setMoveSrc] = useState<{ id: number; patientId: number; name: string } | null>(null)
  const [dragSrc, setDragSrc] = useState<{ id: number; patientId: number; name: string } | null>(null)
  const [confirmDate, setConfirmDate] = useState<string>(currentDate.format('YYYY-MM-DD'))

  const [menuCell, setMenuCell] = useState<{ cell: CellDTO; machineId: number; date: string; shiftId: number } | null>(null)
  const [tempModal, setTempModal] = useState(false)
  const [tempForm, setTempForm] = useState({ patientId: 0, wardId: 0, date: '', mode: 'HD' })
  const [crrtModal, setCrrtModal] = useState(false)
  const [crrtForm, setCrrtForm] = useState({ patientId: 0, wardId: 0, startAt: '', endAt: '' })
  const [planModal, setPlanModal] = useState(false)
  const [planForm, setPlanForm] = useState({ patientId: 0, changeType: 'FREQ', newValue: '', effectiveDate: '' })
  const [outageModal, setOutageModal] = useState(false)
  const [outageForm, setOutageForm] = useState({ machineId: 0, machineCode: '', startDate: '', endDate: '', type: 10, reason: '' })
  const [holidayLoading, setHolidayLoading] = useState(false)

  const [adminOpen, setAdminOpen] = useState(false)
  const [adminTab, setAdminTab] = useState('ward')
  const [adm, setAdm] = useState<AdminState>({ wards: [], machines: [], patients: [], profiles: [], templates: [], shifts: [] })
  const [fWard, setFWard] = useState({ name: '', zoneType: 'A', sort: 1 })
  const [fMachine, setFMachine] = useState({ wardId: 0, code: '', machineType: 'HD', positionIndex: 1 })
  const [fPat, setFPat] = useState({ patientId: 0, name: '', gender: '男', zoneTag: 'A', homeWardId: 0, weeklyCount: 3, freqPattern: 10, shiftId: 0, defaultMode: 'HD', hdfEnabled: false, hdfWeekday: 1, infectionStatus: 'unknown' })

  const [density, setDensity] = useState<'compact' | 'comfortable'>('comfortable')
  const [hideQuality, setHideQuality] = useState(false)
  const [matrixFullscreen, setMatrixFullscreen] = useState(false)

  const dateStr = currentDate.format('YYYY-MM-DD')
  const today = dayjs().format('YYYY-MM-DD')

  const notify = (text: string, isErr?: boolean) => { setMsg(text); setMsgErr(!!isErr); setTimeout(() => setMsg(''), 5000) }

  const fetchData = useCallback(async () => {
    setLoading(true)
    try {
      const [b, cf, q, df, cr] = await Promise.allSettled([getBoard(dateStr), listConflicts(0), getQuality(dateStr, 2), getDiffs(dateStr, 2), listCrrt(dateStr)])
      if (b.status === 'fulfilled') { setBoard(b.value); if (b.value.dates?.length) setConfirmDate(b.value.dates[0]) }
      if (cf.status === 'fulfilled') setConflicts(cf.value.conflicts ?? [])
      if (q.status === 'fulfilled') setQuality(q.value)
      if (df.status === 'fulfilled') setDiffs(df.value.items ?? [])
      if (cr.status === 'fulfilled') setCrrts(cr.value.items ?? [])
    } finally { setLoading(false) }
  }, [dateStr])

  useEffect(() => { fetchData() }, [fetchData])

  const api = {
    generate: async () => {
      setLoading(true)
      try { const r = await generateSchedule({ startDate: dateStr, weeks }); notify(`生成:${r.dialysisDays} 透析日 / ${r.drafts} 草稿 / ${r.conflicts} 冲突`) } catch { notify('生成失败', true) }
      setLoading(false); fetchData()
    },
    seed: async () => { try { const r = await seedDemo(); notify(r.seeded) } catch { notify('种子失败', true) }; fetchData() },
    confirmPlan: async () => { try { const r = await confirmPlan({ weekStart: board?.weekStart, weeks }); notify(`整盘确认:${r.confirmed} 条`) } catch { notify('确认失败', true) }; fetchData() },
    confirmDay: async (level: number) => { try { const r = await confirmDay({ date: confirmDate, level }); notify(`${level === 2 ? '次日' : '当日'}确认:${r.confirmed} 条`) } catch { notify('确认失败', true) }; fetchData() },
    cancel: async (id: number) => { try { await cancelShift(id, '前端操作'); notify('已取消') } catch { notify('取消失败', true) }; setMenuCell(null); fetchData() },
    absent: async (id: number) => { try { await absentShift(id, '前端操作'); notify('已缺席') } catch { notify('操作失败', true) }; setMenuCell(null); fetchData() },
    doMove: async (srcId: number, machineId: number, d: string, shiftId: number) => { try { await moveShift(srcId, { machineId, date: d, shiftId }); notify('已移动') } catch { notify('移动失败', true) }; setMoveSrc(null); setDragSrc(null); fetchData() },
    submitTemp: async () => { try { await insertTemporary(tempForm); notify('临时透析已插入'); setTempModal(false) } catch { notify('插入失败', true) }; fetchData() },
    submitCrrt: async () => { try { await insertCrrt({ patientId: crrtForm.patientId, wardId: crrtForm.wardId, startAt: crrtForm.startAt, endAt: crrtForm.endAt || undefined }); notify('CRRT已安排'); setCrrtModal(false) } catch { notify('CRRT失败', true) }; fetchData() },
    submitOutage: async () => { try { await machineOutage(outageForm.machineId, { startDate: outageForm.startDate, endDate: outageForm.endDate, type: outageForm.type, reason: outageForm.reason }); notify(`${outageForm.machineCode} 停机已登记`); setOutageModal(false) } catch { notify('停机失败', true) }; fetchData() },
    submitHoliday: async () => { setHolidayLoading(true); try { const r = await setHoliday({ date: confirmDate, mode: 10 }) as Record<string, number>; notify(`假日:取消${r.cancelled}/建议${r.suggested}`) } catch { notify('操作失败', true) }; setHolidayLoading(false); fetchData() },
    submitPlan: async () => { try { const r = await planChange(planForm.patientId, { changeType: planForm.changeType, newValue: planForm.newValue, effectiveDate: planForm.effectiveDate || dateStr }) as Record<string, number>; notify(`方案变更:${r.replanned}重排`); setPlanModal(false) } catch { notify('变更失败', true) }; fetchData() },
    doMakeup: async (pid: number) => { try { const r = await makeup(pid, { weekStart: board?.weekStart, weeks: 2 }) as Record<string, number>; notify(`#${pid} 补排${r.placed}次`) } catch { notify('补排失败', true) }; fetchData() },
    start: async (id: number) => { try { await startTreatment(id); notify('已上机') } catch { notify('上机失败', true) }; setMenuCell(null); fetchData() },
    complete: async (id: number) => { try { await completeTreatment(id); notify('已下机') } catch { notify('下机失败', true) }; setMenuCell(null); fetchData() },
  }

  const loadAdmin = async () => {
    try {
      const tok = localStorage.getItem('hdis_access_token') || ''
      const headers = { 'Authorization': `Bearer ${tok}` }
      const get = async (u: string) => (await fetch(u, { headers })).json()
      const [w, m, pt, pr, t, sh] = await Promise.all([get('/api/v2/admin/wards'), get('/api/v2/admin/machines'), get('/api/v2/admin/patients'), get('/api/v2/admin/profiles'), get('/api/v2/admin/templates'), get('/api/v2/admin/shifts')])
      setAdm({ wards: w.data?.items || w.items || [], machines: m.data?.items || m.items || [], patients: pt.data?.items || pt.items || [], profiles: pr.data?.items || pr.items || [], templates: t.data?.items || t.items || [], shifts: sh.data?.items || sh.items || [] })
    } catch { notify('加载管理数据失败', true) }
  }
  const openAdmin = async () => { setAdminOpen(true); await loadAdmin() }

  const adminHeaders = () => {
    const tok = localStorage.getItem('hdis_access_token') || ''
    return { 'Content-Type': 'application/json', 'Authorization': `Bearer ${tok}` }
  }

  const createWard = async () => {
    try { await fetch('/api/v2/admin/wards', { method: 'POST', headers: adminHeaders(), body: JSON.stringify(fWard) }); notify('病区已建'); loadAdmin() } catch { notify('失败', true) }
  }
  const createMachine = async () => {
    try { await fetch('/api/v2/admin/machines', { method: 'POST', headers: adminHeaders(), body: JSON.stringify({ wardId: Number(fMachine.wardId), code: fMachine.code, machineType: fMachine.machineType, positionIndex: Number(fMachine.positionIndex) }) }); notify('机器已建'); loadAdmin() } catch { notify('失败', true) }
  }
  const savePatient = async () => {
    try {
      await upsertPatient({ id: Number(fPat.patientId), name: fPat.name, gender: fPat.gender })
      await upsertProfile({ patientId: Number(fPat.patientId), zoneTag: fPat.zoneTag, homeWardId: fPat.homeWardId || null, weeklyCount: Number(fPat.weeklyCount), freqPattern: Number(fPat.freqPattern), shiftId: fPat.shiftId || null, defaultMode: fPat.defaultMode, hdfEnabled: fPat.hdfEnabled, hdfWeekday: fPat.hdfEnabled ? Number(fPat.hdfWeekday) : null })
      await setInfectionStatus(Number(fPat.patientId), { status: fPat.infectionStatus })
      notify('病人+骨架已录入'); loadAdmin()
    } catch { notify('录入失败', true) }
  }

  const onCellClick = (machineId: number, date: string, shiftId: number, cell: CellDTO | null) => {
    if (moveSrc) {
      if (cell) { notify('目标必须是空格', true); return }
      api.doMove(moveSrc.id, machineId, date, shiftId)
      return
    }
    if (dragSrc) return
    if (!cell || !cell.id) return
    setMenuCell({ cell, machineId, date, shiftId })
  }

  const onDrop = (machineId: number, date: string, shiftId: number, cell: CellDTO | null) => {
    if (!dragSrc || cell) { setDragSrc(null); return }
    api.doMove(dragSrc.id, machineId, date, shiftId)
  }

  const modeColor = (mode: string) => MODE_COLORS[mode] || '#64748b'
  const handleDateChange = (d: Dayjs | null) => {
    if (!d) return
    const monday = d.startOf('week').add(1, 'day')
    setCurrentDate(monday)
  }

  const matrixTopOffset = matrixFullscreen ? 120 : hideQuality ? 180 : 244
  const machineColWidth = density === 'compact' ? 82 : 96
  const minShiftColWidth = density === 'compact' ? 78 : 86
  const cellHeight = density === 'compact' ? 38 : 44
  const patientCardHeight = density === 'compact' ? 30 : 34

  const matrixWrapRef = useRef<HTMLDivElement | null>(null)
  const [matrixWidth, setMatrixWidth] = useState(0)

  useEffect(() => {
    const el = matrixWrapRef.current
    if (!el) { setMatrixWidth(1200); return }
    const update = () => setMatrixWidth(el.clientWidth)
    update()
    const ro = new ResizeObserver(update)
    ro.observe(el)
    return () => ro.disconnect()
  }, [])

  const shiftCount = Math.max(1, (board?.dates?.length || 0) * (board?.shifts?.length || 1))
  const autoShiftColWidth = matrixWidth > 0
    ? Math.floor((matrixWidth - machineColWidth) / shiftCount)
    : minShiftColWidth
  const shiftColWidth = Math.max(minShiftColWidth, autoShiftColWidth)
  const tableMinWidth = machineColWidth + shiftColWidth * shiftCount

  const txName = density === 'compact' ? 'text-[12px]' : 'text-[13px]'
  const txMode = density === 'compact' ? 'text-[10px]' : 'text-[11px]'
  const txStatus = 'text-[11px]'

  return (
    <div className="flex flex-col h-screen overflow-hidden bg-slate-50">
      <div className={`shrink-0 border-b bg-white px-4 ${matrixFullscreen ? 'py-1.5' : 'py-2'}`}>
        <div className="flex flex-wrap items-center gap-1.5 text-sm">
          <h1 className={`font-bold text-slate-800 ${matrixFullscreen ? 'text-base' : 'text-lg'}`}>排班管理</h1>
          {!matrixFullscreen && <span className="text-xs text-slate-400 ml-2">高密度矩阵模式：压缩顶部信息，优先留给排班表</span>}
          <span className="flex-1" />
          <DatePicker value={currentDate} onChange={handleDateChange} allowClear={false} size="small" />
          <Select value={weeks} onChange={setWeeks} size="small" style={{ width: 70 }} options={[{ value: 2, label: '2周' }, { value: 4, label: '4周' }]} />
          <span className="text-[10px] text-slate-300" title="2周/4周为生成与质量评估范围，矩阵仍展示本周">生成范围</span>
          <Button size="small" onClick={fetchData}>刷新</Button>
          <Button size="small" type="primary" onClick={api.generate}>生成排班</Button>
          <Button size="small" onClick={openAdmin}>管理</Button>
          <Button size="small" type={density === 'comfortable' ? 'primary' : 'default'} onClick={() => setDensity('comfortable')}>舒适</Button>
          <Button size="small" type={density === 'compact' ? 'primary' : 'default'} onClick={() => setDensity('compact')}>紧凑</Button>
          <Button size="small" type={matrixFullscreen ? 'primary' : 'default'} onClick={() => setMatrixFullscreen(v => !v)}>{matrixFullscreen ? '退出全屏' : '全屏矩阵'}</Button>
          <Button size="small" onClick={() => setHideQuality(v => !v)}>{hideQuality ? '显示统计' : '隐藏统计'}</Button>
        </div>

        <div className={`flex flex-wrap items-center gap-1.5 text-xs ${matrixFullscreen ? 'hidden' : 'mt-1.5'}`}>
          <span className="text-slate-400">确认</span>
          <Button size="small" onClick={api.confirmPlan}>整盘</Button>
          <Select value={confirmDate} onChange={setConfirmDate} size="small" style={{ width: 118 }} options={board?.dates?.map((d, i) => ({ value: d, label: `${WD[i]} ${d.slice(5)}` })) || []} />
          <Button size="small" onClick={() => api.confirmDay(2)}>次日</Button>
          <Button size="small" onClick={() => api.confirmDay(3)}>当日</Button>
          <span className="text-slate-300 mx-0.5">|</span>
          <Button size="small" danger onClick={api.submitHoliday} loading={holidayLoading}>假日</Button>
          <Button size="small" onClick={() => { setPlanForm({ patientId: 0, changeType: 'FREQ', newValue: '', effectiveDate: board?.dates?.[0] || dateStr }); setPlanModal(true) }}>方案变更</Button>
          <Button size="small" danger onClick={() => { const cw = board?.wards?.[0]; setTempForm({ patientId: 0, wardId: cw?.id || 0, date: board?.dates?.[0] || dateStr, mode: 'HD' }); setTempModal(true) }}>+临时</Button>
          <Button size="small" style={{ background: '#a21caf', borderColor: '#a21caf', color: '#fff' }} onClick={() => { const cw = (board?.wards || []).find(w => w.zoneType === 'C'); setCrrtForm({ patientId: 0, wardId: cw?.id || 0, startAt: `${board?.dates?.[0] || dateStr} 09:00`, endAt: '' }); setCrrtModal(true) }}>+CRRT</Button>
        </div>

        {msg && (
          <div className={`mt-1 text-xs font-medium px-2 py-0.5 rounded ${msgErr ? 'bg-red-50 text-red-700' : 'bg-green-50 text-green-700'}`}>
            {msg}
            {moveSrc && <Button size="small" danger onClick={() => { setMoveSrc(null); notify('已取消移动') }} className="ml-2">取消移动</Button>}
          </div>
        )}
      </div>

      <div className={`shrink-0 flex flex-wrap items-center gap-2 border-b bg-white px-4 ${matrixFullscreen ? 'hidden' : 'py-1.5'} ${hideQuality ? 'hidden' : ''}`}>
        {quality && (
          <div className="flex flex-wrap items-center gap-x-5 gap-y-1 text-xs text-slate-600">
            <span className="font-semibold text-slate-400">质量</span>
            <span>达标率 <b className="text-slate-800">{Math.round(quality.onTargetRate * 100)}%</b></span>
            <span>利用率 <b className="text-slate-800">{Math.round(quality.utilization * 100)}%</b></span>
            <span>稳定率 <b className="text-slate-800">{Math.round(quality.stabilityRate * 100)}%</b></span>
            <span>综合 <b className="text-slate-800">{quality.score}/100</b></span>
            <span>冲突 <b className="text-red-600">{quality.openConflicts}</b></span>
            <span>患者 <b className="text-slate-800">{quality.patientsOnTarget}/{quality.patientsTotal}</b></span>
          </div>
        )}
        <span className="flex-1" />
        <div className="flex items-center gap-1.5 text-xs">
          <span className="text-slate-400">图例</span>
          {['HD', 'HDF', 'CRRT'].map(m => <span key={m} className="px-1.5 py-0.5 rounded border bg-slate-50 text-slate-600">{m}</span>)}
          <span className="text-slate-400">✓=确认级别 · 虚线=草稿 · 绿光=透析中</span>
        </div>
      </div>

      {crrts.length > 0 && !matrixFullscreen && (
        <div className="shrink-0 mx-4 mt-1 border rounded bg-purple-50 px-2 py-1 text-xs flex items-center gap-3 flex-wrap">
          <span className="font-semibold text-purple-800">CRRT({crrts.length})</span>
          {crrts.slice(0, 4).map(x => (
            <span key={x.id} className="text-slate-600">
              <b>{x.patientName || `#${x.patientId}`}</b> {x.machineCode} {String(x.startAt).slice(0, 16).replace('T', ' ')}~{x.endAt ? String(x.endAt).slice(0, 16).replace('T', ' ') : '进行中'}
            </span>
          ))}
        </div>
      )}

      <Spin spinning={loading} wrapperClassName="flex-1 overflow-hidden">
        <Tabs
          className="px-4 pt-1"
          size="small"
          tabBarExtraContent={
            matrixFullscreen ? null : undefined
          }
          items={[
            {
              key: 'board', label: '周排班矩阵',
              children: board ? (
                <div ref={matrixWrapRef} className="overflow-auto rounded border bg-white" style={{ height: `calc(100vh - ${matrixTopOffset}px)` }}>
                  <table
                    className={density === 'compact' ? 'border-collapse text-[12px]' : 'border-collapse text-[13px]'}
                    style={{ minWidth: Math.max(tableMinWidth, matrixWidth || 0), width: Math.max(tableMinWidth, matrixWidth || 0) }}
                  >
                    <thead className="sticky top-0 z-20 bg-gray-100">
                      <tr>
                        <th className="sticky left-0 z-30 bg-gray-100 border px-2" rowSpan={2} style={{ width: machineColWidth, minWidth: machineColWidth }}>病区 / 机器</th>
                        {board.dates?.map((d, i) => {
                          const isToday = d === today
                          return <th key={d} colSpan={board.shifts?.length || 1} className={`border px-1 text-[13px] ${isToday ? 'bg-amber-100' : ''}`} style={{ width: shiftColWidth * (board.shifts?.length || 1), minWidth: shiftColWidth * (board.shifts?.length || 1) }}>{WD[i]}<div className="font-normal text-gray-400">{d.slice(5)}</div></th>
                        })}
                      </tr>
                      <tr>
                        {board.dates?.flatMap(d => (board.shifts || []).map(s => {
                          const isToday = d === today
                          return <th key={d + s.id} className={`border px-1 font-normal text-gray-500 text-[12px] ${isToday ? 'bg-amber-50' : ''}`} style={{ width: shiftColWidth, minWidth: shiftColWidth }}>{s.name.replace('班', '')}</th>
                        }))}
                      </tr>
                    </thead>
                    <tbody>
                      {board.wards?.map(w => (
                        <><tr key={w.id}><td colSpan={1 + (board.shifts?.length || 1) * (board.dates?.length || 1)} className="bg-gray-100 border px-2 py-0.5 text-[12px] font-bold">{w.name} <span className="text-gray-400 font-normal">({(w.machines || []).length}台)</span></td></tr>
                          {(w.machines || []).map((m: MachineDTO) => (
                            <tr key={m.id} className="hover:bg-sky-50">
                              <td className="sticky left-0 z-10 bg-white border px-2 font-bold cursor-pointer hover:bg-rose-50 whitespace-nowrap text-[13px]" style={{ width: machineColWidth, minWidth: machineColWidth }} onClick={() => { setOutageForm({ machineId: m.id, machineCode: m.code, startDate: board.dates?.[0] || dateStr, endDate: board.dates?.[0] || dateStr, type: 10, reason: '' }); setOutageModal(true) }} title="点击登记停机"><span className="text-slate-700">{m.code}</span><span className="ml-1 text-[11px] font-semibold text-slate-400">{m.machineType}</span></td>
                              {board.dates?.flatMap(d => (board.shifts || []).map(s => {
                                const cell = m.cells?.[`${d}|${s.id}`] as CellDTO | undefined
                                const isToday = d === today
                                const isPast = d < today
                                const colTint = isToday ? 'bg-amber-50' : isPast ? 'bg-gray-50' : ''
                                const dropHL = (moveSrc || dragSrc) && !cell
                                return (
                                  <td key={d + s.id}
                                    onClick={() => onCellClick(m.id, d, s.id, cell || null)}
                                    onDragOver={e => { if (!cell) e.preventDefault() }}
                                    onDrop={() => onDrop(m.id, d, s.id, cell || null)}
                                    className={`border align-middle cursor-pointer ${colTint} ${dropHL ? 'bg-green-50' : ''}`}
                                    style={{ width: shiftColWidth, minWidth: shiftColWidth, height: cellHeight, padding: density === 'compact' ? 2 : 3 }}>
                                    {cell ? (
                                      <div draggable onDragStart={() => setDragSrc({ id: cell.id, patientId: cell.patientId, name: cell.patientName })} onDragEnd={() => setDragSrc(null)}
                                        className={`rounded border bg-white leading-tight ${density === 'compact' ? 'px-1 py-[1px]' : 'px-1.5 py-1'} ${cell.status === 10 ? 'border-dashed' : ''} ${cell.status === 50 ? 'shadow-[0_0_9px_1px_rgba(16,185,129,.55)]' : ''} ${cell.status === 60 ? 'opacity-75 border-gray-300' : ''}`}
                                        style={{ borderColor: modeColor(cell.dialysisMode), minHeight: patientCardHeight }}>
                                        <div className="flex items-center gap-0.5">
                                          <span className={`truncate ${txName} font-bold text-slate-900`}>{cell.patientName || `#${cell.patientId}`}</span>
                                          {cell.confirms > 0 && <span className="text-emerald-600 text-[10px]">{'●'.repeat(cell.confirms)}</span>}
                                        </div>
                                        <div className="flex items-center justify-between gap-0.5">
                                          <span className={`rounded font-black ${txMode} ${density === 'compact' ? 'px-1' : 'px-1.5'}`} style={{ color: modeColor(cell.dialysisMode), background: modeColor(cell.dialysisMode) + '1a' }}>{cell.dialysisMode}</span>
                                          {cell.sourceType === 20 ? <span className={`${txStatus} font-black text-amber-600`}>临</span> : cell.status === 80 ? <span className={`${txStatus} font-black text-rose-500`}>缺</span> : cell.status === 50 ? <span className={`${txStatus} font-black text-emerald-600`}>透</span> : cell.status === 60 ? <span className={`${txStatus} font-black text-slate-400`}>完</span> : null}
                                        </div>
                                      </div>
                                    ) : (
                                      <div className={dropHL ? 'h-full rounded border border-dashed border-emerald-300 bg-emerald-50' : 'h-full rounded bg-slate-50/60'} />
                                    )}
                                  </td>
                                )
                              }))}
                            </tr>
                          ))}
                        </>
                      ))}
                    </tbody>
                  </table>
                </div>
              ) : <div className="text-gray-400 text-center py-8">暂无排班数据，请先点击"写演示数据"再"生成排班"</div>,
            },
            {
              key: 'conflicts', label: `冲突(${conflicts.length})`,
              children: conflicts.length > 0 ? (
                <div className="max-h-96 overflow-auto">
                  <Table dataSource={conflicts} rowKey="id" size="small" pagination={false}
                    columns={[
                      { title: '类型', dataIndex: 'conflictType', render: (v: string) => <Tag>{v}</Tag> },
                      { title: '严重度', dataIndex: 'severity', render: (v: number) => <Tag color={v >= 20 ? 'red' : 'orange'}>{v >= 20 ? '报警' : '提示'}</Tag> },
                      { title: '详情', dataIndex: 'detail' },
                      { title: '操作', render: (_: unknown, r: ConflictItem) => <span><Button size="small" type="link" onClick={async () => { await resolveConflict(r.id, 'accept'); fetchData() }}>接受</Button><Button size="small" type="link" onClick={async () => { await resolveConflict(r.id, 'ignore'); fetchData() }}>忽略</Button></span> },
                    ]} />
                </div>
              ) : <div className="text-gray-400 text-center py-4">无待处理冲突</div>,
            },
            {
              key: 'diffs', label: `差异(${diffs.length})`,
              children: diffs.length > 0 ? (
                <div className="max-h-96 overflow-auto">
                  {diffs.map(d => (
                    <div key={d.patientId} className="flex items-center gap-3 py-1 border-b border-rose-100 text-xs">
                      <span className="font-medium">{d.patientName || `#${d.patientId}`}</span>
                      <span className="text-gray-500">应排{d.expected}·已排{d.scheduled}</span>
                      <span className={d.diff > 0 ? 'text-rose-700 font-semibold' : 'text-sky-700'}>{d.diff > 0 ? `少排${d.diff}次` : `多排${-d.diff}次`}</span>
                      {d.diff > 0 && <Button size="small" type="primary" danger onClick={() => api.doMakeup(d.patientId)}>一键补排</Button>}
                    </div>
                  ))}
                </div>
              ) : <div className="text-gray-400 text-center py-4">无差异</div>,
            },
          ]} />
      </Spin>

      <Modal open={!!menuCell} onCancel={() => setMenuCell(null)} footer={null} title="排班操作" width={280}>
        {menuCell && (
          <div className="flex flex-col gap-2">
            <div className="text-sm text-gray-500 mb-2">{menuCell.date} · {STATUS_LABEL[menuCell.cell.status]} · {menuCell.cell.dialysisMode}</div>
            {menuCell.cell.status === 20 && <Button block type="primary" onClick={() => api.start(menuCell.cell.id)} style={{ background: '#059669', borderColor: '#059669' }}>上机(开始治疗)</Button>}
            {menuCell.cell.status === 50 && <Button block type="primary" onClick={() => api.complete(menuCell.cell.id)}>下机(完成治疗)</Button>}
            <Button block onClick={() => { setMoveSrc({ id: menuCell.cell.id, patientId: menuCell.cell.patientId, name: menuCell.cell.patientName }); setMenuCell(null); notify('点击空格移动', false) }}>移动到空格…</Button>
            <Button block danger onClick={() => api.cancel(menuCell.cell.id)}>取消(提前请假)</Button>
            <Button block onClick={() => api.absent(menuCell.cell.id)} style={{ background: '#fef3c7', borderColor: '#f59e0b', color: '#92400e' }}>标记缺席</Button>
          </div>
        )}
      </Modal>

      <Modal open={tempModal} onCancel={() => setTempModal(false)} onOk={api.submitTemp} title="+ 临时透析(急诊加台)">
        <div className="flex flex-col gap-2">
          <label>病人ID <InputNumber value={tempForm.patientId} onChange={v => setTempForm({ ...tempForm, patientId: v || 0 })} className="w-full" /></label>
          <label>病区 <Select value={tempForm.wardId || undefined} onChange={v => setTempForm({ ...tempForm, wardId: v || 0 })} className="w-full" options={board?.wards?.map(w => ({ value: w.id, label: w.name })) || []} /></label>
          <label>日期 <Input value={tempForm.date} onChange={e => setTempForm({ ...tempForm, date: e.target.value })} /></label>
          <label>模式 <Select value={tempForm.mode} onChange={v => setTempForm({ ...tempForm, mode: v })} options={['HD', 'HDF', 'HFD', 'HP', 'HF', 'CRRT'].map(m => ({ value: m, label: m }))} /></label>
        </div>
      </Modal>

      <Modal open={crrtModal} onCancel={() => setCrrtModal(false)} onOk={api.submitCrrt} title="+ CRRT(C区)">
        <div className="flex flex-col gap-2">
          <label>病人ID <InputNumber value={crrtForm.patientId} onChange={v => setCrrtForm({ ...crrtForm, patientId: v || 0 })} className="w-full" /></label>
          <label>C区 <Select value={crrtForm.wardId || undefined} onChange={v => setCrrtForm({ ...crrtForm, wardId: v || 0 })} className="w-full" options={(board?.wards || []).filter(w => w.zoneType === 'C').map(w => ({ value: w.id, label: w.name }))} /></label>
          <label>开始 <Input value={crrtForm.startAt} onChange={e => setCrrtForm({ ...crrtForm, startAt: e.target.value })} placeholder="2026-06-08 09:00" /></label>
          <label>结束 <Input value={crrtForm.endAt} onChange={e => setCrrtForm({ ...crrtForm, endAt: e.target.value })} placeholder="可空=进行中" /></label>
        </div>
      </Modal>

      <Modal open={planModal} onCancel={() => setPlanModal(false)} onOk={api.submitPlan} title="方案变更">
        <div className="flex flex-col gap-2">
          <label>病人ID <InputNumber value={planForm.patientId} onChange={v => setPlanForm({ ...planForm, patientId: v || 0 })} className="w-full" /></label>
          <label>变更项 <Select value={planForm.changeType} onChange={v => setPlanForm({ ...planForm, changeType: v })} options={[{ value: 'FREQ', label: '频率' }, { value: 'SHIFT', label: '班次' }, { value: 'ZONE', label: '分区' }, { value: 'HDF', label: 'HDF' }]} /></label>
          <label>新值 <Input value={planForm.newValue} onChange={e => setPlanForm({ ...planForm, newValue: e.target.value })} placeholder="如10/A/true" /></label>
          <label>生效日 <Input value={planForm.effectiveDate} onChange={e => setPlanForm({ ...planForm, effectiveDate: e.target.value })} placeholder="YYYY-MM-DD" /></label>
        </div>
        <div className="text-xs text-gray-400 mt-2">生效日后未确认排班将取消待重排,已确认报警人工</div>
      </Modal>

      <Modal open={outageModal} onCancel={() => setOutageModal(false)} onOk={api.submitOutage} title={`停机: ${outageForm.machineCode}`}>
        <div className="flex flex-col gap-2">
          <label>起 <Input value={outageForm.startDate} onChange={e => setOutageForm({ ...outageForm, startDate: e.target.value })} /></label>
          <label>止 <Input value={outageForm.endDate} onChange={e => setOutageForm({ ...outageForm, endDate: e.target.value })} /></label>
          <label>类型 <Select value={outageForm.type} onChange={v => setOutageForm({ ...outageForm, type: v })} options={[{ value: 10, label: '临时(≤48h自动迁移)' }, { value: 20, label: '长期/报废(报警人工)' }]} /></label>
          <label>原因 <Input value={outageForm.reason} onChange={e => setOutageForm({ ...outageForm, reason: e.target.value })} placeholder="如故障维修" /></label>
        </div>
      </Modal>

      <Modal open={adminOpen} onCancel={() => setAdminOpen(false)} footer={null} title="资源与病人维护" width={900}>
        <Tabs activeKey={adminTab} onChange={setAdminTab} items={[
          {
            key: 'ward', label: '病区',
            children: <div className="text-sm">
              <div className="flex flex-wrap items-end gap-2 mb-3 bg-gray-50 p-2 rounded">
                <label>名称<Input value={fWard.name} onChange={e => setFWard({ ...fWard, name: e.target.value })} /></label>
                <label>类型<Select value={fWard.zoneType} onChange={v => setFWard({ ...fWard, zoneType: v })} options={[{ value: 'A', label: 'A 门诊' }, { value: 'B', label: 'B 住院' }, { value: 'C', label: 'C 全警戒' }]} /></label>
                <Button type="primary" onClick={createWard}>新增病区</Button>
              </div>
              <Table dataSource={adm.wards} rowKey="id" size="small" pagination={false} columns={[{ title: 'ID', dataIndex: 'id' }, { title: '名称', dataIndex: 'name' }, { title: '类型', dataIndex: 'zoneType' }, { title: '状态', render: (_, r: AdminWard) => r.isDisabled ? '停用' : '启用' }]} />
            </div>,
          },
          {
            key: 'machine', label: '机器',
            children: <div className="text-sm">
              <div className="flex flex-wrap items-end gap-2 mb-3 bg-gray-50 p-2 rounded">
                <label>病区<Select value={fMachine.wardId || undefined} onChange={v => setFMachine({ ...fMachine, wardId: v || 0 })} options={adm.wards.map((w: AdminWard) => ({ value: w.id, label: w.name }))} /></label>
                <label>编号<Input value={fMachine.code} onChange={e => setFMachine({ ...fMachine, code: e.target.value })} /></label>
                <label>机型<Select value={fMachine.machineType} onChange={v => setFMachine({ ...fMachine, machineType: v })} options={['HD', 'HDF', 'CRRT'].map(m => ({ value: m, label: m }))} /></label>
                <label>位序<InputNumber value={fMachine.positionIndex} onChange={v => setFMachine({ ...fMachine, positionIndex: v || 1 })} /></label>
                <Button type="primary" onClick={createMachine}>新增机器</Button>
              </div>
              <Table dataSource={adm.machines} rowKey="id" size="small" pagination={false} columns={[{ title: 'ID', dataIndex: 'id' }, { title: '病区', render: (_: unknown, r: AdminMachine) => adm.wards.find((w: AdminWard) => w.id === r.wardId)?.name }, { title: '编号', dataIndex: 'code' }, { title: '机型', dataIndex: 'machineType' }, { title: '位序', dataIndex: 'positionIndex' }]} />
            </div>,
          },
          {
            key: 'profile', label: '病人/骨架',
            children: <div className="text-sm">
              <div className="grid grid-cols-3 gap-2 mb-3 bg-gray-50 p-2 rounded items-end">
                <label>病人ID<InputNumber value={fPat.patientId} onChange={v => setFPat({ ...fPat, patientId: v || 0 })} className="w-full" /></label>
                <label>姓名<Input value={fPat.name} onChange={e => setFPat({ ...fPat, name: e.target.value })} /></label>
                <label>性别<Select value={fPat.gender} onChange={v => setFPat({ ...fPat, gender: v })} options={[{ value: '男', label: '男' }, { value: '女', label: '女' }]} /></label>
                <label>分区<Select value={fPat.zoneTag} onChange={v => setFPat({ ...fPat, zoneTag: v })} options={['A', 'B', 'C'].map(v => ({ value: v, label: v }))} /></label>
                <label>归属区<Select value={fPat.homeWardId || undefined} onChange={v => setFPat({ ...fPat, homeWardId: v || 0 })} options={adm.wards.map((w: AdminWard) => ({ value: w.id, label: w.name }))} /></label>
                <label>每周次数<InputNumber min={1} max={3} value={fPat.weeklyCount} onChange={v => setFPat({ ...fPat, weeklyCount: v || 0 })} /></label>
                <label>星期组合<Select value={fPat.freqPattern} onChange={v => setFPat({ ...fPat, freqPattern: v })} options={FREQS.map(([v, l]) => ({ value: v, label: l }))} /></label>
                <label>班次<Select value={fPat.shiftId || undefined} onChange={v => setFPat({ ...fPat, shiftId: v || 0 })} options={adm.shifts.map((s: AdminShift) => ({ value: s.id, label: s.name }))} /></label>
                <label>基础模式<Select value={fPat.defaultMode} onChange={v => setFPat({ ...fPat, defaultMode: v })} options={['HD', 'HFD', 'HF'].map(m => ({ value: m, label: m }))} /></label>
                <label>院感<Select value={fPat.infectionStatus} onChange={v => setFPat({ ...fPat, infectionStatus: v })} options={[{ value: 'unknown', label: '未出' }, { value: 'negative', label: '阴性' }, { value: 'positive', label: '阳性' }]} /></label>
                <label className="flex items-center gap-1"><input type="checkbox" checked={fPat.hdfEnabled} onChange={e => setFPat({ ...fPat, hdfEnabled: e.target.checked })} />每两周HDF</label>
                {fPat.hdfEnabled && <label>HDF星期<Select value={fPat.hdfWeekday} onChange={v => setFPat({ ...fPat, hdfWeekday: v })} options={[1, 2, 3, 4, 5, 6].map(d => ({ value: d, label: `周${WD[d - 1]}` }))} /></label>}
                <Button type="primary" onClick={savePatient}>录入病人+骨架</Button>
              </div>
              <Table dataSource={adm.profiles} rowKey="patientId" size="small" pagination={false}
                columns={[
                  { title: '病人', render: (_: unknown, p: AdminProfile) => { const pt = adm.patients.find((x: AdminPatient) => x.id === p.patientId); return <span className={p.patientStatus === 20 ? 'opacity-40' : ''}>#{p.patientId} {pt?.name}{p.patientStatus === 20 && ' (已出组)'}</span> } },
                  { title: '分区', dataIndex: 'zoneTag' },
                  { title: '模式', dataIndex: 'defaultMode' },
                  { title: '组合', render: (_: unknown, r: AdminProfile) => (FREQS.find(f => f[0] === r.freqPattern) || [])[1] },
                  { title: 'HDF', render: (_: unknown, r: AdminProfile) => r.hdfEnabled ? `周${WD[(r.hdfWeekday || 1) - 1]}` : '—' },
                  { title: '院感', render: (_: unknown, r: AdminProfile) => { const pt = adm.patients.find((x: AdminPatient) => x.id === r.patientId); return pt?.infectionStatus === 'positive' ? <span className="text-rose-600 font-semibold">阳</span> : pt?.infectionStatus === 'unknown' ? <span className="text-amber-600">未</span> : '阴' } },
                  { title: '操作', render: (_: unknown, r: AdminProfile) => r.patientStatus !== 20 ? <span className="flex gap-1"><Button size="small" type="link" onClick={async () => { await placePatient(r.patientId, { start: board?.weekStart, weeks: 2 }); notify('已排入'); fetchData() }}>排入</Button><Button size="small" type="link" danger onClick={async () => { await dischargePatient(r.patientId, { reason: '出院' }); notify('已出组'); loadAdmin(); fetchData() }}>出组</Button></span> : null },
                ]} />
            </div>,
          },
          {
            key: 'template', label: '模板',
            children: <div className="text-sm">
              <div className="mb-3 bg-gray-50 p-2 rounded flex items-center gap-3">
                <Button type="primary" onClick={async () => { await rebuildTemplate(); notify('模板已重建'); loadAdmin() }}>由病人骨架重建生效模板</Button>
                <span className="text-gray-500">把现在所有病人的骨架快照成一份新模板(旧模板失效)</span>
              </div>
              <Table dataSource={adm.templates} rowKey="id" size="small" pagination={false} columns={[{ title: 'ID', dataIndex: 'id' }, { title: '名称', dataIndex: 'name' }, { title: '生效', render: (_: unknown, r: AdminTemplate) => r.isActive ? '✅ 生效中' : '—' }]} />
            </div>,
          },
        ]} />
      </Modal>
    </div>
  )
}
