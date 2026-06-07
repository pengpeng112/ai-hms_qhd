import { Fragment, useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { message, Modal, Spin, Table, Tag } from 'antd'
import {
  Calendar as CalendarIcon, ChevronLeft, ChevronRight,
  RefreshCw, Search, Trash2, Plus, LayoutGrid, List,
  PanelRightOpen, PanelRightClose,
  ArrowRightLeft, ClipboardList, History, X,
} from 'lucide-react'
import {
  restApi, getErrorMessage,
  type RestScheduleBed,
  type RestSchedulePendingPatient,
  type RestScheduleWard,
  type RestScheduleWeekResponse,
  type RestScheduleWeekShift,
  type RestShift,
  type RestPatientShift,
} from '@/services'
import ApplyTemplateModal from '@/components/schedule/ApplyTemplateModal'
import CreateScheduleModal from '@/components/schedule/CreateScheduleModal'
import { useScheduleModals } from '@/hooks/useScheduleModals'
import { useScheduleDragDrop } from '@/hooks/useScheduleDragDrop'

// ─── utils（全部用本地时间，避免 UTC 偏移） ───
function pad(n: number) { return n < 10 ? `0${n}` : `${n}` }
function toDateString(d: Date) { return `${d.getFullYear()}-${pad(d.getMonth()+1)}-${pad(d.getDate())}` }
function toApi(dt: string) { return dt }
function startOfWeek(d: Date) { const x = new Date(d); const day = x.getDay() || 7; x.setDate(x.getDate() - day + 1); x.setHours(0,0,0,0); return x }
function addDays(d: Date, n: number) { const x = new Date(d); x.setDate(x.getDate() + n); return x }
function isToday(dt: string) { return dt === toDateString(new Date()) }
function isDateLocked(dt: string) { return dt < toDateString(new Date()) }

const WEEKDAYS = ['周日','周一','周二','周三','周四','周五','周六']
const freq = (p: Pick<RestSchedulePendingPatient,'oddWeekFrequency'|'evenWeekFrequency'>) =>
  p.oddWeekFrequency === p.evenWeekFrequency ? `每周${p.oddWeekFrequency||'--'}次` : `单${p.oddWeekFrequency||0}/双${p.evenWeekFrequency||0}`

// 透析模式色块（用作左侧色条）
const MODE_BORDER: Record<string, string> = {
  'HD': 'border-l-blue-600',
  'HDF': 'border-l-violet-600',
  'HF': 'border-l-emerald-600',
  'HP': 'border-l-orange-500',
  'PE': 'border-l-rose-600',
}
function modeBorder(mode?: string) { return MODE_BORDER[mode || ''] || 'border-l-slate-400' }

// 模式徽章背景（队列和弹窗用）
const MODE_BG: Record<string, string> = {
  'HD': 'bg-blue-600',
  'HDF': 'bg-violet-600',
  'HF': 'bg-emerald-600',
  'HP': 'bg-orange-500',
  'PE': 'bg-rose-600',
}
function modeBg(mode?: string) { return MODE_BG[mode || ''] || 'bg-slate-500' }

/** 过滤乱码/测试病区 */
function isValidWard(w: RestScheduleWard) {
  const n = (w.name || '').trim()
  if (!n || n.length < 2) return false
  if (n.includes('?') || n.includes('�')) return false
  if ([...n].every(ch => {
    const code = ch.charCodeAt(0)
    return code <= 31 || code === 127
  })) return false
  if (/SMOKE_TEST/i.test(n)) return false
  return true
}

// ─── 主组件 ───
export default function Schedule() {
  const [viewMode, setViewMode] = useState<'week'|'day'>('week')
  const [weekStart, setWeekStart] = useState(() => startOfWeek(new Date()))
  const [selectedDay, setSelectedDay] = useState(() => toDateString(new Date()))
  const [wardFilter, setWardFilter] = useState<'ALL' | number>('ALL')
  const [data, setData] = useState<RestScheduleWeekResponse|null>(null)
  const [loading, setLoading] = useState(false)
  const [queueSearch, setQueueSearch] = useState('')
  const [modeFilter, setModeFilter] = useState('ALL')
  const [queueVisible, setQueueVisible] = useState(true)
  const [selectedQueuePatient, setSelectedQueuePatient] = useState<number | null>(null)

  // 使用自定义 hook 管理弹窗状态
  const {
    modal, setModal,
    selPatient, setSelPatient,
    actionMenu, setActionMenu,
    applyTemplateOpen, setApplyTemplateOpen,
    moveModal, setMoveModal,
    moveLoading, setMoveLoading,
    treatModal, setTreatModal,
    treatments, setTreatments,
    treatLoading, setTreatLoading,
    historyModal, setHistoryModal,
    shiftHistory, setShiftHistory,
    historyLoading, setHistoryLoading,
  } = useScheduleModals()

  // 创建排班弹窗状态
  const [createModalOpen, setCreateModalOpen] = useState(false)
  const [createInitial, setCreateInitial] = useState<{
    date?: string
    wardId?: number
    bedId?: number
    shiftId?: number
    patientId?: number
  }>({})

  // 使用自定义 hook 管理拖拽状态
  const {
    dragItem,
    dragOverKey,
    onDragStart,
    onDragOver,
    onDropOnEmpty,
    onDropOnOccupied,
    onDragEnd,
    onDragLeave,
  } = useScheduleDragDrop()

  // 操作菜单（点击已排班卡片弹出）
  const actionMenuRef = useRef<HTMLDivElement>(null)

  // ─── 数据加载 ───
  const days = useMemo(() => Array.from({length:7},(_,i)=>addDays(weekStart,i)), [weekStart])
  const startDate = toDateString(days[0])
  const endDate = toDateString(days[6])

  const loadWeek = useCallback(async () => {
    setLoading(true)
    try {
      const res = await restApi.getScheduleWeek({
        startDate,
        endDate,
        wardId: wardFilter === 'ALL' ? undefined : wardFilter,
      })
      setData(res.data)
    } catch(e) { message.error(getErrorMessage(e)) }
    finally { setLoading(false) }
  }, [endDate, wardFilter, startDate])

  useEffect(() => { void loadWeek() }, [loadWeek])

  // ─── 派生数据 ───
  const wards = useMemo(() => (data?.wards||[]).filter(isValidWard), [data])
  const validWardIds = useMemo(() => new Set(wards.map(w => Number(w.id))), [wards])
  // 过滤掉无效床位：wardId=0 或不在有效病区列表中的
  const beds = useMemo(() => (data?.beds||[]).filter(b => {
    const wid = Number(b.wardId)
    return wid > 0 && validWardIds.has(wid)
  }), [data, validWardIds])
  const shifts = useMemo(() => data?.shifts || [], [data])
  const pShifts = useMemo(() => data?.patientShifts || [], [data])
  const pending = useMemo(() => data?.pendingPatients || [], [data])

  const schedMap = useMemo(() => {
    const m = new Map<string, RestScheduleWeekShift>()
    pShifts.forEach(s => m.set(`${s.bedId}-${s.treatmentTime.split('T')[0]}-${s.shiftId}`, s))
    return m
  }, [pShifts])

  const dayCountMap = useMemo(() => {
    const m = new Map<string, number>()
    pShifts.forEach(s => {
      const dt = s.treatmentTime.split('T')[0]
      m.set(dt, (m.get(dt) || 0) + 1)
    })
    return m
  }, [pShifts])

  const loadMap = useMemo(() => {
    const m = new Map<string,{used:number;total:number}>()
    beds.forEach(() => {
      days.forEach(d => {
        const dt = toDateString(d)
        shifts.forEach(s => {
          const k = `${dt}-${s.id}`
          const c = m.get(k) || {used:0,total:0}
          m.set(k, {...c, total: c.total+1})
        })
      })
    })
    pShifts.forEach(s => {
      const k = `${s.treatmentTime.split('T')[0]}-${s.shiftId}`
      const c = m.get(k)
      if(c) m.set(k, {...c, used: c.used+1})
    })
    return m
  }, [beds,shifts,days,pShifts])

  // 按病区分组
  const groupedBeds = useMemo(() => {
    const wardMap = new Map<string, RestScheduleWard>()
    wards.forEach(w => wardMap.set(String(w.id), w))
    const groups = new Map<string, {ward: RestScheduleWard|undefined; beds: RestScheduleBed[]}>()
    beds.forEach(bed => {
      const wid = String(bed.wardId)
      if (!groups.has(wid)) groups.set(wid, { ward: wardMap.get(wid), beds: [] })
      groups.get(wid)!.beds.push(bed)
    })
    return Array.from(groups.values())
  }, [beds, wards])

  const filteredPending = useMemo(() => {
    const kw = queueSearch.trim().toLowerCase()
    return pending.filter(p => {
      const mm = modeFilter==='ALL'||p.dialysisMode===modeFilter
      const t = `${p.name} ${p.spell||''} ${p.gender} ${p.dialysisMode} ${freq(p)}`.toLowerCase()
      return mm && (!kw || t.includes(kw))
    })
  }, [modeFilter,pending,queueSearch])

  const modeOpts = useMemo(() => {
    const s = new Set<string>()
    pending.forEach(p => { if(p.dialysisMode) s.add(p.dialysisMode) })
    pShifts.forEach(s2 => { if(s2.dialysisMode) s.add(s2.dialysisMode) })
    return ['ALL', ...Array.from(s).sort()]
  }, [pShifts,pending])

  // ─── 操作 ───
  const openModal = (bed: RestScheduleBed, date: string, shift: RestShift, existing?: RestScheduleWeekShift) => {
    const prePatient = selectedQueuePatient
    setModal({ open:true, bed, date, shift, existing:existing||null })
    setSelPatient(existing ? Number(existing.patientId) : (prePatient ?? undefined))
    setSelectedQueuePatient(null)
  }
  const closeModal = () => { setModal({open:false,bed:null,date:'',shift:null,existing:null}); setSelPatient(undefined) }

  const handleSave = async () => {
    const {bed,date,shift,existing} = modal
    if(!bed||!shift) return
    if(!selPatient) { message.error('请选择患者'); return }
    try {
      const p = pending.find(x=>x.id===selPatient)
      if(existing) {
        // 编辑已有排班：禁止换患者，只用 PUT
        await restApi.updatePatientShift(existing.id, {
          bedId: Number(bed.id),
          wardId: Number(bed.wardId),
          shiftId: shift.id,
          treatmentTime: toApi(date),
          patientPlanId: p?.patientPlanId,
          shiftTiming: 20,
        })
        message.success('排班已更新')
      } else {
        await restApi.createPatientShift({ patientId:selPatient, scheduleDate:toApi(date), shiftId:shift.id, bedId:Number(bed.id), wardId:Number(bed.wardId), dialysisMode:p?.dialysisMode||'HD', patientPlanId:p?.patientPlanId, shiftTiming:20, status:1 })
        message.success('排班已创建')
      }
      closeModal(); await loadWeek()
    } catch(e) { message.error(getErrorMessage(e)) }
  }

  const handleDelete = async (item: RestScheduleWeekShift) => {
    if (isDateLocked(item.treatmentTime.split('T')[0])) { message.warning('历史排班不可修改'); return }
    const ok = await new Promise<boolean>(r => Modal.confirm({ title:'取消排班', content:`确认取消 ${item.patientName} 的排班？`, okText:'确认', okButtonProps:{danger:true}, onOk:()=>r(true), onCancel:()=>r(false) }))
    if(!ok) return
    try { await restApi.deletePatientShift(item.id); message.success('已取消'); setActionMenu({visible:false,x:0,y:0,item:null}); await loadWeek() }
    catch(e) { message.error(getErrorMessage(e)) }
  }

  // ─── 操作菜单 ───
  const openActionMenu = (e: React.MouseEvent, item: RestScheduleWeekShift) => {
    e.preventDefault()
    e.stopPropagation()
    setActionMenu({ visible:true, x:e.clientX, y:e.clientY, item })
  }
  const closeActionMenu = useCallback(() => setActionMenu({visible:false, x:0, y:0, item:null}), [setActionMenu])

  // 点击外部关闭菜单
  useEffect(() => {
    if (!actionMenu.visible) return
    const handler = (e: MouseEvent) => {
      if (actionMenuRef.current && !actionMenuRef.current.contains(e.target as Node)) closeActionMenu()
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [actionMenu.visible, closeActionMenu])

  // ─── 换床 ───
  const openMoveModal = (item: RestScheduleWeekShift) => {
    closeActionMenu()
    setMoveModal({ open:true, item, targetBedId:undefined })
  }
  const handleMove = async () => {
    const { item, targetBedId } = moveModal
    if (!item || !targetBedId) { message.error('请选择目标床位'); return }
    const targetBed = beds.find(b => Number(b.id) === targetBedId)
    if (!targetBed) return
    setMoveLoading(true)
    try {
      await restApi.movePatientShift(item.id, { bedId: targetBedId, wardId: Number(targetBed.wardId) })
      message.success('换床成功')
      setMoveModal({open:false, item:null, targetBedId:undefined})
      await loadWeek()
    } catch(e) { message.error(getErrorMessage(e)) }
    finally { setMoveLoading(false) }
  }

  // ─── 治疗记录 ───
  const openTreatModal = async (item: RestScheduleWeekShift) => {
    closeActionMenu()
    setTreatModal({ open:true, patientId:Number(item.patientId), patientName:item.patientName })
    setTreatLoading(true)
    try {
      const res = await restApi.getTreatments({ patientId: String(item.patientId), pageSize:20 })
      setTreatments(res.data?.items || [])
    } catch(e) { message.error(getErrorMessage(e)); setTreatments([]) }
    finally { setTreatLoading(false) }
  }

  // ─── 换床记录（排班历史） ───
  const openHistoryModal = async (item: RestScheduleWeekShift) => {
    closeActionMenu()
    setHistoryModal({ open:true, patientId:Number(item.patientId), patientName:item.patientName })
    setHistoryLoading(true)
    try {
      const res = await restApi.getPatientShifts({ patientId:Number(item.patientId), pageSize:30 })
      setShiftHistory(res.data?.items || [])
    } catch(e) { message.error(getErrorMessage(e)); setShiftHistory([]) }
    finally { setHistoryLoading(false) }
  }



  // ─── 渲染 ───
  const visibleDays = viewMode==='day' ? [new Date(selectedDay+'T00:00:00')] : days
  const colCount = shifts.length || 1
  const secondRowTop = 30

  return (
    <div className="h-full bg-slate-50 p-2 flex flex-col gap-2">
      {/* ── 顶部工具栏 ── */}
      <div className="shrink-0 rounded-lg border border-slate-200 bg-white px-4 py-2 flex items-center justify-between gap-3 shadow-sm">
        <div className="flex items-center gap-3">
          <div className="rounded-lg bg-blue-600 p-1.5 text-white shadow"><CalendarIcon size={16}/></div>
          <div>
            <h2 className="text-sm font-black text-slate-800">排班管理</h2>
            <p className="text-meta font-bold text-slate-400">{startDate} ~ {endDate}</p>
          </div>
          <div className="ml-2 flex items-center rounded-lg border border-slate-200 bg-slate-50 p-0.5">
            <button onClick={()=>setWeekStart(p=>addDays(p,-7))} className="rounded p-1 text-slate-500 hover:bg-white hover:shadow-sm"><ChevronLeft size={14}/></button>
            <button onClick={()=>{setWeekStart(startOfWeek(new Date())); setSelectedDay(toDateString(new Date()))}} className="rounded bg-white px-2.5 py-0.5 text-meta font-black text-blue-600 shadow-sm">本周</button>
            <button onClick={()=>setWeekStart(p=>addDays(p,7))} className="rounded p-1 text-slate-500 hover:bg-white hover:shadow-sm"><ChevronRight size={14}/></button>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <div className="flex rounded-lg border border-slate-200 bg-slate-50 p-0.5">
            <button onClick={()=>setViewMode('week')} className={`flex items-center gap-1 rounded px-2.5 py-1 text-meta font-black transition ${viewMode==='week'?'bg-white text-blue-600 shadow-sm':'text-slate-400 hover:text-slate-600'}`}><LayoutGrid size={12}/>周</button>
            <button onClick={()=>setViewMode('day')} className={`flex items-center gap-1 rounded px-2.5 py-1 text-meta font-black transition ${viewMode==='day'?'bg-white text-blue-600 shadow-sm':'text-slate-400 hover:text-slate-600'}`}><List size={12}/>日</button>
          </div>
          <button onClick={()=>void loadWeek()} className="inline-flex items-center gap-1 rounded-lg border border-slate-200 px-2.5 py-1 text-meta font-black text-slate-600 hover:bg-slate-50"><RefreshCw size={12}/>刷新</button>
          <button
            onClick={() => {
              setCreateInitial({
                date: viewMode === 'day' ? selectedDay : toDateString(new Date()),
                patientId: selectedQueuePatient ?? undefined,
                wardId: wardFilter === 'ALL' ? undefined : Number(wardFilter),
              })
              setCreateModalOpen(true)
            }}
            className="inline-flex items-center gap-1 rounded-lg border border-blue-300 bg-blue-50 px-2.5 py-1 text-meta font-black text-blue-600 hover:bg-blue-100 hover:border-blue-400 transition"
          ><Plus size={12}/>新建排班</button>
          <div className="text-meta text-foreground-muted">
            本周完成度：<span className="font-semibold text-foreground">--</span>
          </div>
          <button onClick={()=>setApplyTemplateOpen(true)} className="inline-flex items-center gap-1 rounded-lg border border-slate-200 px-2 py-1 text-meta font-black text-slate-600 hover:bg-slate-50"><ClipboardList size={12}/>应用模板</button>
          <button onClick={()=>setQueueVisible(v=>!v)} className={`inline-flex items-center gap-1 rounded-lg border px-2 py-1 text-meta font-black transition ${queueVisible?'border-blue-300 bg-blue-50 text-blue-600':'border-slate-200 text-slate-500 hover:bg-slate-50'}`} title={queueVisible?'隐藏待排班':'显示待排班'}>
            {queueVisible ? <PanelRightClose size={13}/> : <PanelRightOpen size={13}/>}
          </button>
        </div>
      </div>

      {/* ── 病区筛选 + 日视图切换 ── */}
      <div className="shrink-0 flex items-center gap-2">
        <div className="flex items-center gap-1 overflow-x-auto flex-1">
          <button
            onClick={()=>setWardFilter('ALL')}
            className={`whitespace-nowrap rounded-lg border px-2.5 py-1 text-meta font-black transition ${wardFilter==='ALL'?'border-blue-500 bg-blue-50 text-blue-700':'border-slate-200 text-slate-500 hover:border-slate-300'}`}
          >全部病区</button>
          {wards.map(w => (
            <button
              key={w.id}
              onClick={()=>setWardFilter(Number(w.id))}
              className={`whitespace-nowrap rounded-lg border px-2.5 py-1 text-meta font-black transition ${wardFilter===Number(w.id)?'border-blue-500 bg-blue-50 text-blue-700':'border-slate-200 text-slate-500 hover:border-slate-300'}`}
            >{w.name}<span className="ml-0.5 text-[9px] opacity-50">{w.bedCount||0}床</span></button>
          ))}
        </div>
        {viewMode==='day' && (
          <div className="flex items-center rounded-lg border border-slate-200 bg-slate-50 p-0.5 shrink-0">
            <button onClick={()=>setSelectedDay(toDateString(addDays(new Date(selectedDay+'T00:00:00'),-1)))} className="rounded p-0.5 text-slate-500 hover:bg-white"><ChevronLeft size={13}/></button>
            <span className="px-2 text-meta font-black text-slate-700">{selectedDay} {WEEKDAYS[new Date(selectedDay+'T00:00:00').getDay()]}</span>
            <button onClick={()=>setSelectedDay(toDateString(addDays(new Date(selectedDay+'T00:00:00'),1)))} className="rounded p-0.5 text-slate-500 hover:bg-white"><ChevronRight size={13}/></button>
          </div>
        )}
      </div>

      {/* ── 主体 ── */}
      <div className="flex-1 min-h-0 flex gap-2">
        {/* 排班表格 */}
        <div className="flex-1 min-w-0 rounded-lg border border-slate-200 bg-white shadow-sm overflow-auto">
          <Spin spinning={loading}>
            <table
              className="w-full border-collapse"
              style={{minWidth: viewMode==='day' ? '400px' : `${62 + visibleDays.length * colCount * 88}px`}}
            >
              <thead>
                {/* 第一层：床位（rowSpan=2）+ 日期+已排班数 */}
                <tr>
                  <th
                    rowSpan={2}
                    className="sticky left-0 z-30 w-[62px] bg-slate-50 border-b border-r border-slate-200 px-1.5 py-1.5 text-meta font-black text-slate-400 align-middle text-center"
                  >床位</th>
                  {visibleDays.map(day => {
                    const dt = toDateString(day)
                    const today = isToday(dt)
                    const dayCount = dayCountMap.get(dt) || 0
                    return (
                      <th
                        key={dt}
                        colSpan={shifts.length || 1}
                        className={[
                          'sticky top-0 z-20 border-b border-r-2 border-slate-300 px-1.5 py-1 text-center whitespace-nowrap',
                          today ? 'bg-red-50 border-t-2 border-t-red-400' : 'bg-slate-50',
                        ].join(' ')}
                      >
                        <span className={`text-[12px] font-black ${today ? 'text-red-700' : 'text-slate-700'}`}>
                          {WEEKDAYS[day.getDay()]} {day.getMonth()+1}/{day.getDate()}
                        </span>
                        <span className={`ml-1 text-[9px] font-bold ${today ? 'text-red-400' : 'text-slate-400'}`}>
                          已排{dayCount}
                        </span>
                      </th>
                    )
                  })}
                </tr>
                {/* 第二层：班次名+容量 */}
                <tr>
                  {visibleDays.map(day => {
                    const dt = toDateString(day)
                    const today = isToday(dt)
                    const subBase = today ? 'bg-red-50/80' : 'bg-slate-50'
                    return shifts.length > 0 ? shifts.map((shift, si) => {
                      const load = loadMap.get(`${dt}-${shift.id}`)
                      const isLastShift = si === shifts.length - 1
                      return (
                        <th
                          key={`${dt}-${shift.id}`}
                          style={{top: secondRowTop, minWidth: 88}}
                          className={`sticky z-20 border-b px-1.5 py-0.5 text-center whitespace-nowrap ${subBase} ${isLastShift ? 'border-r-2 border-r-slate-300' : 'border-r border-r-slate-200'}`}
                        >
                          <span className={`text-meta font-black ${today ? 'text-red-600' : 'text-slate-600'}`}>{shift.name}</span>
                          <span className={`ml-0.5 text-[9px] font-bold ${today ? 'text-red-400' : 'text-slate-400'}`}>{load?.used||0}/{load?.total||0}</span>
                        </th>
                      )
                    }) : (
                      <th
                        key={`${dt}-empty`}
                        style={{top: secondRowTop}}
                        className={`sticky z-20 border-b border-r border-slate-200 px-1.5 py-0.5 text-meta font-black ${subBase} ${today ? 'text-red-500' : 'text-slate-400'}`}
                      >无班次</th>
                    )
                  })}
                </tr>
              </thead>
              <tbody>
                {groupedBeds.length === 0 && (
                  <tr>
                    <td colSpan={1 + visibleDays.length * colCount} className="py-12 text-center text-sm font-bold text-slate-400">暂无床位数据</td>
                  </tr>
                )}
                {groupedBeds.map(({ ward, beds: wbeds }) => (
                  <Fragment key={`ward-${ward?.id ?? 'unknown'}`}>
                    {/* 病区分组标题行 */}
                    <tr className="bg-slate-100/70">
                      <td colSpan={1 + visibleDays.length * colCount} className="px-2 py-1 border-b border-slate-200">
                        <span className="text-meta font-black text-slate-600 tracking-wide">{ward?.name ?? '未知病区'}</span>
                        <span className="ml-1.5 text-[9px] text-slate-400">{wbeds.length} 床</span>
                      </td>
                    </tr>
                    {/* 床位行 */}
                    {wbeds.map((bed, bi) => (
                      <tr key={bed.id} className={bi%2===0 ? '' : 'bg-slate-50/30'}>
                        <td className="sticky left-0 z-20 bg-inherit border-b border-r border-slate-200 px-1.5 py-0.5 text-center w-[62px]">
                          <div className="text-[12px] font-black text-slate-800 leading-tight">{bed.name}</div>
                        </td>
                        {visibleDays.map(day => {
                          const dt = toDateString(day)
                          const today = isToday(dt)
                          const cellBase = today ? 'bg-red-50/20' : ''
                          return shifts.length > 0 ? shifts.map((shift, si) => {
                            const cellKey = `${Number(bed.id)}-${dt}-${shift.id}`
                            const item = schedMap.get(cellKey)
                            const isLastShift = si === shifts.length - 1
                            const borderClass = isLastShift ? 'border-r-2 border-r-slate-300' : 'border-r border-r-slate-100'
                            const isDragOver = dragOverKey === cellKey
                            const isDragging = dragItem?.id === item?.id
                            return (
                              <td
                                key={cellKey}
                                className={`border-b p-0.5 align-middle ${borderClass} ${cellBase} ${isDragOver ? 'bg-blue-100/60' : ''}`}
                                style={{minWidth:88, height:48}}
                                onDragOver={(e)=>onDragOver(e, cellKey)}
                                onDragLeave={onDragLeave}
                              >
                                 {item ? (
                                  <div
                                    draggable={!isDateLocked(dt)}
                                    onDragStart={(e)=>onDragStart(e, item)}
                                    onDragEnd={onDragEnd}
                                    onDrop={(e)=>void onDropOnOccupied(e, item, loadWeek)}
                                    onContextMenu={(e)=>openActionMenu(e, item)}
                                    className={`group relative flex items-center justify-center rounded bg-surface border-l-4 ${modeBorder(item.dialysisMode)} text-foreground px-1 py-0.5 ${isDateLocked(dt) ? 'opacity-70 cursor-not-allowed' : 'cursor-grab active:cursor-grabbing'} h-full hover:bg-surface-sunken hover:shadow-md transition-all ${isDragging ? 'opacity-40 ring-2 ring-blue-400' : ''}`}
                                    title={`${item.patientName} · ${item.dialysisMode||''} · ${item.statusName||''}${isDateLocked(dt) ? ' · 历史锁定' : ''}`}
                                  >
                                    <div className="text-center leading-tight w-full select-none">
                                      <div className="truncate text-body font-medium">{item.patientName}</div>
                                      <div className="text-meta text-foreground-muted">{item.dialysisMode||''}</div>
                                    </div>
                                    {item.sourceType === 'temporary' ? (
                                      <span className="absolute -top-1 -right-1 text-[7px] bg-orange-500 text-white rounded px-0.5 font-bold">临</span>
                                    ) : item.isManualAdjusted ? (
                                      <span className="absolute -top-1 -right-1 text-[7px] bg-amber-100 text-amber-600 rounded px-0.5 font-bold">调整</span>
                                    ) : item.sourceType === 'contract' ? (
                                      <span className="absolute -top-1 -right-1 text-[7px] bg-emerald-100 text-emerald-600 rounded px-0.5 font-bold">合约</span>
                                    ) : item.sourceType === 'template' ? (
                                      <span className="absolute -top-1 -right-1 text-[7px] bg-white/90 text-slate-500 rounded px-0.5 font-bold">模板</span>
                                    ) : null}
                                  </div>
                                ) : (
                                  <div
                                    className={[
                                      'flex h-full items-center justify-center rounded border border-dashed transition-all',
                                      isDateLocked(dt)
                                        ? 'border-slate-200 bg-slate-100/50 text-slate-300 cursor-not-allowed'
                                        : isDragOver
                                          ? 'border-blue-400 bg-blue-100/60 text-blue-500 scale-105 cursor-pointer'
                                          : selectedQueuePatient
                                            ? 'border-green-400 bg-green-50/40 text-green-400 hover:bg-green-50/80 cursor-pointer'
                                            : today
                                              ? 'border-red-200 text-red-200 hover:border-red-300 hover:bg-red-50/40 hover:text-red-300 cursor-pointer'
                                              : 'border-slate-200 text-slate-200 hover:border-blue-300 hover:bg-blue-50/50 hover:text-blue-400 cursor-pointer',
                                    ].join(' ')}
                                    onClick={() => {
                                      if (isDateLocked(dt)) return
                                      setCreateInitial({
                                        date: dt,
                                        wardId: Number(bed.wardId),
                                        bedId: Number(bed.id),
                                        shiftId: shift.id,
                                        patientId: selectedQueuePatient ?? undefined,
                                      })
                                      setCreateModalOpen(true)
                                    }}
                                    onDrop={(e)=>!isDateLocked(dt) && void onDropOnEmpty(e, bed, dt, shift, loadWeek)}
                                  >
                                    <Plus size={11}/>
                                  </div>
                                )}
                              </td>
                            )
                          }) : (
                            <td key={`${bed.id}-${dt}-empty`} className="border-b border-r border-slate-100 text-center text-[9px] font-bold text-slate-300">-</td>
                          )
                        })}
                      </tr>
                    ))}
                  </Fragment>
                ))}
              </tbody>
            </table>
          </Spin>
        </div>

        {/* ── 右侧面板：待排班队列（可隐藏） ── */}
        {queueVisible && (
          <aside className="w-[220px] shrink-0 rounded-lg border border-slate-200 bg-white shadow-sm flex flex-col min-h-0">
            <div className="border-b border-slate-100 px-2 py-1.5">
              <div className="flex items-center justify-between mb-1">
                <p className="text-meta font-black uppercase tracking-widest text-slate-400">待排班</p>
                {/* eslint-disable-next-line no-restricted-syntax -- density:strict 故意小字（待排班数角标） */}
                <span className="rounded-full bg-blue-100 px-1.5 py-0 text-[10px] font-black text-blue-600">{filteredPending.length}</span>
              </div>
              <div className="relative">
                <Search size={12} className="absolute left-2 top-1/2 -translate-y-1/2 text-slate-400"/>
                <input value={queueSearch} onChange={e=>setQueueSearch(e.target.value)} placeholder="搜索患者" className="h-7 w-full rounded border border-slate-200 bg-white pl-7 pr-2 text-meta font-bold outline-none focus:border-blue-400"/>
              </div>
              <div className="mt-1 flex flex-wrap gap-0.5">
                {modeOpts.map(m => (
                  <button key={m} onClick={()=>setModeFilter(m)} className={`rounded px-1.5 py-0.5 text-[9px] font-black ${modeFilter===m?'bg-blue-600 text-white':'bg-white text-slate-400 border border-slate-100'}`}>{m==='ALL'?'全部':m}</button>
                ))}
              </div>
            </div>

            {selectedQueuePatient != null && (
              <div className="mx-1.5 mt-1 rounded bg-green-50 border border-green-300 px-1.5 py-0.5 flex items-center justify-between">
                <span className="text-meta font-black text-green-700">已选·点空格排入</span>
                <button onClick={()=>setSelectedQueuePatient(null)} className="text-green-500 hover:text-green-700"><Plus size={11} className="rotate-45"/></button>
              </div>
            )}

            <div className="flex-1 min-h-0 overflow-y-auto p-1 space-y-0.5">
              {filteredPending.map(p => (
                <div
                  key={p.id}
                  className={[
                    'relative overflow-hidden rounded border px-1 py-0.5 cursor-pointer transition-all',
                    selectedQueuePatient === p.id
                      ? 'border-green-500 bg-green-50 shadow-sm'
                      : 'border-slate-200 bg-white hover:bg-surface-sunken',
                  ].join(' ')}
                  onClick={() => {
                    if (selectedQueuePatient === p.id) {
                      setSelectedQueuePatient(null)
                    } else {
                      setSelectedQueuePatient(p.id)
                      if (modal.open && !modal.existing) setSelPatient(p.id)
                    }
                  }}
                >
                  <div className={`absolute left-0 top-0 bottom-0 w-[3px] rounded-l ${modeBg(p.dialysisMode)}`}/>
                  <div className="pl-1">
                    <div className="flex items-center gap-1">
                      <span className="truncate text-[13px] font-black text-slate-800">{p.name}</span>
                      <span className="shrink-0 text-meta font-bold text-slate-400">{p.gender||''}</span>
                    </div>
                    <div className="flex gap-1 mt-0.5">
                      <span className={`rounded px-1.5 py-0 text-meta font-black text-white ${modeBg(p.dialysisMode)}`}>{p.dialysisMode||'--'}</span>
                      <span className="text-meta font-bold text-slate-400">{freq(p)}</span>
                      {(p.remainingTimes ?? 0) > 0 && (
                        <span className="text-meta font-black text-orange-500">剩 {p.remainingTimes} 次</span>
                      )}
                    </div>
                  </div>
                </div>
              ))}
              {filteredPending.length===0 && (
                <div className="rounded border border-dashed border-slate-200 p-4 text-center text-meta font-bold text-slate-400">暂无待排班患者</div>
              )}
            </div>
          </aside>
        )}
      </div>

      {/* ── 快速排班弹窗 ── */}
      {modal.open && modal.bed && modal.shift && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/40 backdrop-blur-sm p-4" onClick={closeModal}>
          <div className="bg-white rounded-lg shadow-2xl w-full max-w-md overflow-hidden" onClick={e=>e.stopPropagation()}>
            <div className="px-5 py-3 border-b border-slate-100 flex items-center justify-between">
              <div>
                <h3 className="text-sm font-black text-slate-800">{modal.existing ? '修改排班' : '新建排班'}</h3>
                <p className="text-meta text-slate-400 font-bold mt-0.5">{modal.bed.name} · {modal.shift.name} · {modal.date}</p>
              </div>
              <button onClick={closeModal} className="p-1.5 hover:bg-slate-100 rounded-lg text-slate-400"><Plus size={16} className="rotate-45"/></button>
            </div>
            <div className="px-5 py-3 space-y-3">
              <div>
                <label className="text-meta font-black text-slate-500 block mb-1.5">选择患者</label>
                <select value={selPatient||''} onChange={e=>setSelPatient(Number(e.target.value)||undefined)} disabled={!!modal.existing} className="w-full h-8 rounded-lg border border-slate-200 px-3 text-xs font-bold outline-none focus:border-blue-400 disabled:bg-slate-100 disabled:cursor-not-allowed">
                  <option value="">-- 请选择 --</option>
                  {pending.map(p => <option key={p.id} value={p.id}>{p.name} ({p.dialysisMode||'--'}) {freq(p)}</option>)}
                  {modal.existing && !pending.find(p=>p.id===Number(modal.existing!.patientId)) && (
                    <option value={Number(modal.existing.patientId)}>{modal.existing.patientName} (已排班)</option>
                  )}
                </select>
              </div>
              {modal.existing && (
                <div className="rounded-lg bg-amber-50 border border-amber-200 px-3 py-2 text-meta font-bold text-amber-700">
                  当前已排班：{modal.existing.patientName}（{modal.existing.dialysisMode}）
                </div>
              )}
            </div>
            <div className="px-5 py-2.5 bg-slate-50 border-t border-slate-100 flex justify-between">
              {modal.existing && (
                <button onClick={()=>void handleDelete(modal.existing!)} className="px-3 py-1.5 text-meta font-black text-red-600 hover:bg-red-50 rounded-lg transition">取消排班</button>
              )}
              <div className="flex gap-2 ml-auto">
                <button onClick={closeModal} className="px-4 py-1.5 bg-white border border-slate-200 text-slate-600 rounded-lg text-meta font-black hover:bg-slate-50">关闭</button>
                <button onClick={()=>void handleSave()} className="px-5 py-1.5 bg-blue-600 text-white rounded-lg text-meta font-black hover:bg-blue-700 shadow-lg shadow-blue-100">保存</button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* ── 操作菜单（点击已排班卡片弹出） ── */}
      {actionMenu.visible && actionMenu.item && (
        <div
          ref={actionMenuRef}
          className="fixed z-[60] bg-white rounded-lg shadow-2xl border border-slate-200 py-1 min-w-[140px] overflow-hidden"
          style={{ left: actionMenu.x, top: actionMenu.y }}
        >
          <div className="px-3 py-1.5 border-b border-slate-100">
            <p className="text-meta font-black text-slate-800 truncate">{actionMenu.item.patientName}</p>
            <p className="text-[9px] text-slate-400">{actionMenu.item.dialysisMode} · {actionMenu.item.statusName}</p>
          </div>
          <button
            className="w-full flex items-center gap-2 px-3 py-1.5 text-meta font-bold text-slate-600 hover:bg-blue-50 hover:text-blue-700 transition"
            onClick={()=>{
              const itm = actionMenu.item!
              const bed = beds.find(b=>Number(b.id)===itm.bedId)
              const shift = shifts.find(s=>s.id===itm.shiftId)
              if(bed && shift) { closeActionMenu(); openModal(bed, itm.treatmentTime.split('T')[0], shift, itm) }
              else { message.error('无法获取床位或班次信息'); closeActionMenu() }
            }}
          ><Plus size={12}/>修改排班</button>
          <button
            className="w-full flex items-center gap-2 px-3 py-1.5 text-meta font-bold text-slate-600 hover:bg-violet-50 hover:text-violet-700 transition"
            onClick={()=>void openMoveModal(actionMenu.item!)}
          ><ArrowRightLeft size={12}/>换床</button>
          <button
            className="w-full flex items-center gap-2 px-3 py-1.5 text-meta font-bold text-slate-600 hover:bg-emerald-50 hover:text-emerald-700 transition"
            onClick={()=>void openTreatModal(actionMenu.item!)}
          ><ClipboardList size={12}/>治疗记录</button>
          <button
            className="w-full flex items-center gap-2 px-3 py-1.5 text-meta font-bold text-slate-600 hover:bg-amber-50 hover:text-amber-700 transition"
            onClick={()=>void openHistoryModal(actionMenu.item!)}
          ><History size={12}/>换床记录</button>
          <div className="border-t border-slate-100 mt-1"/>
          <button
            className="w-full flex items-center gap-2 px-3 py-1.5 text-meta font-bold text-red-500 hover:bg-red-50 transition"
            onClick={()=>void handleDelete(actionMenu.item!)}
          ><Trash2 size={12}/>取消排班</button>
        </div>
      )}

      {/* ── 换床弹窗 ── */}
      {moveModal.open && moveModal.item && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/40 backdrop-blur-sm p-4" onClick={()=>setMoveModal({open:false,item:null,targetBedId:undefined})}>
          <div className="bg-white rounded-lg shadow-2xl w-full max-w-sm overflow-hidden" onClick={e=>e.stopPropagation()}>
            <div className="px-5 py-3 border-b border-slate-100 flex items-center justify-between">
              <div>
                <h3 className="text-sm font-black text-slate-800">换床</h3>
                <p className="text-meta text-slate-400 font-bold mt-0.5">{moveModal.item.patientName} · {moveModal.item.treatmentTime.split('T')[0]} · {shifts.find(s=>s.id===moveModal.item!.shiftId)?.name||''}</p>
              </div>
              <button onClick={()=>setMoveModal({open:false,item:null,targetBedId:undefined})} className="p-1.5 hover:bg-slate-100 rounded-lg text-slate-400"><X size={16}/></button>
            </div>
            <div className="px-5 py-3 space-y-3">
              <div className="rounded-lg bg-slate-50 border border-slate-200 px-3 py-2 text-meta font-bold text-slate-500">
                当前床位：<span className="text-slate-800 font-black">{moveModal.item.bedName}</span>
              </div>
              <div>
                <label className="text-meta font-black text-slate-500 block mb-1.5">选择目标床位</label>
                <select
                  value={moveModal.targetBedId||''}
                  onChange={e=>setMoveModal(m=>({...m, targetBedId:Number(e.target.value)||undefined}))}
                  className="w-full h-8 rounded-lg border border-slate-200 px-3 text-xs font-bold outline-none focus:border-blue-400"
                >
                  <option value="">-- 请选择 --</option>
                  {groupedBeds.map(({ward, beds:wbeds}) => (
                    <optgroup key={ward?.id??'x'} label={ward?.name??'未知'}>
                      {wbeds.filter(b=>Number(b.id)!==moveModal.item!.bedId).map(b=>(
                        <option key={b.id} value={Number(b.id)}>{b.name}</option>
                      ))}
                    </optgroup>
                  ))}
                </select>
              </div>
            </div>
            <div className="px-5 py-2.5 bg-slate-50 border-t border-slate-100 flex justify-end gap-2">
              <button onClick={()=>setMoveModal({open:false,item:null,targetBedId:undefined})} className="px-4 py-1.5 bg-white border border-slate-200 text-slate-600 rounded-lg text-meta font-black hover:bg-slate-50">取消</button>
              <button onClick={()=>void handleMove()} disabled={moveLoading} className="px-5 py-1.5 bg-violet-600 text-white rounded-lg text-meta font-black hover:bg-violet-700 shadow-lg shadow-violet-100 disabled:opacity-50">{moveLoading?'换床中...':'确认换床'}</button>
            </div>
          </div>
        </div>
      )}

      {/* ── 治疗记录弹窗 ── */}
      {treatModal.open && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/40 backdrop-blur-sm p-4" onClick={()=>setTreatModal({open:false,patientId:undefined,patientName:''})}>
          <div className="bg-white rounded-lg shadow-2xl w-full max-w-2xl max-h-[80vh] overflow-hidden flex flex-col" onClick={e=>e.stopPropagation()}>
            <div className="px-5 py-3 border-b border-slate-100 flex items-center justify-between shrink-0">
              <div>
                <h3 className="text-sm font-black text-slate-800">治疗记录</h3>
                <p className="text-meta text-slate-400 font-bold mt-0.5">{treatModal.patientName}</p>
              </div>
              <button onClick={()=>setTreatModal({open:false,patientId:undefined,patientName:''})} className="p-1.5 hover:bg-slate-100 rounded-lg text-slate-400"><X size={16}/></button>
            </div>
            <div className="flex-1 min-h-0 overflow-auto p-4">
              <Spin spinning={treatLoading}>
                <Table
                  size="small"
                  rowKey="id"
                  pagination={false}
                  dataSource={treatments}
                  locale={{emptyText:'暂无治疗记录'}}
                  columns={[
                    {title:'日期', dataIndex:'treatmentDate', width:100, render:(v:string)=><span className="text-meta font-bold">{v?.split('T')[0]}</span>},
                    {title:'班次', dataIndex:'shiftName', width:60, render:(v:string)=><span className="text-meta">{v||'--'}</span>},
                    {title:'类型', dataIndex:'treatmentType', width:60, render:(v:string)=><span className="text-meta">{v||'--'}</span>},
                    {title:'状态', dataIndex:'status', width:70, render:(v:number)=>{
                      const m:Record<number,{c:string;t:string}>={0:{c:'default',t:'待开始'},1:{c:'processing',t:'进行中'},2:{c:'success',t:'已完成'},3:{c:'error',t:'已取消'}}
                      const s=m[v]||{c:'default',t:'未知'}
                      return <Tag color={s.c} className="text-meta">{s.t}</Tag>
                    }},
                    {title:'医生', dataIndex:'doctorName', width:70, render:(v:string)=><span className="text-meta">{v||'--'}</span>},
                    {title:'时长', dataIndex:'durationMinutes', width:60, render:(v:number)=><span className="text-meta">{v?`${v}分`:'--'}</span>},
                    {title:'备注', dataIndex:'notes', ellipsis:true, render:(v:string)=><span className="text-meta">{v||'--'}</span>},
                  ]}
                />
              </Spin>
            </div>
          </div>
        </div>
      )}

      {/* ── 换床记录弹窗 ── */}
      {historyModal.open && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/40 backdrop-blur-sm p-4" onClick={()=>setHistoryModal({open:false,patientId:undefined,patientName:''})}>
          <div className="bg-white rounded-lg shadow-2xl w-full max-w-2xl max-h-[80vh] overflow-hidden flex flex-col" onClick={e=>e.stopPropagation()}>
            <div className="px-5 py-3 border-b border-slate-100 flex items-center justify-between shrink-0">
              <div>
                <h3 className="text-sm font-black text-slate-800">排班 / 换床记录</h3>
                <p className="text-meta text-slate-400 font-bold mt-0.5">{historyModal.patientName}</p>
              </div>
              <button onClick={()=>setHistoryModal({open:false,patientId:undefined,patientName:''})} className="p-1.5 hover:bg-slate-100 rounded-lg text-slate-400"><X size={16}/></button>
            </div>
            <div className="flex-1 min-h-0 overflow-auto p-4">
              <Spin spinning={historyLoading}>
                <Table
                  size="small"
                  rowKey="id"
                  pagination={false}
                  dataSource={shiftHistory}
                  locale={{emptyText:'暂无排班记录'}}
                  columns={[
                    {title:'日期', dataIndex:'scheduleDate', width:100, render:(v:string)=><span className="text-meta font-bold">{v?.split('T')[0]}</span>},
                    {title:'班次', dataIndex:['shift','name'], width:60, render:(_: unknown, record: RestPatientShift)=><span className="text-meta">{record.shift?.name||'--'}</span>},
                    {title:'床位', dataIndex:['bed','name'], width:60, render:(_: unknown, record: RestPatientShift)=><span className="text-meta font-black">{record.bed?.name||'--'}</span>},
                    {title:'病区', dataIndex:['ward','name'], width:80, render:(_: unknown, record: RestPatientShift)=><span className="text-meta">{record.ward?.name||'--'}</span>},
                    {title:'状态', dataIndex:'status', width:70, render:(v:number)=>{
                      const m:Record<number,{c:string;t:string}>={0:{c:'default',t:'待确认'},1:{c:'processing',t:'已确认'},3:{c:'success',t:'已完成'},4:{c:'error',t:'已取消'},6:{c:'warning',t:'转出'}}
                      const s=m[v]||{c:'default',t:`${v}`}
                      return <Tag color={s.c} className="text-meta">{s.t}</Tag>
                    }},
                    {title:'创建时间', dataIndex:'createTime', width:140, render:(v:string)=><span className="text-meta text-slate-400">{v?.replace('T',' ').slice(0,16)}</span>},
                  ]}
                />
              </Spin>
            </div>
          </div>
        </div>
      )}

      {/* 应用模板弹窗 */}
      <ApplyTemplateModal
        open={applyTemplateOpen}
        onClose={() => setApplyTemplateOpen(false)}
        onSuccess={() => void loadWeek()}
        wardId={wardFilter === 'ALL' ? undefined : Number(wardFilter)}
      />

      {/* 创建排班弹窗 */}
      <CreateScheduleModal
        open={createModalOpen}
        onClose={() => setCreateModalOpen(false)}
        onSuccess={() => void loadWeek()}
        initialDate={createInitial.date}
        initialWardId={createInitial.wardId}
        initialBedId={createInitial.bedId}
        initialShiftId={createInitial.shiftId}
        initialPatientId={createInitial.patientId}
        wards={wards}
        beds={beds}
        shifts={shifts}
        pendingPatients={pending}
      />
    </div>
  )
}
