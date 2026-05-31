import { message } from 'antd'
import { CheckCircle2, ClipboardCheck, ShieldCheck, UserCheck } from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'
import { restApi } from '@/services'
import type { RestTreatment } from '@/services'
import { getErrorMessage } from '@/services/restClient'
import type {
  TreatmentFirstCheckRequest,
  TreatmentSecondCheckRequest,
} from '@/services/restClient'
import type { Patient } from '../types'

interface Props {
  patient: Patient
  treatment: RestTreatment | null
  treatmentLoading?: boolean
  onSaveFirstCheck: (payload: TreatmentFirstCheckRequest) => Promise<void>
  onSaveSecondCheck: (payload: TreatmentSecondCheckRequest) => Promise<void>
}

interface StaffOption {
  id: string
  name: string
}

interface CheckItemState {
  result: boolean
  mistake: string
}

interface FirstCheckFormState {
  operatorId: string
  operateTime: string
  materials: CheckItemState
  param: CheckItemState
  vascular: CheckItemState
  pipeline: CheckItemState
}

interface SecondCheckFormState {
  operatorId: string
  recheckNurseId: string
  qcNurseId: string
  operateTime: string
  dialysisMode: CheckItemState
  prescription: CheckItemState
  anticoagulant: CheckItemState
  vascular: CheckItemState
  lineConnection: CheckItemState
}

function toLocalTimeValue(value?: string) {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  const offset = date.getTimezoneOffset()
  const local = new Date(date.getTime() - offset * 60_000)
  return local.toISOString().slice(0, 16)
}

function parseOptionalNumber(value: string): number | undefined {
  const trimmed = value.trim()
  if (!trimmed) return undefined
  const parsed = Number(trimmed)
  return Number.isFinite(parsed) ? parsed : undefined
}

function mapFirstCheckForm(treatment: RestTreatment | null): FirstCheckFormState {
  const first = treatment?.firstCheck
  return {
    operatorId: first?.operatorId ? String(first.operatorId) : '',
    operateTime: toLocalTimeValue(first?.operateTime),
    materials: { result: first?.materialsResult ?? true, mistake: first?.materialsMistake || '' },
    param: { result: first?.paramResult ?? true, mistake: first?.paramMistake || '' },
    vascular: { result: first?.vascularAccessResult ?? true, mistake: first?.vascularAccessMistake || '' },
    pipeline: { result: first?.pipelineResult ?? true, mistake: first?.pipelineMistake || '' },
  }
}

function mapSecondCheckForm(treatment: RestTreatment | null): SecondCheckFormState {
  const second = treatment?.secondCheck
  return {
    operatorId: second?.operatorId ? String(second.operatorId) : '',
    recheckNurseId: second?.recheckNurseId ? String(second.recheckNurseId) : '',
    qcNurseId: second?.qcNurseId ? String(second.qcNurseId) : '',
    operateTime: toLocalTimeValue(second?.operateTime),
    dialysisMode: { result: second?.dialysisModeResult ?? second?.paramResult ?? true, mistake: second?.dialysisModeMistake || second?.paramMistake || '' },
    prescription: { result: second?.prescriptionResult ?? second?.paramResult ?? true, mistake: second?.prescriptionMistake || second?.paramMistake || '' },
    anticoagulant: { result: second?.anticoagulantResult ?? true, mistake: second?.anticoagulantMistake || '' },
    vascular: { result: second?.vascularAccessResult ?? true, mistake: second?.vascularAccessMistake || '' },
    lineConnection: { result: second?.lineConnectionResult ?? second?.pipelineResult ?? true, mistake: second?.lineConnectionMistake || second?.pipelineMistake || '' },
  }
}

function CheckResultRow({
  label,
  value,
  onChange,
}: {
  label: string
  value: CheckItemState
  onChange: (value: CheckItemState) => void
}) {
  return (
    <div className="grid grid-cols-[160px_1fr_1fr] gap-4 items-center">
      <div className="text-sm font-semibold text-slate-700">{label}</div>
      <div className="flex items-center gap-4">
        <label className="inline-flex items-center gap-2 text-sm text-slate-600">
          <input
            type="radio"
            checked={value.result}
            onChange={() => onChange({ ...value, result: true, mistake: '' })}
          />
          正常
        </label>
        <label className="inline-flex items-center gap-2 text-sm text-slate-600">
          <input
            type="radio"
            checked={!value.result}
            onChange={() => onChange({ ...value, result: false })}
          />
          异常
        </label>
      </div>
      <input
        value={value.mistake}
        onChange={(e) => onChange({ ...value, mistake: e.target.value })}
        placeholder="异常原因"
        className="h-10 rounded-xl border border-slate-200 px-3 text-sm font-medium outline-none"
      />
    </div>
  )
}

function StaffSelect({
  label,
  value,
  options,
  onChange,
}: {
  label: string
  value: string
  options: StaffOption[]
  onChange: (value: string) => void
}) {
  return (
    <label className="block">
      <div className="text-[11px] font-semibold uppercase tracking-wide text-slate-400 mb-2">{label}</div>
      <select
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="h-11 w-full rounded-2xl border border-slate-200 bg-white px-4 text-sm font-semibold outline-none"
      >
        <option value="">请选择</option>
        {options.map((item) => (
          <option key={item.id} value={item.id}>
            {item.name}
          </option>
        ))}
      </select>
    </label>
  )
}

export default function Verification({
  patient,
  treatment,
  treatmentLoading = false,
  onSaveFirstCheck,
  onSaveSecondCheck,
}: Props) {
  const [staffOptions, setStaffOptions] = useState<StaffOption[]>([])
  const [loadingStaff, setLoadingStaff] = useState(false)
  const [savingFirst, setSavingFirst] = useState(false)
  const [savingSecond, setSavingSecond] = useState(false)
  const [firstForm, setFirstForm] = useState<FirstCheckFormState>(mapFirstCheckForm(treatment))
  const [secondForm, setSecondForm] = useState<SecondCheckFormState>(mapSecondCheckForm(treatment))

  useEffect(() => {
    setFirstForm(mapFirstCheckForm(treatment))
    setSecondForm(mapSecondCheckForm(treatment))
  }, [treatment])

  useEffect(() => {
    const loadStaff = async () => {
      setLoadingStaff(true)
      try {
        const users = await restApi.getUserList({ status: 'active' })
        setStaffOptions(
          users.items.map((item) => ({
            id: item.id,
            name: item.realName || item.username,
          }))
        )
      } catch (error) {
        console.error('[Verification] load staff failed', error)
        message.error(getErrorMessage(error))
      } finally {
        setLoadingStaff(false)
      }
    }
    void loadStaff()
  }, [])

  const disinfectionOperator = useMemo(() => {
    return staffOptions.find((item) => item.id === firstForm.operatorId)?.name || patient.name
  }, [firstForm.operatorId, patient.name, staffOptions])

  const handleSaveFirst = async () => {
    try {
      setSavingFirst(true)
      await onSaveFirstCheck({
        operatorId: parseOptionalNumber(firstForm.operatorId),
        operateTime: firstForm.operateTime ? new Date(firstForm.operateTime).toISOString() : undefined,
        materialsResult: firstForm.materials.result,
        materialsMistake: firstForm.materials.result ? '' : firstForm.materials.mistake.trim() || undefined,
        paramResult: firstForm.param.result,
        paramMistake: firstForm.param.result ? '' : firstForm.param.mistake.trim() || undefined,
        vascularAccessResult: firstForm.vascular.result,
        vascularAccessMistake: firstForm.vascular.result ? '' : firstForm.vascular.mistake.trim() || undefined,
        pipelineResult: firstForm.pipeline.result,
        pipelineMistake: firstForm.pipeline.result ? '' : firstForm.pipeline.mistake.trim() || undefined,
      })
    } catch (error) {
      console.error('[Verification] save first failed', error)
      message.error(getErrorMessage(error))
    } finally {
      setSavingFirst(false)
    }
  }

  const handleSaveSecond = async () => {
    try {
      setSavingSecond(true)
      await onSaveSecondCheck({
        operatorId: parseOptionalNumber(secondForm.operatorId),
        recheckNurseId: parseOptionalNumber(secondForm.recheckNurseId),
        qcNurseId: parseOptionalNumber(secondForm.qcNurseId),
        operateTime: secondForm.operateTime ? new Date(secondForm.operateTime).toISOString() : undefined,
        dialysisModeResult: secondForm.dialysisMode.result,
        dialysisModeMistake: secondForm.dialysisMode.result ? '' : secondForm.dialysisMode.mistake.trim() || undefined,
        prescriptionResult: secondForm.prescription.result,
        prescriptionMistake: secondForm.prescription.result ? '' : secondForm.prescription.mistake.trim() || undefined,
        anticoagulantResult: secondForm.anticoagulant.result,
        anticoagulantMistake: secondForm.anticoagulant.result ? '' : secondForm.anticoagulant.mistake.trim() || undefined,
        vascularAccessResult: secondForm.vascular.result,
        vascularAccessMistake: secondForm.vascular.result ? '' : secondForm.vascular.mistake.trim() || undefined,
        lineConnectionResult: secondForm.lineConnection.result,
        lineConnectionMistake: secondForm.lineConnection.result ? '' : secondForm.lineConnection.mistake.trim() || undefined,
      })
    } catch (error) {
      console.error('[Verification] save second failed', error)
      message.error(getErrorMessage(error))
    } finally {
      setSavingSecond(false)
    }
  }

  return (
    <div className="space-y-6 pb-8">
      {treatmentLoading ? (
        <section className="rounded-3xl border border-blue-100 bg-blue-50 px-6 py-4 text-sm font-semibold text-blue-700">
          正在加载新患者治疗数据，核对表单已切换为空状态。
        </section>
      ) : null}

      <div className="grid grid-cols-1 xl:grid-cols-3 gap-6">
        <section className="rounded-3xl border border-slate-200 bg-white shadow-sm overflow-hidden">
          <div className="px-5 py-4 border-b border-slate-200 flex items-center gap-2">
            <UserCheck size={16} className="text-blue-600" />
            <h3 className="text-sm font-black text-slate-800">首次核对</h3>
          </div>
          <div className="p-5 space-y-5">
            <div className="grid grid-cols-2 gap-4">
              <StaffSelect label="核对人" value={firstForm.operatorId} options={staffOptions} onChange={(value) => setFirstForm((current) => ({ ...current, operatorId: value }))} />
              <label className="block">
                <div className="text-[11px] font-semibold uppercase tracking-wide text-slate-400 mb-2">核对时间</div>
                <input
                  type="datetime-local"
                  value={firstForm.operateTime}
                  onChange={(e) => setFirstForm((current) => ({ ...current, operateTime: e.target.value }))}
                  className="h-11 w-full rounded-2xl border border-slate-200 bg-white px-4 text-sm font-semibold outline-none"
                />
              </label>
            </div>
            <div className="space-y-4">
              <CheckResultRow label="耗材规格核对" value={firstForm.materials} onChange={(value) => setFirstForm((current) => ({ ...current, materials: value }))} />
              <CheckResultRow label="处方参数核对" value={firstForm.param} onChange={(value) => setFirstForm((current) => ({ ...current, param: value }))} />
              <CheckResultRow label="血管通路核对" value={firstForm.vascular} onChange={(value) => setFirstForm((current) => ({ ...current, vascular: value }))} />
              <CheckResultRow label="管路连接核对" value={firstForm.pipeline} onChange={(value) => setFirstForm((current) => ({ ...current, pipeline: value }))} />
            </div>
            <button
              type="button"
              onClick={() => void handleSaveFirst()}
              disabled={treatmentLoading || savingFirst || loadingStaff}
              className="w-full h-12 rounded-2xl bg-blue-600 text-sm font-semibold text-white disabled:opacity-60"
            >
              {treatmentLoading ? '治疗加载中...' : savingFirst ? '保存中...' : '保存首次核对'}
            </button>
          </div>
        </section>

        <section className="rounded-3xl border border-slate-200 bg-white shadow-sm overflow-hidden">
          <div className="px-5 py-4 border-b border-slate-200 flex items-center gap-2">
            <ShieldCheck size={16} className="text-orange-500" />
            <h3 className="text-sm font-black text-slate-800">二次核对</h3>
          </div>
          <div className="p-5 space-y-5">
            <div className="grid grid-cols-1 gap-4">
              <StaffSelect label="上机护士" value={secondForm.operatorId} options={staffOptions} onChange={(value) => setSecondForm((current) => ({ ...current, operatorId: value }))} />
              <div className="grid grid-cols-2 gap-4">
                <StaffSelect label="复核护士" value={secondForm.recheckNurseId} options={staffOptions} onChange={(value) => setSecondForm((current) => ({ ...current, recheckNurseId: value }))} />
                <StaffSelect label="质控护士" value={secondForm.qcNurseId} options={staffOptions} onChange={(value) => setSecondForm((current) => ({ ...current, qcNurseId: value }))} />
              </div>
              <label className="block">
                <div className="text-[11px] font-semibold uppercase tracking-wide text-slate-400 mb-2">核对时间</div>
                <input
                  type="datetime-local"
                  value={secondForm.operateTime}
                  onChange={(e) => setSecondForm((current) => ({ ...current, operateTime: e.target.value }))}
                  className="h-11 w-full rounded-2xl border border-slate-200 bg-white px-4 text-sm font-semibold outline-none"
                />
              </label>
            </div>
            <div className="space-y-4">
              <CheckResultRow label="透析模式核对" value={secondForm.dialysisMode} onChange={(value) => setSecondForm((current) => ({ ...current, dialysisMode: value }))} />
              <CheckResultRow label="处方内容核对" value={secondForm.prescription} onChange={(value) => setSecondForm((current) => ({ ...current, prescription: value }))} />
              <CheckResultRow label="抗凝方案核对" value={secondForm.anticoagulant} onChange={(value) => setSecondForm((current) => ({ ...current, anticoagulant: value }))} />
              <CheckResultRow label="血管通路复核" value={secondForm.vascular} onChange={(value) => setSecondForm((current) => ({ ...current, vascular: value }))} />
              <CheckResultRow label="管路连接复核" value={secondForm.lineConnection} onChange={(value) => setSecondForm((current) => ({ ...current, lineConnection: value }))} />
            </div>
            <button
              type="button"
              onClick={() => void handleSaveSecond()}
              disabled={treatmentLoading || savingSecond || loadingStaff}
              className="w-full h-12 rounded-2xl bg-orange-500 text-sm font-semibold text-white disabled:opacity-60"
            >
              {treatmentLoading ? '治疗加载中...' : savingSecond ? '保存中...' : '保存二次核对'}
            </button>
          </div>
        </section>

        <section className="rounded-3xl border border-emerald-200 bg-white shadow-sm overflow-hidden">
          <div className="px-5 py-4 border-b border-emerald-100 flex items-center justify-between gap-3">
            <div className="flex items-center gap-2">
              <ClipboardCheck size={16} className="text-emerald-600" />
              <h3 className="text-sm font-black text-slate-800">消毒登记</h3>
            </div>
            <span className="rounded-full bg-emerald-100 px-3 py-1 text-[11px] font-black uppercase tracking-wide text-emerald-700">
              未接入保存
            </span>
          </div>
          <div className="p-5 space-y-4">
            <label className="block">
              <div className="text-[11px] font-semibold uppercase tracking-wide text-slate-400 mb-2">消毒方式</div>
              <input value="机表消毒" readOnly disabled className="h-11 w-full rounded-2xl border border-slate-200 bg-slate-100 px-4 text-sm font-semibold text-slate-500 outline-none disabled:cursor-not-allowed" />
            </label>
            <label className="block">
              <div className="text-[11px] font-semibold uppercase tracking-wide text-slate-400 mb-2">消毒液</div>
              <input value="500mg/L 含氯消毒液" readOnly disabled className="h-11 w-full rounded-2xl border border-slate-200 bg-slate-100 px-4 text-sm font-semibold text-slate-500 outline-none disabled:cursor-not-allowed" />
            </label>
            <div className="rounded-2xl bg-emerald-50 border border-emerald-100 p-4 flex items-center justify-between">
              <div>
                <div className="text-[11px] font-semibold uppercase tracking-wide text-emerald-600">登记患者</div>
                <div className="mt-1 text-sm font-bold text-slate-800">{patient.name}</div>
                <div className="mt-1 text-xs text-slate-500">登记人：{disinfectionOperator}</div>
              </div>
              <CheckCircle2 size={20} className="text-emerald-600" />
            </div>
            <div className="text-xs text-slate-400">当前页面仅保留展示态，消毒登记未绑定独立保存接口，避免产生未持久化的假数据。</div>
          </div>
        </section>
      </div>
    </div>
  )
}
