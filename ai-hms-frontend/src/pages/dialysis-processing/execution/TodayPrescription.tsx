import { message } from 'antd'
import { Activity, Box, Clock, Droplets, Package, RotateCw, Settings } from 'lucide-react'
import { type ReactNode, useEffect, useMemo, useState } from 'react'
import type { RestTreatment } from '@/services'
import { getErrorMessage } from '@/services/restClient'
import { prescriptionApi } from '@/services/orderApi'
import { patientApi, type TreatmentPlan } from '@/services/patientApi'
import type {
  Prescription,
  PrescriptionMaterial,
  PrescriptionUpdateRequest,
} from '@/services/orderApi'
import type { Patient } from '../types'

interface Props {
  patient: Patient
  treatment: RestTreatment | null
  treatmentLoading?: boolean
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

function mapPlanToForm(plan: TreatmentPlan | null): PrescriptionFormState {
  if (!plan) return EMPTY_FORM
  return {
    duration: toText(plan.duration),
    dryWeight: toText(plan.dryWeight),
    extraWeight: toText(plan.extraWeight),
    dialysisMode: plan.dialysisMode?.mode || '',
    bloodFlow: toText(plan.dialysisMode?.bloodFlow),
    substituteVolume: toText(plan.dialysisMode?.substituteVolume),
    initialDrug: plan.anticoagulant?.initialDrug || '',
    initialDose: plan.anticoagulant?.initialDose || '',
    maintenanceDrug: plan.anticoagulant?.maintenanceDrug || '',
    maintenanceDose: plan.anticoagulant?.maintenanceDose || '',
    infusionRate: plan.anticoagulant?.infusionRate || '',
    infusionTime: plan.anticoagulant?.infusionTime || '',
    dialysateType: plan.parameters?.dialysateType || '',
    dialysateGroup: plan.parameters?.dialysateGroup || '',
    flowRate: toText(plan.parameters?.flowRate),
    na: toText(plan.parameters?.na),
    ca: toText(plan.parameters?.ca),
    k: toText(plan.parameters?.k),
    hco3: toText(plan.parameters?.hco3),
    glucose: plan.parameters?.glucose || '',
    conductivity: toText(plan.parameters?.conductivity),
    temp: toText(plan.parameters?.temp),
    volume: toText(plan.parameters?.volume),
    notes: plan.notes || '',
  }
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

function MetricCard({ label, value, unit, primary }: { label: string; value: string; unit?: string; primary?: boolean }) {
  return (
    <div className={`rounded-xl border p-3.5 ${primary ? 'border-blue-500 bg-blue-600 text-white' : 'border-slate-200 bg-white'}`}>
      <div className={`text-[11px] font-semibold ${primary ? 'text-blue-100' : 'text-slate-400'}`}>{label}</div>
      <div className={`mt-1.5 text-xl font-black ${primary ? 'text-white' : 'text-slate-900'}`}>
        {value || '--'}
        {unit ? <span className={`ml-1 text-[10px] font-semibold ${primary ? 'text-blue-100' : 'text-slate-400'}`}>{unit}</span> : null}
      </div>
    </div>
  )
}

function EditableField({
  label,
  value,
  unit,
  disabled,
  onChange,
  compact,
}: {
  label: string
  value: string
  unit?: string
  disabled: boolean
  onChange: (value: string) => void
  compact?: boolean
}) {
  return (
    <label className="block min-w-0">
      <div className={`${compact ? 'mb-1' : 'mb-2'} text-[11px] font-semibold text-slate-400`}>{label}</div>
      <div className="relative">
        <input
          value={disabled ? value || '--' : value}
          onChange={(e) => onChange(e.target.value)}
          disabled={disabled}
          className={`${compact ? 'h-9' : 'h-10'} w-full rounded-lg border px-3 text-sm font-bold outline-none ${unit ? 'pr-14' : ''} ${
            disabled ? 'border-transparent bg-transparent text-slate-900' : 'border-blue-300 bg-white text-slate-900 focus:border-blue-500'
          }`}
        />
        {unit ? <span className="absolute right-3 top-1/2 -translate-y-1/2 text-xs font-medium text-slate-400">{unit}</span> : null}
      </div>
    </label>
  )
}

function ValueGroup({ title, children }: { title: string; children: ReactNode }) {
  return (
    <div className="rounded-xl border border-slate-200 bg-slate-50 p-3.5">
      <div className="mb-2.5 text-[11px] font-bold uppercase tracking-wide text-slate-500">{title}</div>
      <div className="space-y-3">{children}</div>
    </div>
  )
}

function SectionCard({
  title,
  icon,
  children,
  right,
}: {
  title: string
  icon: ReactNode
  children: ReactNode
  right?: ReactNode
}) {
  return (
    <section className="overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm">
      <div className="flex items-center justify-between border-b border-slate-100 px-4 py-3">
        <div className="flex items-center gap-2 text-sm font-bold text-slate-800">
          {icon}
          {title}
        </div>
        {right}
      </div>
      {children}
    </section>
  )
}

export default function TodayPrescription({ patient, treatment, treatmentLoading = false }: Props) {
  const [editing, setEditing] = useState(false)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [executing, setExecuting] = useState(false)
  const [treatmentPlan, setTreatmentPlan] = useState<TreatmentPlan | null>(null)
  const [prescriptions, setPrescriptions] = useState<Prescription[]>([])
  const [currentPrescription, setCurrentPrescription] = useState<Prescription | null>(null)
  const [form, setForm] = useState<PrescriptionFormState>(EMPTY_FORM)

  useEffect(() => {
    const loadData = async () => {
      setLoading(true)
      try {
        const [plan, prescList] = await Promise.all([
          patientApi.getTreatmentPlan(patient.id).catch(() => null),
          prescriptionApi.list(patient.id).catch(() => [] as Prescription[]),
        ])
        setTreatmentPlan(plan)
        setPrescriptions(prescList)

        const today = new Date().toISOString().slice(0, 10)
        const todayItem = prescList.find((item) => normalizeDateYMD(item.prescriptionDate) === today)
        const active = todayItem ?? prescList[0] ?? null
        setCurrentPrescription(active)

        if (plan) {
          setForm(mapPlanToForm(plan))
        } else if (active) {
          setForm(mapPrescriptionToForm(active))
        } else {
          setForm(EMPTY_FORM)
        }
      } catch (error) {
        console.error('[TodayPrescription] load failed', error)
        message.error(getErrorMessage(error))
      } finally {
        setLoading(false)
      }
    }
    void loadData()
  }, [patient.id])

  useEffect(() => {
    if (treatmentPlan && !currentPrescription) {
      setForm(mapPlanToForm(treatmentPlan))
    }
  }, [treatmentPlan, currentPrescription])

  const metrics = useMemo(() => {
    const preWeight = treatment?.beforeSigns?.weight || 0
    const extra = parseOptionalNumber(form.extraWeight) ?? 0
    const preNetWeight = preWeight - extra

    const dryWeight = parseOptionalNumber(form.dryWeight) ?? treatmentPlan?.dryWeight ?? patient.dryWeight

    const lastPostWeight = 0
    const weightChange = preNetWeight - lastPostWeight

    const preBp =
      treatment?.beforeSigns?.sbp && treatment?.beforeSigns?.dbp
        ? `${treatment.beforeSigns.sbp}/${treatment.beforeSigns.dbp}`
        : ''

    const vascularAccess = treatmentPlan?.vascularAccessId ? '已配置' : '未配置'

    return {
      dialysisMethod: form.dialysisMode || treatmentPlan?.dialysisMode?.mode || patient.treatmentPlan,
      preWeight,
      preNetWeight,
      lastPostWeight,
      weightChange,
      dryWeight,
      targetUf: form.extraWeight,
      preBp,
      duration: form.duration,
      bloodFlow: form.bloodFlow,
      vascularAccess,
    }
  }, [form, treatmentPlan, patient.dryWeight, patient.treatmentPlan, treatment])

  const materials: PrescriptionMaterial[] = currentPrescription?.materials || []
  const planMaterials = treatmentPlan?.materials || []
  const displayMaterials = materials.length > 0 ? materials : planMaterials.map((m) => ({
    id: m.id || '',
    name: m.name || '',
    category: m.category || '',
    count: m.count ?? 0,
    code: m.code || '',
    brand: m.brand || '',
    spec: m.spec || '',
    note: m.note || '',
  }))
  const canExecute = currentPrescription?.status === '待执行'
  const hasAnticoagulant = form.initialDrug || form.initialDose || form.maintenanceDrug || form.maintenanceDose || form.infusionRate || form.infusionTime
  const formattedUpdatedAt = currentPrescription?.updatedAt
    ? new Date(currentPrescription.updatedAt).toLocaleString('zh-CN')
    : '--'

  const updateField = (key: keyof PrescriptionFormState, value: string) => {
    setForm((current) => ({ ...current, [key]: value }))
  }

  const handleExtractToday = async () => {
    try {
      setLoading(true)
      const extracted = await prescriptionApi.extract(patient.id, new Date().toISOString().slice(0, 10))
      setCurrentPrescription(extracted)
      if (treatmentPlan) {
        setForm(mapPlanToForm(treatmentPlan))
      } else {
        setForm(mapPrescriptionToForm(extracted))
      }
      setPrescriptions((items) => {
        const exists = items.some((item) => item.id === extracted.id)
        return exists ? items.map((item) => (item.id === extracted.id ? extracted : item)) : [extracted, ...items]
      })
      message.success('已提取今日处方')
    } catch (error) {
      console.error('[TodayPrescription] extract failed', error)
      message.error(getErrorMessage(error))
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
      message.error(getErrorMessage(error))
    } finally {
      setSaving(false)
    }
  }

  const handleExecute = async () => {
    if (!currentPrescription) {
      message.warning('暂无可执行处方')
      return
    }
    try {
      setExecuting(true)
      const executed = await prescriptionApi.execute(patient.id, currentPrescription.id)
      setCurrentPrescription(executed)
      setPrescriptions((items) => items.map((item) => (item.id === executed.id ? executed : item)))
      message.success('处方已执行')
    } catch (error) {
      console.error('[TodayPrescription] execute failed', error)
      message.error(getErrorMessage(error))
    } finally {
      setExecuting(false)
    }
  }

  return (
    <div className="space-y-4 pb-8">
      {treatmentLoading ? (
        <section className="rounded-lg border border-blue-100 bg-blue-50 px-4 py-3 text-sm font-semibold text-blue-700">
          正在加载新患者治疗上下文，处方概览已清空旧治疗数据。
        </section>
      ) : null}

      <div className="flex items-start justify-between gap-4">
        <div>
          <div className="flex items-center gap-2">
            <h2 className="text-lg font-black text-slate-900">处方摘要</h2>
            <span className="text-xs text-slate-400">
              {patient.name} · {patient.gender}/{patient.age}岁 · {patient.bedId}
            </span>
          </div>
          <div className="mt-1.5 flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-slate-500">
            <span>来源方案: {treatmentPlan ? `${treatmentPlan.dialysisMode?.mode || '-'}（干体重: ${treatmentPlan.dryWeight}kg）` : '未加载'}</span>
            <span>处方总数: {prescriptions.length}</span>
            <span>来源: {currentPrescription ? '处方' : '治疗方案回填'}</span>
          </div>
        </div>
        <div className="flex shrink-0 items-center gap-3">
          <div className="grid grid-cols-2 divide-x divide-slate-200 rounded-lg border border-slate-200 bg-white px-4 py-2.5 text-center shadow-sm">
            <div className="px-3">
              <div className="text-[11px] font-semibold text-slate-400">干体重</div>
              <div className="mt-0.5 text-base font-black text-blue-700">{metrics.dryWeight || '--'} <span className="text-[10px]">KG</span></div>
            </div>
            <div className="px-3">
              <div className="text-[11px] font-semibold text-slate-400">治疗方式</div>
              <div className="mt-0.5 text-base font-black text-slate-900">{metrics.dialysisMethod || '--'}</div>
            </div>
          </div>
          <button
            type="button"
            onClick={() => void handleExtractToday()}
            disabled={loading}
            className="inline-flex h-9 items-center gap-2 rounded-lg border border-blue-200 bg-blue-50 px-4 text-xs font-bold text-blue-700 transition hover:bg-blue-100 disabled:opacity-50"
          >
            <RotateCw size={14} />
            提取今日处方
          </button>
        </div>
      </div>

      {loading ? (
        <div className="rounded-lg border border-slate-200 bg-white p-10 text-center text-slate-500">
          正在加载治疗方案和处方...
        </div>
      ) : (
        <>
          <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-5">
            <MetricCard label="透析方法" value={metrics.dialysisMethod} primary />
            <MetricCard label="本次体重增加量" value={toText(metrics.weightChange)} unit="kg" />
            <MetricCard label="目标超滤量" value={metrics.targetUf} unit="L" />
            <MetricCard label="透前血压" value={metrics.preBp || '--'} unit="mmHg" />
            <MetricCard label="透析时间" value={metrics.duration} unit="H" />
          </div>

          <div className="grid gap-4 xl:grid-cols-2">
            <SectionCard title="核心处方参数" icon={<Activity size={16} className="text-blue-600" />}>
              <div className="space-y-3 p-4">
                <ValueGroup title="治疗与容量">
                  <div className="grid grid-cols-2 gap-x-6 gap-y-2.5 sm:grid-cols-3">
                    <EditableField label="透析方法" value={form.dialysisMode} disabled={!editing} compact onChange={(v) => updateField('dialysisMode', v)} />
                    <EditableField label="干体重" value={form.dryWeight} unit="kg" disabled={!editing} compact onChange={(v) => updateField('dryWeight', v)} />
                    <EditableField label="本次增量" value={toText(metrics.weightChange)} unit="kg" disabled compact onChange={() => undefined} />
                    <EditableField label="目标超滤" value={form.extraWeight} unit="L" disabled={!editing} compact onChange={(v) => updateField('extraWeight', v)} />
                    <EditableField label="透析时间" value={form.duration} unit="H" disabled={!editing} compact onChange={(v) => updateField('duration', v)} />
                    <EditableField label="透前体重" value={toText(metrics.preWeight)} unit="kg" disabled compact onChange={() => undefined} />
                  </div>
                </ValueGroup>
                <ValueGroup title="血流与通路">
                  <div className="grid grid-cols-2 gap-x-6 gap-y-2.5 sm:grid-cols-3">
                    <EditableField label="血管通路" value={metrics.vascularAccess} disabled compact onChange={() => undefined} />
                    <EditableField label="标准血流" value={form.bloodFlow} unit="ml/min" disabled={!editing} compact onChange={(v) => updateField('bloodFlow', v)} />
                    <EditableField label="置换量" value={form.substituteVolume} unit="L" disabled={!editing} compact onChange={(v) => updateField('substituteVolume', v)} />
                  </div>
                </ValueGroup>
              </div>
            </SectionCard>

            <SectionCard
              title="抗凝方案"
              icon={<Droplets size={16} className={hasAnticoagulant ? 'text-orange-500' : 'text-slate-400'} />}
            >
              <div className="p-4">
                {hasAnticoagulant ? (
                  <div className="space-y-3">
                    <div className="rounded-lg bg-orange-50 px-3 py-2 text-sm font-bold text-orange-800">
                      药剂名称：{form.initialDrug || '--'}
                    </div>
                    <div className="grid grid-cols-2 gap-x-6 gap-y-2.5 sm:grid-cols-3">
                      <EditableField label="首剂量" value={form.initialDose} disabled={!editing} compact onChange={(v) => updateField('initialDose', v)} />
                      <EditableField label="维持剂" value={form.maintenanceDrug} disabled={!editing} compact onChange={(v) => updateField('maintenanceDrug', v)} />
                      <EditableField label="维持量" value={form.maintenanceDose} disabled={!editing} compact onChange={(v) => updateField('maintenanceDose', v)} />
                      <EditableField label="注入速率" value={form.infusionRate} disabled={!editing} compact onChange={(v) => updateField('infusionRate', v)} />
                      <EditableField label="注入时间" value={form.infusionTime} disabled={!editing} compact onChange={(v) => updateField('infusionTime', v)} />
                    </div>
                  </div>
                ) : (
                  <div className="py-6 text-center text-sm text-slate-400">未配置抗凝方案</div>
                )}
              </div>
            </SectionCard>
          </div>

          <SectionCard title="透析液及机器设定" icon={<Settings size={16} className="text-blue-600" />}>
            <div className="grid grid-cols-1 gap-4 p-4 md:grid-cols-3">
              <ValueGroup title="液体">
                <div className="grid grid-cols-1 gap-2.5">
                  <EditableField label="透析液" value={form.dialysateType} disabled={!editing} compact onChange={(v) => updateField('dialysateType', v)} />
                  <EditableField label="液温" value={form.temp} unit="℃" disabled={!editing} compact onChange={(v) => updateField('temp', v)} />
                  <EditableField label="液量" value={form.volume} unit="L" disabled={!editing} compact onChange={(v) => updateField('volume', v)} />
                </div>
              </ValueGroup>
              <ValueGroup title="浓度">
                <div className="grid grid-cols-1 gap-2.5">
                  <EditableField label="Na 浓度" value={form.na} unit="mmol/L" disabled={!editing} compact onChange={(v) => updateField('na', v)} />
                  <EditableField label="Ca 浓度" value={form.ca} unit="mmol/L" disabled={!editing} compact onChange={(v) => updateField('ca', v)} />
                  <EditableField label="K 浓度" value={form.k} unit="mmol/L" disabled={!editing} compact onChange={(v) => updateField('k', v)} />
                  <EditableField label="HCO₃ 浓度" value={form.hco3} unit="mmol/L" disabled={!editing} compact onChange={(v) => updateField('hco3', v)} />
                  <EditableField label="葡萄糖" value={form.glucose} unit="mmol/L" disabled={!editing} compact onChange={(v) => updateField('glucose', v)} />
                </div>
              </ValueGroup>
              <ValueGroup title="机器参数">
                <div className="grid grid-cols-1 gap-2.5">
                  <EditableField label="电导度" value={form.conductivity} unit="mS/cm" disabled={!editing} compact onChange={(v) => updateField('conductivity', v)} />
                  <EditableField label="血流" value={form.bloodFlow} unit="ml/min" disabled={!editing} compact onChange={(v) => updateField('bloodFlow', v)} />
                  <EditableField label="透析液流速" value={form.flowRate} unit="ml/min" disabled={!editing} compact onChange={(v) => updateField('flowRate', v)} />
                  <EditableField label="透析液分组" value={form.dialysateGroup} disabled={!editing} compact onChange={(v) => updateField('dialysateGroup', v)} />
                  <EditableField label="置换量" value={form.substituteVolume} unit="L" disabled={!editing} compact onChange={(v) => updateField('substituteVolume', v)} />
                  <EditableField label="备注" value={form.notes} disabled={!editing} compact onChange={(v) => updateField('notes', v)} />
                </div>
              </ValueGroup>
            </div>
          </SectionCard>

          <div className="grid gap-4 xl:grid-cols-2">
            <SectionCard
              title="透析材料清单"
              icon={<Package size={16} className="text-blue-600" />}
              right={<span className="text-[11px] text-slate-400">来源: {currentPrescription ? '处方' : '治疗方案'}</span>}
            >
              <div className="overflow-x-auto">
                <table className="w-full min-w-[700px] text-left">
                  <thead className="sticky top-0 bg-slate-50 text-xs text-slate-500">
                    <tr>
                      <th className="px-4 py-2.5 font-semibold">#</th>
                      <th className="px-4 py-2.5 font-semibold">材料名称</th>
                      <th className="px-4 py-2.5 font-semibold">分类</th>
                      <th className="px-4 py-2.5 font-semibold">数量</th>
                      <th className="px-4 py-2.5 font-semibold">编码</th>
                      <th className="px-4 py-2.5 font-semibold">品牌</th>
                      <th className="px-4 py-2.5 font-semibold">规格</th>
                      <th className="px-4 py-2.5 font-semibold">备注</th>
                    </tr>
                  </thead>
                  <tbody>
                    {displayMaterials.length > 0 ? (
                      displayMaterials.map((item, index) => (
                        <tr key={`${item.id}-${index}`} className="border-t border-slate-50 text-xs">
                          <td className="px-4 py-2 text-slate-900">{index + 1}</td>
                          <td className="px-4 py-2 font-semibold text-slate-800">{item.name}</td>
                          <td className="px-4 py-2 text-slate-600">{item.category || '--'}</td>
                          <td className="px-4 py-2 text-slate-600">{item.count}</td>
                          <td className="px-4 py-2 text-slate-600">{item.code || '--'}</td>
                          <td className="px-4 py-2 text-slate-600">{item.brand || '--'}</td>
                          <td className="px-4 py-2 text-slate-600">{item.spec || '--'}</td>
                          <td className="px-4 py-2 text-slate-600">{item.note || '--'}</td>
                        </tr>
                      ))
                    ) : (
                      <tr>
                        <td colSpan={8} className="px-4 py-8 text-center text-xs text-slate-400">
                          当前处方和治疗方案均无耗材明细，可先检查治疗方案配置。
                        </td>
                      </tr>
                    )}
                  </tbody>
                </table>
              </div>
            </SectionCard>

            <SectionCard title="处方动态调整记录" icon={<Clock size={16} className="text-slate-400" />}>
              <div className="overflow-x-auto">
                <table className="w-full min-w-[400px] text-left">
                  <thead className="sticky top-0 bg-slate-50 text-xs text-slate-500">
                    <tr>
                      <th className="px-4 py-2.5 font-semibold">#</th>
                      <th className="px-4 py-2.5 font-semibold">调整内容</th>
                      <th className="px-4 py-2.5 font-semibold">调整人</th>
                      <th className="px-4 py-2.5 font-semibold">调整时间</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr>
                      <td colSpan={4} className="px-4 py-12 text-center">
                        <div className="text-xs font-medium text-slate-400">暂无动态调整记录</div>
                        <div className="mt-1 text-[11px] text-slate-300">修改处方后，建议记录调整内容、调整人和时间。</div>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </SectionCard>
          </div>

          <div className="sticky bottom-0 z-10 flex items-center justify-between rounded-xl border border-slate-200 bg-white/95 px-4 py-3.5 shadow-lg backdrop-blur">
            <div className="flex items-center gap-5 text-xs text-slate-400">
              <span className="font-semibold">UPDATE: {formattedUpdatedAt}</span>
              <span className="inline-flex items-center gap-1.5">
                STATUS:
                <span className={`rounded-full px-2.5 py-0.5 text-[11px] font-bold ${
                  currentPrescription?.status === '待执行' ? 'bg-blue-50 text-blue-700' :
                  currentPrescription?.status === '已执行' ? 'bg-emerald-50 text-emerald-700' :
                  'bg-slate-100 text-slate-600'
                }`}>
                  {currentPrescription?.status || '--'}
                </span>
              </span>
            </div>
            <div className="flex items-center gap-3">
              {!currentPrescription ? (
                <button
                  type="button"
                  onClick={() => void handleExtractToday()}
                  disabled={loading}
                  className="inline-flex items-center gap-2 rounded-lg bg-blue-600 px-5 py-2 text-sm font-bold text-white transition hover:bg-blue-700 disabled:opacity-50"
                >
                  <RotateCw size={14} />
                  提取今日处方
                </button>
              ) : editing ? (
                <>
                  <button
                    type="button"
                    onClick={() => { setEditing(false); if (currentPrescription) { setForm(mapPrescriptionToForm(currentPrescription)) } }}
                    className="rounded-lg border border-slate-200 bg-white px-4 py-2 text-sm font-semibold text-slate-600 transition hover:bg-slate-50"
                  >
                    取消
                  </button>
                  <button
                    type="button"
                    onClick={() => void handleSave()}
                    disabled={saving}
                    className="rounded-lg bg-blue-600 px-5 py-2 text-sm font-bold text-white transition hover:bg-blue-700 disabled:opacity-50"
                  >
                    {saving ? '保存中...' : '保存处方'}
                  </button>
                </>
              ) : (
                <button
                  type="button"
                  onClick={() => setEditing(true)}
                  className="inline-flex items-center gap-2 rounded-lg border border-slate-300 bg-white px-4 py-2 text-sm font-semibold text-slate-700 transition hover:bg-slate-50"
                >
                  <Settings size={14} />
                  修改处方
                </button>
              )}
              <button
                type="button"
                onClick={() => void handleExecute()}
                disabled={!canExecute || executing}
                title={canExecute ? undefined : '仅待执行处方可核对并执行'}
                className="inline-flex items-center gap-2 rounded-lg bg-blue-600 px-5 py-2 text-sm font-bold text-white transition hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-40"
              >
                <Box size={15} />
                {executing ? '执行中...' : '核对并执行'}
              </button>
            </div>
          </div>
        </>
      )}
    </div>
  )
}
