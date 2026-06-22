/**
 * RNa 智能钠处方面板（钠清除比处方）
 *
 * 移植自专利设计稿原型 `透析处方助手_v1.html` + `model_v2.py`
 * 物理已按 α=0.92(对流占比) / W=0.93(血浆水分率) 修正——可移动钠 m=(α/W)·C≈0.989·C，
 * 略低于血清钠（旧"透析液钠 +6"为反临床错误，已弃）。
 *
 * 操作模型：
 *   ① 病人（透前血钠/体重/身高…，多取自 LIS / 患者档案）
 *   ② 你来定：左右滑动「超滤量 V_UF」+ 拨「RNa 钠清除比 ⇄ 透后血清钠」（同一旋钮，联动）
 *   ③ 系统给：自动算出「透析液钠 C_d」（抄到机器）+ 对流/弥散分解 + 安全灯
 *
 * 核心公式（per-session 闭式）：
 *   搬水 Na_target = V_UF · C_pre / W
 *   本次清钠 M_target = RNa · Na_target = (C_pre·V_UF + δ·TBW) / W,  δ = C_pre − C_post
 *   对流 conv = (α/W)·C̄·V_UF ;  弥散 diff = M_target − conv
 *   C_d = (α/W)·C̄·(1 + Q/D) − M_target/(D·T),  C̄=(C_pre+C_post)/2, Q=V_UF/T
 *   联动：RNa→δ=(RNa−1)·C_pre·V_UF/TBW→C_post ; C_post→δ=C_pre−C_post→RNa=1+δ·TBW/(C_pre·V_UF)
 *
 * 使用：
 *   <RNaPrescriptionPanel
 *     data={{ cPre, preWeight, dryWeight, ufOverride, patient: { height, age, isMale } }}
 *     cPreSource="lab_report"
 *     onAdopt={(cd) => setFieldValue('parameters.na', cd)}
 *   />
 */

import { useEffect, useMemo, useState } from 'react'
import { Button, InputNumber, Slider, Collapse, Tag, Tooltip } from 'antd'
import { ThunderboltFilled, InfoCircleOutlined } from '@ant-design/icons'
import { apiClient } from '@services/restClient'

// ─────────────────────────────────────────────────────────────────────────────
// 常量（与 model_v2.py 一致）
// ─────────────────────────────────────────────────────────────────────────────
const W = 0.93            // 血浆水分率（目标钠按血浆水基准 ÷W）
const DEFAULT_ALPHA = 0.92 // Donnan 系数 = 对流占比；弥散占比 = 1−α
const DEFAULT_D = 12       // 钠弥散度 (L/h)
const DEFAULT_T = 4        // 透析时长 (h)
const DEFAULT_FLOOR = 135.5 // 透后血清钠安全地板
const DEFAULT_CDMAX = 148   // 透析机透析液钠上限
const DEFAULT_DCAP = 3.0    // 单次降钠上限

// ─────────────────────────────────────────────────────────────────────────────
// 类型
// ─────────────────────────────────────────────────────────────────────────────
export interface RNaPatientInfo {
  height: number   // cm
  age: number      // 岁
  isMale: boolean
}

export interface RNaInputData {
  cPre?: number        // 透前血清钠 (mmol/L)
  preWeight?: number   // 透前体重 (kg)
  dryWeight: number    // 干体重 (kg)
  ufOverride?: number  // 医生设定的本次超滤目标 (L)；提供时作为 V_UF 初值
  patient: RNaPatientInfo
}

export type DataSource = 'before_check' | 'lab_report' | 'plan' | 'manual'

export interface RNaHistoryPoint {
  date: string
  cPre: number
  cdUsed?: number
  deltaUsed?: number
}

export interface RNaPrescriptionPanelProps {
  data: RNaInputData
  cPreSource?: DataSource
  preWeightSource?: DataSource
  recentHistory?: RNaHistoryPoint[]
  onAdopt: (cd: number) => void
  /** 默认钠清除比（钠目标旋钮初值，通常 1.0=等钠维持）*/
  defaultRNa?: number
  /** 受控超滤量（L）：传入则 V_UF 滑块由父组件管理，作为处方超滤目标的唯一来源 */
  vuf?: number
  onVufChange?: (v: number) => void
  /** 传入则"采纳"时向后端 /calculate-na 写一条服务端留痕（fire-and-forget）*/
  patientId?: string
}

const SOURCE_LABEL: Record<DataSource, { text: string; color: string }> = {
  before_check: { text: '今日透前', color: 'green' },
  lab_report:   { text: 'LIS·最近血钠', color: 'blue' },
  plan:         { text: '治疗方案', color: 'default' },
  manual:       { text: '手动输入', color: 'orange' },
}

// ─────────────────────────────────────────────────────────────────────────────
// 计算（移植自 v1 HTML 的 compute()）
// ─────────────────────────────────────────────────────────────────────────────
function watsonTBW(dryWeight: number, heightCm: number, age: number, isMale: boolean): number {
  const t = isMale
    ? 2.447 - 0.09156 * age + 0.1074 * heightCm + 0.3362 * dryWeight
    : -2.097 + 0.1069 * heightCm + 0.2466 * dryWeight
  return Math.max(10, t)
}

interface AdvancedParams {
  D: number; T: number; alpha: number; floor: number; cdmax: number; dcap: number
}

interface RNaResult {
  rNa: number
  cPost: number
  delta: number
  tbw: number
  naTarget: number   // 搬水
  deload: number     // 脱载
  mTarget: number    // 本次清钠
  conv: number       // 对流
  diff: number       // 弥散
  cd: number
  cdAdopt: number    // 取整 0.5 后的机器设定值
}

function computeRNa(
  cPre: number, vuf: number, tbw: number,
  driver: 'rna' | 'cpost', rNaInput: number, cPostInput: number,
  adv: AdvancedParams
): RNaResult {
  const R = adv.alpha / W
  let rNa: number, cPost: number, delta: number
  if (driver === 'rna') {
    rNa = rNaInput
    delta = vuf > 0 ? ((rNa - 1) * cPre * vuf) / tbw : 0
    cPost = cPre - delta
  } else {
    cPost = cPostInput
    delta = cPre - cPost
    rNa = vuf > 0 ? 1 + (delta * tbw) / (cPre * vuf) : 1
  }
  const naTarget = (vuf * cPre) / W           // 搬水
  const mTarget = rNa * naTarget              // 本次清钠
  const deload = mTarget - naTarget           // 脱载
  const cbar = (cPre + cPost) / 2
  const q = vuf / adv.T
  const cd = R * cbar * (1 + q / adv.D) - mTarget / (adv.D * adv.T)
  const conv = R * cbar * vuf                 // 对流
  const diff = mTarget - conv                 // 弥散
  return {
    rNa, cPost, delta, tbw, naTarget, deload, mTarget, conv, diff,
    cd, cdAdopt: Math.round(cd * 2) / 2,
  }
}

// ─────────────────────────────────────────────────────────────────────────────
// 组件
// ─────────────────────────────────────────────────────────────────────────────
export function RNaPrescriptionPanel({
  data, cPreSource, onAdopt, defaultRNa = 1.0,
  vuf: vufProp, onVufChange, patientId,
}: RNaPrescriptionPanelProps) {
  // 透前血钠（缺失则手动）
  const [manualCPre, setManualCPre] = useState<number | undefined>(undefined)
  const cPre = manualCPre ?? data.cPre
  const isManualCPre = manualCPre !== undefined

  // 超滤量：受控（父组件=处方超滤目标唯一来源）或内部状态
  const autoUf = Math.max(0, (data.preWeight ?? data.dryWeight) - data.dryWeight)
  const [innerVuf, setInnerVuf] = useState<number>(
    Number((data.ufOverride ?? autoUf).toFixed(1))
  )
  const vuf = vufProp ?? innerVuf
  const setVuf = (v: number) => {
    if (onVufChange) {
      onVufChange(v)
      return
    }
    setInnerVuf(v)
  }

  // 钠目标旋钮（RNa ⇄ C_post 联动）
  const [driver, setDriver] = useState<'rna' | 'cpost'>('rna')
  const [rNaInput, setRNaInput] = useState<number>(defaultRNa)
  const [cPostInput, setCPostInput] = useState<number>(cPre ?? 140)
  const [rnaDirty, setRnaDirty] = useState(false)

  // 异步 defaultRNa 同步：阶段变化时默认值更新，但用户手动调整后不再覆盖
  useEffect(() => {
    if (!rnaDirty) {
      setRNaInput(defaultRNa)
    }
  }, [defaultRNa, rnaDirty])

  // 高级参数
  const [adv, setAdv] = useState<AdvancedParams>({
    D: DEFAULT_D, T: DEFAULT_T, alpha: DEFAULT_ALPHA,
    floor: DEFAULT_FLOOR, cdmax: DEFAULT_CDMAX, dcap: DEFAULT_DCAP,
  })

  const tbw = watsonTBW(data.dryWeight, data.patient.height, data.patient.age, data.patient.isMale)

  const result = useMemo<RNaResult | null>(() => {
    if (cPre === undefined || vuf <= 0) return null
    return computeRNa(cPre, vuf, tbw, driver, rNaInput, cPostInput, adv)
  }, [cPre, vuf, tbw, driver, rNaInput, cPostInput, adv])

  // 拨 RNa → 同步显示 C_post；拨 C_post → 同步显示 RNa
  const shownRNa = result ? result.rNa : rNaInput
  const shownCPost = result ? result.cPost : cPostInput

  const setRNa = (v: number | null) => {
    if (v == null) return
    setDriver('rna'); setRNaInput(v); setRnaDirty(true)
  }
  const setCPost = (v: number | null) => {
    if (v == null) return
    setDriver('cpost'); setCPostInput(v)
  }

  // ─── 安全灯（移植自 v1）───────────────────────────────────────────────
  type Chip = { level: 'ok' | 'warn' | 'bad'; icon: string; text: string }
  const chips: Chip[] = []
  if (result) {
    const { cPost, cd, delta, rNa } = result
    if (cPost < adv.floor) chips.push({ level: 'bad', icon: '🔴', text: `透后血钠 ${cPost.toFixed(1)} 低于地板 ${adv.floor}：会致低钠，调高 RNa/透后或减脱载` })
    else if (cPost < adv.floor + 1.5) chips.push({ level: 'warn', icon: '🟡', text: `透后 ${cPost.toFixed(1)} 已贴近地板 ${adv.floor}，谨慎` })
    else chips.push({ level: 'ok', icon: '🟢', text: `透后血钠 ${cPost.toFixed(1)} ＞ 地板 ${adv.floor}` })

    if (cd > adv.cdmax) chips.push({ level: 'warn', icon: '🟡', text: `透析液钠 ${cd.toFixed(1)} 超机器上限 ${adv.cdmax}：延长时间或降 RNa，别硬拉` })
    else if (cd < 128) chips.push({ level: 'warn', icon: '🟡', text: `透析液钠 ${cd.toFixed(1)} 偏低，谨慎` })
    else chips.push({ level: 'ok', icon: '🟢', text: `透析液钠 ${cd.toFixed(1)} 在安全区 [128, ${adv.cdmax}]` })

    if (delta > adv.dcap) chips.push({ level: 'warn', icon: '🟡', text: `单次降血钠 ${delta.toFixed(1)} ＞ ${adv.dcap}：太猛，建议分几次透析（疗程棘轮）` })
    else if (rNa < 1) chips.push({ level: 'warn', icon: '🟡', text: `RNa ${rNa.toFixed(2)} ＜ 1：在加钠/欠清除（一般不是想要的）` })
    else chips.push({ level: 'ok', icon: '🟢', text: `单次降血钠 ${delta.toFixed(1)} ≤ ${adv.dcap}，平稳` })
  }

  const chipColor = (lv: Chip['level']) =>
    lv === 'ok' ? { bg: '#E7F4EC', fg: '#1E8449' }
      : lv === 'warn' ? { bg: '#FBF3DF', fg: '#8a6400' }
        : { bg: '#FBE9E7', fg: '#C0392B' }

  // 分解条比例
  const conv = result ? Math.max(0, result.conv) : 0
  const diff = result ? result.diff : 0
  const totBar = conv + Math.abs(diff) || 1

  // 采纳时向后端写一条服务端留痕（fire-and-forget，失败不打扰医生）
  const auditAdopt = (r: RNaResult) => {
    if (!patientId || cPre === undefined) return
    apiClient
      .post(`/api/v1/patients/${patientId}/prescriptions/calculate-na`, {
        cPre,
        preWeight: data.preWeight ?? data.dryWeight + vuf,
        dryWeight: data.dryWeight,
        heightCm: data.patient.height,
        ageYears: data.patient.age,
        isMale: data.patient.isMale,
        vuf,
        driver,
        rNa: r.rNa,
        cPost: r.cPost,
        alpha: adv.alpha,
        d: adv.D,
        t: adv.T,
      })
      .catch(() => { /* 留痕失败不影响开单 */ })
  }

  return (
    <div style={{
      background: '#fff', border: '1px solid #DDE3EA', borderRadius: 14, overflow: 'hidden',
      fontFamily: '"Microsoft YaHei","PingFang SC","Segoe UI",sans-serif', color: '#1B2A41',
    }}>
      {/* 头部 */}
      <div style={{ background: '#1B2A41', color: '#fff', padding: '12px 16px' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <span style={{ fontSize: 16 }}>⚗</span>
          <div style={{ fontWeight: 700, fontSize: 15 }}>RNa 智能钠处方</div>
          <span style={{ marginLeft: 6, fontSize: 10, background: '#0E7C7B', color: '#fff', padding: '1px 6px', borderRadius: 4 }}>
            钠清除比 · α=0.92
          </span>
        </div>
        <div style={{ fontSize: 11, opacity: 0.8, marginTop: 3 }}>
          滑动超滤量 + 拨钠目标 → 透析液钠自动出（对流:弥散 = 92:8）
        </div>
      </div>

      {/* ② 你来定 */}
      <div style={{ padding: '12px 16px' }}>
        <div style={{ fontSize: 13, color: '#0E7C7B', fontWeight: 700, marginBottom: 10 }}>② 你来定（2 项）</div>

        {/* 透前血钠（缺失时手动）*/}
        <div style={{
          display: 'flex', alignItems: 'center', gap: 8, marginBottom: 12,
          fontSize: 12, color: '#5A6B7B',
        }}>
          <span>透前血清钠 C_pre：</span>
          {isManualCPre || cPre === undefined ? (
            <InputNumber
              size="small" value={manualCPre ?? data.cPre} step={1} min={100} max={175}
              onChange={(v) => setManualCPre(v ?? undefined)} style={{ width: 90 }}
              addonAfter="mmol/L"
            />
          ) : (
            <>
              <b style={{ fontSize: 15, color: '#1B2A41' }}>{cPre.toFixed(1)}</b>
              <span>mmol/L</span>
              {cPreSource && <Tag color={SOURCE_LABEL[cPreSource].color} style={{ marginLeft: 2 }}>{SOURCE_LABEL[cPreSource].text}</Tag>}
              <Button type="link" size="small" onClick={() => setManualCPre(cPre)}>改</Button>
            </>
          )}
          <Tooltip title={`Watson 总体水 TBW ≈ ${tbw.toFixed(1)} L（${data.patient.isMale ? '男' : '女'} ${data.patient.age}岁 ${data.patient.height}cm）`}>
            <span style={{ marginLeft: 'auto', fontSize: 11, color: '#9aa6b2' }}>
              TBW ≈ {tbw.toFixed(1)} L <InfoCircleOutlined />
            </span>
          </Tooltip>
        </div>

        {/* 超滤量 V_UF 滑块 */}
        <div style={{ background: '#EEF4F4', border: '1px solid #CFE3E2', borderRadius: 12, padding: '10px 12px' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline', fontSize: 12, color: '#5A6B7B' }}>
            <span>超滤量 V_UF（搬水）</span>
            <span style={{ fontSize: 11 }}>透前−干重 = {autoUf.toFixed(1)} L</span>
          </div>
          <div style={{ fontSize: 28, fontWeight: 700, color: '#1B2A41', lineHeight: 1.2 }}>
            {vuf.toFixed(1)} <span style={{ fontSize: 14, color: '#5A6B7B', fontWeight: 400 }}>L</span>
          </div>
          <Slider min={0} max={5} step={0.1} value={vuf} onChange={setVuf}
            styles={{ track: { background: '#0E7C7B' } }} />
        </div>

        {/* 钠目标：RNa ⇄ C_post 联动 */}
        <div style={{ marginTop: 12 }}>
          <div style={{ fontSize: 12, color: '#5A6B7B', marginBottom: 6 }}>钠目标（拨任一个，另一个联动）</div>
          <Slider
            min={0.8} max={1.4} step={0.01} value={Number(shownRNa.toFixed(2))}
            onChange={setRNa} styles={{ track: { background: '#0E7C7B' } }}
            marks={{ 0.8: '0.8', 1.0: '1.0', 1.2: '1.2', 1.4: '1.4' }}
          />
          <div style={{ display: 'grid', gridTemplateColumns: '1fr auto 1fr', alignItems: 'center', gap: 8, marginTop: 8 }}>
            <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', background: '#fff', border: '1px solid #CFE3E2', borderRadius: 10, padding: '6px 4px' }}>
              <InputNumber
                variant="borderless" value={Number(shownRNa.toFixed(2))} step={0.01} min={0.8} max={1.6}
                onChange={setRNa} controls={false}
                style={{ width: '100%', fontSize: 20, fontWeight: 700, textAlign: 'center' }}
              />
              <span style={{ fontSize: 11, color: '#5A6B7B' }}>RNa 钠清除比</span>
            </div>
            <div style={{ fontSize: 11, color: '#0E7C7B', textAlign: 'center', whiteSpace: 'nowrap' }}>⇄<br />同一旋钮</div>
            <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', background: '#fff', border: '1px solid #CFE3E2', borderRadius: 10, padding: '6px 4px' }}>
              <InputNumber
                variant="borderless" value={Number(shownCPost.toFixed(1))} step={0.5}
                onChange={setCPost} controls={false}
                style={{ width: '100%', fontSize: 20, fontWeight: 700, textAlign: 'center' }}
              />
              <span style={{ fontSize: 11, color: '#5A6B7B' }}>透后血清钠</span>
            </div>
          </div>
        </div>
      </div>

      {/* ③ 系统给 */}
      <div style={{ padding: '0 16px 14px' }}>
        <div style={{ fontSize: 13, color: '#0E7C7B', fontWeight: 700, margin: '4px 0 10px' }}>③ 系统给（自动）</div>

        <div style={{ background: 'linear-gradient(135deg,#0E7C7B,#125f5e)', color: '#fff', borderRadius: 14, padding: 16, textAlign: 'center' }}>
          <div style={{ fontSize: 13, opacity: 0.85 }}>★ 透析液钠 C_d（抄到机器）</div>
          <div style={{ fontSize: 46, fontWeight: 800, lineHeight: 1.1, margin: '2px 0' }}>
            {result ? result.cd.toFixed(1) : '—'}
            <span style={{ fontSize: 16, fontWeight: 500, opacity: 0.9 }}> mmol/L</span>
          </div>
          {result && (
            <div style={{ fontSize: 12, opacity: 0.9, marginTop: 4 }}>
              本次清钠 <b>{result.mTarget.toFixed(0)}</b> mmol（搬水 {result.naTarget.toFixed(0)} + 脱载 {result.deload.toFixed(0)}）
              · RNa {result.rNa.toFixed(2)} · 透后 {result.cPost.toFixed(1)}
            </div>
          )}
          {/* 对流/弥散分解条 */}
          {result && (
            <div style={{ display: 'flex', height: 22, borderRadius: 6, overflow: 'hidden', marginTop: 12, fontSize: 11 }}>
              <div style={{ background: '#0a5e5d', flex: conv / totBar, display: 'flex', alignItems: 'center', justifyContent: 'center', whiteSpace: 'nowrap' }}>
                对流 {conv.toFixed(0)}
              </div>
              <div style={{ background: diff >= 0 ? '#B8860B' : '#C0392B', flex: Math.abs(diff) / totBar, display: 'flex', alignItems: 'center', justifyContent: 'center', whiteSpace: 'nowrap' }}>
                弥散 {diff >= 0 ? '+' : ''}{diff.toFixed(0)}
              </div>
            </div>
          )}
        </div>

        {/* 安全灯 */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 6, marginTop: 12 }}>
          {chips.map((c, i) => {
            const col = chipColor(c.level)
            return (
              <div key={i} style={{ display: 'flex', gap: 8, alignItems: 'flex-start', fontSize: 13, padding: '8px 10px', borderRadius: 9, background: col.bg, color: col.fg }}>
                <span style={{ fontSize: 15, lineHeight: 1.2 }}>{c.icon}</span>
                <span>{c.text}</span>
              </div>
            )
          })}
        </div>

        {/* 采纳按钮 */}
        <Button
          type="primary" block size="large" disabled={!result}
          icon={<ThunderboltFilled />}
          onClick={() => { if (result) { onAdopt(result.cdAdopt); auditAdopt(result) } }}
          style={{ marginTop: 12, height: 50, background: '#0E7C7B', borderColor: '#0E7C7B', fontWeight: 700 }}
        >
          {result ? `采纳：将透析液钠设为 ${result.cdAdopt.toFixed(1)} mmol/L` : '请补全透前血钠与超滤量'}
        </Button>

        {/* 高级参数 */}
        <Collapse
          ghost
          size="small"
          style={{ marginTop: 4 }}
          items={[{
            key: 'adv',
            label: <span style={{ fontSize: 12, color: '#5A6B7B' }}>高级参数（默认即可）</span>,
            children: (
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 8 }}>
                <AdvNum label="D 钠弥散度 (L/h)" value={adv.D} step={1} onChange={(v) => setAdv({ ...adv, D: v })} />
                <AdvNum label="T 时长 (h)" value={adv.T} step={0.5} onChange={(v) => setAdv({ ...adv, T: v })} />
                <AdvNum label="α 对流占比" value={adv.alpha} step={0.01} onChange={(v) => setAdv({ ...adv, alpha: v })} />
                <AdvNum label="地板 C_floor" value={adv.floor} step={0.5} onChange={(v) => setAdv({ ...adv, floor: v })} />
                <AdvNum label="C_d 机器上限" value={adv.cdmax} step={1} onChange={(v) => setAdv({ ...adv, cdmax: v })} />
                <AdvNum label="单次降钠上限 δ_cap" value={adv.dcap} step={0.5} onChange={(v) => setAdv({ ...adv, dcap: v })} />
              </div>
            ),
          }]}
        />

        <div style={{ fontSize: 11, color: '#9aa6b2', textAlign: 'center', marginTop: 6, lineHeight: 1.6 }}>
          设计稿原型，非临床定案。落地以在线电导/离子清除反馈闭环校准；高钠或大幅脱载需分次。
        </div>
      </div>
    </div>
  )
}

function AdvNum({ label, value, step, onChange }: {
  label: string; value: number; step: number; onChange: (v: number) => void
}) {
  return (
    <label style={{ display: 'flex', flexDirection: 'column', gap: 3 }}>
      <span style={{ fontSize: 11, color: '#5A6B7B' }}>{label}</span>
      <InputNumber size="small" value={value} step={step} onChange={(v) => onChange(v ?? value)} style={{ width: '100%' }} />
    </label>
  )
}
