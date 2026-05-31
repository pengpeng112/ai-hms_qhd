import { message } from 'antd'
import { Activity, AlertCircle, Clock, Droplets, Edit3, Gauge, Plus, Trash2 } from 'lucide-react'
import { type ReactNode, useEffect, useMemo, useState } from 'react'
import { restApi } from '@/services'
import { getErrorMessage } from '@/services/restClient'
import type { RestTreatment, TreatmentDuringParamRequest, RestDuringParam } from '@/services/restClient'
import type { Patient } from '../types'

interface Props {
  patient: Patient
  treatment: RestTreatment | null
  treatmentLoading?: boolean
  onCreate: (payload: TreatmentDuringParamRequest) => Promise<void>
  onUpdate: (paramId: number, payload: TreatmentDuringParamRequest) => Promise<void>
  onDelete: (paramId: number) => Promise<void>
}

interface MonitorFormState {
  recordTime: string
  sbp: string
  dbp: string
  heartRate: string
  respiration: string
  spO2: string
  ufVolume: string
  bloodFlow: string
  dialysateFlow: string
  venousPressure: string
  arterialPressure: string
  tmp: string
  temperature: string
  conductivity: string
  ufRate: string
  notes: string
}

const EMPTY_FORM: MonitorFormState = {
  recordTime: '',
  sbp: '',
  dbp: '',
  heartRate: '',
  respiration: '',
  spO2: '',
  ufVolume: '',
  bloodFlow: '',
  dialysateFlow: '',
  venousPressure: '',
  arterialPressure: '',
  tmp: '',
  temperature: '',
  conductivity: '',
  ufRate: '',
  notes: '',
}

function toText(value?: number | string | null) {
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

function formatDateTime(value?: string) {
  if (!value) return '--'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString('zh-CN', { hour12: false })
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

function mapParamToForm(param: RestDuringParam | null): MonitorFormState {
  if (!param) return { ...EMPTY_FORM, recordTime: toDateTimeLocal(new Date().toISOString()) }
  return {
    recordTime: toDateTimeLocal(param.recordTime),
    sbp: toText(param.sbp),
    dbp: toText(param.dbp),
    heartRate: toText(param.heartRate),
    respiration: toText(param.respiration),
    spO2: toText(param.spO2),
    ufVolume: toText(param.ufVolume),
    bloodFlow: toText(param.bloodFlow),
    dialysateFlow: toText(param.dialysateFlow),
    venousPressure: toText(param.venousPressure),
    arterialPressure: toText(param.arterialPressure),
    tmp: toText(param.tmp),
    temperature: toText(param.temperature),
    conductivity: toText(param.conductivity),
    ufRate: toText(param.ufRate),
    notes: param.notes || '',
  }
}

function buildPayload(form: MonitorFormState): TreatmentDuringParamRequest {
  return {
    recordTime: toIsoOrUndefined(form.recordTime),
    sbp: parseOptionalNumber(form.sbp),
    dbp: parseOptionalNumber(form.dbp),
    heartRate: parseOptionalNumber(form.heartRate),
    respiration: parseOptionalNumber(form.respiration),
    spO2: parseOptionalNumber(form.spO2),
    ufVolume: parseOptionalNumber(form.ufVolume),
    bloodFlow: parseOptionalNumber(form.bloodFlow),
    dialysateFlow: parseOptionalNumber(form.dialysateFlow),
    venousPressure: parseOptionalNumber(form.venousPressure),
    arterialPressure: parseOptionalNumber(form.arterialPressure),
    tmp: parseOptionalNumber(form.tmp),
    temperature: parseOptionalNumber(form.temperature),
    conductivity: parseOptionalNumber(form.conductivity),
    ufRate: parseOptionalNumber(form.ufRate),
    notes: form.notes.trim() || undefined,
  }
}

function NumericField({ label, value, onChange, unit }: { label: string; value: string; onChange: (value: string) => void; unit?: string }) {
  return (
    <label className="block min-w-0">
      <span className="mb-2 block text-xs font-semibold text-slate-400">{label}</span>
      <div className="relative">
        <input value={value} onChange={(e) => onChange(e.target.value)} className="h-11 w-full rounded-lg border border-slate-200 px-4 pr-16 text-sm font-semibold text-slate-700 outline-none transition focus:border-blue-400" />
        {unit ? <span className="absolute right-4 top-1/2 -translate-y-1/2 text-xs text-slate-400">{unit}</span> : null}
      </div>
    </label>
  )
}

function MetricCard({ label, value, unit, icon, tone }: { label: string; value: string; unit?: string; icon: ReactNode; tone: string }) {
  return (
    <div className="rounded-lg border border-slate-200 bg-white px-5 py-4 shadow-sm">
      <div className="flex items-center gap-4">
        <div className={`flex h-12 w-12 items-center justify-center rounded-lg ${tone}`}>{icon}</div>
        <div>
          <div className="text-xs font-bold text-slate-400">{label}</div>
          <div className="mt-2 text-xl font-black text-slate-900">
            {value || '--'} <span className="text-xs font-semibold text-slate-400">{unit}</span>
          </div>
        </div>
      </div>
    </div>
  )
}

export default function MidMonitoring({ patient, treatment, treatmentLoading = false, onCreate, onUpdate, onDelete }: Props) {
  const [users, setUsers] = useState<Record<string, string>>({})
  const [modalOpen, setModalOpen] = useState(false)
  const [saving, setSaving] = useState(false)
  const [editingParam, setEditingParam] = useState<RestDuringParam | null>(null)
  const [form, setForm] = useState<MonitorFormState>(mapParamToForm(null))

  useEffect(() => {
    const loadUsers = async () => {
      try {
        const list = await restApi.getUserList({ status: 'active' })
        setUsers(Object.fromEntries(list.items.map((item: { id: number | string; realName?: string; username?: string }) => [String(item.id), item.realName || item.username || String(item.id)])))
      } catch (error) {
        console.error('[MidMonitoring] load users failed', error)
      }
    }
    void loadUsers()
  }, [])

  const rows = useMemo(
    () => [...(treatment?.duringParams || [])].sort((a, b) => new Date(b.recordTime).getTime() - new Date(a.recordTime).getTime()),
    [treatment]
  )

  const stats = useMemo(() => {
    const latest = rows[0]
    const meanArterial = rows
      .filter((item) => item.sbp !== undefined && item.dbp !== undefined)
      .map((item) => Math.round(((item.sbp || 0) + 2 * (item.dbp || 0)) / 3))
    const avgBp = meanArterial.length > 0 ? Math.round(meanArterial.reduce((sum, item) => sum + item, 0) / meanArterial.length).toString() : '--'
    return {
      avgBp,
      latestTmp: latest?.tmp !== undefined ? String(latest.tmp) : '--',
      currentFlow: latest?.bloodFlow !== undefined ? String(latest.bloodFlow) : '--',
      ufRate: latest?.ufRate !== undefined ? String(latest.ufRate) : '--',
      alarmCount: rows.filter((item) => item.notes?.trim()).length,
      currentUf: latest?.ufVolume !== undefined ? String(latest.ufVolume) : '--',
      latestTime: formatDateTime(latest?.recordTime),
    }
  }, [rows])

  const updateField = (key: keyof MonitorFormState, value: string) => {
    setForm((current) => ({ ...current, [key]: value }))
  }

  const closeModal = () => {
    setModalOpen(false)
    setEditingParam(null)
    setForm(mapParamToForm(null))
  }

  const openCreate = () => {
    setEditingParam(null)
    setForm(mapParamToForm(null))
    setModalOpen(true)
  }

  const openEdit = (param: RestDuringParam) => {
    setEditingParam(param)
    setForm(mapParamToForm(param))
    setModalOpen(true)
  }

  const handleSave = async () => {
    try {
      setSaving(true)
      if (editingParam) {
        await onUpdate(editingParam.id, buildPayload(form))
        message.success('透中监测已更新')
      } else {
        await onCreate(buildPayload(form))
        message.success('透中监测已新增')
      }
      closeModal()
    } catch (error) {
      console.error('[MidMonitoring] save failed', error)
      message.error(getErrorMessage(error))
    } finally {
      setSaving(false)
    }
  }

  const handleRemove = async (paramId: number) => {
    try {
      await onDelete(paramId)
      message.success('透中监测已删除')
    } catch (error) {
      console.error('[MidMonitoring] delete failed', error)
      message.error(getErrorMessage(error))
    }
  }

  return (
    <div className="space-y-5 pb-8">
      {treatmentLoading ? (
        <section className="rounded-lg border border-blue-100 bg-blue-50 px-6 py-4 text-sm font-semibold text-blue-700">正在加载新患者治疗数据，透中监测列表已隐藏旧记录。</section>
      ) : null}

      <div className="grid grid-cols-1 gap-4 md:grid-cols-5">
        <MetricCard label="平均动脉压" value={stats.avgBp} unit="mmHg" icon={<Activity size={22} className="text-blue-600" />} tone="bg-blue-50" />
        <MetricCard label="实时跨膜压" value={stats.latestTmp} unit="mmHg" icon={<Gauge size={22} className="text-orange-500" />} tone="bg-orange-50" />
        <MetricCard label="当前血流量" value={stats.currentFlow} unit="ml/min" icon={<Droplets size={22} className="text-indigo-600" />} tone="bg-indigo-50" />
        <MetricCard label="超滤速率" value={stats.ufRate} unit="ml/h" icon={<Clock size={22} className="text-emerald-600" />} tone="bg-emerald-50" />
        <MetricCard label="异常预警" value={String(stats.alarmCount)} unit="项" icon={<AlertCircle size={22} className="text-rose-600" />} tone="bg-rose-50" />
      </div>

      <section className="overflow-hidden rounded-lg border border-slate-200 bg-white shadow-sm">
        <div className="flex items-center justify-between border-b border-slate-100 px-6 py-4">
          <div className="flex items-center gap-3">
            <span className="h-7 w-1 rounded-full bg-blue-600" />
            <h3 className="text-base font-black text-slate-900">实时监测记录流水</h3>
            <span className="rounded-md bg-slate-100 px-2 py-1 text-xs font-black tracking-widest text-slate-400">REAL-TIME FEED</span>
          </div>
          <button type="button" onClick={openCreate} disabled={treatmentLoading} className="inline-flex items-center gap-2 rounded-lg bg-blue-600 px-5 py-2.5 text-sm font-bold text-white disabled:opacity-60">
            <Plus size={16} />录入新监测点
          </button>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full min-w-[1500px] text-left">
            <thead className="bg-slate-50 text-xs text-slate-600">
              <tr className="text-center text-blue-600">
                <th className="px-4 py-3" colSpan={2}>基础</th>
                <th className="px-4 py-3" colSpan={6}>生命体征指标</th>
                <th className="px-4 py-3" colSpan={4}>透析核心压力与流量</th>
                <th className="px-4 py-3" colSpan={3}>末尾项</th>
              </tr>
              <tr>
                <th className="px-4 py-3">序号</th>
                <th className="px-4 py-3">观测时间</th>
                <th className="px-4 py-3">收缩压<br />(mmHg)</th>
                <th className="px-4 py-3">舒张压<br />(mmHg)</th>
                <th className="px-4 py-3 text-rose-600">心率 (bpm)</th>
                <th className="px-4 py-3">呼吸 (次/分)</th>
                <th className="px-4 py-3">血氧 (%)</th>
                <th className="px-4 py-3">累计超滤 (L)</th>
                <th className="px-4 py-3 text-blue-600">血流量<br />(ml/min)</th>
                <th className="px-4 py-3 text-blue-600">动脉压<br />(mmHg)</th>
                <th className="px-4 py-3 text-blue-600">静脉压<br />(mmHg)</th>
                <th className="px-4 py-3">跨膜压</th>
                <th className="px-4 py-3">备注</th>
                <th className="px-4 py-3">记录人</th>
                <th className="px-4 py-3 text-right">操作</th>
              </tr>
            </thead>
            <tbody>
              {rows.length > 0 ? rows.map((row, index) => (
                <tr key={row.id} className="border-t border-slate-100 text-sm">
                  <td className="px-4 py-4 text-slate-500">{index + 1}</td>
                  <td className="px-4 py-4 font-semibold text-slate-800">{formatDateTime(row.recordTime)}</td>
                  <td className="px-4 py-4 text-slate-600">{toText(row.sbp) || '--'}</td>
                  <td className="px-4 py-4 text-slate-600">{toText(row.dbp) || '--'}</td>
                  <td className="px-4 py-4 text-rose-600">{toText(row.heartRate) || '--'}</td>
                  <td className="px-4 py-4 text-slate-600">{toText(row.respiration) || '--'}</td>
                  <td className="px-4 py-4 text-slate-600">{toText(row.spO2) || '--'}</td>
                  <td className="px-4 py-4 text-blue-600">{toText(row.ufVolume) || '--'}</td>
                  <td className="px-4 py-4 text-blue-600">{toText(row.bloodFlow) || '--'}</td>
                  <td className="px-4 py-4 text-blue-600">{toText(row.arterialPressure) || '--'}</td>
                  <td className="px-4 py-4 text-blue-600">{toText(row.venousPressure) || '--'}</td>
                  <td className="px-4 py-4 text-slate-600">{toText(row.tmp) || '--'}</td>
                  <td className="px-4 py-4 text-slate-500">{row.notes || '--'}</td>
                  <td className="px-4 py-4 text-slate-500">{users[String(row.creatorId)] || '--'}</td>
                  <td className="px-4 py-4">
                    <div className="flex items-center justify-end gap-2">
                      <button type="button" onClick={() => openEdit(row)} className="flex h-8 w-8 items-center justify-center rounded-lg border border-slate-200 text-slate-500 hover:bg-slate-50" aria-label="编辑监测点"><Edit3 size={14} /></button>
                      <button type="button" onClick={() => void handleRemove(row.id)} className="flex h-8 w-8 items-center justify-center rounded-lg border border-slate-200 text-slate-500 hover:bg-slate-50" aria-label="删除监测点"><Trash2 size={14} /></button>
                    </div>
                  </td>
                </tr>
              )) : (
                <tr><td colSpan={15} className="px-6 py-16 text-center text-sm text-slate-300">暂无监测记录</td></tr>
              )}
            </tbody>
          </table>
        </div>
        <div className="flex flex-wrap items-center justify-between gap-3 border-t border-slate-100 px-6 py-4 text-xs font-semibold text-slate-500">
          <div className="flex items-center gap-6"><span className="text-emerald-600">● 机位状态：正常传输中</span><span>同步间隔：60分钟/点</span><span>最新监测时间：{stats.latestTime}</span><span>超滤进度：{stats.currentUf} L</span></div>
          <div className="flex gap-2"><button type="button" disabled className="rounded-md border border-slate-200 px-3 py-1 text-slate-400">上一页</button><button type="button" disabled className="rounded-md border border-slate-200 px-3 py-1 text-slate-400">下一页</button></div>
        </div>
      </section>

      <div className="rounded-lg border border-amber-200 bg-amber-50 px-5 py-4 text-sm font-semibold text-amber-700">
        监测数据支持自动采集与人工录入结合。最新监测时间：{stats.latestTime}，如需修改历史监测记录，请联系护士长权限核准。
      </div>

      {modalOpen ? (
        <div className="fixed inset-0 z-[100] flex items-center justify-center bg-slate-900/50 px-4 backdrop-blur-sm">
          <div className="w-full max-w-5xl rounded-lg bg-white p-8 shadow-2xl">
            <div className="mb-6 flex items-start justify-between">
              <div><h3 className="text-2xl font-black text-slate-900">{editingParam ? '编辑透中监测' : '新增透中监测'}</h3><div className="mt-2 text-sm font-semibold text-slate-400">{patient.name}</div></div>
              <button type="button" onClick={closeModal} className="text-sm font-semibold text-slate-500">关闭</button>
            </div>
            <div className="grid grid-cols-1 gap-4 md:grid-cols-4">
              <label className="block md:col-span-2"><span className="mb-2 block text-xs font-semibold text-slate-400">记录时间</span><input type="datetime-local" value={form.recordTime} onChange={(e) => updateField('recordTime', e.target.value)} className="h-11 w-full rounded-lg border border-slate-200 px-4 text-sm font-semibold text-slate-700 outline-none transition focus:border-blue-400" /></label>
              <NumericField label="收缩压" value={form.sbp} onChange={(value) => updateField('sbp', value)} unit="mmHg" />
              <NumericField label="舒张压" value={form.dbp} onChange={(value) => updateField('dbp', value)} unit="mmHg" />
              <NumericField label="心率" value={form.heartRate} onChange={(value) => updateField('heartRate', value)} unit="次/分" />
              <NumericField label="呼吸" value={form.respiration} onChange={(value) => updateField('respiration', value)} unit="次/分" />
              <NumericField label="血氧" value={form.spO2} onChange={(value) => updateField('spO2', value)} unit="%" />
              <NumericField label="超滤量" value={form.ufVolume} onChange={(value) => updateField('ufVolume', value)} unit="L" />
              <NumericField label="血流量" value={form.bloodFlow} onChange={(value) => updateField('bloodFlow', value)} unit="ml/min" />
              <NumericField label="透析液流量" value={form.dialysateFlow} onChange={(value) => updateField('dialysateFlow', value)} unit="ml/min" />
              <NumericField label="静脉压" value={form.venousPressure} onChange={(value) => updateField('venousPressure', value)} unit="mmHg" />
              <NumericField label="动脉压" value={form.arterialPressure} onChange={(value) => updateField('arterialPressure', value)} unit="mmHg" />
              <NumericField label="跨膜压" value={form.tmp} onChange={(value) => updateField('tmp', value)} unit="mmHg" />
              <NumericField label="机温" value={form.temperature} onChange={(value) => updateField('temperature', value)} unit="℃" />
              <NumericField label="电导度" value={form.conductivity} onChange={(value) => updateField('conductivity', value)} />
              <NumericField label="超滤率" value={form.ufRate} onChange={(value) => updateField('ufRate', value)} unit="ml/h" />
              <label className="block md:col-span-4"><span className="mb-2 block text-xs font-semibold text-slate-400">备注</span><textarea value={form.notes} onChange={(e) => updateField('notes', e.target.value)} rows={4} className="w-full resize-none rounded-lg border border-slate-200 px-4 py-3 text-sm font-semibold text-slate-700 outline-none transition focus:border-blue-400" /></label>
            </div>
            <div className="mt-6 flex justify-end gap-3"><button type="button" onClick={closeModal} className="rounded-lg border border-slate-200 px-6 py-3 text-sm font-semibold text-slate-600">取消</button><button type="button" onClick={() => void handleSave()} disabled={treatmentLoading || saving} className="rounded-lg bg-blue-600 px-6 py-3 text-sm font-bold text-white disabled:opacity-60">{treatmentLoading ? '治疗加载中...' : saving ? '保存中...' : '保存监测'}</button></div>
          </div>
        </div>
      ) : null}
    </div>
  )
}
