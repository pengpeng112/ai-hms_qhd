import { message } from 'antd'
import { Clock, Heart, Scale, Thermometer, AlertTriangle } from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'
import type { RestTreatment } from '@/services'
import { getErrorMessage } from '@/services/restClient'
import type { TreatmentAfterSignsRequest } from '@/services/restClient'
import type { Patient } from '../types'
import { useAuth } from '@/contexts/AuthContext'

interface Props {
  patient: Patient
  treatment: RestTreatment | null
  treatmentLoading?: boolean
  onSave: (payload: TreatmentAfterSignsRequest) => Promise<void>
  onSubmit: (payload: TreatmentAfterSignsRequest) => Promise<void>
}

interface PostAssessmentFormState {
  startTime: string
  endTime: string
  realUfVolume: string
  realSubstituteVolume: string
  weight: string
  extraWeight: string
  lossWeight: string
  postNetWeight: string
  sbp: string
  dbp: string
  heartRate: string
  respiration: string
  temperature: string
  realIntake: string
  pressurePoint: string
  complication: string
  hasDialysisEvent: boolean
  dialyzerCoag: string
  lineACoag: string
  lineVCoag: string
  symptoms: string
  notes: string
}

const COAG_OPTIONS = ['0级', '1级', '2级', '3级']

const EMPTY_FORM: PostAssessmentFormState = {
  startTime: '',
  endTime: '',
  realUfVolume: '',
  realSubstituteVolume: '',
  weight: '',
  extraWeight: '',
  lossWeight: '',
  postNetWeight: '',
  sbp: '',
  dbp: '',
  heartRate: '',
  respiration: '',
  temperature: '',
  realIntake: '',
  pressurePoint: '',
  complication: '',
  hasDialysisEvent: false,
  dialyzerCoag: '',
  lineACoag: '',
  lineVCoag: '',
  symptoms: '',
  notes: '',
}

function toText(value?: string | number | null) {
  if (value === undefined || value === null) return ''
  return String(value)
}

function toDateTimeLocal(value?: string) {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  const local = new Date(date.getTime() - date.getTimezoneOffset() * 60000)
  return local.toISOString().slice(0, 16)
}

function parseOptionalNumber(value: string) {
  const trimmed = value.trim()
  if (!trimmed) return undefined
  const parsed = Number(trimmed)
  return Number.isFinite(parsed) ? parsed : undefined
}

function toIsoOrUndefined(value: string) {
  if (!value.trim()) return undefined
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? undefined : date.toISOString()
}

function getSymptomItemValue(items: Array<{ code: string; value: string }> | undefined, code: string) {
  return items?.find((item) => item.code === code)?.value || ''
}

function parseBp(bp?: string) {
  if (!bp) return { sbp: '', dbp: '' }
  const [sbp, dbp] = bp.split('/')
  return { sbp: sbp?.trim() || '', dbp: dbp?.trim() || '' }
}

function mapTreatmentToForm(treatment: RestTreatment | null): PostAssessmentFormState {
  if (!treatment) return {
    ...EMPTY_FORM,
    endTime: toDateTimeLocal(new Date().toISOString()),
  }
  const endBp = parseBp(treatment.endBp)
  const after = treatment.afterSigns
  const symItems = treatment.afterSymptomItems
  return {
    startTime: toDateTimeLocal(treatment.startTime),
    endTime: toDateTimeLocal(treatment.endTime || new Date().toISOString()),
    realUfVolume: toText(after?.realUfVolume),
    realSubstituteVolume: toText(after?.realSubstituteVolume),
    weight: toText(after?.weight),
    extraWeight: toText(after?.extraWeight),
    lossWeight: toText(after?.lossWeight ?? treatment.weightLossKg),
    postNetWeight: '',
    sbp: toText(after?.sbp ?? endBp.sbp),
    dbp: toText(after?.dbp ?? endBp.dbp),
    heartRate: toText(after?.heartRate ?? getSymptomItemValue(symItems, 'heart_rate')),
    respiration: toText(after?.respiration ?? getSymptomItemValue(symItems, 'respiration')),
    temperature: toText(after?.temperature ?? getSymptomItemValue(symItems, 'temperature')),
    realIntake: toText(after?.realIntake ?? getSymptomItemValue(symItems, 'real_intake')),
    pressurePoint: toText(after?.pressurePoint || getSymptomItemValue(symItems, 'bp_site')),
    complication: after?.complication || getSymptomItemValue(symItems, 'complication') || treatment.complications || '',
    hasDialysisEvent: Boolean(
      after?.complication || getSymptomItemValue(symItems, 'complication') || treatment.complications
    ),
    dialyzerCoag: getSymptomItemValue(symItems, 'dialyzer_coag'),
    lineACoag: getSymptomItemValue(symItems, 'line_a_coag'),
    lineVCoag: getSymptomItemValue(symItems, 'line_v_coag'),
    symptoms: after?.symptoms || getSymptomItemValue(symItems, 'symptoms'),
    notes: after?.notes || getSymptomItemValue(symItems, 'notes') || treatment.treatmentSummary || treatment.notes || '',
  }
}

function buildPayload(form: PostAssessmentFormState): TreatmentAfterSignsRequest {
  const symptomItems = [
    ['bp_site', form.pressurePoint],
    ['real_intake', form.realIntake],
    ['symptoms', form.symptoms],
    ['notes', form.notes],
    ['heart_rate', form.heartRate],
    ['respiration', form.respiration],
    ['temperature', form.temperature],
    ['dialyzer_coag', form.dialyzerCoag],
    ['line_a_coag', form.lineACoag],
    ['line_v_coag', form.lineVCoag],
  ].map(([code, value]) => ({ code, value: value.trim() })).filter((item) => item.value)

  return {
    startTime: toIsoOrUndefined(form.startTime),
    endTime: toIsoOrUndefined(form.endTime),
    realUfVolume: parseOptionalNumber(form.realUfVolume),
    realSubstituteVolume: parseOptionalNumber(form.realSubstituteVolume),
    weight: parseOptionalNumber(form.weight),
    extraWeight: parseOptionalNumber(form.extraWeight),
    lossWeight: parseOptionalNumber(form.lossWeight),
    sbp: parseOptionalNumber(form.sbp),
    dbp: parseOptionalNumber(form.dbp),
    heartRate: parseOptionalNumber(form.heartRate),
    respiration: parseOptionalNumber(form.respiration),
    temperature: parseOptionalNumber(form.temperature),
    realIntake: parseOptionalNumber(form.realIntake),
    pressurePoint: form.pressurePoint.trim() || undefined,
    complication: form.hasDialysisEvent ? (form.complication.trim() || undefined) : '',
    symptoms: form.symptoms.trim() || undefined,
    notes: form.notes.trim() || undefined,
    symptomItems,
  }
}

function Field({ label, value, onChange, unit, type = 'text', required, placeholder }: {
  label: string; value: string; onChange: (value: string) => void; unit?: string; type?: 'text' | 'datetime-local'
  required?: boolean; placeholder?: string
}) {
  return (
    <label className="block min-w-0">
      <span className={`mb-2 block text-xs font-semibold ${required ? 'text-rose-500' : 'text-slate-400'}`}>
        {required ? `* ${label}` : label}
      </span>
      <div className="relative">
        <input
          type={type}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          placeholder={placeholder}
          className="h-10 w-full rounded-lg border border-slate-200 px-3 pr-14 text-sm font-semibold text-slate-800 outline-none transition focus:border-blue-400"
        />
        {unit ? <span className="absolute right-3 top-1/2 -translate-y-1/2 text-xs text-slate-400">{unit}</span> : null}
      </div>
    </label>
  )
}

function CoagSelect({ label, value, onChange }: { label: string; value: string; onChange: (v: string) => void }) {
  return (
    <div>
      <div className="mb-2 text-xs font-semibold text-slate-500">{label}</div>
      <div className="flex gap-2">
        {COAG_OPTIONS.map((opt) => (
          <button
            key={opt}
            type="button"
            onClick={() => onChange(value === opt ? '' : opt)}
            className={`rounded-lg border px-3 py-2 text-xs font-semibold transition ${
              value === opt
                ? 'border-blue-400 bg-blue-50 text-blue-700 shadow-sm'
                : 'border-slate-200 bg-white text-slate-500 hover:border-slate-300'
            }`}
          >
            {opt}
          </button>
        ))}
      </div>
    </div>
  )
}

export default function PostAssessment({ patient, treatment, treatmentLoading = false, onSave, onSubmit }: Props) {
  const { user: currentUser } = useAuth()
  const [saving, setSaving] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [form, setForm] = useState<PostAssessmentFormState>(mapTreatmentToForm(treatment))

  useEffect(() => {
    setForm(mapTreatmentToForm(treatment))
  }, [treatment])

  const updateField = (key: keyof PostAssessmentFormState, value: string | boolean) =>
    setForm((current) => ({ ...current, [key]: value }))

  const toggleDialysisEvent = (has: boolean) => {
    setForm((current) => ({ ...current, hasDialysisEvent: has, complication: has ? current.complication : '' }))
  }

  // 自动计算：透后净重 = 透后体重 - 额外体重
  const calcPostNetWeight = useMemo(() => {
    const w = parseOptionalNumber(form.weight)
    const e = parseOptionalNumber(form.extraWeight) ?? 0
    if (w === undefined) return ''
    return (w - e).toFixed(1)
  }, [form.weight, form.extraWeight])

  // 自动计算：体重丢失 = 透前体重 - 透后净重
  const calcLossWeight = useMemo(() => {
    const preWeight = treatment?.beforeSigns?.weight
    const postNet = parseOptionalNumber(calcPostNetWeight)
    if (preWeight === undefined || preWeight === null || postNet === undefined) return ''
    return (preWeight - postNet).toFixed(1)
  }, [treatment, calcPostNetWeight])

  // 同步自动计算结果到form
  useEffect(() => {
    if (calcPostNetWeight && calcPostNetWeight !== form.postNetWeight) {
      setForm((current) => ({ ...current, postNetWeight: calcPostNetWeight }))
    }
  }, [calcPostNetWeight])

  useEffect(() => {
    if (calcLossWeight && calcLossWeight !== form.lossWeight) {
      setForm((current) => ({ ...current, lossWeight: calcLossWeight }))
    }
  }, [calcLossWeight])

  const hasEvent = useMemo(() => form.hasDialysisEvent, [form.hasDialysisEvent])

  const handleSave = async () => {
    try {
      setSaving(true)
      await onSave(buildPayload(form))
      message.success('透后评估已保存')
    } catch (error) {
      console.error('[PostAssessment] save failed', error)
      message.error(getErrorMessage(error))
    } finally {
      setSaving(false)
    }
  }

  const handleSubmit = async () => {
    try {
      setSubmitting(true)
      await onSubmit(buildPayload(form))
      message.success('透后评估已提交')
    } catch (error) {
      console.error('[PostAssessment] submit failed', error)
      message.error(getErrorMessage(error))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="space-y-5 pb-8">
      {treatmentLoading ? (
        <section className="rounded-lg border border-blue-100 bg-blue-50 px-6 py-4 text-sm font-semibold text-blue-700">
          正在加载新患者治疗数据，透后评估表单已重置为空状态。
        </section>
      ) : null}

      <section className="rounded-lg border border-slate-200 bg-white px-6 py-4 shadow-sm">
        <div className="grid grid-cols-1 gap-4 md:grid-cols-[auto_1fr_1fr] md:items-center">
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-indigo-600 text-white">
              <Clock size={18} />
            </div>
            <div>
              <div className="text-xs font-semibold text-slate-400">治疗时间</div>
              <div className="mt-1 text-sm font-black text-slate-900">
                {form.startTime || '--'} ~ {form.endTime || '--'}
              </div>
              <div className="text-xs text-slate-400">结束时间取超滤量稳定后自动判断</div>
            </div>
          </div>
          <Field label="实际超滤量" value={form.realUfVolume} onChange={(value) => updateField('realUfVolume', value)} unit="ML" placeholder="取超滤量最大值" />
          <Field label="实际置换液量" value={form.realSubstituteVolume} onChange={(value) => updateField('realSubstituteVolume', value)} unit="ML" placeholder="取置换液量最大值" />
        </div>
      </section>

      <section className="overflow-hidden rounded-lg border border-blue-100 bg-white shadow-sm">
        <div className="flex items-center justify-between border-b border-slate-100 px-5 py-4">
          <div className="flex items-center gap-2">
            <span className="h-6 w-1 rounded-full bg-blue-600" />
            <Scale size={16} className="text-blue-600" />
            <h3 className="text-sm font-black text-slate-800">体重与生命体征</h3>
          </div>
          <button type="button" disabled title="功能待后端接口就绪" className="rounded-lg bg-blue-50 px-3 py-2 text-xs font-bold text-blue-600 opacity-60">
            查看下机图片
          </button>
        </div>
        <div className="grid grid-cols-1 gap-5 p-5 md:grid-cols-2 xl:grid-cols-4">
          <div>
            <Field label="透后体重" value={form.weight} onChange={(value) => updateField('weight', value)} unit="KG" placeholder="优先从体重秤获取" />
            <p className="mt-1 text-xs text-slate-400">来自透析室体重秤</p>
          </div>
          <div>
            <Field label="透后净重" value={form.postNetWeight} onChange={(value) => updateField('postNetWeight', value)} unit="KG" placeholder="自动计算" />
            <p className="mt-1 text-xs text-slate-400">= 透后体重 - 额外体重</p>
          </div>
          <div>
            <Field label="体重丢失" value={form.lossWeight} onChange={(value) => updateField('lossWeight', value)} unit="KG" placeholder="自动计算" />
            <p className="mt-1 text-xs text-slate-400">= 透前体重 - 透后净重</p>
          </div>
          <Field label="额外体重" value={form.extraWeight} onChange={(value) => updateField('extraWeight', value)} unit="KG" />
          <div className="xl:col-span-1">
            <div className="mb-2 text-xs font-semibold text-slate-400">透后血压 (MMHG)</div>
            <div className="grid grid-cols-[1fr_auto_1fr] items-center gap-3">
              <input value={form.sbp} onChange={(e) => updateField('sbp', e.target.value)} placeholder="收缩压" className="h-10 rounded-lg border border-slate-200 px-3 text-sm font-semibold outline-none" />
              <span className="text-slate-300">/</span>
              <input value={form.dbp} onChange={(e) => updateField('dbp', e.target.value)} placeholder="舒张压" className="h-10 rounded-lg border border-slate-200 px-3 text-sm font-semibold outline-none" />
            </div>
            <p className="mt-1 text-xs text-slate-400">优先取独立血压仪数据</p>
          </div>
          <Field label="测压部位" value={form.pressurePoint} onChange={(value) => updateField('pressurePoint', value)} />
          <div>
            <Field label="透后心率" value={form.heartRate} onChange={(value) => updateField('heartRate', value)} placeholder="优先取血压仪数据" required />
            <p className="mt-1 text-xs text-slate-400">优先取独立血压仪数据</p>
          </div>
          <Field label="透后呼吸" value={form.respiration} onChange={(value) => updateField('respiration', value)} unit="次/分" required />
          <Field label="透后体温" value={form.temperature} onChange={(value) => updateField('temperature', value)} unit="℃" required placeholder="点击输入" />
          <Field label="实际摄入" value={form.realIntake} onChange={(value) => updateField('realIntake', value)} unit="ML" />
          <Field label="血压实际测量时间" type="datetime-local" value={form.endTime} onChange={(value) => updateField('endTime', value)} />
        </div>
      </section>

      <section className="rounded-lg border border-slate-200 bg-white p-5 shadow-sm">
        <div className="mb-5 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <span className="h-6 w-1 rounded-full bg-orange-500" />
            <Heart size={16} className="text-orange-500" />
            <h3 className="text-sm font-black text-slate-800">临床观察与记录</h3>
          </div>
          <span className={`rounded-full px-3 py-1 text-xs font-bold ${hasEvent ? 'bg-amber-100 text-amber-700' : 'bg-slate-100 text-slate-500'}`}>
            {hasEvent ? '已记录事件' : '无透析负面事件'}
          </span>
        </div>
        <div className="grid grid-cols-1 gap-5 xl:grid-cols-2">
          <div className="space-y-5">
            <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
              <CoagSelect label="透析器凝血分级" value={form.dialyzerCoag} onChange={(v) => updateField('dialyzerCoag', v)} />
              <CoagSelect label="血路管A端凝血分级" value={form.lineACoag} onChange={(v) => updateField('lineACoag', v)} />
              <CoagSelect label="血路管V端凝血分级" value={form.lineVCoag} onChange={(v) => updateField('lineVCoag', v)} />
            </div>
            <div>
              <div className="mb-2 flex items-center gap-4">
                <span className="text-sm font-bold text-slate-700">发生透析事件</span>
                <div className="grid grid-cols-2 rounded-lg bg-slate-100 p-1 text-xs font-bold">
                  <button
                    type="button"
                    onClick={() => toggleDialysisEvent(false)}
                    className={`rounded-md px-3 py-1.5 ${!form.hasDialysisEvent ? 'bg-white text-slate-700 shadow-sm' : 'text-slate-500'}`}
                  >
                    否
                  </button>
                  <button
                    type="button"
                    onClick={() => toggleDialysisEvent(true)}
                    className={`rounded-md px-3 py-1.5 ${form.hasDialysisEvent ? 'bg-rose-500 text-white shadow-sm' : 'text-slate-500'}`}
                  >
                    是
                  </button>
                </div>
              </div>
              {form.hasDialysisEvent ? (
                <textarea
                  value={form.complication}
                  onChange={(e) => updateField('complication', e.target.value)}
                  rows={3}
                  placeholder="请填写透析事件说明..."
                  className="w-full resize-none rounded-lg border border-rose-200 bg-rose-50 px-4 py-3 text-sm font-semibold text-slate-700 outline-none"
                />
              ) : null}
            </div>
            <label className="block">
              <span className="mb-2 block text-sm font-bold text-slate-700">透后备注</span>
              <textarea value={form.notes} onChange={(e) => updateField('notes', e.target.value)} rows={3} placeholder="透后备注、护理观察或交接提醒..." className="w-full resize-none rounded-lg border border-slate-200 px-4 py-3 text-sm font-semibold text-slate-700 outline-none" />
            </label>
          </div>
          <div className="space-y-4">
            <div className="grid grid-cols-[80px_1fr] items-center gap-4">
              <span className="text-sm font-bold text-slate-700">内瘘情况:</span>
              <input value={form.symptoms} onChange={(e) => updateField('symptoms', e.target.value)} placeholder="杂音强、震颤强..." className="h-10 rounded-lg border border-slate-200 px-3 text-sm font-semibold outline-none" />
            </div>
            <label className="flex items-center gap-2 rounded-lg border border-blue-100 bg-blue-50 px-4 py-3 text-sm font-semibold text-blue-700">
              <input type="checkbox" checked readOnly />是否进行内瘘/导管护理健康指导
            </label>
            <div className="rounded-lg border border-amber-100 bg-amber-50 px-4 py-3 text-sm text-amber-700 flex items-center gap-2">
              <AlertTriangle size={14} />
              其他意外情况（管路折叠、渗脱等）字段待后端结构确认
            </div>
          </div>
        </div>
      </section>

      <div className="flex items-center justify-between bg-white px-4 py-4 shadow-sm">
        <div className="flex items-center gap-6 text-sm text-slate-500">
          <span className="flex items-center gap-1">
            <Clock size={14} />
            评估时间：{new Date().toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })}
          </span>
          <span className="inline-flex items-center gap-2">
            <Thermometer size={16} />
            下机护士：{currentUser?.name || '未登录'}
          </span>
          <span>{patient.name}</span>
        </div>
        <div className="flex gap-3">
          <button type="button" onClick={() => void handleSave()} disabled={treatmentLoading || saving || submitting} className="rounded-lg border border-slate-200 px-8 py-2.5 text-sm font-semibold text-slate-500 disabled:opacity-60">
            {saving ? '暂存中...' : '暂存报告'}
          </button>
          <button type="button" onClick={() => void handleSubmit()} disabled={treatmentLoading || saving || submitting} className="rounded-lg bg-blue-600 px-8 py-2.5 text-sm font-bold text-white disabled:opacity-60">
            {submitting ? '提交中...' : '提交结项'}
          </button>
        </div>
      </div>
    </div>
  )
}
