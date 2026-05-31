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
    <div className={`rounded-lg border px-4 py-4 shadow-sm ${primary ? 'border-blue-600 bg-blue-600 text-white' : 'border-slate-200 bg-white'}`}>
      <div className={`text-xs font-semibold ${primary ? 'text-blue-100' : 'text-slate-400'}`}>{label}</div>
      <div className={`mt-2 text-2xl font-black ${primary ? 'text-white' : 'text-slate-900'}`}>
        {value || '--'}
        {unit ? <span className={`ml-1 text-xs ${primary ? 'text-blue-100' : 'text-slate-400'}`}>{unit}</span> : null}
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
}: {
  label: string
  value: string
  unit?: string
  disabled: boolean
  onChange: (value: string) => void
}) {
  return (
    <label className="block min-w-0">
      <div className="mb-2 text-xs font-semibold text-slate-400">{label}</div>
      <div className="relative">
        <input
          value={disabled ? value || '--' : value}
          onChange={(e) => onChange(e.target.value)}
          disabled={disabled}
          className={`h-10 w-full rounded-md border px-3 text-sm font-bold outline-none ${unit ? 'pr-14' : ''} ${
            disabled ? 'border-transparent bg-transparent text-slate-900' : 'border-blue-300 bg-white text-slate-900'
          }`}
        />
        {unit ? <span className="absolute right-3 top-1/2 -translate-y-1/2 text-xs font-medium text-slate-400">{unit}</span> : null}
      </div>
    </label>
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
    <section className="overflow-hidden rounded-lg border border-slate-200 bg-white shadow-sm">
      <div className="flex items-center justify-between border-b border-slate-200 px-4 py-3">
        <div className="flex items-center gap-2 text-sm font-black text-slate-800">
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

  // 加载治疗方案和处方
  useEffect(() => {
    const loadData = async () => {
      setLoading(true)
      try {
        // 并行加载治疗方案和处方列表
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

        // 优先从治疗方案加载表单默认值
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

  // 当治疗方案加载后，同处方数据合并（处方为准，方案回填）
  useEffect(() => {
    if (treatmentPlan && !currentPrescription) {
      setForm(mapPlanToForm(treatmentPlan))
    }
  }, [treatmentPlan, currentPrescription])

  // 计算派生指标
  const metrics = useMemo(() => {
    // 透前净重 = 透前体重 - 额外体重
    const preWeight = treatment?.beforeSigns?.weight || 0
    const extra = parseOptionalNumber(form.extraWeight) ?? 0
    const preNetWeight = preWeight - extra

    // 干体重从治疗方案读取
    const dryWeight = parseOptionalNumber(form.dryWeight) ?? treatmentPlan?.dryWeight ?? patient.dryWeight

    // 较前增量 = 透前净重 - 上次透后体重（暂取0，后续接入）
    const lastPostWeight = 0
    const weightChange = preNetWeight - lastPostWeight

    // 透前血压从透前评估读取
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
      // 合并治疗方案数据到处方表单
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
    <div className="space-y-5 pb-8">
      {treatmentLoading ? (
        <section className="rounded-lg border border-blue-100 bg-blue-50 px-6 py-4 text-sm font-semibold text-blue-700">
          正在加载新患者治疗上下文，处方概览已清空旧治疗数据。
        </section>
      ) : null}

      <div className="flex items-center justify-between gap-4">
        <div>
          <div className="flex flex-wrap items-center gap-2">
            <h2 className="text-xl font-black text-slate-900">{patient.name}</h2>
            <span className="rounded-md bg-blue-600 px-2.5 py-1 text-sm font-bold text-white">{patient.status || '未排床'}</span>
            <span className="text-xs font-semibold text-slate-500">ID: {patient.id}</span>
            <span className="text-xs font-semibold text-slate-500">{patient.gender} / {patient.age}岁</span>
            <span className="text-xs font-semibold text-slate-500">费用: {patient.costType || '--'}</span>
            <span className="text-xs font-semibold text-slate-500">透析龄: {patient.dialysisAge || '--'}</span>
          </div>
          <div className="mt-2 text-xs text-slate-400">
            来源方案: {treatmentPlan ? `${treatmentPlan.dialysisMode?.mode || '-'} (干体重: ${treatmentPlan.dryWeight}kg)` : '未加载'}
            {' | '}处方总数：{prescriptions.length}
          </div>
        </div>
        <div className="flex shrink-0 items-center gap-3">
          <div className="grid grid-cols-2 divide-x divide-slate-200 rounded-lg border border-slate-200 bg-white px-5 py-3 text-center shadow-sm">
            <div className="px-4">
              <div className="text-xs font-semibold text-slate-400">干体重</div>
              <div className="mt-1 font-black text-blue-700">{metrics.dryWeight || '--'} <span className="text-xs">KG</span></div>
            </div>
            <div className="px-4">
              <div className="text-xs font-semibold text-slate-400">治疗方案</div>
              <div className="mt-1 font-black text-slate-900">{metrics.dialysisMethod || '--'}</div>
            </div>
          </div>
          <button
            type="button"
            onClick={() => void handleExtractToday()}
            disabled={loading}
            className="inline-flex h-11 items-center gap-2 rounded-lg border border-slate-200 bg-white px-4 text-sm font-semibold text-slate-700 disabled:opacity-60"
          >
            <RotateCw size={15} />
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
          <div className="grid grid-cols-1 gap-4 md:grid-cols-5">
            <MetricCard label="透析方法" value={metrics.dialysisMethod} primary />
            <MetricCard label="本次体重增加量" value={toText(metrics.weightChange)} unit="KG" />
            <MetricCard label="目标超滤量" value={metrics.targetUf} unit="L" />
            <MetricCard label="透前血压" value={metrics.preBp} unit="MMHG" />
            <MetricCard label="透析时间" value={metrics.duration} unit="H" />
          </div>

          <div className="grid gap-4 xl:grid-cols-2">
            <SectionCard title="体重详情与基础标准" icon={<Activity size={16} className="text-blue-600" />}>
              <div className="grid grid-cols-2 gap-x-10 gap-y-5 p-5 md:grid-cols-3">
                <EditableField label="透前体重" value={toText(metrics.preWeight)} unit="kg" disabled onChange={() => undefined} />
                <EditableField label="透前净重" value={toText(metrics.preNetWeight)} unit="kg" disabled onChange={() => undefined} />
                <EditableField label="上次透后体重" value={toText(metrics.lastPostWeight)} unit="kg" disabled onChange={() => undefined} />
                <EditableField label="较前增量" value={toText(metrics.weightChange)} unit="kg" disabled onChange={() => undefined} />
                <EditableField label="干体重" value={form.dryWeight} unit="kg" disabled={!editing} onChange={(value) => updateField('dryWeight', value)} />
                <EditableField label="标准血流" value={form.bloodFlow} unit="ml/min" disabled={!editing} onChange={(value) => updateField('bloodFlow', value)} />
              </div>
            </SectionCard>

            <SectionCard title="抗凝方案设定" icon={<Droplets size={16} className="text-blue-600" />}>
              <div className="grid grid-cols-2 gap-x-10 gap-y-5 p-5 md:grid-cols-3">
                <EditableField label="药剂名称" value={form.initialDrug} disabled={!editing} onChange={(value) => updateField('initialDrug', value)} />
                <EditableField label="首剂量" value={form.initialDose} disabled={!editing} onChange={(value) => updateField('initialDose', value)} />
                <EditableField label="维持剂" value={form.maintenanceDrug} disabled={!editing} onChange={(value) => updateField('maintenanceDrug', value)} />
                <EditableField label="维持量" value={form.maintenanceDose} disabled={!editing} onChange={(value) => updateField('maintenanceDose', value)} />
                <EditableField label="注入速率" value={form.infusionRate} disabled={!editing} onChange={(value) => updateField('infusionRate', value)} />
                <EditableField label="注入时间" value={form.infusionTime} disabled={!editing} onChange={(value) => updateField('infusionTime', value)} />
              </div>
            </SectionCard>
          </div>

          <SectionCard title="透析液及通路设定" icon={<Settings size={16} className="text-blue-600" />}>
            <div className="grid grid-cols-2 gap-x-10 gap-y-5 p-5 md:grid-cols-4 xl:grid-cols-8">
              <EditableField label="血管通路" value={metrics.vascularAccess} disabled onChange={() => undefined} />
              <EditableField label="透析液" value={form.dialysateType} disabled={!editing} onChange={(value) => updateField('dialysateType', value)} />
              <EditableField label="透析液流速" value={form.flowRate} unit="ml/min" disabled={!editing} onChange={(value) => updateField('flowRate', value)} />
              <EditableField label="Na浓度" value={form.na} unit="mmol/L" disabled={!editing} onChange={(value) => updateField('na', value)} />
              <EditableField label="Ca浓度" value={form.ca} unit="mmol/L" disabled={!editing} onChange={(value) => updateField('ca', value)} />
              <EditableField label="K浓度" value={form.k} unit="mmol/L" disabled={!editing} onChange={(value) => updateField('k', value)} />
              <EditableField label="HCO3浓度" value={form.hco3} unit="mmol/L" disabled={!editing} onChange={(value) => updateField('hco3', value)} />
              <EditableField label="葡萄糖" value={form.glucose} unit="mmol/L" disabled={!editing} onChange={(value) => updateField('glucose', value)} />
              <EditableField label="电导度" value={form.conductivity} unit="mS/cm" disabled={!editing} onChange={(value) => updateField('conductivity', value)} />
              <EditableField label="液温" value={form.temp} unit="℃" disabled={!editing} onChange={(value) => updateField('temp', value)} />
              <EditableField label="透析液量" value={form.volume} unit="L" disabled={!editing} onChange={(value) => updateField('volume', value)} />
              <EditableField label="透析液分组" value={form.dialysateGroup} disabled={!editing} onChange={(value) => updateField('dialysateGroup', value)} />
              <EditableField label="置换量" value={form.substituteVolume} unit="L" disabled={!editing} onChange={(value) => updateField('substituteVolume', value)} />
              <EditableField label="备注" value={form.notes} disabled={!editing} onChange={(value) => updateField('notes', value)} />
            </div>
          </SectionCard>

          <SectionCard
            title="透析材料清单"
            icon={<Package size={16} className="text-blue-600" />}
            right={<div className="text-xs font-semibold text-slate-400">
              来源: {currentPrescription ? '处方' : '治疗方案'}
            </div>}
          >
            <div className="overflow-x-auto">
              <table className="w-full min-w-[900px] text-left">
                <thead className="bg-slate-50 text-xs text-slate-500">
                  <tr>
                    <th className="px-6 py-3">#</th>
                    <th className="px-6 py-3">材料名称</th>
                    <th className="px-6 py-3">分类</th>
                    <th className="px-6 py-3">数量</th>
                    <th className="px-6 py-3">编码</th>
                    <th className="px-6 py-3">品牌</th>
                    <th className="px-6 py-3">规格</th>
                    <th className="px-6 py-3">备注</th>
                  </tr>
                </thead>
                <tbody>
                  {displayMaterials.length > 0 ? (
                    displayMaterials.map((item, index) => (
                      <tr key={`${item.id}-${index}`} className="border-t border-slate-100 text-sm">
                        <td className="px-6 py-3 text-slate-900">{index + 1}</td>
                        <td className="px-6 py-3 font-semibold text-slate-800">{item.name}</td>
                        <td className="px-6 py-3 text-slate-600">{item.category || '--'}</td>
                        <td className="px-6 py-3 text-slate-600">{item.count}</td>
                        <td className="px-6 py-3 text-slate-600">{item.code || '--'}</td>
                        <td className="px-6 py-3 text-slate-600">{item.brand || '--'}</td>
                        <td className="px-6 py-3 text-slate-600">{item.spec || '--'}</td>
                        <td className="px-6 py-3 text-slate-600">{item.note || '--'}</td>
                      </tr>
                    ))
                  ) : (
                    <tr>
                      <td colSpan={8} className="px-6 py-8 text-center text-sm text-slate-400">
                        当前处方和治疗方案均无耗材明细
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          </SectionCard>

          <SectionCard title="当日处方动态调整记录" icon={<Clock size={16} className="text-amber-500" />}>
            <div className="overflow-x-auto">
              <table className="w-full min-w-[760px] text-left">
                <thead className="bg-slate-50 text-xs text-slate-500">
                  <tr>
                    <th className="px-6 py-3">#</th>
                    <th className="px-6 py-3">调整内容</th>
                    <th className="px-6 py-3">调整人</th>
                    <th className="px-6 py-3">调整时间</th>
                    <th className="px-6 py-3">操作</th>
                  </tr>
                </thead>
                <tbody>
                  <tr>
                    <td colSpan={5} className="px-6 py-8 text-center text-sm text-slate-300">暂无数据</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </SectionCard>

          <div className="sticky bottom-0 z-10 flex items-center justify-between border-t border-slate-200 bg-white/95 px-4 py-4 shadow-lg backdrop-blur">
            <div className="flex items-center gap-5 text-xs font-bold text-slate-400">
              <span>UPDATE: {formattedUpdatedAt}</span>
              <span className="inline-flex items-center gap-2">STATUS: <span className="rounded-full bg-emerald-50 px-3 py-1 text-emerald-700">{currentPrescription?.status || '--'}</span></span>
            </div>
            <div className="flex items-center gap-3">
              {editing ? (
                <>
                  <button type="button" onClick={() => { setEditing(false); if (currentPrescription) { setForm(mapPrescriptionToForm(currentPrescription)) } }} className="rounded-lg border border-slate-200 bg-white px-5 py-2 text-sm font-semibold text-slate-700">取消</button>
                  <button type="button" onClick={() => void handleSave()} disabled={saving} className="rounded-lg bg-blue-600 px-5 py-2 text-sm font-bold text-white disabled:opacity-60">{saving ? '保存中...' : '保存处方'}</button>
                </>
              ) : (
                <button type="button" onClick={() => setEditing(true)} className="inline-flex items-center gap-2 rounded-lg border border-slate-300 bg-white px-5 py-2 text-sm font-semibold text-slate-700"><Settings size={14} />修改处方</button>
              )}
              <button type="button" onClick={() => void handleExecute()} disabled={!canExecute || executing} title={canExecute ? undefined : '仅待执行处方可核对并执行'} className="inline-flex items-center gap-2 rounded-lg bg-blue-600 px-6 py-2 text-sm font-bold text-white disabled:cursor-not-allowed disabled:opacity-50">
                <Box size={15} />{executing ? '执行中...' : '核对并执行'}
              </button>
            </div>
          </div>
        </>
      )}
    </div>
  )
}
