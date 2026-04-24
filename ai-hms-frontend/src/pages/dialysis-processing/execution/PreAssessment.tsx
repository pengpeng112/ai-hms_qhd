import { message } from 'antd'
import { Activity, Heart, Scale, Stethoscope, Thermometer } from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'
import type { RestTreatment } from '@/services'
import type { TreatmentBeforeSignsRequest } from '@/services/restClient'
import type { Patient, PreAssessmentFormValue } from '../types'

function Card({ title, icon, children }: { title: string; icon: React.ReactNode; children: React.ReactNode }) {
  return (
    <section className="bg-white rounded-3xl border border-slate-200 shadow-sm p-6">
      <div className="flex items-center gap-2 mb-5">
        {icon}
        <h3 className="text-sm font-black text-slate-800 tracking-wide">{title}</h3>
      </div>
      {children}
    </section>
  )
}

function Input({
  label,
  value,
  suffix,
  onChange,
  disabled,
}: {
  label: string
  value: string
  suffix?: string
  onChange?: (value: string) => void
  disabled?: boolean
}) {
  return (
    <label className="block">
      <div className="text-[11px] font-semibold uppercase tracking-wide text-slate-400 mb-2">{label}</div>
      <div className="flex items-center rounded-2xl border border-slate-200 bg-slate-50 px-4 h-12">
        <input
          value={value}
          onChange={(e) => onChange?.(e.target.value)}
          disabled={disabled}
          className="w-full bg-transparent outline-none text-sm font-bold text-slate-800 disabled:text-slate-400"
        />
        {suffix ? <span className="text-xs font-semibold text-slate-400">{suffix}</span> : null}
      </div>
    </label>
  )
}

const EMPTY_FORM: PreAssessmentFormValue = {
  weight: '',
  extraWeight: '',
  targetUf: '',
  sbp: '',
  dbp: '',
  heartRate: '',
  respiration: '',
  temperature: '',
  pressurePoint: '',
  aSite: '',
  vSite: '',
  consciousness: '',
  nurseLevel: '',
  notes: '',
  symptoms: [],
}

function getSymptomItemValue(items: Array<{ code: string; value: string }> | undefined, code: string) {
  return items?.find((item) => item.code === code)?.value ?? ''
}

function toText(value?: number | string | null) {
  if (value === undefined || value === null) return ''
  return String(value)
}

function mapTreatmentToForm(treatment: RestTreatment | null): PreAssessmentFormValue {
  if (!treatment) return EMPTY_FORM
  const before = treatment.beforeSigns
  const symptomItems = treatment.beforeSymptomItems
  return {
    weight: toText(before?.weight),
    extraWeight: toText(before?.extraWeight),
    targetUf: getSymptomItemValue(symptomItems, 'uf_volume'),
    sbp: toText(before?.sbp),
    dbp: toText(before?.dbp),
    heartRate: toText(before?.heartRate),
    respiration: toText(before?.respiration),
    temperature: toText(before?.temperature),
    pressurePoint: toText(before?.pressurePoint),
    aSite: getSymptomItemValue(symptomItems, 'a_site'),
    vSite: getSymptomItemValue(symptomItems, 'v_site'),
    consciousness: getSymptomItemValue(symptomItems, 'consciousness'),
    nurseLevel: getSymptomItemValue(symptomItems, 'nurse_level'),
    notes: before?.notes ?? '',
    symptoms: (before?.symptoms || getSymptomItemValue(symptomItems, 'symptoms'))
      .split(/[，,；;、]/)
      .map((item) => item.trim())
      .filter(Boolean),
  }
}

function parseOptionalNumber(value: string): number | undefined {
  const trimmed = value.trim()
  if (!trimmed) return undefined
  const parsed = Number(trimmed)
  return Number.isFinite(parsed) ? parsed : undefined
}

interface Props {
  patient: Patient
  treatment: RestTreatment | null
  saving?: boolean
  onSave: (payload: TreatmentBeforeSignsRequest) => Promise<void>
}

export default function PreAssessment({ patient, treatment, saving = false, onSave }: Props) {
  const [form, setForm] = useState<PreAssessmentFormValue>(EMPTY_FORM)

  useEffect(() => {
    setForm(mapTreatmentToForm(treatment))
  }, [treatment])

  const weightGain = useMemo(() => {
    const weight = Number(form.weight)
    if (!Number.isFinite(weight) || !patient.dryWeight) return ''
    return (weight - patient.dryWeight).toFixed(1)
  }, [form.weight, patient.dryWeight])

  const updateField = (key: keyof PreAssessmentFormValue, value: string | string[]) => {
    setForm((current) => ({ ...current, [key]: value }))
  }

  const handleSave = async () => {
    try {
      await onSave({
        weight: parseOptionalNumber(form.weight),
        extraWeight: parseOptionalNumber(form.extraWeight),
        sbp: parseOptionalNumber(form.sbp),
        dbp: parseOptionalNumber(form.dbp),
        heartRate: parseOptionalNumber(form.heartRate),
        respiration: parseOptionalNumber(form.respiration),
        temperature: parseOptionalNumber(form.temperature),
        pressurePoint: form.pressurePoint.trim() || undefined,
        notes: form.notes.trim() || undefined,
        symptomItems: [
          { code: 'uf_volume', value: form.targetUf.trim() },
          { code: 'a_site', value: form.aSite.trim() },
          { code: 'v_site', value: form.vSite.trim() },
          { code: 'consciousness', value: form.consciousness.trim() },
          { code: 'nurse_level', value: form.nurseLevel.trim() },
          { code: 'symptoms', value: form.symptoms.join('，') },
        ].filter((item) => item.value),
      })
    } catch (error) {
      console.error('[PreAssessment] save failed', error)
      message.error('透前评估保存失败')
    }
  }

  return (
    <div className="space-y-6 pb-8">
      <div className="grid grid-cols-1 xl:grid-cols-4 gap-6">
        <Card title="体重与容量评估" icon={<Scale size={18} className="text-blue-600" />}>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Input label="透前体重" value={form.weight} onChange={(value) => updateField('weight', value)} suffix="kg" />
            <Input label="干体重" value={String(patient.dryWeight)} suffix="kg" disabled />
            <Input label="体重增长" value={weightGain} suffix="kg" disabled />
            <Input label="目标超滤量" value={form.targetUf} onChange={(value) => updateField('targetUf', value)} suffix="L" />
          </div>
        </Card>

        <Card title="生命体征" icon={<Activity size={18} className="text-rose-500" />}>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Input label="透前收缩压" value={form.sbp} onChange={(value) => updateField('sbp', value)} suffix="mmHg" />
            <Input label="透前舒张压" value={form.dbp} onChange={(value) => updateField('dbp', value)} suffix="mmHg" />
            <Input label="测压部位" value={form.pressurePoint} onChange={(value) => updateField('pressurePoint', value)} />
            <Input label="透前心率" value={form.heartRate} onChange={(value) => updateField('heartRate', value)} suffix="次/分" />
            <Input label="呼吸" value={form.respiration} onChange={(value) => updateField('respiration', value)} suffix="次/分" />
            <Input label="透前体温" value={form.temperature} onChange={(value) => updateField('temperature', value)} suffix="℃" />
          </div>
        </Card>

        <Card title="通路与症状" icon={<Stethoscope size={18} className="text-emerald-600" />}>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Input label="A 端位点" value={form.aSite} onChange={(value) => updateField('aSite', value)} />
            <Input label="V 端位点" value={form.vSite} onChange={(value) => updateField('vSite', value)} />
            <Input label="神志状态" value={form.consciousness} onChange={(value) => updateField('consciousness', value)} />
            <Input label="护理分级" value={form.nurseLevel} onChange={(value) => updateField('nurseLevel', value)} />
          </div>
          <div className="mt-5">
            <div className="text-[11px] font-semibold uppercase tracking-wide text-slate-400 mb-2">透前症状记录</div>
            <div className="min-h-14 rounded-2xl border border-slate-200 bg-slate-50 p-3 flex flex-wrap gap-2">
              {form.symptoms.map((item) => (
                <span key={item} className="inline-flex items-center rounded-xl bg-white border border-slate-200 px-3 py-1 text-xs font-semibold text-slate-700">
                  {item}
                  <button
                    type="button"
                    onClick={() => updateField('symptoms', form.symptoms.filter((symptom) => symptom !== item))}
                    className="ml-2 text-slate-400 hover:text-slate-600"
                  >
                    ×
                  </button>
                </span>
              ))}
              <button
                type="button"
                onClick={() => updateField('symptoms', [...form.symptoms, `新增症状 ${form.symptoms.length + 1}`])}
                className="inline-flex items-center rounded-xl border border-dashed border-blue-300 px-3 py-1 text-xs font-semibold text-blue-600"
              >
                添加
              </button>
            </div>
          </div>
          <div className="mt-5">
            <div className="text-[11px] font-semibold uppercase tracking-wide text-slate-400 mb-2">透前备注</div>
            <textarea
              value={form.notes}
              onChange={(e) => updateField('notes', e.target.value)}
              rows={4}
              className="w-full rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm font-semibold text-slate-700 outline-none resize-none"
            />
          </div>
        </Card>
      </div>

      <section className="bg-slate-900 rounded-[32px] text-white px-8 py-6 shadow-xl">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
          <div className="flex items-center gap-3">
            <Heart className="text-rose-300" size={18} />
            <div>
              <div className="text-[11px] uppercase tracking-wide text-slate-400">患者状态</div>
              <div className="text-sm font-bold">{patient.status}</div>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <Thermometer className="text-amber-300" size={18} />
            <div>
              <div className="text-[11px] uppercase tracking-wide text-slate-400">透析方案</div>
              <div className="text-sm font-bold">{patient.treatmentPlan}</div>
            </div>
          </div>
          <div className="md:col-span-2 flex items-center justify-end gap-3">
            <button type="button" className="px-5 py-3 rounded-2xl bg-slate-800 text-sm font-semibold border border-slate-700">暂存</button>
            <button
              type="button"
              onClick={() => void handleSave()}
              disabled={saving}
              className="px-5 py-3 rounded-2xl bg-blue-600 text-sm font-semibold disabled:opacity-60"
            >
              {saving ? '保存中...' : '提交透前评估'}
            </button>
          </div>
        </div>
      </section>
    </div>
  )
}
