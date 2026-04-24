import { message } from 'antd'
import { AlertTriangle, Heart, Scale, Thermometer } from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'
import type { RestTreatment } from '@/services'
import type { TreatmentAfterSignsRequest } from '@/services/restClient'
import type { Patient } from '../types'

interface Props {
  patient: Patient
  treatment: RestTreatment | null
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
  sbp: string
  dbp: string
  heartRate: string
  respiration: string
  temperature: string
  realIntake: string
  pressurePoint: string
  complication: string
  symptoms: string
  notes: string
}

const EMPTY_FORM: PostAssessmentFormState = {
  startTime: '',
  endTime: '',
  realUfVolume: '',
  realSubstituteVolume: '',
  weight: '',
  extraWeight: '',
  lossWeight: '',
  sbp: '',
  dbp: '',
  heartRate: '',
  respiration: '',
  temperature: '',
  realIntake: '',
  pressurePoint: '',
  complication: '',
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

function getSymptomItemValue(
  items: Array<{ code: string; value: string }> | undefined,
  code: string
) {
  return items?.find((item) => item.code === code)?.value || ''
}

function parseBp(bp?: string) {
  if (!bp) return { sbp: '', dbp: '' }
  const [sbp, dbp] = bp.split('/')
  return { sbp: sbp?.trim() || '', dbp: dbp?.trim() || '' }
}

function mapTreatmentToForm(treatment: RestTreatment | null): PostAssessmentFormState {
  if (!treatment) {
    return {
      ...EMPTY_FORM,
      endTime: toDateTimeLocal(new Date().toISOString()),
    }
  }

  const endBp = parseBp(treatment.endBp)

  return {
    startTime: toDateTimeLocal(treatment.startTime),
    endTime: toDateTimeLocal(treatment.endTime || new Date().toISOString()),
    realUfVolume: '',
    realSubstituteVolume: '',
    weight: '',
    extraWeight: '',
    lossWeight: toText(treatment.weightLossKg),
    sbp: endBp.sbp,
    dbp: endBp.dbp,
    heartRate: getSymptomItemValue(treatment.afterSymptomItems, 'heart_rate'),
    respiration: getSymptomItemValue(treatment.afterSymptomItems, 'respiration'),
    temperature: getSymptomItemValue(treatment.afterSymptomItems, 'temperature'),
    realIntake: getSymptomItemValue(treatment.afterSymptomItems, 'real_intake'),
    pressurePoint: getSymptomItemValue(treatment.afterSymptomItems, 'bp_site'),
    complication: getSymptomItemValue(treatment.afterSymptomItems, 'complication') || treatment.complications || '',
    symptoms: getSymptomItemValue(treatment.afterSymptomItems, 'symptoms'),
    notes:
      getSymptomItemValue(treatment.afterSymptomItems, 'notes') ||
      treatment.treatmentSummary ||
      treatment.notes ||
      '',
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
  ]
    .map(([code, value]) => ({ code, value: value.trim() }))
    .filter((item) => item.value)

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
    complication: form.complication.trim() || undefined,
    symptoms: form.symptoms.trim() || undefined,
    notes: form.notes.trim() || undefined,
    symptomItems,
  }
}

function Field({
  label,
  value,
  onChange,
  unit,
  type = 'text',
}: {
  label: string
  value: string
  onChange: (value: string) => void
  unit?: string
  type?: 'text' | 'datetime-local'
}) {
  return (
    <label className="block">
      <div className="mb-2 text-[11px] font-semibold uppercase tracking-wide text-slate-400">{label}</div>
      <div className="relative">
        <input
          type={type}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          className="h-11 w-full rounded-2xl border border-slate-200 px-4 pr-14 text-sm font-semibold text-slate-800 outline-none transition focus:border-blue-400"
        />
        {unit ? <span className="absolute right-4 top-1/2 -translate-y-1/2 text-xs text-slate-400">{unit}</span> : null}
      </div>
    </label>
  )
}

export default function PostAssessment({ patient, treatment, onSave, onSubmit }: Props) {
  const [saving, setSaving] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [form, setForm] = useState<PostAssessmentFormState>(mapTreatmentToForm(treatment))

  useEffect(() => {
    setForm(mapTreatmentToForm(treatment))
  }, [treatment])

  const updateField = (key: keyof PostAssessmentFormState, value: string) => {
    setForm((current) => ({ ...current, [key]: value }))
  }

  const hasEvent = useMemo(
    () => Boolean(form.complication.trim() || form.symptoms.trim()),
    [form.complication, form.symptoms]
  )

  const handleSave = async () => {
    try {
      setSaving(true)
      await onSave(buildPayload(form))
      message.success('透后评估已保存')
    } catch (error) {
      console.error('[PostAssessment] save failed', error)
      message.error('透后评估保存失败')
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
      message.error('透后评估提交失败')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="space-y-6 pb-8">
      <section className="overflow-hidden rounded-3xl border border-slate-200 bg-white shadow-sm">
        <div className="flex items-center gap-2 border-b border-slate-200 px-6 py-4">
          <Scale size={16} className="text-blue-600" />
          <h3 className="text-sm font-black text-slate-800">体重与生命体征</h3>
        </div>
        <div className="grid grid-cols-1 gap-4 p-6 lg:grid-cols-4">
          <Field label="上机时间" type="datetime-local" value={form.startTime} onChange={(value) => updateField('startTime', value)} />
          <Field label="下机时间" type="datetime-local" value={form.endTime} onChange={(value) => updateField('endTime', value)} />
          <Field label="实际超滤" value={form.realUfVolume} onChange={(value) => updateField('realUfVolume', value)} unit="ml" />
          <Field label="实际置换量" value={form.realSubstituteVolume} onChange={(value) => updateField('realSubstituteVolume', value)} unit="ml" />
          <Field label="透后体重" value={form.weight} onChange={(value) => updateField('weight', value)} unit="kg" />
          <Field label="额外体重" value={form.extraWeight} onChange={(value) => updateField('extraWeight', value)} unit="kg" />
          <Field label="体重丢失" value={form.lossWeight} onChange={(value) => updateField('lossWeight', value)} unit="kg" />
          <Field label="实际摄入" value={form.realIntake} onChange={(value) => updateField('realIntake', value)} unit="ml" />
          <Field label="透后收缩压" value={form.sbp} onChange={(value) => updateField('sbp', value)} unit="mmHg" />
          <Field label="透后舒张压" value={form.dbp} onChange={(value) => updateField('dbp', value)} unit="mmHg" />
          <Field label="透后心率" value={form.heartRate} onChange={(value) => updateField('heartRate', value)} unit="次/分" />
          <Field label="透后呼吸" value={form.respiration} onChange={(value) => updateField('respiration', value)} unit="次/分" />
          <Field label="透后体温" value={form.temperature} onChange={(value) => updateField('temperature', value)} unit="℃" />
          <Field label="测压部位" value={form.pressurePoint} onChange={(value) => updateField('pressurePoint', value)} />
          <Field label="目标方案" value={patient.treatmentPlan} onChange={() => undefined} />
        </div>
      </section>

      <section className="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
        <div className="mb-5 flex items-center gap-2">
          <Heart size={16} className="text-rose-500" />
          <h3 className="text-sm font-black text-slate-800">内瘘与临床观察</h3>
        </div>
        <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
          <label className="block">
            <div className="mb-2 text-[11px] font-semibold uppercase tracking-wide text-slate-400">并发症</div>
            <textarea
              rows={5}
              value={form.complication}
              onChange={(e) => updateField('complication', e.target.value)}
              className="w-full resize-none rounded-2xl border border-slate-200 px-4 py-3 text-sm font-semibold text-slate-700 outline-none transition focus:border-blue-400"
            />
          </label>
          <label className="block">
            <div className="mb-2 text-[11px] font-semibold uppercase tracking-wide text-slate-400">症状及处理</div>
            <textarea
              rows={5}
              value={form.symptoms}
              onChange={(e) => updateField('symptoms', e.target.value)}
              className="w-full resize-none rounded-2xl border border-slate-200 px-4 py-3 text-sm font-semibold text-slate-700 outline-none transition focus:border-blue-400"
            />
          </label>
        </div>
      </section>

      <section className="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
        <div className="mb-5 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <AlertTriangle size={16} className="text-amber-500" />
            <h3 className="text-sm font-black text-slate-800">透析事件</h3>
          </div>
          <span
            className={`rounded-full px-3 py-1 text-xs font-semibold ${
              hasEvent ? 'bg-amber-100 text-amber-700' : 'bg-slate-100 text-slate-500'
            }`}
          >
            {hasEvent ? '已记录事件' : '无事件'}
          </span>
        </div>
        <label className="block">
          <div className="mb-2 text-[11px] font-semibold uppercase tracking-wide text-slate-400">透后备注</div>
          <textarea
            rows={4}
            value={form.notes}
            onChange={(e) => updateField('notes', e.target.value)}
            className="w-full resize-none rounded-2xl border border-slate-200 px-4 py-3 text-sm font-semibold text-slate-700 outline-none transition focus:border-blue-400"
          />
        </label>
      </section>

      <div className="rounded-3xl bg-slate-900 px-6 py-5 text-white shadow-lg">
        <div className="flex items-center justify-between gap-4">
          <div className="flex items-center gap-3">
            <Thermometer size={18} className="text-slate-300" />
            <span className="text-sm font-semibold">{patient.name} 的透后评估已接入真实保存</span>
          </div>
          <div className="flex gap-3">
            <button
              type="button"
              onClick={() => void handleSave()}
              disabled={saving || submitting}
              className="rounded-2xl border border-slate-700 bg-slate-800 px-5 py-3 text-sm font-semibold disabled:opacity-60"
            >
              {saving ? '保存中...' : '暂存'}
            </button>
            <button
              type="button"
              onClick={() => void handleSubmit()}
              disabled={saving || submitting}
              className="rounded-2xl bg-blue-600 px-5 py-3 text-sm font-semibold disabled:opacity-60"
            >
              {submitting ? '提交中...' : '提交透后评估'}
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
