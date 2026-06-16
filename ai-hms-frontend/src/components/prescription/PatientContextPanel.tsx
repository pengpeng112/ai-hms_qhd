import { useEffect, useRef, useState, type CSSProperties } from 'react'
import { Spin, Tooltip } from 'antd'
import { ReloadOutlined } from '@ant-design/icons'
import { fetchPatientContext } from './PatientContextUtils'

// ─── 类型定义（对应后端 PrescriptionContextResponse）────────────────────────

interface LabIndicator {
  conceptId: string
  displayName: string
  value: number | null
  unit: string
  targetLow: number
  targetHigh: number
  status: 'normal' | 'watch' | 'high' | 'low' | 'critical_high' | 'critical_low' | 'missing'
  statusLabel: string
  testedAt: string | null
  daysAgo: number | null
  actionHint?: string
}

interface WeightUFContext {
  preWeight: number | null
  dryWeight: number | null
  weightGain: number | null
  weightGainPct: number | null
  ufTarget: number | null
  ufRatePerKg: number | null
  ufRateStatus: 'safe' | 'watch' | 'danger' | ''
  preWeightAt: string | null
}

interface TrendPoint {
  date: string
  value: number
  aux?: number
}

interface SodiumClearanceContext {
  lastRatio: number | null
  lastDate: string | null
  targetLow: number
  targetHigh: number
  status: 'good' | 'low' | 'high'
  statusLabel: string
  trend: TrendPoint[]
}

interface VitalSignsContext {
  lastSystolic: number | null
  lastDiastolic: number | null
  lastHeartRate: number | null
  lastDate: string | null
  bpStatus: 'normal' | 'high' | 'low'
  bpStatusLabel: string
  hrStatus: 'normal' | 'high' | 'low'
  hrStatusLabel: string
  bpTrend: TrendPoint[]
  hrTrend: TrendPoint[]
  sysLow: number
  sysHigh: number
  diaLow: number
  diaHigh: number
  hrLow: number
  hrHigh: number
}

interface LastTreatmentContext {
  treatmentDate: string | null
  actualUFml: number | null
  plannedUFml: number | null
  ufDiffMl: number | null
  actualMinutes: number | null
  plannedMinutes: number | null
  alarmCount: number
  alarmSummary: string
  outcome: 'completed' | 'early_stop' | 'not_found'
}

interface PrescriptionHint {
  priority: number
  icon: string
  title: string
  description: string
}

export interface PatientDemographics {
  heightCm: number | null
  ageYears: number | null
  isMale: boolean
  genderText: string
}

export interface PrescriptionContextData {
  patientId: string
  patientName: string
  bedCode: string
  demographics: PatientDemographics
  weight: WeightUFContext
  labs: LabIndicator[]
  lastTreatment: LastTreatmentContext
  sodiumClearance: SodiumClearanceContext
  vitals: VitalSignsContext
  hints: PrescriptionHint[]
  generatedAt: string
}

interface PatientContextPanelProps {
  patientId: string
  /** 由父组件预取的数据（与 RNa 面板共享同一来源，避免重复请求）；不传则自取 */
  externalData?: PrescriptionContextData | null
  /** 父组件提供的刷新回调；不传则使用内部刷新 */
  onRefresh?: () => void
  /** 本次抗凝方案（由页面从治疗方案/上次处方推导后传入）*/
  anticoagulant?: AnticoagulantInfo
}

function fmtLab(v: number): string {
  return v % 1 === 0 ? v.toFixed(0) : v.toFixed(2)
}

const FLAG_COLORS = {
  bad:  { bg: '#FBE9E7', text: '#C0392B' },
  warn: { bg: '#FBF3DF', text: '#8a6400' },
  ok:   { bg: '#E7F4EC', text: '#1E8449' },
  info: { bg: '#eef2f7', text: '#5A6B7B' },
}

// ─── 抗凝方案 ────────────────────────────────────────────────────────────────
// 临床五类：肝素 / 低分子肝素 / 萘莫司他 / 相对无肝素 / 无肝素
// 来源：原治疗方案 或 上一次透析处方；无既往（新患者/方案变更）需医生本次填写

export type AnticoagulantType =
  | '肝素' | '低分子肝素' | '萘莫司他' | '相对无肝素' | '无肝素' | '未设定'

export interface AnticoagulantInfo {
  type: AnticoagulantType
  initialDose?: string
  maintenanceDose?: string
  drugText?: string  // 原始药名摘要
  source: string     // 继承上次处方 / 沿用治疗方案 / 需医生填写
  needsInput: boolean
}

const ANTICOAG_COLORS: Record<AnticoagulantType, { bg: string; border: string; text: string }> = {
  '肝素':     { bg: '#eff6ff', border: '#93c5fd', text: '#1d4ed8' },
  '低分子肝素': { bg: '#f0fdfa', border: '#5eead4', text: '#0f766e' },
  '萘莫司他':  { bg: '#faf5ff', border: '#d8b4fe', text: '#7e22ce' },
  '相对无肝素': { bg: '#fffbeb', border: '#fcd34d', text: '#b45309' },
  '无肝素':    { bg: '#fff7ed', border: '#fdba74', text: '#c2410c' },
  '未设定':    { bg: '#fef2f2', border: '#fca5a5', text: '#dc2626' },
}

// ─── 颜色工具 ────────────────────────────────────────────────────────────────

const STATUS_COLORS: Record<string, { bg: string; text: string; border: string }> = {
  normal:       { bg: '#f0fdf4', text: '#16a34a', border: '#bbf7d0' },
  watch:        { bg: '#fefce8', text: '#b45309', border: '#fde68a' },
  high:         { bg: '#fef2f2', text: '#dc2626', border: '#fecaca' },
  low:          { bg: '#fdf2f8', text: '#be185d', border: '#f5d0fe' },
  critical_high:{ bg: '#dc2626', text: '#fff',    border: '#b91c1c' },
  critical_low: { bg: '#dc2626', text: '#fff',    border: '#b91c1c' },
  missing:      { bg: '#f8fafc', text: '#94a3b8', border: '#e2e8f0' },
}

const HINT_COLORS = {
  1: { bg: '#fff1f2', border: '#fda4af', text: '#9f1239' },
  2: { bg: '#fffbeb', border: '#fcd34d', text: '#78350f' },
  3: { bg: '#f0f9ff', border: '#bae6fd', text: '#0369a1' },
}

// ─── 小组件 ──────────────────────────────────────────────────────────────────

function SectionTitle({ icon, children }: { icon: string; children: React.ReactNode }) {
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 10 }}>
      <span style={{ fontSize: 12 }}>{icon}</span>
      <span style={{
        fontSize: 10, fontWeight: 700, letterSpacing: '0.07em',
        textTransform: 'uppercase', color: '#8c959f',
      }}>
        {children}
      </span>
    </div>
  )
}

function StatusBadge({ status, label }: { status: string; label: string }) {
  const c = STATUS_COLORS[status] ?? STATUS_COLORS.missing
  return (
    <span style={{
      display: 'inline-block', padding: '2px 7px', borderRadius: 10,
      fontSize: 10, fontWeight: 700,
      background: c.bg, color: c.text, border: `1px solid ${c.border}`,
    }}>
      {label}
    </span>
  )
}

function WeightCard({ label, value, unit, tag, highlight = false }: {
  label: string; value: string; unit: string; tag?: string; highlight?: boolean
}) {
  return (
    <div style={{
      background: highlight ? '#f0f9f9' : '#f8fafc',
      border: `1px solid ${highlight ? '#a7d8d7' : '#e5eaf0'}`,
      borderRadius: 10, padding: '10px 12px',
    }}>
      <div style={{ fontSize: 10, color: '#9ca3af', marginBottom: 2 }}>
        {label}
        {tag && (
          <span style={{
            marginLeft: 4, padding: '1px 5px', borderRadius: 4,
            fontSize: 9, fontWeight: 600, background: '#e0f2fe', color: '#0369a1',
          }}>
            {tag}
          </span>
        )}
      </div>
      <div style={{
        fontSize: 20, fontWeight: 700, color: highlight ? '#0E7C7B' : '#1B2A41',
        fontFamily: 'monospace', lineHeight: 1.1,
      }}>
        {value}
      </div>
      <div style={{ fontSize: 11, color: '#6b7280', fontFamily: 'monospace' }}>{unit}</div>
    </div>
  )
}

interface RefBand {
  low: number
  high: number
  color?: string
  edge?: string // 边界虚线颜色
}

// Sparkline 迷你趋势折线（带可选参考带 + 第二条线）
function Sparkline({
  points, width = 200, height = 44, color = '#0E7C7B', auxColor = '#94a3b8',
  bands = [],
}: {
  points: TrendPoint[]
  width?: number
  height?: number
  color?: string
  auxColor?: string
  bands?: RefBand[]
}) {
  if (points.length === 0) return null
  const pad = 4
  const hasAux = points.some(p => p.aux != null)

  // 计算 y 轴范围（包含主值、辅值、所有参考带）
  const allVals: number[] = []
  points.forEach(p => { allVals.push(p.value); if (p.aux != null) allVals.push(p.aux) })
  bands.forEach(b => { allVals.push(b.low, b.high) })
  const min = Math.min(...allVals)
  const max = Math.max(...allVals)
  const range = max - min || 1

  const x = (i: number) => pad + (i / (points.length - 1)) * (width - 2 * pad)
  const y = (v: number) => pad + (1 - (v - min) / range) * (height - 2 * pad)

  const linePath = (key: 'value' | 'aux') =>
    points
      .map((p, i) => {
        const v = key === 'value' ? p.value : p.aux
        if (v == null) return ''
        return `${i === 0 ? 'M' : 'L'} ${x(i).toFixed(1)} ${y(v).toFixed(1)}`
      })
      .join(' ')

  return (
    <svg width={width} height={height} style={{ display: 'block' }}>
      {/* 参考带（正常范围）*/}
      {bands.map((b, bi) => (
        <g key={bi}>
          <rect
            x={pad} y={y(b.high)} width={width - 2 * pad} height={Math.abs(y(b.low) - y(b.high))}
            fill={b.color ?? 'rgba(34,197,94,0.10)'}
          />
          <line x1={pad} y1={y(b.high)} x2={width - pad} y2={y(b.high)} stroke={b.edge ?? '#cbd5e1'} strokeWidth={0.8} strokeDasharray="3 2" />
          <line x1={pad} y1={y(b.low)} x2={width - pad} y2={y(b.low)} stroke={b.edge ?? '#cbd5e1'} strokeWidth={0.8} strokeDasharray="3 2" />
        </g>
      ))}
      {/* 辅线（舒张压）*/}
      {hasAux && (
        <path d={linePath('aux')} fill="none" stroke={auxColor} strokeWidth={1.5} strokeDasharray="3 2" />
      )}
      {/* 主线 */}
      <path d={linePath('value')} fill="none" stroke={color} strokeWidth={2} strokeLinejoin="round" />
      {/* 数据点 */}
      {points.map((p, i) => (
        <circle key={i} cx={x(i)} cy={y(p.value)} r={i === points.length - 1 ? 3 : 1.6}
          fill={i === points.length - 1 ? color : '#fff'} stroke={color} strokeWidth={1.2} />
      ))}
    </svg>
  )
}

// ─── 主组件 ──────────────────────────────────────────────────────────────────

export function PatientContextPanel({ patientId, externalData, onRefresh, anticoagulant }: PatientContextPanelProps) {
  const controlled = externalData !== undefined
  const [innerData, setInnerData] = useState<PrescriptionContextData | null>(null)
  const [loading, setLoading] = useState(!controlled)
  const [error, setError] = useState<string | null>(null)
  const [expanded, setExpanded] = useState(false)
  const [highlight, setHighlight] = useState<string | null>(null)
  const blockRefs = useRef<Record<string, HTMLDivElement | null>>({})

  const innerFetch = async () => {
    setLoading(true)
    setError(null)
    try {
      setInnerData(await fetchPatientContext(patientId))
    } catch {
      setError('数据加载失败')
    } finally {
      setLoading(false)
    }
  }

  // 受控模式（父组件预取）时不自取；否则按 patientId 自取
  useEffect(() => {
    if (!controlled) void innerFetch()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [patientId, controlled])

  const data = controlled ? externalData : innerData
  const refresh = onRefresh ?? innerFetch

  if (!controlled && loading) {
    return (
      <div style={{ padding: 24, textAlign: 'center' }}>
        <Spin size="small" /> <span style={{ marginLeft: 8, fontSize: 12, color: '#94a3b8' }}>加载开单参考...</span>
      </div>
    )
  }

  if (error || !data) {
    return (
      <div style={{ padding: 16, fontSize: 12, color: '#94a3b8', textAlign: 'center' }}>
        {error ?? '暂无数据'} &nbsp;
        <span style={{ color: '#0E7C7B', cursor: 'pointer' }} onClick={refresh}>重试</span>
      </div>
    )
  }

  const { weight, labs, lastTreatment: lastTx, sodiumClearance: naClr, vitals, hints } = data

  // 概要标签 → 点击展开并定位高亮对应详情块
  type FlagKey = 'labs' | 'tx' | 'na' | 'bp'
  type Flag = { key: FlagKey; level: 'bad' | 'warn' | 'ok' | 'info'; text: string }
  const flags: Flag[] = []
  labs.forEach((l) => {
    if (l.value == null) return
    const name = l.displayName.split(' ')[0]
    if (l.status === 'critical_high' || l.status === 'high') flags.push({ key: 'labs', level: 'bad', text: `${name} ${fmtLab(l.value)} 偏高` })
    else if (l.status === 'critical_low' || l.status === 'low') flags.push({ key: 'labs', level: 'bad', text: `${name} ${fmtLab(l.value)} 偏低` })
    else if (l.status === 'watch') flags.push({ key: 'labs', level: 'warn', text: `${name} ${fmtLab(l.value)} 临界` })
  })
  if (lastTx.outcome !== 'not_found') {
    const short = lastTx.actualMinutes != null && lastTx.plannedMinutes != null && lastTx.actualMinutes < lastTx.plannedMinutes - 5
    if (short || lastTx.alarmCount > 0) {
      const parts = [short ? '上次短透' : '上次治疗', lastTx.alarmCount > 0 ? `${lastTx.alarmCount}报警` : ''].filter(Boolean)
      flags.push({ key: 'tx', level: 'info', text: parts.join(' · ') })
    }
  }
  if (vitals.lastSystolic != null) {
    flags.push({
      key: 'bp',
      level: vitals.bpStatus === 'normal' ? 'ok' : 'warn',
      text: `均压 ${Math.round(vitals.lastSystolic)}/${Math.round(vitals.lastDiastolic ?? 0)}·心率 ${Math.round(vitals.lastHeartRate ?? 0)}`,
    })
  }
  if (naClr.lastRatio != null) {
    flags.push({ key: 'na', level: naClr.status === 'good' ? 'ok' : 'warn', text: `钠清除比 ${naClr.lastRatio.toFixed(2)}` })
  }

  const openBlock = (key?: FlagKey) => {
    setExpanded(true)
    if (key) {
      setHighlight(key)
      window.setTimeout(() => blockRefs.current[key]?.scrollIntoView({ behavior: 'smooth', block: 'nearest' }), 60)
      window.setTimeout(() => setHighlight(null), 1600)
    }
  }
  const blockStyle = (key: FlagKey): CSSProperties => ({
    padding: '12px 16px',
    borderBottom: '1px solid #f0f0f0',
    transition: 'box-shadow .3s',
    boxShadow: highlight === key ? 'inset 0 0 0 2px #0E7C7B' : 'none',
  })
  const setBlockRef = (key: FlagKey) => (el: HTMLDivElement | null) => { blockRefs.current[key] = el }

  return (
    <div style={{
      background: '#fff',
      border: '1px solid #e5eaf0',
      borderRadius: 14,
      overflow: 'hidden',
      fontSize: 12,
    }}>
      {/* 头部 */}
      <div style={{
        background: 'linear-gradient(135deg, #1B2A41 0%, #243447 100%)',
        padding: '10px 16px',
        display: 'flex', alignItems: 'center', justifyContent: 'space-between',
      }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <span style={{ fontSize: 15 }}>📋</span>
          <div>
            <div style={{ color: '#fff', fontWeight: 700, fontSize: 13, lineHeight: 1.2 }}>
              开单参考 · 患者数据汇总
            </div>
            <div style={{ color: '#8fa6c2', fontSize: 10, marginTop: 1 }}>只读参考，不影响处方</div>
          </div>
        </div>
        <Tooltip title={`数据更新于 ${data.generatedAt}`}>
          <span
            style={{ color: '#0E7C7B', cursor: 'pointer', fontSize: 14 }}
            onClick={refresh}
          >
            <ReloadOutlined />
          </span>
        </Tooltip>
      </div>

      {/* ① 高优先级提示（如有）*/}
      {hints.filter(h => h.priority === 1).length > 0 && (
        <div style={{ padding: '10px 16px 0 16px' }}>
          {hints.filter(h => h.priority === 1).map((hint, i) => (
            <div key={i} style={{
              display: 'flex', alignItems: 'flex-start', gap: 8,
              padding: '7px 10px', borderRadius: 8, marginBottom: 6,
              background: HINT_COLORS[1].bg, borderLeft: `3px solid ${HINT_COLORS[1].border}`,
              color: HINT_COLORS[1].text,
            }}>
              <span style={{ fontSize: 14, flexShrink: 0 }}>{hint.icon}</span>
              <div>
                <strong>{hint.title}</strong>：{hint.description}
              </div>
            </div>
          ))}
        </div>
      )}

      {/* ② 体重 & 超滤 */}
      <div style={{ padding: '12px 16px', borderBottom: '1px solid #f0f0f0' }}>
        <SectionTitle icon="⚖">体重 &amp; 超滤目标</SectionTitle>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: 8 }}>
          <WeightCard
            label="透前体重"
            value={weight.preWeight != null ? weight.preWeight.toFixed(1) : '—'}
            unit="kg"
            tag="今日透前"
          />
          <WeightCard
            label="干体重"
            value={weight.dryWeight != null ? weight.dryWeight.toFixed(1) : '—'}
            unit="kg"
            tag="治疗方案"
          />
          <WeightCard
            label="超滤目标"
            value={weight.ufTarget != null ? weight.ufTarget.toFixed(1) : '—'}
            unit={`L（+${weight.weightGainPct?.toFixed(1) ?? '—'}%）`}
            tag="自动算"
            highlight
          />
        </div>
        {weight.ufRatePerKg != null && (
          <div style={{
            marginTop: 8, padding: '6px 10px', borderRadius: 6,
            display: 'flex', alignItems: 'center', gap: 8,
            background: weight.ufRateStatus === 'safe' ? '#f0fdf4' : weight.ufRateStatus === 'watch' ? '#fffbeb' : '#fef2f2',
            borderLeft: `3px solid ${weight.ufRateStatus === 'safe' ? '#22c55e' : weight.ufRateStatus === 'watch' ? '#f59e0b' : '#ef4444'}`,
            fontSize: 11, color: '#475569',
          }}>
            超滤速率（4h 估算）：
            <span style={{ fontFamily: 'monospace', fontWeight: 700, fontSize: 13 }}>
              {weight.ufRatePerKg.toFixed(1)} mL/kg/h
            </span>
            <span style={{ fontSize: 10, color: '#94a3b8' }}>（警戒11·报警13）</span>
            <span style={{ marginLeft: 'auto', fontWeight: 700, color: weight.ufRateStatus === 'safe' ? '#16a34a' : weight.ufRateStatus === 'watch' ? '#b45309' : '#dc2626' }}>
              {weight.ufRateStatus === 'safe' ? '✓ 安全' : weight.ufRateStatus === 'watch' ? '⚠ 警戒' : '✗ 报警'}
            </span>
          </div>
        )}
      </div>

      {/* 本次抗凝方案 */}
      {anticoagulant && (
        <div style={{ padding: '12px 16px', borderBottom: '1px solid #f0f0f0' }}>
          <SectionTitle icon="💉">本次抗凝方案</SectionTitle>
          <div style={{
            display: 'flex', alignItems: 'center', gap: 12,
            padding: '10px 12px', borderRadius: 10,
            background: ANTICOAG_COLORS[anticoagulant.type].bg,
            border: `1px solid ${ANTICOAG_COLORS[anticoagulant.type].border}`,
          }}>
            <div style={{
              fontSize: 15, fontWeight: 800, whiteSpace: 'nowrap',
              color: ANTICOAG_COLORS[anticoagulant.type].text,
            }}>
              {anticoagulant.type}
            </div>
            <div style={{ flex: 1, minWidth: 0, fontSize: 11, color: '#475569' }}>
              {anticoagulant.needsInput ? (
                <span style={{ color: '#dc2626', fontWeight: 600 }}>
                  无既往抗凝记录，请在本次处方中填写
                </span>
              ) : (
                <div style={{ display: 'flex', flexWrap: 'wrap', gap: '2px 12px' }}>
                  {anticoagulant.initialDose && <span>首剂 <strong>{anticoagulant.initialDose}</strong></span>}
                  {anticoagulant.maintenanceDose && <span>维持 <strong>{anticoagulant.maintenanceDose}</strong></span>}
                  {anticoagulant.drugText && <span style={{ color: '#94a3b8' }}>（{anticoagulant.drugText}）</span>}
                </div>
              )}
            </div>
            <span style={{
              fontSize: 10, fontWeight: 600, whiteSpace: 'nowrap',
              padding: '2px 7px', borderRadius: 10,
              background: anticoagulant.needsInput ? '#fee2e2' : '#f1f5f9',
              color: anticoagulant.needsInput ? '#dc2626' : '#64748b',
            }}>
              {anticoagulant.source}
            </span>
          </div>
          {(anticoagulant.type === '无肝素' || anticoagulant.type === '相对无肝素') && (
            <div style={{ fontSize: 10, color: '#b45309', marginTop: 6, lineHeight: 1.5 }}>
              ⚠ 无/相对无肝素方案多用于高出血风险患者，注意管路凝血监测与生理盐水冲洗。
            </div>
          )}
        </div>
      )}

      {/* 概要标签行（点击任一标签展开并定位详情）*/}
      <div style={{ padding: '10px 16px', borderBottom: '1px solid #f0f0f0' }}>
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: 6 }}>
          {flags.map((f, i) => (
            <span key={i} onClick={() => openBlock(f.key)} style={{
              fontSize: 11, padding: '3px 9px', borderRadius: 20, fontWeight: 600, cursor: 'pointer',
              background: FLAG_COLORS[f.level].bg, color: FLAG_COLORS[f.level].text,
            }}>
              {f.text}
            </span>
          ))}
        </div>
        <div onClick={() => setExpanded((e) => !e)} style={{
          marginTop: 10, fontSize: 12, color: '#0E7C7B', fontWeight: 600, cursor: 'pointer',
          border: '1px dashed #bcdad9', borderRadius: 8, padding: '7px 10px',
          textAlign: 'center', background: '#f7fbfb',
        }}>
          {expanded ? '▴ 收起既往与趋势' : '▾ 展开既往与趋势（12次血压/心率 · 钠清除比曲线 · 检验 · 上次治疗）'}
        </div>
      </div>

      {expanded && (
        <>
          {/* ③ 近期检验 */}
          <div ref={setBlockRef('labs')} style={blockStyle('labs')}>
            <SectionTitle icon="🧪">近期检验</SectionTitle>
        <table style={{ width: '100%', borderCollapse: 'collapse' }}>
          <thead>
            <tr>
              {['指标', '结果', '目标范围', '状态', '日期'].map(h => (
                <th key={h} style={{
                  fontSize: 9, fontWeight: 700, color: '#94a3b8',
                  textTransform: 'uppercase', letterSpacing: '0.06em',
                  padding: '3px 6px', textAlign: 'left',
                  borderBottom: '1px solid #e5eaf0',
                }}>
                  {h}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {labs.map(lab => (
              <>
                <tr key={lab.conceptId}>
                  <td style={{ padding: '7px 6px', fontWeight: 600, color: '#374151', fontSize: 12 }}>
                    {lab.displayName}
                  </td>
                  <td style={{ padding: '7px 6px' }}>
                    {lab.value != null ? (
                      <span style={{
                        fontFamily: 'monospace', fontWeight: 700, fontSize: 13,
                        color: STATUS_COLORS[lab.status]?.text ?? '#374151',
                      }}>
                        {lab.value % 1 === 0 ? lab.value.toFixed(0) : lab.value.toFixed(2)}
                        <span style={{ fontSize: 10, color: '#9ca3af', marginLeft: 2 }}>{lab.unit}</span>
                      </span>
                    ) : (
                      <span style={{ color: '#cbd5e1', fontSize: 11 }}>无数据</span>
                    )}
                  </td>
                  <td style={{ padding: '7px 6px' }}>
                    <span style={{ fontSize: 10, color: '#9ca3af' }}>
                      {lab.targetLow > 0 ? `${lab.targetLow} – ` : '> '}{lab.targetHigh < 9 ? lab.targetHigh : lab.targetLow}
                    </span>
                  </td>
                  <td style={{ padding: '7px 6px' }}>
                    <StatusBadge status={lab.status} label={lab.statusLabel} />
                  </td>
                  <td style={{ padding: '7px 6px', textAlign: 'right' }}>
                    <Tooltip title={lab.daysAgo != null ? `${lab.daysAgo}天前` : undefined}>
                      <span style={{ fontSize: 10, color: lab.daysAgo != null && lab.daysAgo > 30 ? '#f59e0b' : '#9ca3af' }}>
                        {lab.testedAt ?? '—'}
                      </span>
                    </Tooltip>
                  </td>
                </tr>
                {lab.actionHint && (lab.status === 'high' || lab.status === 'low' || lab.status === 'critical_high' || lab.status === 'critical_low' || lab.status === 'watch') && (
                  <tr key={`${lab.conceptId}-hint`}>
                    <td colSpan={5} style={{ padding: '0 6px 8px 6px' }}>
                      <div style={{ fontSize: 10, color: STATUS_COLORS[lab.status]?.text }}>
                        → {lab.actionHint}
                      </div>
                    </td>
                  </tr>
                )}
              </>
            ))}
          </tbody>
        </table>
      </div>

      {/* ④ 上次治疗 */}
      {lastTx.outcome !== 'not_found' && (
        <div ref={setBlockRef('tx')} style={blockStyle('tx')}>
          <SectionTitle icon="🕐">上次治疗（{lastTx.treatmentDate ?? '—'}）</SectionTitle>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: 8, marginBottom: 8 }}>
            {/* 实际超滤 */}
            <div style={{ background: '#f8fafc', border: '1px solid #e5eaf0', borderRadius: 8, padding: '8px 10px' }}>
              <div style={{ fontSize: 10, color: '#94a3b8' }}>实际超滤</div>
              <div style={{
                fontFamily: 'monospace', fontSize: 16, fontWeight: 700,
                color: lastTx.ufDiffMl != null && lastTx.ufDiffMl < -200 ? '#b45309' : '#1B2A41',
              }}>
                {lastTx.actualUFml != null ? (lastTx.actualUFml / 1000).toFixed(1) : '—'} L
              </div>
              {lastTx.ufDiffMl != null && (
                <div style={{ fontSize: 10, color: '#9ca3af' }}>
                  差 {lastTx.ufDiffMl > 0 ? '+' : ''}{lastTx.ufDiffMl} mL
                </div>
              )}
            </div>
            {/* 实际时长 */}
            <div style={{ background: '#f8fafc', border: '1px solid #e5eaf0', borderRadius: 8, padding: '8px 10px' }}>
              <div style={{ fontSize: 10, color: '#94a3b8' }}>实际时长</div>
              <div style={{
                fontFamily: 'monospace', fontSize: 16, fontWeight: 700,
                color: lastTx.actualMinutes != null && lastTx.plannedMinutes != null && lastTx.actualMinutes < lastTx.plannedMinutes - 10
                  ? '#b45309' : '#1B2A41',
              }}>
                {lastTx.actualMinutes != null
                  ? `${Math.floor(lastTx.actualMinutes / 60)}h${lastTx.actualMinutes % 60}min`
                  : '—'}
              </div>
              {lastTx.plannedMinutes != null && (
                <div style={{ fontSize: 10, color: '#9ca3af' }}>目标 {lastTx.plannedMinutes} min</div>
              )}
            </div>
            {/* 结果 */}
            <div style={{ background: '#f8fafc', border: '1px solid #e5eaf0', borderRadius: 8, padding: '8px 10px' }}>
              <div style={{ fontSize: 10, color: '#94a3b8' }}>结果</div>
              <div style={{ fontWeight: 700, fontSize: 13, color: lastTx.outcome === 'early_stop' ? '#dc2626' : '#16a34a' }}>
                {lastTx.outcome === 'early_stop' ? '提前终止' : '顺利完成'}
              </div>
            </div>
          </div>
          {lastTx.alarmCount > 0 && (
            <div style={{
              display: 'flex', alignItems: 'center', gap: 8,
              padding: '6px 10px', background: '#fefce8', borderRadius: 8,
              borderLeft: '3px solid #f59e0b', fontSize: 11, color: '#78350f',
            }}>
              <span style={{ fontFamily: 'monospace', fontSize: 18, fontWeight: 800, color: '#b45309' }}>
                {lastTx.alarmCount}
              </span>
              <span>次报警{lastTx.alarmSummary ? `：${lastTx.alarmSummary}` : ''}</span>
            </div>
          )}
        </div>
      )}

      {/* ⑤ 钠清除比趋势（专利核心指标）*/}
      <div ref={setBlockRef('na')} style={blockStyle('na')}>
        <SectionTitle icon="🧂">钠清除比 · 脱钠空间评估</SectionTitle>
        <div style={{ display: 'flex', alignItems: 'center', gap: 14 }}>
          <div style={{ flexShrink: 0 }}>
            <div style={{ fontSize: 10, color: '#94a3b8' }}>
              上次（{naClr.lastDate ?? '—'}）
            </div>
            <div style={{
              fontFamily: 'monospace', fontSize: 26, fontWeight: 800, lineHeight: 1.1,
              color: naClr.status === 'good' ? '#0E7C7B' : naClr.status === 'high' ? '#b45309' : '#be185d',
            }}>
              {naClr.lastRatio != null ? naClr.lastRatio.toFixed(2) : '—'}
            </div>
            <div style={{ fontSize: 10, marginTop: 2 }}>
              <StatusBadge
                status={naClr.status === 'good' ? 'normal' : naClr.status === 'high' ? 'watch' : 'low'}
                label={naClr.statusLabel}
              />
            </div>
            <div style={{ fontSize: 9, color: '#cbd5e1', marginTop: 3 }}>
              目标 {naClr.targetLow}–{naClr.targetHigh}
            </div>
          </div>
          <div style={{ flex: 1, minWidth: 0 }}>
            <Sparkline
              points={naClr.trend}
              width={260}
              color="#0E7C7B"
              bands={[{ low: naClr.targetLow, high: naClr.targetHigh, color: 'rgba(14,124,123,0.08)', edge: '#a7d8d7' }]}
            />
            <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 9, color: '#cbd5e1', marginTop: 2 }}>
              <span>{naClr.trend[0]?.date}</span>
              <span style={{ color: '#94a3b8' }}>近 {naClr.trend.length} 次</span>
              <span>{naClr.trend[naClr.trend.length - 1]?.date}</span>
            </div>
          </div>
        </div>
        <div style={{ fontSize: 10, color: '#94a3b8', marginTop: 6, lineHeight: 1.5 }}>
          趋势缓降趋平 = 钠池接近排空（疗程棘轮自限）；持续高位 = 仍有脱钠空间，可上调 RNa。
        </div>
      </div>

      {/* ⑥ 血压心率趋势（上次透析平均）*/}
      <div ref={setBlockRef('bp')} style={blockStyle('bp')}>
        <SectionTitle icon="🫀">血压 / 心率 · 上次透析平均</SectionTitle>
        <div style={{ display: 'flex', gap: 16 }}>
          {/* 血压 */}
          <div style={{ flex: 1, minWidth: 0 }}>
            <div style={{ display: 'flex', alignItems: 'baseline', gap: 6, marginBottom: 4 }}>
              <span style={{
                fontFamily: 'monospace', fontSize: 18, fontWeight: 800,
                color: vitals.bpStatus === 'normal' ? '#1B2A41' : vitals.bpStatus === 'high' ? '#dc2626' : '#be185d',
              }}>
                {vitals.lastSystolic != null ? `${Math.round(vitals.lastSystolic)}/${Math.round(vitals.lastDiastolic ?? 0)}` : '—'}
              </span>
              <span style={{ fontSize: 10, color: '#94a3b8' }}>mmHg</span>
              <StatusBadge
                status={vitals.bpStatus === 'normal' ? 'normal' : vitals.bpStatus === 'high' ? 'watch' : 'low'}
                label={vitals.bpStatusLabel}
              />
            </div>
            <Sparkline
              points={vitals.bpTrend}
              width={230}
              color="#C0392B"
              auxColor="#e89a9a"
              bands={[
                { low: vitals.sysLow, high: vitals.sysHigh, color: 'rgba(192,57,43,0.06)', edge: '#f0b8b0' },
                { low: vitals.diaLow, high: vitals.diaHigh, color: 'rgba(148,163,184,0.10)', edge: '#cbd5e1' },
              ]}
            />
            <div style={{ fontSize: 9, color: '#cbd5e1', marginTop: 2 }}>
              收缩压(实线 90–150)/舒张压(虚线 60–90) · 每点=单次透析均值 · 近 {vitals.bpTrend.length} 次
            </div>
          </div>
          {/* 心率 */}
          <div style={{ flex: 1, minWidth: 0 }}>
            <div style={{ display: 'flex', alignItems: 'baseline', gap: 6, marginBottom: 4 }}>
              <span style={{
                fontFamily: 'monospace', fontSize: 18, fontWeight: 800,
                color: vitals.hrStatus === 'normal' ? '#1B2A41' : vitals.hrStatus === 'high' ? '#dc2626' : '#be185d',
              }}>
                {vitals.lastHeartRate != null ? Math.round(vitals.lastHeartRate) : '—'}
              </span>
              <span style={{ fontSize: 10, color: '#94a3b8' }}>bpm</span>
              <StatusBadge
                status={vitals.hrStatus === 'normal' ? 'normal' : vitals.hrStatus === 'high' ? 'watch' : 'low'}
                label={vitals.hrStatusLabel}
              />
            </div>
            <Sparkline
              points={vitals.hrTrend}
              width={230}
              color="#0E7C7B"
              bands={[{ low: vitals.hrLow, high: vitals.hrHigh, color: 'rgba(14,124,123,0.08)', edge: '#a7d8d7' }]}
            />
            <div style={{ fontSize: 9, color: '#cbd5e1', marginTop: 2 }}>
              心率（正常 60–90）· 每点=单次透析均值 · 近 {vitals.hrTrend.length} 次
            </div>
          </div>
        </div>
        <div style={{ fontSize: 10, color: '#94a3b8', marginTop: 6, lineHeight: 1.5 }}>
          血压趋势下行 + 透析中低血压 = 干体重可能过低或脱钠过猛，强力脱钠需谨慎。
        </div>
      </div>

      {/* ⑦ 中/低优先级提示 */}
      {hints.filter(h => h.priority > 1).length > 0 && (
        <div style={{ padding: '10px 16px' }}>
          <SectionTitle icon="💡">本次处方关注点</SectionTitle>
          {hints.filter(h => h.priority > 1).map((hint, i) => {
            const colors = HINT_COLORS[hint.priority as 2 | 3] ?? HINT_COLORS[3]
            return (
              <div key={i} style={{
                display: 'flex', alignItems: 'flex-start', gap: 8,
                padding: '7px 10px', borderRadius: 8, marginBottom: 6,
                background: colors.bg, borderLeft: `3px solid ${colors.border}`,
                color: colors.text, fontSize: 11,
              }}>
                <span style={{ fontSize: 13, flexShrink: 0 }}>{hint.icon}</span>
                <div>
                  <strong>{hint.title}</strong>：{hint.description}
                </div>
              </div>
            )
          })}
        </div>
      )}
        </>
      )}

      {/* 底部时间戳 */}
      <div style={{
        padding: '6px 16px', background: '#f8fafc', borderTop: '1px solid #e5eaf0',
        display: 'flex', justifyContent: 'space-between', alignItems: 'center',
        fontSize: 10, color: '#94a3b8',
      }}>
        <span>
          <span style={{ display: 'inline-block', width: 6, height: 6, borderRadius: '50%', background: '#22c55e', marginRight: 4, verticalAlign: 'middle' }} />
          数据已同步 · {data.generatedAt}
        </span>
        <span style={{ color: '#0E7C7B', cursor: 'pointer', fontWeight: 600 }} onClick={refresh}>
          ↻ 刷新
        </span>
      </div>
    </div>
  )
}
