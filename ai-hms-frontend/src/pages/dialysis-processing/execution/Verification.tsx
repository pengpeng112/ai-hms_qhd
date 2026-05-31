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
import { useAuth } from '@/contexts/AuthContext'

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

function formatDateTimeForPicker(date: Date) {
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
    <label className="block min-w-0">
      <span className="mb-2 block text-xs font-semibold text-slate-400">{label}</span>
      <select
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="h-10 w-full rounded-lg border border-slate-200 bg-white px-3 text-sm font-semibold text-slate-800 outline-none"
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

function DateTimeInput({ label, value, onChange }: { label: string; value: string; onChange: (value: string) => void }) {
  return (
    <label className="block min-w-0">
      <span className="mb-2 block text-xs font-semibold text-slate-400">{label}</span>
      <input
        type="datetime-local"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="h-10 w-full rounded-lg border border-slate-200 bg-white px-3 text-sm font-semibold text-slate-800 outline-none"
      />
    </label>
  )
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
    <div className="rounded-lg border border-slate-100 bg-white p-4">
      <div className="flex items-center justify-between gap-3">
        <div className="text-sm font-bold text-slate-800">{label}</div>
        <div className="grid grid-cols-2 rounded-lg bg-slate-100 p-1 text-xs font-bold">
          <button
            type="button"
            onClick={() => onChange({ result: true, mistake: '' })}
            className={`rounded-md px-3 py-1.5 ${value.result ? 'bg-blue-600 text-white shadow-sm' : 'text-slate-500'}`}
          >
            正常
          </button>
          <button
            type="button"
            onClick={() => onChange({ ...value, result: false })}
            className={`rounded-md px-3 py-1.5 ${!value.result ? 'bg-rose-500 text-white shadow-sm' : 'text-slate-500'}`}
          >
            异常
          </button>
        </div>
      </div>
      {!value.result ? (
        <textarea
          value={value.mistake}
          onChange={(e) => onChange({ ...value, mistake: e.target.value })}
          rows={2}
          placeholder="请输入异常原因"
          className="mt-3 w-full resize-none rounded-lg border border-rose-200 bg-rose-50 px-3 py-2 text-sm outline-none"
        />
      ) : null}
    </div>
  )
}

function RoleCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border border-slate-200 bg-white px-5 py-4 text-center shadow-sm">
      <div className="text-xs font-semibold text-slate-400">{label}</div>
      <div className="mt-2 text-sm font-black text-slate-900">{value || '请选择'}</div>
    </div>
  )
}

export default function Verification({
  patient,
  treatment,
  treatmentLoading = false,
  onSaveFirstCheck,
  onSaveSecondCheck,
}: Props) {
  const { user: currentUser } = useAuth()
  const [staffOptions, setStaffOptions] = useState<StaffOption[]>([])
  const [loadingStaff, setLoadingStaff] = useState(false)
  const [savingFirst, setSavingFirst] = useState(false)
  const [savingSecond, setSavingSecond] = useState(false)
  const [savingAll, setSavingAll] = useState(false)
  const [firstForm, setFirstForm] = useState<FirstCheckFormState>(mapFirstCheckForm(treatment))
  const [secondForm, setSecondForm] = useState<SecondCheckFormState>(mapSecondCheckForm(treatment))
  const [disinfectionTime, setDisinfectionTime] = useState('')

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

  const staffNameById = useMemo(() => {
    return new Map(staffOptions.map((item) => [item.id, item.name]))
  }, [staffOptions])

  // 当前用户的员工ID
  const currentUserId = useMemo(() => {
    const found = staffOptions.find(
      (s) => s.name === currentUser?.name || s.id === currentUser?.id
    )
    return found?.id || ''
  }, [staffOptions, currentUser])

  // 默认时间计算
  const defaultTimes = useMemo(() => {
    const startTime = treatment?.startTime ? new Date(treatment.startTime) : null

    // 首次核对时间 = 签到时间（取开始时间替代）+ 1分钟
    const firstCheckTime = startTime
      ? formatDateTimeForPicker(new Date(startTime.getTime() + 1 * 60_000))
      : ''

    // 二次核对时间 = 开始治疗时间 + 10分钟
    const secondCheckTime = startTime
      ? formatDateTimeForPicker(new Date(startTime.getTime() + 10 * 60_000))
      : ''

    // 消毒时间 = 开始治疗时间 + 5分钟
    const disinfectionTime = startTime
      ? formatDateTimeForPicker(new Date(startTime.getTime() + 5 * 60_000))
      : ''

    return { firstCheckTime, secondCheckTime, disinfectionTime }
  }, [treatment])

  // 当表单未填写时使用默认值（仅在加载后首次设置）
  useEffect(() => {
    if (loadingStaff || staffOptions.length === 0) return
    setFirstForm((current) => ({
      ...current,
      operatorId: current.operatorId || currentUserId,
      operateTime: current.operateTime || defaultTimes.firstCheckTime,
    }))
    setSecondForm((current) => ({
      ...current,
      operateTime: current.operateTime || defaultTimes.secondCheckTime,
    }))
    setDisinfectionTime((current) => current || defaultTimes.disinfectionTime)
  }, [loadingStaff, staffOptions.length, currentUserId, defaultTimes])

  const disinfectionOperator = staffNameById.get(firstForm.operatorId) || patient.name

  const buildFirstPayload = (): TreatmentFirstCheckRequest => ({
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

  const buildSecondPayload = (): TreatmentSecondCheckRequest => ({
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

  const handleSaveFirst = async () => {
    try {
      setSavingFirst(true)
      await onSaveFirstCheck(buildFirstPayload())
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
      await onSaveSecondCheck(buildSecondPayload())
    } catch (error) {
      console.error('[Verification] save second failed', error)
      message.error(getErrorMessage(error))
    } finally {
      setSavingSecond(false)
    }
  }

  const handleSaveAll = async () => {
    try {
      setSavingAll(true)
      await onSaveFirstCheck(buildFirstPayload())
      await onSaveSecondCheck(buildSecondPayload())
    } catch (error) {
      console.error('[Verification] save all failed', error)
      message.error(getErrorMessage(error))
    } finally {
      setSavingAll(false)
    }
  }

  return (
    <div className="space-y-5 pb-8">
      {treatmentLoading ? (
        <section className="rounded-lg border border-blue-100 bg-blue-50 px-6 py-4 text-sm font-semibold text-blue-700">
          正在加载新患者治疗数据，核对表单已切换为空状态。
        </section>
      ) : null}

      <div className="flex flex-wrap items-center gap-2 text-sm font-semibold text-slate-600">
        <span className="text-xl font-black text-slate-900">{patient.name}</span>
        <span>ID: {patient.id}</span>
        <span>{patient.gender} / {patient.age}岁</span>
        <span>治疗方案: {patient.treatmentPlan || '--'}</span>
      </div>

      <div className="grid grid-cols-1 gap-5 xl:grid-cols-3">
        <section className="overflow-hidden rounded-lg border border-slate-200 bg-white shadow-sm">
          <div className="flex items-center gap-2 border-b border-slate-100 px-5 py-4">
            <UserCheck size={16} className="text-blue-600" />
            <h3 className="text-sm font-black text-slate-800">首次核对</h3>
            <span className="ml-auto text-xs text-slate-400">
              默认取病区责任护士
            </span>
          </div>
          <div className="space-y-5 p-5">
            <div className="rounded-lg border border-slate-100 bg-slate-50 px-4 py-4 text-sm leading-6 text-slate-600">
              核对项目：透析模式、处方参数、耗材规格、患者身份、管路连接安全性。
            </div>
            <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
              <StaffSelect label="核对人" value={firstForm.operatorId} options={staffOptions} onChange={(value) => setFirstForm((current) => ({ ...current, operatorId: value }))} />
              <DateTimeInput label="核对时间" value={firstForm.operateTime} onChange={(value) => setFirstForm((current) => ({ ...current, operateTime: value }))} />
            </div>
            <div className="space-y-3">
              <CheckResultRow label="耗材规格（透析器/血路管）" value={firstForm.materials} onChange={(value) => setFirstForm((current) => ({ ...current, materials: value }))} />
              <CheckResultRow label="处方参数（透析方式/处方内容）" value={firstForm.param} onChange={(value) => setFirstForm((current) => ({ ...current, param: value }))} />
              <CheckResultRow label="血管通路与患者身份" value={firstForm.vascular} onChange={(value) => setFirstForm((current) => ({ ...current, vascular: value }))} />
              <CheckResultRow label="管路连接与预冲" value={firstForm.pipeline} onChange={(value) => setFirstForm((current) => ({ ...current, pipeline: value }))} />
            </div>
            <button type="button" onClick={() => void handleSaveFirst()} disabled={treatmentLoading || savingFirst || loadingStaff} className="h-11 w-full rounded-lg bg-blue-600 text-sm font-bold text-white disabled:opacity-60">
              {treatmentLoading ? '治疗加载中...' : savingFirst ? '保存中...' : '确认并完成核对'}
            </button>
          </div>
        </section>

        <section className="overflow-hidden rounded-lg border border-slate-200 bg-white shadow-sm">
          <div className="flex items-center gap-2 border-b border-slate-100 px-5 py-4">
            <ShieldCheck size={16} className="text-orange-500" />
            <h3 className="text-sm font-black text-slate-800">二次核对</h3>
            <span className="ml-auto text-xs text-slate-400">
              不可与首次核对人相同
            </span>
          </div>
          <div className="space-y-5 p-5">
            <div className="rounded-lg border border-slate-100 bg-slate-50 px-4 py-4 text-sm leading-6 text-slate-600">
              二次核对重点：核实透析参数、处方调整项、耗材效期及批号一致性。
            </div>
            <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
              <StaffSelect label="核对人" value={secondForm.operatorId} options={staffOptions.filter((s) => s.id !== firstForm.operatorId)} onChange={(value) => setSecondForm((current) => ({ ...current, operatorId: value }))} />
              <DateTimeInput label="核对时间" value={secondForm.operateTime} onChange={(value) => setSecondForm((current) => ({ ...current, operateTime: value }))} />
            </div>
            <div className="space-y-3">
              <CheckResultRow label="透析模式" value={secondForm.dialysisMode} onChange={(value) => setSecondForm((current) => ({ ...current, dialysisMode: value }))} />
              <CheckResultRow label="处方内容" value={secondForm.prescription} onChange={(value) => setSecondForm((current) => ({ ...current, prescription: value }))} />
              <CheckResultRow label="抗凝剂" value={secondForm.anticoagulant} onChange={(value) => setSecondForm((current) => ({ ...current, anticoagulant: value }))} />
              <CheckResultRow label="血管通路" value={secondForm.vascular} onChange={(value) => setSecondForm((current) => ({ ...current, vascular: value }))} />
              <CheckResultRow label="管路连接" value={secondForm.lineConnection} onChange={(value) => setSecondForm((current) => ({ ...current, lineConnection: value }))} />
            </div>
            <button type="button" onClick={() => void handleSaveSecond()} disabled={treatmentLoading || savingSecond || loadingStaff} className="h-11 w-full rounded-lg bg-orange-500 text-sm font-bold text-white disabled:opacity-60">
              {treatmentLoading ? '治疗加载中...' : savingSecond ? '保存中...' : '确认并完成核对'}
            </button>
          </div>
        </section>

        <section className="overflow-hidden rounded-lg border border-emerald-200 bg-emerald-50/30 shadow-sm">
          <div className="flex items-center gap-2 border-b border-emerald-100 px-5 py-4">
            <ClipboardCheck size={16} className="text-emerald-600" />
            <h3 className="text-sm font-black text-slate-800">机器消毒登记</h3>
          </div>
          <div className="space-y-6 p-5">
            <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
              <label className="block">
                <span className="mb-2 block text-xs font-semibold text-slate-400">消毒类型</span>
                <input value="机器" readOnly disabled className="h-10 w-full rounded-lg border border-emerald-200 bg-white px-3 text-sm font-bold text-emerald-700" />
              </label>
              <label className="block">
                <span className="mb-2 block text-xs font-semibold text-slate-400">消毒液</span>
                <input value="500mg/L含氯消毒液" readOnly disabled className="h-10 w-full rounded-lg border border-emerald-200 bg-white px-3 text-sm font-bold text-emerald-700" />
              </label>
            </div>
            <DateTimeInput label="消毒时间" value={disinfectionTime || defaultTimes.disinfectionTime} onChange={setDisinfectionTime} />
            <div className="mt-28 rounded-lg border border-emerald-200 bg-emerald-50 px-4 py-4">
              <div className="flex items-center justify-between">
                <div>
                  <div className="text-xs font-semibold text-emerald-600">登记人</div>
                  <div className="mt-1 text-sm font-black text-slate-900">{disinfectionOperator}</div>
                </div>
                <CheckCircle2 size={20} className="text-emerald-500" />
              </div>
            </div>
            <button type="button" disabled title="功能待后端接口就绪" className="h-11 w-full cursor-not-allowed rounded-lg bg-emerald-600 text-sm font-bold text-white opacity-60">
              功能待后端接口就绪
            </button>
          </div>
        </section>
      </div>

      <section className="space-y-4">
        <div className="flex items-center justify-between">
          <h3 className="border-l-4 border-slate-700 pl-3 text-sm font-black text-slate-800">人员配置与角色登记</h3>
          <span className="text-xs text-slate-400">默认取本床位所在病区责任护士</span>
        </div>
        <div className="grid grid-cols-1 gap-4 md:grid-cols-5">
          <RoleCard label="预冲护士" value={staffNameById.get(firstForm.operatorId) || currentUser?.name || ''} />
          <RoleCard label="穿刺/注射" value={staffNameById.get(firstForm.operatorId) || currentUser?.name || ''} />
          <RoleCard label="上机护士" value={staffNameById.get(firstForm.operatorId) || currentUser?.name || ''} />
          <RoleCard label="质控护士" value={staffNameById.get(secondForm.qcNurseId) || ''} />
          <RoleCard label="质检医生" value="" />
        </div>
      </section>

      <div className="flex items-center justify-between border-t border-slate-200 pt-4">
        <div className="text-xs text-slate-400">系统已根据当前排班与核对记录自动建议登记人员，请在提交前进行最后确认。</div>
        <div className="flex gap-3">
          <button type="button" disabled className="rounded-lg border border-slate-200 bg-white px-8 py-2 text-sm font-semibold text-slate-400">暂存修改</button>
          <button type="button" onClick={() => void handleSaveAll()} disabled={treatmentLoading || loadingStaff || savingAll} className="rounded-lg bg-blue-600 px-8 py-2 text-sm font-bold text-white disabled:opacity-60">
            {savingAll ? '提交中...' : '提交并生效'}
          </button>
        </div>
      </div>
    </div>
  )
}
