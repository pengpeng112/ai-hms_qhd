import { message } from 'antd'
import { Activity, Clock, Droplets, Edit3, Plus, Trash2 } from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'
import { restApi } from '@/services'
import type { RestTreatment, TreatmentDuringParamRequest, RestDuringParam } from '@/services/restClient'
import type { Patient } from '../types'

interface Props {
  patient: Patient
  treatment: RestTreatment | null
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
  if (!value) return '-'
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
  if (!param) {
    return {
      ...EMPTY_FORM,
      recordTime: toDateTimeLocal(new Date().toISOString()),
    }
  }

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

function NumericField({
  label,
  value,
  onChange,
  unit,
}: {
  label: string
  value: string
  onChange: (value: string) => void
  unit?: string
}) {
  return (
    <label className="block">
      <div className="mb-2 text-[11px] font-semibold uppercase tracking-wide text-slate-400">{label}</div>
      <div className="relative">
        <input
          value={value}
          onChange={(e) => onChange(e.target.value)}
          className="h-11 w-full rounded-2xl border border-slate-200 px-4 pr-14 text-sm font-semibold text-slate-700 outline-none transition focus:border-blue-400"
        />
        {unit ? <span className="absolute right-4 top-1/2 -translate-y-1/2 text-xs text-slate-400">{unit}</span> : null}
      </div>
    </label>
  )
}

export default function MidMonitoring({
  patient,
  treatment,
  onCreate,
  onUpdate,
  onDelete,
}: Props) {
  const [users, setUsers] = useState<Record<string, string>>({})
  const [modalOpen, setModalOpen] = useState(false)
  const [saving, setSaving] = useState(false)
  const [editingParam, setEditingParam] = useState<RestDuringParam | null>(null)
  const [form, setForm] = useState<MonitorFormState>(mapParamToForm(null))

  useEffect(() => {
    const loadUsers = async () => {
      try {
        const list = await restApi.getUserList({ status: 'active' })
        setUsers(
          Object.fromEntries(list.map((item) => [String(item.id), item.realName || item.username || String(item.id)]))
        )
      } catch (error) {
        console.error('[MidMonitoring] load users failed', error)
      }
    }

    void loadUsers()
  }, [])

  const rows = useMemo(
    () =>
      [...(treatment?.duringParams || [])].sort(
        (a, b) => new Date(b.recordTime).getTime() - new Date(a.recordTime).getTime()
      ),
    [treatment]
  )

  const stats = useMemo(() => {
    const latest = rows[0]
    const tmpList = rows.map((item) => item.tmp).filter((item): item is number => item !== undefined)
    const avgTmp = tmpList.length > 0 ? (tmpList.reduce((sum, item) => sum + item, 0) / tmpList.length).toFixed(0) : '-'
    const latestBp =
      latest?.sbp !== undefined && latest?.dbp !== undefined ? `${latest.sbp}/${latest.dbp}` : '-'

    return {
      currentUf: latest?.ufVolume !== undefined ? String(latest.ufVolume) : '-',
      currentFlow: latest?.bloodFlow !== undefined ? String(latest.bloodFlow) : '-',
      avgTmp,
      latestBp,
      latestTime: formatDateTime(latest?.recordTime),
    }
  }, [rows])

  const updateField = (key: keyof MonitorFormState, value: string) => {
    setForm((current) => ({ ...current, [key]: value }))
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
      setModalOpen(false)
      setEditingParam(null)
      setForm(mapParamToForm(null))
    } catch (error) {
      console.error('[MidMonitoring] save failed', error)
      message.error(editingParam ? '透中监测更新失败' : '透中监测新增失败')
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
      message.error('透中监测删除失败')
    }
  }

  return (
    <div className="space-y-6 pb-8">
      <div className="grid grid-cols-1 gap-4 md:grid-cols-5">
        {[
          ['当前患者', patient.name, ''],
          ['最新血压', stats.latestBp, 'mmHg'],
          ['实时超滤', stats.currentUf, 'L'],
          ['当前血流量', stats.currentFlow, 'ml/min'],
          ['平均跨膜压', stats.avgTmp, 'mmHg'],
        ].map(([label, value, unit]) => (
          <div key={label} className="rounded-2xl border border-slate-200 bg-white px-4 py-4 shadow-sm">
            <div className="text-[11px] font-semibold uppercase tracking-wide text-slate-400">{label}</div>
            <div className="mt-2 text-2xl font-black text-slate-800">
              {value}
              {unit ? <span className="ml-1 text-xs font-medium text-slate-400">{unit}</span> : null}
            </div>
          </div>
        ))}
      </div>

      <section className="overflow-hidden rounded-3xl border border-slate-200 bg-white shadow-sm">
        <div className="flex items-center justify-between border-b border-slate-200 px-6 py-4">
          <div className="flex items-center gap-2">
            <Activity size={16} className="text-blue-600" />
            <h3 className="text-sm font-black text-slate-800">透中监测记录</h3>
          </div>
          <button
            type="button"
            onClick={openCreate}
            className="inline-flex items-center gap-2 rounded-2xl bg-blue-600 px-5 py-3 text-sm font-semibold text-white"
          >
            <Plus size={16} />
            新增监测点
          </button>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full min-w-[1400px] text-left">
            <thead className="bg-slate-50 text-xs uppercase tracking-wide text-slate-500">
              <tr>
                <th className="px-6 py-3">记录时间</th>
                <th className="px-6 py-3">血压</th>
                <th className="px-6 py-3">心率</th>
                <th className="px-6 py-3">呼吸</th>
                <th className="px-6 py-3">血氧</th>
                <th className="px-6 py-3">超滤量</th>
                <th className="px-6 py-3">血流量</th>
                <th className="px-6 py-3">静脉压</th>
                <th className="px-6 py-3">动脉压</th>
                <th className="px-6 py-3">跨膜压</th>
                <th className="px-6 py-3">备注</th>
                <th className="px-6 py-3">记录人</th>
                <th className="px-6 py-3 text-right">操作</th>
              </tr>
            </thead>
            <tbody>
              {rows.length > 0 ? (
                rows.map((row) => (
                  <tr key={row.id} className="border-t border-slate-100 text-sm">
                    <td className="px-6 py-4 font-semibold text-slate-800">{formatDateTime(row.recordTime)}</td>
                    <td className="px-6 py-4 text-slate-600">
                      {row.sbp !== undefined && row.dbp !== undefined ? `${row.sbp}/${row.dbp}` : '-'}
                    </td>
                    <td className="px-6 py-4 text-slate-600">{toText(row.heartRate) || '-'}</td>
                    <td className="px-6 py-4 text-slate-600">{toText(row.respiration) || '-'}</td>
                    <td className="px-6 py-4 text-slate-600">{toText(row.spO2) || '-'}</td>
                    <td className="px-6 py-4 text-slate-600">{toText(row.ufVolume) || '-'}</td>
                    <td className="px-6 py-4 text-slate-600">{toText(row.bloodFlow) || '-'}</td>
                    <td className="px-6 py-4 text-slate-600">{toText(row.venousPressure) || '-'}</td>
                    <td className="px-6 py-4 text-slate-600">{toText(row.arterialPressure) || '-'}</td>
                    <td className="px-6 py-4 text-slate-600">{toText(row.tmp) || '-'}</td>
                    <td className="px-6 py-4 text-slate-500">{row.notes || '-'}</td>
                    <td className="px-6 py-4 text-slate-500">{users[String(row.creatorId)] || '--'}</td>
                    <td className="px-6 py-4">
                      <div className="flex items-center justify-end gap-2">
                        <button
                          type="button"
                          onClick={() => openEdit(row)}
                          className="flex h-9 w-9 items-center justify-center rounded-xl border border-slate-200 text-slate-500 transition hover:bg-slate-50"
                        >
                          <Edit3 size={14} />
                        </button>
                        <button
                          type="button"
                          onClick={() => void handleRemove(row.id)}
                          className="flex h-9 w-9 items-center justify-center rounded-xl border border-slate-200 text-slate-500 transition hover:bg-slate-50"
                        >
                          <Trash2 size={14} />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))
              ) : (
                <tr>
                  <td colSpan={13} className="px-6 py-10 text-center text-sm text-slate-400">
                    当前治疗暂无透中监测记录
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </section>

      <div className="rounded-3xl border border-slate-200 bg-slate-900 px-6 py-5 text-white shadow-lg">
        <div className="flex flex-wrap items-center gap-4">
          <div className="flex items-center gap-3">
            <Clock size={18} className="text-slate-300" />
            <div className="text-sm font-semibold">最新监测时间：{stats.latestTime}</div>
          </div>
          <div className="flex items-center gap-3">
            <Droplets size={18} className="text-slate-300" />
            <div className="text-sm font-semibold">超滤进度：{stats.currentUf} L</div>
          </div>
        </div>
      </div>

      {modalOpen ? (
        <div className="fixed inset-0 z-[100] flex items-center justify-center bg-slate-900/50 px-4 backdrop-blur-sm">
          <div className="w-full max-w-5xl rounded-[32px] border border-slate-200 bg-white p-8 shadow-2xl">
            <div className="mb-6 flex items-center justify-between">
              <div>
                <h3 className="text-xl font-black text-slate-800">
                  {editingParam ? '编辑透中监测' : '新增透中监测'}
                </h3>
                <div className="mt-1 text-sm text-slate-400">{patient.name}</div>
              </div>
              <button
                type="button"
                onClick={() => {
                  setModalOpen(false)
                  setEditingParam(null)
                  setForm(mapParamToForm(null))
                }}
                className="text-sm font-semibold text-slate-500"
              >
                关闭
              </button>
            </div>

            <div className="grid grid-cols-1 gap-4 md:grid-cols-4">
              <label className="block md:col-span-2">
                <div className="mb-2 text-[11px] font-semibold uppercase tracking-wide text-slate-400">记录时间</div>
                <input
                  type="datetime-local"
                  value={form.recordTime}
                  onChange={(e) => updateField('recordTime', e.target.value)}
                  className="h-11 w-full rounded-2xl border border-slate-200 px-4 text-sm font-semibold text-slate-700 outline-none transition focus:border-blue-400"
                />
              </label>
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
              <label className="block md:col-span-4">
                <div className="mb-2 text-[11px] font-semibold uppercase tracking-wide text-slate-400">备注</div>
                <textarea
                  value={form.notes}
                  onChange={(e) => updateField('notes', e.target.value)}
                  rows={4}
                  className="w-full rounded-2xl border border-slate-200 px-4 py-3 text-sm font-semibold text-slate-700 outline-none transition focus:border-blue-400"
                />
              </label>
            </div>

            <div className="mt-6 flex justify-end gap-3">
              <button
                type="button"
                onClick={() => {
                  setModalOpen(false)
                  setEditingParam(null)
                  setForm(mapParamToForm(null))
                }}
                className="rounded-2xl border border-slate-200 px-5 py-3 text-sm font-semibold text-slate-600"
              >
                取消
              </button>
              <button
                type="button"
                onClick={() => void handleSave()}
                disabled={saving}
                className="rounded-2xl bg-blue-600 px-5 py-3 text-sm font-semibold text-white disabled:opacity-60"
              >
                {saving ? '保存中...' : editingParam ? '保存修改' : '保存监测'}
              </button>
            </div>
          </div>
        </div>
      ) : null}
    </div>
  )
}
