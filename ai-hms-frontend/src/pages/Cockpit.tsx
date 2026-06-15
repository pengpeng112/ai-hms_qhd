// Cockpit - 今日治疗驾驶舱（护士床位墙 v1）
// 设计：打开系统先看「我今天要做的事」——读当日治疗状态出工作流卡墙，点卡进执行。
// 不依赖医嘱粒度/DBA，仅读老库治疗状态(0待上机/1透中/2已下机/3中断)+ join 患者床位。
// 待下机(🟡)、透前(🟠)、当班过滤、医生墙 为后续 refinement（依赖派生超滤/老库原始状态/人力排班）。

import { useEffect, useMemo, useState, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { Activity, Clock, RefreshCw, ArrowRight, AlertTriangle, BedDouble } from 'lucide-react'
import { restApi } from '@/services'

type BedState = 'pending' | 'preTreatment' | 'inProgress' | 'readyOff' | 'completed' | 'interrupted'

interface BedCard {
  treatmentId: number
  patientId: string
  patientName: string
  bedNumber: string
  mode: string
  shiftName: string
  queueNo: string
  doctorName: string
  startTime?: string
  state: BedState
}

// 治疗状态(应用层 0/1/2/3) → 床位墙工作流卡态
const STATE_META: Record<BedState, { label: string; dot: string; chip: string; card: string; sort: number }> = {
  pending:      { label: '待上机', dot: 'bg-sky-500',     chip: 'bg-sky-50 text-sky-600 border-sky-200',             card: 'border-l-sky-400',     sort: 0 },
  preTreatment: { label: '透前',   dot: 'bg-orange-500',  chip: 'bg-orange-50 text-orange-600 border-orange-200',     card: 'border-l-orange-400',  sort: 1 },
  inProgress:   { label: '透析中', dot: 'bg-emerald-500', chip: 'bg-emerald-50 text-emerald-600 border-emerald-200', card: 'border-l-emerald-400', sort: 2 },
  readyOff:     { label: '待下机', dot: 'bg-amber-500',   chip: 'bg-amber-50 text-amber-600 border-amber-200',       card: 'border-l-amber-400',   sort: 3 },
  completed:    { label: '已下机', dot: 'bg-slate-400',   chip: 'bg-slate-50 text-slate-500 border-slate-200',       card: 'border-l-slate-300',   sort: 4 },
  interrupted:  { label: '中断',   dot: 'bg-rose-500',    chip: 'bg-rose-50 text-rose-600 border-rose-200',          card: 'border-l-rose-400',    sort: 5 },
}

// 透中且「时长到」(已透≥处方时长) → 派生待下机(🟡)
function isTimeDone(startTime: string | undefined, durationMinutes: number | undefined): boolean {
  if (!startTime || !durationMinutes || durationMinutes <= 0) return false
  const elapsed = (Date.now() - new Date(startTime).getTime()) / 60000
  return elapsed >= durationMinutes
}

// 卡态判定：优先老库原始状态码(10/20/30/40/50/60),回退应用层 status(0/1/2/3)
function deriveState(status: number, legacy: string, startTime?: string, durationMinutes?: number): BedState {
  const l = (legacy || '').trim()
  if (status === 3 || l === '50') return 'interrupted'
  if (status === 2 || l === '40' || l === '60' || l === '100') return 'completed'
  if (l === '20') return 'preTreatment'
  if (status === 1 || l === '30') {
    return isTimeDone(startTime, durationMinutes) ? 'readyOff' : 'inProgress'
  }
  return 'pending'
}

function todayParam(): string {
  const d = new Date()
  const p = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())}`
}

function fmtTime(iso?: string): string {
  if (!iso) return '--:--'
  const d = new Date(iso)
  if (isNaN(d.getTime())) return '--:--'
  const p = (n: number) => String(n).padStart(2, '0')
  return `${p(d.getHours())}:${p(d.getMinutes())}`
}

export default function Cockpit() {
  const navigate = useNavigate()
  const [cards, setCards] = useState<BedCard[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [lastRefresh, setLastRefresh] = useState<Date>(new Date())

  const loadData = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const today = todayParam()
      const [treatRes, patRes] = await Promise.all([
        restApi.getTreatments({ treatmentDate: today, pageSize: 200 }),
        restApi.getPatientList({ page: 1, pageSize: 500 }),
      ])

      const patientMap = new Map<string, { name: string; bed: string; mode: string }>()
      for (const p of patRes.data.items || []) {
        patientMap.set(String(p.id), {
          name: p.name || `#${p.id}`,
          bed: p.bedNumber || '',
          mode: p.defaultMode || '',
        })
      }

      const next: BedCard[] = (treatRes.data.items || []).map((t) => {
        const pinfo = patientMap.get(String(t.patientId))
        return {
          treatmentId: t.id,
          patientId: String(t.patientId),
          patientName: pinfo?.name || `#${t.patientId}`,
          bedNumber: pinfo?.bed || '--',
          mode: pinfo?.mode || '',
          shiftName: t.shiftName || '',
          queueNo: t.queueNo || '',
          doctorName: t.doctorName || '',
          startTime: t.startTime,
          state: deriveState(t.status, t.legacyStatus || '', t.startTime, t.durationMinutes),
        }
      })

      next.sort((a, b) => {
        const s = STATE_META[a.state].sort - STATE_META[b.state].sort
        if (s !== 0) return s
        return a.bedNumber.localeCompare(b.bedNumber, 'zh')
      })

      setCards(next)
      setLastRefresh(new Date())
    } catch (e) {
      setError(e instanceof Error ? e.message : '加载失败')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { loadData() }, [loadData])

  const counts = useMemo(() => {
    const c: Record<BedState, number> = { pending: 0, preTreatment: 0, inProgress: 0, readyOff: 0, completed: 0, interrupted: 0 }
    for (const card of cards) c[card.state]++
    return c
  }, [cards])

  const goExecute = (patientId: string) => navigate(`/dialysis-processing?patientId=${encodeURIComponent(patientId)}`)

  return (
    <div className="h-full overflow-y-auto bg-slate-50 p-6">
      {/* Header */}
      <div className="flex items-center justify-between mb-5">
        <div className="flex items-center gap-3">
          <div className="w-11 h-11 rounded-2xl bg-blue-600 flex items-center justify-center shadow-sm">
            <Activity className="text-white" size={22} />
          </div>
          <div>
            <h1 className="text-xl font-black text-slate-800">今日治疗驾驶舱</h1>
            <p className="text-xs text-slate-400 mt-0.5">
              {todayParam()} · 护士床位墙 · 共 {cards.length} 位患者
              <span className="ml-2 text-slate-300">更新于 {fmtTime(lastRefresh.toISOString())}</span>
            </p>
          </div>
        </div>
        <button
          type="button"
          onClick={loadData}
          disabled={loading}
          className="flex items-center gap-2 px-4 py-2 rounded-xl border border-slate-200 bg-white text-sm font-bold text-slate-600 hover:bg-slate-50 disabled:opacity-50 transition-all"
        >
          <RefreshCw size={15} className={loading ? 'animate-spin' : ''} /> 刷新
        </button>
      </div>

      {/* Summary */}
      <div className="grid grid-cols-3 lg:grid-cols-6 gap-3 mb-5">
        {(['pending', 'preTreatment', 'inProgress', 'readyOff', 'completed', 'interrupted'] as BedState[]).map((s) => (
          <div key={s} className="bg-white rounded-2xl border border-slate-100 px-4 py-3.5 flex items-center justify-between">
            <div className="flex items-center gap-2">
              <span className={`w-2.5 h-2.5 rounded-full ${STATE_META[s].dot}`} />
              <span className="text-[13px] font-bold text-slate-500">{STATE_META[s].label}</span>
            </div>
            <span className="text-2xl font-black text-slate-800 tabular-nums">{counts[s]}</span>
          </div>
        ))}
      </div>

      {/* Error */}
      {error && (
        <div className="flex items-center gap-2 rounded-2xl border border-rose-200 bg-rose-50 px-5 py-4 mb-5 text-sm">
          <AlertTriangle size={18} className="text-rose-500 shrink-0" />
          <span className="text-rose-600 font-medium">{error}</span>
          <button type="button" onClick={loadData} className="ml-auto text-rose-600 font-bold underline">重试</button>
        </div>
      )}

      {/* Bed wall */}
      {loading && cards.length === 0 ? (
        <div className="py-24 text-center text-slate-400 font-bold">加载中…</div>
      ) : cards.length === 0 && !error ? (
        <div className="py-24 flex flex-col items-center text-slate-300">
          <BedDouble size={40} className="mb-3" />
          <p className="font-bold">今日暂无治疗安排</p>
        </div>
      ) : (
        <div className="grid grid-cols-2 md:grid-cols-3 xl:grid-cols-4 gap-4">
          {cards.map((card) => {
            const meta = STATE_META[card.state]
            return (
              <button
                type="button"
                key={card.treatmentId}
                onClick={() => goExecute(card.patientId)}
                className={`text-left bg-white rounded-2xl border border-slate-100 border-l-4 ${meta.card} p-4 hover:shadow-lg hover:-translate-y-0.5 transition-all group`}
              >
                <div className="flex items-start justify-between mb-3">
                  <div className="flex items-center gap-2.5 min-w-0">
                    <div className="w-10 h-10 rounded-xl bg-slate-700 text-white flex items-center justify-center font-black text-sm shrink-0">
                      {card.bedNumber}
                    </div>
                    <div className="min-w-0">
                      <p className="font-black text-slate-800 truncate">{card.patientName}</p>
                      <p className="text-[11px] text-slate-400 truncate">
                        {[card.mode, card.shiftName].filter(Boolean).join(' · ') || '—'}
                      </p>
                    </div>
                  </div>
                  <span className={`shrink-0 px-2 py-0.5 rounded-md text-[11px] font-black border ${meta.chip}`}>
                    {meta.label}
                  </span>
                </div>
                <div className="flex items-center justify-between text-[12px] text-slate-400">
                  <span className="flex items-center gap-1">
                    <Clock size={12} />
                    {card.state === 'pending' ? (card.queueNo ? `队列 ${card.queueNo}` : '待上机') : fmtTime(card.startTime)}
                  </span>
                  <span className="flex items-center gap-1 text-blue-500 font-bold opacity-0 group-hover:opacity-100 transition-opacity">
                    进执行 <ArrowRight size={13} />
                  </span>
                </div>
              </button>
            )
          })}
        </div>
      )}
    </div>
  )
}
