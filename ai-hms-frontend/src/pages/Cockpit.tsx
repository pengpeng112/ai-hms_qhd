// Cockpit - 今日治疗驾驶舱（护士床位墙 + 医生分诊墙 v1）
// 设计：打开系统先看「我今天要做的事」。护士镜头=床位工作流卡墙；医生镜头=分诊决策卡墙。
// 数据：读当日治疗(状态+透前体征) + 患者(床位) + 当日处方状态(批量)，全部聚合，点卡进执行/处方。
// 后续 refinement：报警卡(依赖监控告警)、当班过滤(依赖人力排班)、执行/处方深链(执行页未读URL参)。

import { useEffect, useMemo, useState, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { message } from 'antd'
import { Activity, Clock, RefreshCw, ArrowRight, AlertTriangle, BedDouble, Stethoscope, Users, PenLine, LogIn } from 'lucide-react'
import { restApi, getErrorMessage } from '@/services'
import { getMyDuties, getCheckInStatus, checkIn as checkInApi, type ResolvedDuty } from '@/services/smartScheduleApi'
import InfectiousAlertCards from '@/components/infectious/InfectiousAlertCards'
import WqAlertCards from '@/components/water-quality/WqAlertCards'
import VascularAlertCards from '@/components/vascular/VascularAlertCards'
import ActrButton from '@/components/actr/ActrButton'
import AdverseAlertCards from '@/components/adverse/AdverseAlertCards'

// ===== 护士镜头：床位工作流卡态 =====
type BedState = 'pending' | 'preTreatment' | 'inProgress' | 'readyOff' | 'completed' | 'interrupted'

const STATE_META: Record<BedState, { label: string; dot: string; chip: string; card: string; sort: number }> = {
  pending:      { label: '待上机', dot: 'bg-sky-500',     chip: 'bg-sky-50 text-sky-600 border-sky-200',             card: 'border-l-sky-400',     sort: 0 },
  preTreatment: { label: '透前',   dot: 'bg-orange-500',  chip: 'bg-orange-50 text-orange-600 border-orange-200',     card: 'border-l-orange-400',  sort: 1 },
  inProgress:   { label: '透析中', dot: 'bg-emerald-500', chip: 'bg-emerald-50 text-emerald-600 border-emerald-200', card: 'border-l-emerald-400', sort: 2 },
  readyOff:     { label: '待下机', dot: 'bg-amber-500',   chip: 'bg-amber-50 text-amber-600 border-amber-200',       card: 'border-l-amber-400',   sort: 3 },
  completed:    { label: '已下机', dot: 'bg-slate-400',   chip: 'bg-slate-50 text-slate-500 border-slate-200',       card: 'border-l-slate-300',   sort: 4 },
  interrupted:  { label: '中断',   dot: 'bg-rose-500',    chip: 'bg-rose-50 text-rose-600 border-rose-200',          card: 'border-l-rose-400',    sort: 5 },
}

// ===== 医生镜头：分诊决策卡态（报警 > 异常值 > 待体征 > 待开方 > 待签 > 已签）=====
type DoctorState = 'alarm' | 'abnormal' | 'needVitals' | 'needRx' | 'needSign' | 'signed'

const DOCTOR_META: Record<DoctorState, { label: string; dot: string; chip: string; card: string; sort: number }> = {
  alarm:      { label: '报警',   dot: 'bg-rose-600',    chip: 'bg-rose-100 text-rose-700 border-rose-300',         card: 'border-l-rose-500',    sort: 0 },
  abnormal:   { label: '异常值', dot: 'bg-orange-500',  chip: 'bg-orange-50 text-orange-600 border-orange-200',     card: 'border-l-orange-400',  sort: 1 },
  needVitals: { label: '待体征', dot: 'bg-slate-400',   chip: 'bg-slate-50 text-slate-500 border-slate-200',        card: 'border-l-slate-300',   sort: 2 },
  needRx:     { label: '待开方', dot: 'bg-sky-500',     chip: 'bg-sky-50 text-sky-600 border-sky-200',             card: 'border-l-sky-400',     sort: 3 },
  needSign:   { label: '待签',   dot: 'bg-violet-500',  chip: 'bg-violet-50 text-violet-600 border-violet-200',     card: 'border-l-violet-400',  sort: 4 },
  signed:     { label: '已签',   dot: 'bg-emerald-500', chip: 'bg-emerald-50 text-emerald-600 border-emerald-200', card: 'border-l-emerald-400', sort: 5 },
}

interface BedCard {
  treatmentId: number
  patientId: string
  patientName: string
  bedNumber: string
  mode: string
  shiftName: string
  queueNo: string
  startTime?: string
  state: BedState
  kioskCheckedIn: boolean
  kioskSelfMeasured: boolean
  wardId?: number | null
}

interface DoctorCard {
  treatmentId: number
  patientId: string
  patientName: string
  bedNumber: string
  mode: string
  doctorName: string
  state: DoctorState
  note: string // 异常值摘要 或 处方提示
  prescriptionId?: string // 待签卡一键签发用
  wardId?: number | null
}

function isTimeDone(startTime: string | undefined, durationMinutes: number | undefined): boolean {
  if (!startTime || !durationMinutes || durationMinutes <= 0) return false
  const elapsed = (Date.now() - new Date(startTime).getTime()) / 60000
  return elapsed >= durationMinutes
}

// 已签到：优先看 signInTime（与生产库现有数据一致）；兼容 legacyStatus >= 10。
function parseLegacyCheckedIn(signInTime?: string, legacy?: string): boolean {
  if (signInTime) return true
  const n = parseInt((legacy || '').trim(), 10)
  return Number.isFinite(n) && n >= 10
}

// 已自测：kiosk 写入的透前体征有实测值（体重或收缩压任一 >0）。
function hasSelfSigns(before?: { weight?: number; sbp?: number }): boolean {
  return !!before && ((before.weight ?? 0) > 0 || (before.sbp ?? 0) > 0)
}

function deriveBedState(status: number, legacy: string, startTime?: string, durationMinutes?: number): BedState {
  const l = (legacy || '').trim()
  if (status === 3 || l === '50') return 'interrupted'
  if (status === 2 || l === '40' || l === '60' || l === '100') return 'completed'
  if (l === '20') return 'preTreatment'
  if (status === 1 || l === '30') return isTimeDone(startTime, durationMinutes) ? 'readyOff' : 'inProgress'
  return 'pending'
}

// 临床分级（成人常规护栏，仅提示，需个体化）：
// 报警=治疗中断 或 危急体征(SBP<80/>200·DBP>120·HR<40/>140)；异常值=临界体征。
function assessClinical(
  before: { sbp?: number; dbp?: number; heartRate?: number } | undefined,
  status: number, legacy: string,
): { level: 'alarm' | 'abnormal' | null; note: string } {
  if (status === 3 || (legacy || '').trim() === '50') return { level: 'alarm', note: '治疗中断' }
  const { sbp, dbp, heartRate } = before || {}
  const crit: string[] = []
  const warn: string[] = []
  const has = (v?: number): v is number => typeof v === 'number' && v > 0
  if (has(sbp)) { if (sbp < 80 || sbp > 200) crit.push(`收缩压 ${sbp}`); else if (sbp < 90 || sbp > 180) warn.push(`收缩压 ${sbp}`) }
  if (has(dbp)) { if (dbp > 120) crit.push(`舒张压 ${dbp}`); else if (dbp > 110) warn.push(`舒张压 ${dbp}`) }
  if (has(heartRate)) { if (heartRate < 40 || heartRate > 140) crit.push(`心率 ${heartRate}`); else if (heartRate < 50 || heartRate > 120) warn.push(`心率 ${heartRate}`) }
  if (crit.length) return { level: 'alarm', note: crit.join(' · ') }
  if (warn.length) return { level: 'abnormal', note: warn.join(' · ') }
  return { level: null, note: '' }
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

type Lens = 'nurse' | 'doctor'

export default function Cockpit() {
  const navigate = useNavigate()
  const [lens, setLens] = useState<Lens>('nurse')
  const [bedCards, setBedCards] = useState<BedCard[]>([])
  const [doctorCards, setDoctorCards] = useState<DoctorCard[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [lastRefresh, setLastRefresh] = useState<Date>(new Date())

  const [checkedIn, setCheckedIn] = useState<boolean | null>(null)
  const [myDuties, setMyDuties] = useState<ResolvedDuty[]>([])
  const [checkingIn, setCheckingIn] = useState(false)

  const loadDuty = useCallback(async () => {
    const today = todayParam()
    try {
      const [status, duties] = await Promise.all([getCheckInStatus(today), getMyDuties(today)])
      setCheckedIn(!!status?.checkedIn)
      setMyDuties(duties || [])
    } catch {
      setCheckedIn(null)
    }
  }, [])

  useEffect(() => { loadDuty() }, [loadDuty])

  const doCheckIn = async () => {
    if (myDuties.length === 0) return
    setCheckingIn(true)
    try {
      for (const duty of myDuties) {
        const isDoctor = duty.dutyRole === '当班医生'
        await checkInApi({
          wardId: duty.wardId,
          operatorType: isDoctor ? 10 : 20,
          type: duty.dutyRole === '主班护士' || isDoctor ? 10 : 20,
        })
      }
      message.success('已接班，当日权限已激活')
      await loadDuty()
    } catch (e) {
      message.error(getErrorMessage(e))
    } finally {
      setCheckingIn(false)
    }
  }

  const loadData = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const today = todayParam()
      const [treatRes, patRes, rxRes] = await Promise.all([
        restApi.getTreatments({ treatmentDate: today, pageSize: 200 }),
        restApi.getPatientList({ page: 1, pageSize: 500 }),
        restApi.getPrescriptionDayStatus(today),
      ])

      const patientMap = new Map<string, { name: string; bed: string; mode: string }>()
      for (const p of patRes.data.items || []) {
        patientMap.set(String(p.id), { name: p.name || `#${p.id}`, bed: p.bedNumber || '', mode: p.defaultMode || '' })
      }
      const rxMap = new Map<string, { signed: boolean; prescriptionId?: string }>()
      for (const r of rxRes) rxMap.set(String(r.patientId), { signed: r.signed, prescriptionId: r.prescriptionId })

      const treatments = treatRes.data.items || []

      // 护士床位墙
      const beds: BedCard[] = treatments.map((t) => {
        const pinfo = patientMap.get(String(t.patientId))
        return {
          treatmentId: t.id,
          patientId: String(t.patientId),
          patientName: pinfo?.name || `#${t.patientId}`,
          bedNumber: pinfo?.bed || '--',
          mode: pinfo?.mode || '',
          shiftName: t.shiftName || '',
          queueNo: t.queueNo || '',
          startTime: t.startTime,
          state: deriveBedState(t.status, t.legacyStatus || '', t.startTime, t.durationMinutes),
          kioskCheckedIn: parseLegacyCheckedIn(t.signInTime, t.legacyStatus || ''),
          kioskSelfMeasured: hasSelfSigns(t.beforeSigns),
          wardId: t.wardId,
        }
      })
      beds.sort((a, b) => STATE_META[a.state].sort - STATE_META[b.state].sort || a.bedNumber.localeCompare(b.bedNumber, 'zh'))

      // 医生分诊墙
      const docs: DoctorCard[] = treatments.map((t) => {
        const pinfo = patientMap.get(String(t.patientId))
        const clin = assessClinical(t.beforeSigns, t.status, t.legacyStatus || '')
        const rx = rxMap.get(String(t.patientId))
        // 透前体重(自测/手录)是当日处方确认的前置；血压可后补，不纳入门禁。
        const vitalsReady = (t.beforeSigns?.weight ?? 0) > 0
        let state: DoctorState
        let note = ''
        if (clin.level === 'alarm') {
          state = 'alarm'; note = clin.note
        } else if (clin.level === 'abnormal') {
          state = 'abnormal'; note = clin.note
        } else if (!vitalsReady && !(rx && rx.signed)) {
          state = 'needVitals'; note = '待透前体重(自测/手录)'
        } else if (!rx) {
          state = 'needRx'; note = '今日尚无处方'
        } else if (!rx.signed) {
          state = 'needSign'; note = '处方待签发'
        } else {
          state = 'signed'; note = '处方已签发'
        }
        return {
          treatmentId: t.id,
          patientId: String(t.patientId),
          patientName: pinfo?.name || `#${t.patientId}`,
          bedNumber: pinfo?.bed || '--',
          mode: pinfo?.mode || '',
          doctorName: t.doctorName || '',
          state,
          note,
          prescriptionId: rx?.prescriptionId,
          wardId: t.wardId,
        }
      })
      docs.sort((a, b) => DOCTOR_META[a.state].sort - DOCTOR_META[b.state].sort || a.bedNumber.localeCompare(b.bedNumber, 'zh'))

      setBedCards(beds)
      setDoctorCards(docs)
      setLastRefresh(new Date())
    } catch (e) {
      setError(e instanceof Error ? e.message : '加载失败')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { loadData() }, [loadData])

  const [scope, setScope] = useState<'all' | 'mine'>('all')
  const myWardIds = useMemo(() => new Set(myDuties.map((d) => d.wardId)), [myDuties])
  const canMine = checkedIn === true && myWardIds.size > 0
  const inScope = useCallback(
    (wardId?: number | null) => scope === 'all' || (canMine && wardId != null && myWardIds.has(wardId)),
    [scope, canMine, myWardIds],
  )
  const visibleBeds = useMemo(() => bedCards.filter((card) => inScope(card.wardId)), [bedCards, inScope])
  const visibleDocs = useMemo(() => doctorCards.filter((card) => inScope(card.wardId)), [doctorCards, inScope])

  const bedCounts = useMemo(() => {
    const c: Record<BedState, number> = { pending: 0, preTreatment: 0, inProgress: 0, readyOff: 0, completed: 0, interrupted: 0 }
    for (const card of visibleBeds) c[card.state]++
    return c
  }, [visibleBeds])

  const docCounts = useMemo(() => {
    const c: Record<DoctorState, number> = { alarm: 0, abnormal: 0, needVitals: 0, needRx: 0, needSign: 0, signed: 0 }
    for (const card of visibleDocs) c[card.state]++
    return c
  }, [visibleDocs])

  const total = lens === 'nurse' ? visibleBeds.length : visibleDocs.length
  const goExecute = (patientId: string) => navigate(`/dialysis-processing?patientId=${encodeURIComponent(patientId)}`)

  const [signingId, setSigningId] = useState<string | null>(null)
  const handleSign = async (card: DoctorCard) => {
    if (!card.prescriptionId || signingId) return
    setSigningId(card.patientId)
    try {
      await restApi.signPrescription(card.patientId, card.prescriptionId)
      message.success(`${card.patientName} 处方已签发`)
      await loadData()
    } catch (e) {
      message.error(getErrorMessage(e))
    } finally {
      setSigningId(null)
    }
  }

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
              {todayParam()} · {lens === 'nurse' ? '护士床位墙' : '医生分诊墙'} · 共 {total} 位患者
              <span className="ml-2 text-slate-300">更新于 {fmtTime(lastRefresh.toISOString())}</span>
            </p>
          </div>
        </div>
        <div className="flex items-center gap-3">
          <div className="flex items-center bg-white border border-slate-200 rounded-xl p-1">
            <button type="button" onClick={() => setScope('all')}
              className={`px-3 py-1.5 rounded-lg text-sm font-bold transition-all ${scope === 'all' ? 'bg-slate-700 text-white shadow-sm' : 'text-slate-500 hover:text-slate-700'}`}>
              全室
            </button>
            <button type="button" onClick={() => canMine && setScope('mine')} disabled={!canMine}
              title={canMine ? '只看我当班室的病人' : '接班后可用（今日有当班排班）'}
              className={`px-3 py-1.5 rounded-lg text-sm font-bold transition-all ${scope === 'mine' ? 'bg-blue-600 text-white shadow-sm' : 'text-slate-500 hover:text-slate-700'} ${!canMine ? 'opacity-40 cursor-not-allowed' : ''}`}>
              我的病人
            </button>
          </div>
          {/* 镜头切换 */}
          <div className="flex items-center bg-white border border-slate-200 rounded-xl p-1">
            <button type="button" onClick={() => setLens('nurse')}
              className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm font-bold transition-all ${lens === 'nurse' ? 'bg-blue-600 text-white shadow-sm' : 'text-slate-500 hover:text-slate-700'}`}>
              <Users size={15} /> 护士墙
            </button>
            <button type="button" onClick={() => setLens('doctor')}
              className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm font-bold transition-all ${lens === 'doctor' ? 'bg-blue-600 text-white shadow-sm' : 'text-slate-500 hover:text-slate-700'}`}>
              <Stethoscope size={15} /> 医生墙
            </button>
          </div>
          <button type="button" onClick={loadData} disabled={loading}
            className="flex items-center gap-2 px-4 py-2 rounded-xl border border-slate-200 bg-white text-sm font-bold text-slate-600 hover:bg-slate-50 disabled:opacity-50 transition-all">
            <RefreshCw size={15} className={loading ? 'animate-spin' : ''} /> 刷新
          </button>
        </div>
      </div>

      {/* 传染病预警卡 */}
      <InfectiousAlertCards />

      {/* 水质预警卡 */}
      <WqAlertCards />

      {/* 血管通路告警卡 */}
      <VascularAlertCards />

      {/* 不良事件告警卡 */}
      <AdverseAlertCards />

      {checkedIn === false && myDuties.length > 0 && (
        <div className="flex items-center justify-between gap-3 rounded-2xl border border-amber-200 bg-amber-50 px-5 py-3.5 mb-5">
          <div className="flex items-center gap-3 text-[13px]">
            <LogIn size={18} className="text-amber-500 shrink-0" />
            <div>
              <span className="font-black text-amber-700">今日尚未接班</span>
              <span className="text-amber-600 ml-2">
                你今日当班：{myDuties.map((d) => d.dutyRole).join('、')}。确认接班后激活「当班」权限与「我的病人」过滤。
              </span>
            </div>
          </div>
          <button type="button" onClick={doCheckIn} disabled={checkingIn}
            className="shrink-0 px-4 py-2 rounded-xl bg-amber-500 text-white text-sm font-bold hover:bg-amber-600 disabled:opacity-50">
            {checkingIn ? '接班中...' : '确认接班'}
          </button>
        </div>
      )}

      {/* Summary */}
      {lens === 'nurse' ? (
        <div className="grid grid-cols-3 lg:grid-cols-6 gap-3 mb-5">
          {(['pending', 'preTreatment', 'inProgress', 'readyOff', 'completed', 'interrupted'] as BedState[]).map((s) => (
            <div key={s} className="bg-white rounded-2xl border border-slate-100 px-4 py-3.5 flex items-center justify-between">
              <div className="flex items-center gap-2">
                <span className={`w-2.5 h-2.5 rounded-full ${STATE_META[s].dot}`} />
                <span className="text-[13px] font-bold text-slate-500">{STATE_META[s].label}</span>
              </div>
              <span className="text-2xl font-black text-slate-800 tabular-nums">{bedCounts[s]}</span>
            </div>
          ))}
        </div>
      ) : (
        <div className="grid grid-cols-3 lg:grid-cols-5 gap-3 mb-5">
          {(['alarm', 'abnormal', 'needVitals', 'needRx', 'needSign', 'signed'] as DoctorState[]).map((s) => (
            <div key={s} className="bg-white rounded-2xl border border-slate-100 px-5 py-3.5 flex items-center justify-between">
              <div className="flex items-center gap-2">
                <span className={`w-2.5 h-2.5 rounded-full ${DOCTOR_META[s].dot}`} />
                <span className="text-[13px] font-bold text-slate-500">{DOCTOR_META[s].label}</span>
              </div>
              <span className="text-2xl font-black text-slate-800 tabular-nums">{docCounts[s]}</span>
            </div>
          ))}
        </div>
      )}

      {/* Error */}
      {error && (
        <div className="flex items-center gap-2 rounded-2xl border border-rose-200 bg-rose-50 px-5 py-4 mb-5 text-sm">
          <AlertTriangle size={18} className="text-rose-500 shrink-0" />
          <span className="text-rose-600 font-medium">{error}</span>
          <button type="button" onClick={loadData} className="ml-auto text-rose-600 font-bold underline">重试</button>
        </div>
      )}

      {/* Wall */}
      {loading && total === 0 ? (
        <div className="py-24 text-center text-slate-400 font-bold">加载中…</div>
      ) : total === 0 && !error ? (
        <div className="py-24 flex flex-col items-center text-slate-300">
          <BedDouble size={40} className="mb-3" />
          <p className="font-bold">今日暂无治疗安排</p>
        </div>
      ) : lens === 'nurse' ? (
        <div className="grid grid-cols-2 md:grid-cols-3 xl:grid-cols-4 gap-4">
          {visibleBeds.map((card) => {
            const meta = STATE_META[card.state]
            return (
              <button type="button" key={card.treatmentId} onClick={() => goExecute(card.patientId)}
                className={`text-left bg-white rounded-2xl border border-slate-100 border-l-4 ${meta.card} p-4 hover:shadow-lg hover:-translate-y-0.5 transition-all group`}>
                <div className="flex items-start justify-between mb-3">
                  <div className="flex items-center gap-2.5 min-w-0">
                    <div className="w-10 h-10 rounded-xl bg-slate-700 text-white flex items-center justify-center font-black text-sm shrink-0">{card.bedNumber}</div>
                    <div className="min-w-0">
                      <p className="font-black text-slate-800 truncate">{card.patientName}</p>
                      <p className="text-[11px] text-slate-400 truncate">{[card.mode, card.shiftName].filter(Boolean).join(' · ') || '—'}</p>
                    </div>
                  </div>
                  <div className="flex shrink-0 items-center gap-1">
                    {card.kioskSelfMeasured ? (
                      <span className="px-1.5 py-0.5 rounded text-[10px] font-black border bg-emerald-50 text-emerald-600 border-emerald-200">已自测</span>
                    ) : card.kioskCheckedIn ? (
                      <span className="px-1.5 py-0.5 rounded text-[10px] font-black border bg-sky-50 text-sky-600 border-sky-200">已签到</span>
                    ) : null}
                    <span className={`px-2 py-0.5 rounded-md text-[11px] font-black border ${meta.chip}`}>{meta.label}</span>
                  </div>
                </div>
                <div className="flex items-center justify-between text-[12px] text-slate-400">
                  <span className="flex items-center gap-1"><Clock size={12} />{card.state === 'pending' ? (card.queueNo ? `队列 ${card.queueNo}` : '待上机') : fmtTime(card.startTime)}</span>
                  <span className="flex items-center gap-1 text-blue-500 font-bold opacity-0 group-hover:opacity-100 transition-opacity">进执行 <ArrowRight size={13} /></span>
                </div>
              </button>
            )
          })}
        </div>
      ) : (
        <div className="grid grid-cols-2 md:grid-cols-3 xl:grid-cols-4 gap-4">
          {visibleDocs.map((card) => {
            const meta = DOCTOR_META[card.state]
            return (
              <div key={card.treatmentId} role="button" tabIndex={0} onClick={() => goExecute(card.patientId)}
                onKeyDown={(e) => { if (e.key === 'Enter') goExecute(card.patientId) }}
                className={`text-left bg-white rounded-2xl border border-slate-100 border-l-4 ${meta.card} p-4 hover:shadow-lg hover:-translate-y-0.5 transition-all group cursor-pointer`}>
                <div className="flex items-start justify-between mb-3">
                  <div className="flex items-center gap-2.5 min-w-0">
                    <div className="w-10 h-10 rounded-xl bg-slate-700 text-white flex items-center justify-center font-black text-sm shrink-0">{card.bedNumber}</div>
                    <div className="min-w-0">
                      <p className="font-black text-slate-800 truncate">{card.patientName}</p>
                      <p className="text-[11px] text-slate-400 truncate">{[card.mode, card.doctorName].filter(Boolean).join(' · ') || '—'}</p>
                    </div>
                  </div>
                  <span className={`shrink-0 px-2 py-0.5 rounded-md text-[11px] font-black border ${meta.chip}`}>{meta.label}</span>
                </div>
                <div className="flex items-center justify-between text-[12px] gap-2">
                  <span className={`truncate ${card.state === 'alarm' ? 'text-rose-600 font-black' : card.state === 'abnormal' ? 'text-orange-500 font-bold' : 'text-slate-400'}`}>{card.note}</span>
                  <div className="flex items-center gap-1 shrink-0">
                    <ActrButton patientId={card.patientId} prescriptionId={card.prescriptionId} />
                    {card.state === 'needSign' && card.prescriptionId ? (
                      <button type="button" onClick={(e) => { e.stopPropagation(); handleSign(card) }} disabled={signingId === card.patientId}
                        className="flex items-center gap-1 px-2.5 py-1 rounded-lg bg-violet-600 text-white font-bold hover:bg-violet-700 disabled:opacity-50 transition-all">
                        <PenLine size={12} /> {signingId === card.patientId ? '签发中' : '签发'}
                      </button>
                    ) : (
                      <span className="flex items-center gap-1 text-blue-500 font-bold opacity-0 group-hover:opacity-100 transition-opacity">开方 <ArrowRight size={13} /></span>
                    )}
                  </div>
                </div>
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}
