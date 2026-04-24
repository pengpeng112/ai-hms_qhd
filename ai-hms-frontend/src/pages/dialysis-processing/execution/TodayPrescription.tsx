import { message } from 'antd'
import { Droplets, Package, Settings } from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'
import type { RestTreatment } from '@/services'
import { prescriptionApi } from '@/services/orderApi'
import type {
  Prescription,
  PrescriptionMaterial,
  PrescriptionUpdateRequest,
} from '@/services/orderApi'
import type { Patient } from '../types'

interface Props {
  patient: Patient
  treatment: RestTreatment | null
}

interface PrescriptionFormState {
  duration: string
  dryWeight: string
  extraWeight: string
  dialysisMode: string
  bloodFlow: string
  substituteVolume: string
  initialDrug: string
  initialDose: string
  maintenanceDrug: string
  maintenanceDose: string
  infusionRate: string
  infusionTime: string
  dialysateType: string
  dialysateGroup: string
  flowRate: string
  na: string
  ca: string
  k: string
  hco3: string
  glucose: string
  conductivity: string
  temp: string
  volume: string
  notes: string
}

const EMPTY_FORM: PrescriptionFormState = {
  duration: '',
  dryWeight: '',
  extraWeight: '',
  dialysisMode: '',
  bloodFlow: '',
  substituteVolume: '',
  initialDrug: '',
  initialDose: '',
  maintenanceDrug: '',
  maintenanceDose: '',
  infusionRate: '',
  infusionTime: '',
  dialysateType: '',
  dialysateGroup: '',
  flowRate: '',
  na: '',
  ca: '',
  k: '',
  hco3: '',
  glucose: '',
  conductivity: '',
  temp: '',
  volume: '',
  notes: '',
}

function normalizeDateYMD(value?: string) {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value.slice(0, 10)
  return date.toISOString().slice(0, 10)
}

function toText(value?: string | number | null) {
  if (value === undefined || value === null) return ''
  return String(value)
}

function parseOptionalNumber(value: string): number | undefined {
  const trimmed = value.trim()
  if (!trimmed) return undefined
  const parsed = Number(trimmed)
  return Number.isFinite(parsed) ? parsed : undefined
}

function mapPrescriptionToForm(prescription: Prescription | null): PrescriptionFormState {
  if (!prescription) return EMPTY_FORM
  return {
    duration: toText(prescription.duration),
    dryWeight: toText(prescription.dryWeight),
    extraWeight: toText(prescription.extraWeight),
    dialysisMode: prescription.dialysisMode?.mode || '',
    bloodFlow: toText(prescription.dialysisMode?.bloodFlow),
    substituteVolume: toText(prescription.dialysisMode?.substituteVolume),
    initialDrug: prescription.anticoagulant?.initialDrug || '',
    initialDose: prescription.anticoagulant?.initialDose || '',
    maintenanceDrug: prescription.anticoagulant?.maintenanceDrug || '',
    maintenanceDose: prescription.anticoagulant?.maintenanceDose || '',
    infusionRate: prescription.anticoagulant?.infusionRate || '',
    infusionTime: prescription.anticoagulant?.infusionTime || '',
    dialysateType: prescription.parameters?.dialysateType || '',
    dialysateGroup: prescription.parameters?.dialysateGroup || '',
    flowRate: toText(prescription.parameters?.flowRate),
    na: toText(prescription.parameters?.na),
    ca: toText(prescription.parameters?.ca),
    k: toText(prescription.parameters?.k),
    hco3: toText(prescription.parameters?.hco3),
    glucose: prescription.parameters?.glucose || '',
    conductivity: toText(prescription.parameters?.conductivity),
    temp: toText(prescription.parameters?.temp),
    volume: toText(prescription.parameters?.volume),
    notes: prescription.notes || '',
  }
}

function MetricCard({ label, value, unit }: { label: string; value: string; unit?: string }) {
  return (
    <div className="rounded-2xl border border-slate-200 bg-white px-4 py-4 shadow-sm">
      <div className="text-[11px] font-semibold uppercase tracking-wide text-slate-400">{label}</div>
      <div className="mt-2 text-2xl font-black text-slate-800">
        {value || '-'}
        {unit ? <span className="ml-1 text-xs text-slate-400">{unit}</span> : null}
      </div>
    </div>
  )
}

function EditableField({
  label,
  value,
  disabled,
  onChange,
}: {
  label: string
  value: string
  disabled: boolean
  onChange: (value: string) => void
}) {
  return (
    <label className="block">
      <div className="text-[11px] font-semibold uppercase tracking-wide text-slate-400 mb-2">{label}</div>
      <input
        value={value}
        onChange={(e) => onChange(e.target.value)}
        disabled={disabled}
        className={`h-11 w-full rounded-2xl border px-4 text-sm font-semibold outline-none ${
          disabled ? 'border-slate-200 bg-slate-50 text-slate-700' : 'border-blue-300 bg-white'
        }`}
      />
    </label>
  )
}

export default function TodayPrescription({ patient, treatment }: Props) {
  const [editing, setEditing] = useState(false)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [prescriptions, setPrescriptions] = useState<Prescription[]>([])
  const [currentPrescription, setCurrentPrescription] = useState<Prescription | null>(null)
  const [form, setForm] = useState<PrescriptionFormState>(EMPTY_FORM)

  useEffect(() => {
    const loadPrescriptions = async () => {
      setLoading(true)
      try {
        const list = await prescriptionApi.list(patient.id)
        setPrescriptions(list)
        const today = new Date().toISOString().slice(0, 10)
        const todayItem = list.find((item) => normalizeDateYMD(item.prescriptionDate) === today)
        const active = todayItem ?? list[0] ?? null
        setCurrentPrescription(active)
        setForm(mapPrescriptionToForm(active))
      } catch (error) {
        console.error('[TodayPrescription] load failed', error)
        message.error('处方加载失败')
      } finally {
        setLoading(false)
      }
    }
    void loadPrescriptions()
  }, [patient.id])

  useEffect(() => {
    setForm(mapPrescriptionToForm(currentPrescription))
  }, [currentPrescription])

  const metrics = useMemo(() => {
    const beforeWeight = treatment?.beforeSigns?.weight
    const dryWeight = parseOptionalNumber(form.dryWeight) ?? patient.dryWeight
    const weightGain =
      beforeWeight !== undefined && dryWeight !== undefined ? (beforeWeight - dryWeight).toFixed(1) : ''
    const preBp =
      treatment?.beforeSigns?.sbp && treatment?.beforeSigns?.dbp
        ? `${treatment.beforeSigns.sbp}/${treatment.beforeSigns.dbp}`
        : ''
    return {
      dialysisMethod: form.dialysisMode || patient.treatmentPlan,
      preWeight: toText(beforeWeight),
      dryWeight: toText(dryWeight),
      weightGain,
      targetUf: form.extraWeight,
      preBp,
      duration: form.duration,
      bloodFlow: form.bloodFlow,
    }
  }, [form, patient.dryWeight, patient.treatmentPlan, treatment])

  const materials: PrescriptionMaterial[] = currentPrescription?.materials || []

  const updateField = (key: keyof PrescriptionFormState, value: string) => {
    setForm((current) => ({ ...current, [key]: value }))
  }

  const handleExtractToday = async () => {
    try {
      setLoading(true)
      const extracted = await prescriptionApi.extract(patient.id, new Date().toISOString().slice(0, 10))
      setCurrentPrescription(extracted)
      setForm(mapPrescriptionToForm(extracted))
      setPrescriptions((items) => {
        const exists = items.some((item) => item.id === extracted.id)
        return exists ? items.map((item) => (item.id === extracted.id ? extracted : item)) : [extracted, ...items]
      })
      message.success('已提取今日处方')
    } catch (error) {
      console.error('[TodayPrescription] extract failed', error)
      message.error('提取今日处方失败')
    } finally {
      setLoading(false)
    }
  }

  const handleSave = async () => {
    if (!currentPrescription) {
      message.warning('暂无可编辑处方，请先提取今日处方')
      return
    }
    try {
      setSaving(true)
      const payload: PrescriptionUpdateRequest = {
        duration: parseOptionalNumber(form.duration),
        dryWeight: parseOptionalNumber(form.dryWeight),
        extraWeight: parseOptionalNumber(form.extraWeight),
        dialysisMode: {
          ...currentPrescription.dialysisMode,
          mode: form.dialysisMode,
          bloodFlow: parseOptionalNumber(form.bloodFlow) ?? 0,
          substituteVolume: parseOptionalNumber(form.substituteVolume),
        },
        anticoagulant: {
          ...currentPrescription.anticoagulant,
          initialDrug: form.initialDrug,
          initialDose: form.initialDose,
          maintenanceDrug: form.maintenanceDrug,
          maintenanceDose: form.maintenanceDose,
          infusionRate: form.infusionRate,
          infusionTime: form.infusionTime,
        },
        parameters: {
          ...currentPrescription.parameters,
          dialysateType: form.dialysateType,
          dialysateGroup: form.dialysateGroup,
          flowRate: parseOptionalNumber(form.flowRate) ?? 0,
          na: parseOptionalNumber(form.na) ?? 0,
          ca: parseOptionalNumber(form.ca) ?? 0,
          k: parseOptionalNumber(form.k) ?? 0,
          hco3: parseOptionalNumber(form.hco3) ?? 0,
          glucose: form.glucose,
          conductivity: parseOptionalNumber(form.conductivity) ?? 0,
          temp: parseOptionalNumber(form.temp) ?? 0,
          volume: parseOptionalNumber(form.volume) ?? 0,
        },
        notes: form.notes,
      }
      const updated = await prescriptionApi.update(patient.id, currentPrescription.id, payload)
      setCurrentPrescription(updated)
      setPrescriptions((items) => items.map((item) => (item.id === updated.id ? updated : item)))
      setEditing(false)
      message.success('处方已保存')
    } catch (error) {
      console.error('[TodayPrescription] save failed', error)
      message.error('处方保存失败')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="space-y-6 pb-8">
      <div className="flex justify-between items-center rounded-3xl border border-slate-200 bg-white px-6 py-4 shadow-sm">
        <div>
          <div className="text-xs font-semibold uppercase tracking-wide text-slate-400">当前患者</div>
          <div className="mt-1 text-xl font-black text-slate-800">{patient.name}</div>
          <div className="mt-1 text-xs text-slate-400">处方总数：{prescriptions.length}</div>
        </div>
        <div className="flex gap-3">
          <button
            type="button"
            onClick={() => void handleExtractToday()}
            className="px-4 py-2 rounded-xl border border-slate-200 text-sm font-semibold text-slate-700 bg-white"
          >
            提取今日处方
          </button>
          {editing ? (
            <>
              <button
                type="button"
                onClick={() => {
                  setEditing(false)
                  setForm(mapPrescriptionToForm(currentPrescription))
                }}
                className="px-4 py-2 rounded-xl border border-slate-200 text-sm font-semibold text-slate-700 bg-white"
              >
                取消
              </button>
              <button
                type="button"
                onClick={() => void handleSave()}
                disabled={saving}
                className="px-4 py-2 rounded-xl bg-blue-600 text-sm font-semibold text-white disabled:opacity-60"
              >
                {saving ? '保存中...' : '保存处方'}
              </button>
            </>
          ) : (
            <button
              type="button"
              onClick={() => setEditing(true)}
              className="px-4 py-2 rounded-xl bg-slate-100 text-sm font-semibold text-slate-700"
            >
              调整处方
            </button>
          )}
        </div>
      </div>

      {loading ? (
        <div className="rounded-3xl border border-slate-200 bg-white p-10 text-center text-slate-500">
          正在加载处方...
        </div>
      ) : !currentPrescription ? (
        <div className="rounded-3xl border border-slate-200 bg-white p-10 text-center text-slate-500">
          未找到当日处方，可先点击“提取今日处方”。
        </div>
      ) : (
        <>
          <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
            <MetricCard label="透析方法" value={metrics.dialysisMethod} />
            <MetricCard label="透前体重" value={metrics.preWeight} unit="kg" />
            <MetricCard label="干体重" value={metrics.dryWeight} unit="kg" />
            <MetricCard label="较干体重增量" value={metrics.weightGain} unit="kg" />
            <MetricCard label="透前血压" value={metrics.preBp} unit="mmHg" />
          </div>

          <section className="rounded-3xl border border-slate-200 bg-white shadow-sm overflow-hidden">
            <div className="px-6 py-4 border-b border-slate-200 flex items-center gap-2">
              <Droplets size={16} className="text-blue-600" />
              <h3 className="text-sm font-black text-slate-800">抗凝方案</h3>
            </div>
            <div className="grid grid-cols-1 lg:grid-cols-4 gap-4 p-6">
              <EditableField label="透析时长" value={form.duration} disabled={!editing} onChange={(value) => updateField('duration', value)} />
              <EditableField label="标准血流量" value={form.bloodFlow} disabled={!editing} onChange={(value) => updateField('bloodFlow', value)} />
              <EditableField label="首剂名称" value={form.initialDrug} disabled={!editing} onChange={(value) => updateField('initialDrug', value)} />
              <EditableField label="首剂量" value={form.initialDose} disabled={!editing} onChange={(value) => updateField('initialDose', value)} />
              <EditableField label="维持剂" value={form.maintenanceDrug} disabled={!editing} onChange={(value) => updateField('maintenanceDrug', value)} />
              <EditableField label="维持量" value={form.maintenanceDose} disabled={!editing} onChange={(value) => updateField('maintenanceDose', value)} />
              <EditableField label="注入速率" value={form.infusionRate} disabled={!editing} onChange={(value) => updateField('infusionRate', value)} />
              <EditableField label="注入时间" value={form.infusionTime} disabled={!editing} onChange={(value) => updateField('infusionTime', value)} />
            </div>
          </section>

          <section className="rounded-3xl border border-slate-200 bg-white shadow-sm overflow-hidden">
            <div className="px-6 py-4 border-b border-slate-200 flex items-center gap-2">
              <Settings size={16} className="text-emerald-600" />
              <h3 className="text-sm font-black text-slate-800">透析参数与通路</h3>
            </div>
            <div className="grid grid-cols-1 lg:grid-cols-4 gap-4 p-6">
              <EditableField label="透析模式" value={form.dialysisMode} disabled={!editing} onChange={(value) => updateField('dialysisMode', value)} />
              <EditableField label="干体重" value={form.dryWeight} disabled={!editing} onChange={(value) => updateField('dryWeight', value)} />
              <EditableField label="超滤目标" value={form.extraWeight} disabled={!editing} onChange={(value) => updateField('extraWeight', value)} />
              <EditableField label="置换量" value={form.substituteVolume} disabled={!editing} onChange={(value) => updateField('substituteVolume', value)} />
              <EditableField label="透析液类型" value={form.dialysateType} disabled={!editing} onChange={(value) => updateField('dialysateType', value)} />
              <EditableField label="透析液分组" value={form.dialysateGroup} disabled={!editing} onChange={(value) => updateField('dialysateGroup', value)} />
              <EditableField label="透析液流量" value={form.flowRate} disabled={!editing} onChange={(value) => updateField('flowRate', value)} />
              <EditableField label="Na" value={form.na} disabled={!editing} onChange={(value) => updateField('na', value)} />
              <EditableField label="Ca" value={form.ca} disabled={!editing} onChange={(value) => updateField('ca', value)} />
              <EditableField label="K" value={form.k} disabled={!editing} onChange={(value) => updateField('k', value)} />
              <EditableField label="HCO3" value={form.hco3} disabled={!editing} onChange={(value) => updateField('hco3', value)} />
              <EditableField label="葡萄糖" value={form.glucose} disabled={!editing} onChange={(value) => updateField('glucose', value)} />
              <EditableField label="电导率" value={form.conductivity} disabled={!editing} onChange={(value) => updateField('conductivity', value)} />
              <EditableField label="温度" value={form.temp} disabled={!editing} onChange={(value) => updateField('temp', value)} />
              <EditableField label="透析液总量" value={form.volume} disabled={!editing} onChange={(value) => updateField('volume', value)} />
              <EditableField label="备注" value={form.notes} disabled={!editing} onChange={(value) => updateField('notes', value)} />
            </div>
          </section>

          <section className="rounded-3xl border border-slate-200 bg-white shadow-sm overflow-hidden">
            <div className="px-6 py-4 border-b border-slate-200 flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Package size={16} className="text-indigo-600" />
                <h3 className="text-sm font-black text-slate-800">处方耗材</h3>
              </div>
              <div className="text-xs text-slate-400">
                处方日期：{normalizeDateYMD(currentPrescription.prescriptionDate)}
              </div>
            </div>
            <div className="overflow-x-auto">
              <table className="w-full text-left min-w-[900px]">
                <thead className="bg-slate-50 text-xs text-slate-500 uppercase tracking-wide">
                  <tr>
                    <th className="px-6 py-3">名称</th>
                    <th className="px-6 py-3">分类</th>
                    <th className="px-6 py-3">数量</th>
                    <th className="px-6 py-3">编码</th>
                    <th className="px-6 py-3">品牌</th>
                    <th className="px-6 py-3">规格</th>
                    <th className="px-6 py-3">备注</th>
                  </tr>
                </thead>
                <tbody>
                  {materials.length > 0 ? (
                    materials.map((item) => (
                      <tr key={item.id} className="border-t border-slate-100 text-sm">
                        <td className="px-6 py-4 font-semibold text-slate-800">{item.name}</td>
                        <td className="px-6 py-4 text-slate-600">{item.category}</td>
                        <td className="px-6 py-4 text-slate-600">{item.count}</td>
                        <td className="px-6 py-4 text-slate-600">{item.code || '-'}</td>
                        <td className="px-6 py-4 text-slate-600">{item.brand || '-'}</td>
                        <td className="px-6 py-4 text-slate-600">{item.spec || '-'}</td>
                        <td className="px-6 py-4 text-slate-600">{item.note || '-'}</td>
                      </tr>
                    ))
                  ) : (
                    <tr>
                      <td colSpan={7} className="px-6 py-8 text-center text-sm text-slate-400">
                        当前处方没有耗材明细
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          </section>
        </>
      )}
    </div>
  )
}
