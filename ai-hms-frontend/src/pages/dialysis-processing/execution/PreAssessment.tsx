import { message } from 'antd'
import { Activity, AlertTriangle, Scale, Stethoscope, X, User } from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'
import type { RestTreatment } from '@/services'
import { restApi, getErrorMessage } from '@/services/restClient'
import type { TreatmentBeforeSignsRequest, VascularAccessApi } from '@/services/restClient'
import type { Patient, PreAssessmentFormValue } from '../types'
import { useAuth } from '@/contexts/AuthContext'

const FISTULA_DEFAULTS = ['杂音强', '震颤强', '搏动强']
const BP_SITES = ['右上肢', '左上肢', '右下肢', '左下肢', '其他']
const CONSCIOUSNESS_OPTS = ['清醒', '嗜睡', '昏睡', '昏迷']
const NURSING_LEVEL_OPTS = ['危重', '重症', '其他']

function AssessmentSection({ title, icon, children }: { title: string; icon: React.ReactNode; children: React.ReactNode }) {
  return (
    <section className="rounded-lg border border-slate-200 bg-white p-4 shadow-sm">
      <div className="mb-4 flex items-center gap-2 border-b border-slate-100 pb-3">
        {icon}
        <h3 className="text-base font-bold text-slate-800">{title}</h3>
      </div>
      {children}
    </section>
  )
}

function FieldInput({
  label,
  value,
  suffix,
  onChange,
  disabled,
  placeholder,
}: {
  label: string
  value: string
  suffix?: string
  onChange?: (value: string) => void
  disabled?: boolean
  placeholder?: string
}) {
  return (
    <label className="block min-w-0">
      <span className="mb-1.5 block text-sm font-medium text-slate-600">{label}</span>
      <span className="flex h-10 items-center rounded-md border border-slate-200 bg-white px-3 focus-within:border-blue-400 focus-within:ring-2 focus-within:ring-blue-100">
        <input
          value={value}
          onChange={(e) => onChange?.(e.target.value)}
          disabled={disabled}
          placeholder={placeholder}
          className="min-w-0 flex-1 bg-transparent text-sm font-medium text-slate-800 outline-none placeholder:text-slate-300 disabled:text-slate-500"
        />
        {suffix ? <span className="ml-2 min-w-10 text-right text-xs font-medium text-slate-400">{suffix}</span> : null}
      </span>
    </label>
  )
}

function BpInput({
  sbp,
  dbp,
  onSbpChange,
  onDbpChange,
}: {
  sbp: string
  dbp: string
  onSbpChange?: (v: string) => void
  onDbpChange?: (v: string) => void
}) {
  return (
    <label className="block min-w-0">
      <span className="mb-1.5 block text-sm font-medium text-slate-600">透前血压</span>
      <span className="flex h-10 items-center rounded-md border border-slate-200 bg-white px-3 focus-within:border-blue-400 focus-within:ring-2 focus-within:ring-blue-100">
        <input
          value={sbp}
          onChange={(e) => onSbpChange?.(e.target.value)}
          placeholder="收缩压"
          className="min-w-0 flex-1 bg-transparent text-sm font-medium text-slate-800 outline-none placeholder:text-slate-300"
        />
        <span className="mx-1.5 text-slate-400">/</span>
        <input
          value={dbp}
          onChange={(e) => onDbpChange?.(e.target.value)}
          placeholder="舒张压"
          className="min-w-0 flex-1 bg-transparent text-sm font-medium text-slate-800 outline-none placeholder:text-slate-300"
        />
        <span className="ml-2 min-w-10 text-right text-xs font-medium text-slate-400">mmHg</span>
      </span>
    </label>
  )
}

function ChipSelect({
  label,
  options,
  value,
  onChange,
}: {
  label: string
  options: string[]
  value: string
  onChange: (v: string) => void
}) {
  return (
    <label className="block min-w-0">
      <span className="mb-1.5 block text-sm font-medium text-slate-600">{label}</span>
      <div className="flex flex-wrap gap-1.5">
        {options.map((opt) => (
          <button
            key={opt}
            type="button"
            onClick={() => onChange(opt)}
            className={`rounded-md border px-3 py-1.5 text-xs font-semibold transition ${
              value === opt
                ? 'border-blue-400 bg-blue-50 text-blue-700'
                : 'border-slate-200 bg-white text-slate-600 hover:border-slate-300'
            }`}
          >
            {opt}
          </button>
        ))}
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
  fallRisk: '',
  painScore: '',
  notes: '',
  symptoms: [],
  fistulaStatus: [...FISTULA_DEFAULTS],
  skinRecord: '',
  refuseMeasure: false,
  bedridden: false,
}

function getSymptomItemValue(items: Array<{ code: string; value: string }> | undefined, code: string) {
  return items?.find((item) => item.code === code)?.value ?? ''
}

function toText(value?: number | string | null) {
  if (value === undefined || value === null) return ''
  return String(value)
}

function mapTreatmentToForm(treatment: RestTreatment | null): PreAssessmentFormValue {
  if (!treatment) return { ...EMPTY_FORM }
  const before = treatment.beforeSigns
  const symptomItems = treatment.beforeSymptomItems

  const fistulaRaw = getSymptomItemValue(symptomItems, 'fistula_tags')
  const fistulaParsed = fistulaRaw
    ? fistulaRaw.split(/[，,；;、]/).map((s) => s.trim()).filter(Boolean)
    : [...FISTULA_DEFAULTS]

  const symptomsRaw = before?.symptoms || getSymptomItemValue(symptomItems, 'symptoms')
  const symptomsParsed = symptomsRaw
    ? symptomsRaw.split(/[，,；;、]/).map((s) => s.trim()).filter(Boolean)
    : []

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
    fallRisk: getSymptomItemValue(symptomItems, 'fall_risk'),
    painScore: getSymptomItemValue(symptomItems, 'pain_score'),
    notes: '',
    symptoms: symptomsParsed,
    fistulaStatus: fistulaParsed,
    skinRecord: getSymptomItemValue(symptomItems, 'skin_record'),
    refuseMeasure: getSymptomItemValue(symptomItems, 'pre_weight_refused') === '是',
    bedridden: getSymptomItemValue(symptomItems, 'pre_bedridden') === '是',
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
  treatmentLoading?: boolean
  onSave: (payload: TreatmentBeforeSignsRequest) => Promise<void>
}

export default function PreAssessment({
  patient,
  treatment,
  saving = false,
  treatmentLoading = false,
  onSave,
}: Props) {
  const { user: currentUser } = useAuth()
  const [form, setForm] = useState<PreAssessmentFormValue>({ ...EMPTY_FORM })
  const [newSymptom, setNewSymptom] = useState('')
  const [vascularAccesses, setVascularAccesses] = useState<VascularAccessApi[]>([])
  const [skinRecordHistory, setSkinRecordHistory] = useState<string[]>([])

  useEffect(() => {
    if (!patient.id) return
    let cancelled = false
    const load = async () => {
      try {
        const data = await restApi.getVascularAccesses(patient.id)
        if (!cancelled) {
          setVascularAccesses(data.filter((v) => !v.isDisabled))
        }
      } catch (error) {
        if (!cancelled) {
          console.error('加载血管通路数据失败:', error)
        }
      }
    }
    load()
    return () => { cancelled = true }
  }, [patient.id])

  // 从血管通路数据中提取 A端和 V端位点选项
  const aSiteOptions = useMemo(() => {
    const sites = new Set<string>()
    vascularAccesses.forEach((v) => {
      if (v.aPuncturePosition && Array.isArray(v.aPuncturePosition)) {
        v.aPuncturePosition.forEach((s) => sites.add(s))
      }
    })
    return Array.from(sites)
  }, [vascularAccesses])

  const vSiteOptions = useMemo(() => {
    const sites = new Set<string>()
    vascularAccesses.forEach((v) => {
      if (v.vPuncturePosition && Array.isArray(v.vPuncturePosition)) {
        v.vPuncturePosition.forEach((s) => sites.add(s))
      }
    })
    return Array.from(sites)
  }, [vascularAccesses])

  useEffect(() => {
    queueMicrotask(() => {
      setForm(mapTreatmentToForm(treatment))
      setNewSymptom('')
    })
  }, [treatment])

  const weightGain = useMemo(() => {
    const w = Number(form.weight)
    if (!Number.isFinite(w) || !patient.dryWeight) return ''
    return (w - patient.dryWeight).toFixed(1)
  }, [form.weight, patient.dryWeight])

  const weightWarning = useMemo(() => {
    const g = Number(weightGain)
    return Number.isFinite(g) && g > 4 ? '体重增长超过正常范围' : ''
  }, [weightGain])

  const calculatedTargetUf = useMemo(() => {
    const gain = Number(weightGain)
    if (!Number.isFinite(gain)) return ''
    return (gain + 0.2).toFixed(1)
  }, [weightGain])

  const [targetUfManuallyEdited, setTargetUfManuallyEdited] = useState(false)

  const handleTargetUfChange = (value: string) => {
    setTargetUfManuallyEdited(true)
    updateField('targetUf', value)
  }

  const displayTargetUf = targetUfManuallyEdited ? form.targetUf : (calculatedTargetUf || form.targetUf)

  const updateField = (key: keyof PreAssessmentFormValue, value: string | string[] | boolean) => {
    setForm((current) => ({ ...current, [key]: value }))
  }

  const handleAddSymptom = () => {
    const s = newSymptom.trim()
    if (!s) return
    updateField('symptoms', [...form.symptoms, s])
    setNewSymptom('')
  }

  const handleRemoveSymptom = (item: string) => {
    updateField('symptoms', form.symptoms.filter((s) => s !== item))
  }

  const handleSave = async () => {
    try {
      const symptomItems: Array<{ code: string; value: string }> = [
        { code: 'uf_volume', value: form.targetUf.trim() },
        { code: 'a_site', value: form.aSite.trim() },
        { code: 'v_site', value: form.vSite.trim() },
        { code: 'consciousness', value: form.consciousness.trim() },
        { code: 'nurse_level', value: form.nurseLevel.trim() },
        { code: 'fall_risk', value: form.fallRisk.trim() },
        { code: 'pain_score', value: form.painScore.trim() },
        { code: 'fistula_tags', value: form.fistulaStatus.join('，') },
        { code: 'symptoms', value: form.symptoms.join('，') },
        { code: 'skin_record', value: form.skinRecord.trim() },
        { code: 'pre_weight_refused', value: form.refuseMeasure ? '是' : '否' },
        { code: 'pre_bedridden', value: form.bedridden ? '是' : '否' },
      ].filter((item) => item.value)

      // 保存皮肤记录到历史
      if (form.skinRecord.trim() && !skinRecordHistory.includes(form.skinRecord.trim())) {
        setSkinRecordHistory((prev) => [form.skinRecord.trim(), ...prev].slice(0, 20))
      }

      await onSave({
        weight: form.refuseMeasure ? undefined : parseOptionalNumber(form.weight),
        extraWeight: parseOptionalNumber(form.extraWeight),
        sbp: parseOptionalNumber(form.sbp),
        dbp: parseOptionalNumber(form.dbp),
        heartRate: parseOptionalNumber(form.heartRate),
        respiration: parseOptionalNumber(form.respiration),
        temperature: parseOptionalNumber(form.temperature),
        pressurePoint: form.pressurePoint.trim() || undefined,
        notes: form.notes.trim() || undefined,
        symptomItems,
      })
    } catch (error) {
      console.error('[PreAssessment] save failed', error)
      message.error(getErrorMessage(error))
    }
  }

  const startTime = treatment?.startTime
    ? new Date(treatment.startTime).toLocaleString('zh-CN', { hour: '2-digit', minute: '2-digit' })
    : '未开始'

  return (
    <div className="space-y-5 pb-6">
      {treatmentLoading ? (
        <section className="rounded-lg border border-blue-100 bg-blue-50 px-4 py-3 text-sm font-semibold text-blue-700">
          正在切换患者并加载今日治疗，透前评估已清空旧患者数据。
        </section>
      ) : null}

      <AssessmentSection title="体重与容量评估" icon={<Scale size={18} className="text-blue-600" />}>
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-5">
          <FieldInput label="* 透前体重" value={form.refuseMeasure ? '拒测' : form.weight} onChange={(v) => updateField('weight', v)} suffix="KG" disabled={form.refuseMeasure} placeholder="优先从体重秤获取" />
          <FieldInput label="额外体重" value={form.extraWeight} onChange={(v) => updateField('extraWeight', v)} suffix="KG" placeholder="来自治疗方案" />
          <FieldInput label="干体重" value={String(patient.dryWeight || '')} suffix="KG" disabled placeholder="来自患者方案" />
          <FieldInput label="体重增长" value={weightGain} suffix="KG" disabled placeholder="自动计算" />
          <FieldInput label="* 目标超滤量" value={displayTargetUf} onChange={handleTargetUfChange} suffix="L" placeholder="默认=体重增长+0.2" />
        </div>
        <div className="mt-2 flex items-center gap-4 text-xs text-slate-600">
          <label className="inline-flex items-center gap-1"><input type="checkbox" checked={form.refuseMeasure} onChange={(e) => updateField('refuseMeasure', e.target.checked)} />患者拒测</label>
          <label className="inline-flex items-center gap-1"><input type="checkbox" checked={form.bedridden} onChange={(e) => updateField('bedridden', e.target.checked)} />卧床</label>
          {weightWarning ? <span className="inline-flex items-center gap-1 text-amber-600"><AlertTriangle size={13} />{weightWarning}</span> : null}
        </div>
      </AssessmentSection>

      <AssessmentSection title="生命体征监测" icon={<Activity size={18} className="text-rose-500" />}>
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-5">
          <div>
            <BpInput sbp={form.sbp} dbp={form.dbp} onSbpChange={(v) => updateField('sbp', v)} onDbpChange={(v) => updateField('dbp', v)} />
            <p className="mt-1 text-xs text-slate-400">优先取独立血压仪数据</p>
          </div>
          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-600">测压部位</label>
            <select
              value={form.pressurePoint}
              onChange={(e) => updateField('pressurePoint', e.target.value)}
              className="h-10 w-full rounded-md border border-slate-200 bg-white px-3 text-sm font-medium text-slate-800 outline-none focus:border-blue-400 focus:ring-2 focus:ring-blue-100"
            >
              <option value="">请选择</option>
              {BP_SITES.map((site) => (
                <option key={site} value={site}>{site}</option>
              ))}
            </select>
          </div>
          <div>
            <FieldInput label="* 透前心率" value={form.heartRate} onChange={(v) => updateField('heartRate', v)} suffix="次/分" placeholder="优先取血压仪数据" />
            <p className="mt-1 text-xs text-slate-400">优先取独立血压仪数据</p>
          </div>
          <FieldInput label="* 透前体温" value={form.temperature} onChange={(v) => updateField('temperature', v)} suffix="℃" placeholder="点击输入" />
          <FieldInput label="呼吸" value={form.respiration} onChange={(v) => updateField('respiration', v)} suffix="次/分" placeholder="点击输入" />
          <FieldInput label="疼痛评分" value={form.painScore} onChange={(v) => updateField('painScore', v)} suffix="分" />
        </div>
      </AssessmentSection>

      <AssessmentSection title="血管通路与神志状态" icon={<Stethoscope size={18} className="text-emerald-600" />}>
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-4">
          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-600">A端位点</label>
            <select
              value={form.aSite}
              onChange={(e) => updateField('aSite', e.target.value)}
              className="h-10 w-full rounded-md border border-slate-200 bg-white px-3 text-sm font-medium text-slate-800 outline-none focus:border-blue-400 focus:ring-2 focus:ring-blue-100"
            >
              <option value="">请选择</option>
              {aSiteOptions.map((site) => (
                <option key={site} value={site}>{site}</option>
              ))}
            </select>
          </div>
          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-600">V端位点</label>
            <select
              value={form.vSite}
              onChange={(e) => updateField('vSite', e.target.value)}
              className="h-10 w-full rounded-md border border-slate-200 bg-white px-3 text-sm font-medium text-slate-800 outline-none focus:border-blue-400 focus:ring-2 focus:ring-blue-100"
            >
              <option value="">请选择</option>
              {vSiteOptions.map((site) => (
                <option key={site} value={site}>{site}</option>
              ))}
            </select>
          </div>
          <ChipSelect label="神志状态" options={CONSCIOUSNESS_OPTS} value={form.consciousness} onChange={(v) => updateField('consciousness', v)} />
          <ChipSelect label="护理分级" options={NURSING_LEVEL_OPTS} value={form.nurseLevel} onChange={(v) => updateField('nurseLevel', v)} />
        </div>
        <div className="mt-4 grid grid-cols-1 gap-4 xl:grid-cols-2">
          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-600">内瘘情况描述</label>
            <div className="flex min-h-10 flex-wrap gap-1.5 rounded-md border border-slate-200 bg-white p-2">
              {form.fistulaStatus.map((item) => (
                <span key={item} className="inline-flex items-center rounded-md bg-blue-50 px-2.5 py-1 text-xs font-medium text-blue-700">
                  {item}<button type="button" onClick={() => updateField('fistulaStatus', form.fistulaStatus.filter((s) => s !== item))} className="ml-1 text-blue-400"><X size={12} /></button>
                </span>
              ))}
              <button type="button" onClick={() => { const s = prompt('新增内瘘描述'); if (s?.trim()) updateField('fistulaStatus', [...form.fistulaStatus, s.trim()]) }} className="text-xs font-semibold text-slate-400">新增描述...</button>
            </div>
          </div>
          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-600">透前皮肤记录</label>
            <div className="flex min-h-10 flex-wrap items-center gap-1.5 rounded-md border border-slate-200 bg-white p-2">
              {skinRecordHistory.slice(0, 5).map((item) => (
                <button
                  key={item}
                  type="button"
                  onClick={() => updateField('skinRecord', item)}
                  className={`rounded-md px-2.5 py-1 text-xs font-medium transition ${
                    form.skinRecord === item
                      ? 'bg-emerald-50 text-emerald-700 border border-emerald-200'
                      : 'bg-slate-50 text-slate-600 border border-slate-200 hover:bg-slate-100'
                  }`}
                >
                  {item}
                </button>
              ))}
              <input
                value={form.skinRecord}
                onChange={(e) => updateField('skinRecord', e.target.value)}
                placeholder="填写皮肤记录..."
                className="min-w-[140px] flex-1 text-sm outline-none"
              />
            </div>
          </div>
        </div>
        <div className="mt-4 grid grid-cols-1 gap-4 xl:grid-cols-2">
          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-600">透前症状历史</label>
            <div className="flex min-h-10 flex-wrap gap-1.5 rounded-md border border-slate-200 bg-white p-2">
              {form.symptoms.map((item) => (
                <span key={item} className="inline-flex items-center rounded-md bg-slate-50 px-2.5 py-1 text-xs font-medium text-slate-700">
                  {item}<button type="button" onClick={() => handleRemoveSymptom(item)} className="ml-1 text-slate-400"><X size={12} /></button>
                </span>
              ))}
              <input value={newSymptom} onChange={(e) => setNewSymptom(e.target.value)} onKeyDown={(e) => { if (e.key === 'Enter') handleAddSymptom() }} placeholder="新增症状历史..." className="min-w-[140px] flex-1 text-sm outline-none" />
            </div>
          </div>
          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-600">透前备注</label>
            <textarea value={form.notes} onChange={(e) => updateField('notes', e.target.value)} rows={3} className="w-full resize-none rounded-md border border-slate-200 bg-white px-3 py-2 text-sm text-slate-700 outline-none" placeholder="记录透前特殊情况、观察重点或护理提醒..." />
          </div>
        </div>
      </AssessmentSection>

      <section className="rounded-lg bg-slate-950 px-6 py-5 text-white shadow-lg">
        <div className="grid gap-4 md:grid-cols-4">
          <div><div className="text-xs text-slate-400">透析机器开始时间</div><div className="mt-1 font-bold">{startTime}</div></div>
          <div><div className="text-xs text-slate-400">接诊医生</div><div className="mt-1 font-bold">{treatment?.doctorName || currentUser?.name || '未分配'}</div></div>
          <div><div className="text-xs text-slate-400">当前记录患者</div><div className="mt-1 font-bold">{patient.name}（{patient.bedId}）</div></div>
          <div><div className="text-xs text-slate-400">最后评估人</div><div className="mt-1 font-bold">{currentUser?.name || '未登录'}</div></div>
        </div>
      </section>

      <div className="flex items-center justify-between rounded-lg bg-white px-4 py-3 shadow-sm">
        <div className="flex items-center gap-6 text-sm text-slate-500">
          <span className="flex items-center gap-1">
            <span className="text-xs text-slate-400">接诊医生:</span>
            <span className="font-medium text-slate-700">{treatment?.doctorName || currentUser?.name || '未分配'}</span>
          </span>
          <span className="flex items-center gap-1">
            <User size={14} className="text-slate-400" />
            <span className="text-xs text-slate-400">评估人:</span>
            <span className="font-medium text-slate-700">{currentUser?.name || '未知'}</span>
          </span>
          <span className="text-xs text-slate-400">称重照片历史</span>
        </div>
        <div className="flex gap-3">
          <button type="button" disabled className="rounded-lg border border-slate-200 px-5 py-2 text-sm font-semibold text-slate-400">暂存草稿</button>
          <button type="button" onClick={() => void handleSave()} disabled={saving || treatmentLoading} className="rounded-lg bg-blue-600 px-6 py-2 text-sm font-bold text-white disabled:opacity-60">
            {treatmentLoading ? '治疗加载中...' : saving ? '提交中...' : '提交透前评估'}
          </button>
        </div>
      </div>
    </div>
  )
}
