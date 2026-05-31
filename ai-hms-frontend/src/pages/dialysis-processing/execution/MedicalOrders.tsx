import { message } from 'antd'
import { CheckCircle2, Edit3, FileText, PauseCircle, Plus, Search, X, Zap } from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'
import { orderApi } from '@/services/orderApi'
import { getErrorMessage } from '@/services/restClient'
import type {
  Order,
  OrderCreateRequest,
  OrderUpdateRequest,
} from '@/services/orderApi'
import type { Patient } from '../types'

interface Props {
  patient: Patient
}

interface OrderFormState {
  type: '长期' | '临时'
  category: string
  name: string
  content: string
  dose: string
  unit: string
  route: string
  timing: string
  execTiming: string
  spec: string
  frequency: string
  priority: string
  startTime: string
  endTime: string
  notes: string
}

const EMPTY_FORM: OrderFormState = {
  type: '长期',
  category: '',
  name: '',
  content: '',
  dose: '',
  unit: '',
  route: '',
  timing: '',
  execTiming: '',
  spec: '',
  frequency: '',
  priority: '普通',
  startTime: '',
  endTime: '',
  notes: '',
}

function formatDateTime(value?: string) {
  if (!value) return '--'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString('zh-CN', { hour12: false })
}

function toDateTimeLocal(value?: string) {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  const local = new Date(date.getTime() - date.getTimezoneOffset() * 60000)
  return local.toISOString().slice(0, 16)
}

function toIsoOrUndefined(value: string) {
  if (!value.trim()) return undefined
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? undefined : date.toISOString()
}

function toOptionalText(value: string) {
  const trimmed = value.trim()
  return trimmed ? trimmed : undefined
}

function buildOrderSummary(order: Order) {
  if (order.content?.trim()) return order.content.trim()
  const parts = [order.name, order.dose && order.unit ? `${order.dose}${order.unit}` : order.dose, order.route]
    .map((item) => item?.trim())
    .filter(Boolean)
  return parts.join(' / ') || '--'
}

function buildUsage(order: Order) {
  const parts = [order.route, order.dose && order.unit ? `${order.dose}${order.unit}` : order.dose, order.frequency]
    .map((item) => item?.trim())
    .filter(Boolean)
  return parts.join('，') || order.timing || '--'
}

function mapOrderToForm(order: Order | null): OrderFormState {
  if (!order) return EMPTY_FORM
  return {
    type: order.type,
    category: order.category || '',
    name: order.name || '',
    content: order.content || '',
    dose: order.dose || '',
    unit: order.unit || '',
    route: order.route || '',
    timing: order.timing || '',
    execTiming: order.execTiming || '',
    spec: order.spec || '',
    frequency: order.frequency || '',
    priority: order.priority || '普通',
    startTime: toDateTimeLocal(order.startTime),
    endTime: toDateTimeLocal(order.endTime),
    notes: order.notes || '',
  }
}

function buildPayload(form: OrderFormState): OrderCreateRequest | OrderUpdateRequest {
  return {
    type: form.type,
    category: toOptionalText(form.category),
    name: toOptionalText(form.name),
    content: toOptionalText(form.content),
    dose: toOptionalText(form.dose),
    unit: toOptionalText(form.unit),
    route: toOptionalText(form.route),
    timing: toOptionalText(form.timing),
    execTiming: toOptionalText(form.execTiming),
    spec: toOptionalText(form.spec),
    frequency: toOptionalText(form.frequency),
    priority: toOptionalText(form.priority),
    startTime: toIsoOrUndefined(form.startTime),
    endTime: toIsoOrUndefined(form.endTime),
    notes: toOptionalText(form.notes),
  }
}

function TextField({
  label,
  value,
  onChange,
  type = 'text',
  placeholder,
}: {
  label: string
  value: string
  onChange: (value: string) => void
  type?: 'text' | 'datetime-local'
  placeholder?: string
}) {
  return (
    <label className="block min-w-0">
      <span className="mb-2 block text-xs font-semibold text-slate-400">{label}</span>
      <input
        type={type}
        value={value}
        placeholder={placeholder}
        onChange={(e) => onChange(e.target.value)}
        className="h-11 w-full rounded-lg border border-slate-200 px-4 text-sm font-semibold text-slate-700 outline-none transition focus:border-blue-400"
      />
    </label>
  )
}

function SelectField({
  label,
  value,
  options,
  onChange,
}: {
  label: string
  value: string
  options: string[]
  onChange: (value: string) => void
}) {
  return (
    <label className="block min-w-0">
      <span className="mb-2 block text-xs font-semibold text-slate-400">{label}</span>
      <select
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="h-11 w-full rounded-lg border border-slate-200 bg-white px-4 text-sm font-semibold text-slate-700 outline-none transition focus:border-blue-400"
      >
        {options.map((item) => (
          <option key={item} value={item}>{item}</option>
        ))}
      </select>
    </label>
  )
}

function StatusIcon({ enabled }: { enabled: boolean }) {
  return enabled ? <CheckCircle2 size={16} className="text-emerald-500" /> : <span className="text-slate-300">--</span>
}

export default function MedicalOrders({ patient }: Props) {
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [orders, setOrders] = useState<Order[]>([])
  const [modalOpen, setModalOpen] = useState(false)
  const [editingOrder, setEditingOrder] = useState<Order | null>(null)
  const [form, setForm] = useState<OrderFormState>(EMPTY_FORM)
  const [sourceMode, setSourceMode] = useState<'direct' | 'template'>('direct')

  useEffect(() => {
    const loadOrders = async () => {
      setLoading(true)
      try {
        const data = await orderApi.list(patient.id, { includeExpired: false })
        setOrders(data)
      } catch (error) {
        console.error('[MedicalOrders] load failed', error)
        message.error(getErrorMessage(error))
      } finally {
        setLoading(false)
      }
    }
    void loadOrders()
  }, [patient.id])

  const sortedOrders = useMemo(
    () =>
      [...orders].sort((a, b) => {
        const left = new Date(b.startTime || b.createdAt).getTime()
        const right = new Date(a.startTime || a.createdAt).getTime()
        return left - right
      }),
    [orders]
  )

  const stats = useMemo(() => {
    const longTerm = orders.filter((item) => item.type === '长期').length
    const temporary = orders.filter((item) => item.type === '临时').length
    const active = orders.filter((item) => item.status !== '已停止').length
    return { longTerm, temporary, active }
  }, [orders])

  const updateField = (key: keyof OrderFormState, value: string) => {
    setForm((current) => ({ ...current, [key]: value }))
  }

  const closeModal = () => {
    setModalOpen(false)
    setEditingOrder(null)
    setForm(EMPTY_FORM)
    setSourceMode('direct')
  }

  const openCreate = () => {
    setEditingOrder(null)
    setForm(EMPTY_FORM)
    setSourceMode('direct')
    setModalOpen(true)
  }

  const openEdit = (order: Order) => {
    setEditingOrder(order)
    setForm(mapOrderToForm(order))
    setSourceMode('direct')
    setModalOpen(true)
  }

  const handleSave = async () => {
    if (sourceMode === 'template') {
      message.info('功能待后端接口就绪')
      return
    }
    if (!form.content.trim() && !form.name.trim()) {
      message.warning('请至少填写医嘱内容或名称')
      return
    }

    try {
      setSaving(true)
      if (editingOrder) {
        const updated = await orderApi.update(patient.id, editingOrder.id, buildPayload(form) as OrderUpdateRequest)
        setOrders((items) => items.map((item) => (item.id === updated.id ? updated : item)))
        message.success('医嘱已更新')
      } else {
        const created = await orderApi.create(patient.id, buildPayload(form) as OrderCreateRequest)
        setOrders((items) => [created, ...items])
        message.success('医嘱已新增')
      }
      closeModal()
    } catch (error) {
      console.error('[MedicalOrders] save failed', error)
      message.error(getErrorMessage(error))
    } finally {
      setSaving(false)
    }
  }

  const handleStop = async (order: Order) => {
    try {
      const updatedList = await orderApi.stop(patient.id, order.id, '前端停嘱')
      setOrders(updatedList)
      message.success('医嘱已停用')
    } catch (error) {
      console.error('[MedicalOrders] stop failed', error)
      message.error(getErrorMessage(error))
    }
  }

  return (
    <div className="space-y-5 pb-8">
      <div className="rounded-lg border border-slate-200 bg-white px-5 py-4 shadow-sm">
        <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
          <div className="flex items-center gap-3">
            <FileText size={17} className="text-blue-600" />
            <span className="text-sm font-black text-slate-800">透析医嘱</span>
            <span className="rounded-full bg-slate-100 px-3 py-1 text-xs font-bold text-slate-600">{orders.length} 条</span>
            <span className="rounded-full bg-blue-50 px-3 py-1 text-xs font-bold text-blue-700">长期 {stats.longTerm}</span>
            <span className="rounded-full bg-amber-50 px-3 py-1 text-xs font-bold text-amber-700">临时 {stats.temporary}</span>
            <span className="rounded-full bg-emerald-50 px-3 py-1 text-xs font-bold text-emerald-700">有效 {stats.active}</span>
            <span className="text-xs text-slate-400">{patient.name} · ID: {patient.id}</span>
          </div>
          <button type="button" onClick={openCreate} className="inline-flex items-center justify-center gap-2 rounded-lg bg-blue-600 px-5 py-2.5 text-sm font-bold text-white shadow-sm">
            <Plus size={16} />
            新增透析医嘱
          </button>
        </div>
      </div>

      <section className="overflow-hidden rounded-lg border border-slate-200 bg-white shadow-sm">
        <div className="flex items-center justify-between border-b border-slate-100 px-5 py-3">
          <div className="flex items-center gap-2 text-sm font-black text-slate-800">
            <FileText size={16} className="text-blue-600" />
            透析医嘱明细
          </div>
          <div className="flex items-center gap-2 text-xs font-semibold text-slate-400">
            <Search size={13} />
            横向滚动查看完整字段
          </div>
        </div>
        {loading ? (
          <div className="p-10 text-center text-slate-500">正在加载透析医嘱...</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full min-w-[1380px] text-left">
              <thead className="bg-slate-50 text-xs text-slate-500">
                <tr>
                  <th className="px-5 py-3">#</th>
                  <th className="px-5 py-3">类型</th>
                  <th className="px-5 py-3">医嘱内容</th>
                  <th className="px-5 py-3">使用描述</th>
                  <th className="px-5 py-3">医生/下嘱时间</th>
                  <th className="px-5 py-3">最近执行</th>
                  <th className="px-5 py-3">核对</th>
                  <th className="px-5 py-3">执行</th>
                  <th className="px-5 py-3">本周执行</th>
                  <th className="px-5 py-3">最后修改</th>
                  <th className="px-5 py-3 text-right">操作</th>
                </tr>
              </thead>
              <tbody>
                {sortedOrders.length > 0 ? (
                  sortedOrders.map((order, index) => (
                    <tr key={order.id} className="border-t border-slate-100 text-sm">
                      <td className="px-5 py-4 text-slate-500">{index + 1}</td>
                      <td className="px-5 py-4">
                        <span className={`rounded-md px-2 py-1 text-xs font-bold ${order.type === '长期' ? 'bg-blue-50 text-blue-700' : 'bg-amber-50 text-amber-700'}`}>
                          {order.type}
                        </span>
                      </td>
                      <td className="px-5 py-4">
                        <div className="font-black text-slate-900">{buildOrderSummary(order)}</div>
                        <div className="mt-1 text-xs text-blue-500">{order.category || '--'}</div>
                      </td>
                      <td className="px-5 py-4 text-slate-600">{buildUsage(order)}</td>
                      <td className="px-5 py-4">
                        <div className="font-bold text-slate-800">{order.doctorName || '--'}</div>
                        <div className="mt-1 text-xs text-slate-400">{formatDateTime(order.startTime || order.createdAt)}</div>
                      </td>
                      <td className="px-5 py-4 text-slate-500">{order.executedAt ? formatDateTime(order.executedAt) : '--'}</td>
                      <td className="px-5 py-4"><StatusIcon enabled={order.status === '已执行' || order.status === '执行中'} /></td>
                      <td className="px-5 py-4"><StatusIcon enabled={order.status === '已执行'} /></td>
                      <td className="px-5 py-4 text-slate-500">-- 次</td>
                      <td className="px-5 py-4 text-xs text-slate-500">{formatDateTime(order.updatedAt)}</td>
                      <td className="px-5 py-4">
                        <div className="flex items-center justify-end gap-2">
                          <button type="button" onClick={() => openEdit(order)} className="flex h-8 w-8 items-center justify-center rounded-lg border border-slate-200 text-slate-500 hover:bg-slate-50" aria-label="编辑医嘱">
                            <Edit3 size={14} />
                          </button>
                          <button type="button" onClick={() => void handleStop(order)} disabled={order.status === '已停止'} className="flex h-8 w-8 items-center justify-center rounded-lg border border-slate-200 text-slate-500 hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-50" aria-label="停用医嘱">
                            <PauseCircle size={14} />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))
                ) : (
                  <tr>
                    <td colSpan={11} className="px-6 py-10 text-center text-sm text-slate-400">当前患者暂无透析医嘱</td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        )}
      </section>

      <section className="overflow-hidden rounded-lg border border-slate-200 bg-white shadow-sm">
        <div className="flex items-center gap-2 border-b border-slate-100 px-5 py-3 text-sm font-black text-slate-800">
          <Zap size={16} className="text-amber-500" />
          抗凝剂
        </div>
        <div className="overflow-x-auto">
          <table className="w-full min-w-[900px] text-left">
            <thead className="bg-slate-50 text-xs text-slate-500">
              <tr><th className="px-6 py-3">类别</th><th className="px-6 py-3">名称</th><th className="px-6 py-3">剂量</th><th className="px-6 py-3">医生</th><th className="px-6 py-3">执行</th><th className="px-6 py-3">核对</th></tr>
            </thead>
            <tbody>
              <tr><td colSpan={6} className="px-6 py-10 text-center text-sm text-slate-400">暂无抗凝剂接口数据</td></tr>
            </tbody>
          </table>
        </div>
      </section>

      {modalOpen ? (
        <div className="fixed inset-0 z-[100] flex items-center justify-center bg-slate-900/55 px-4 backdrop-blur-sm">
          <div className="flex max-h-[88vh] w-full max-w-4xl flex-col overflow-hidden rounded-lg bg-white shadow-2xl">
            <div className="flex items-center justify-between px-8 py-6">
              <div className="inline-grid grid-cols-2 rounded-lg bg-slate-100 p-1 text-sm font-bold">
                {(['长期', '临时'] as const).map((item) => (
                  <button key={item} type="button" onClick={() => updateField('type', item)} className={`rounded-md px-8 py-2 ${form.type === item ? 'bg-white text-blue-600 shadow-sm' : 'text-slate-400'}`}>{item}医嘱</button>
                ))}
              </div>
              <button type="button" onClick={closeModal} className="text-slate-400 hover:text-slate-600" aria-label="关闭">
                <X size={20} />
              </button>
            </div>

            <div className="border-b border-slate-100 px-8">
              <div className="flex gap-8 text-sm font-black">
                <button type="button" onClick={() => setSourceMode('direct')} className={`border-b-2 px-0 py-3 ${sourceMode === 'direct' ? 'border-blue-600 text-blue-600' : 'border-transparent text-slate-400'}`}>直接新增录入</button>
                <button type="button" onClick={() => setSourceMode('template')} className={`border-b-2 px-0 py-3 ${sourceMode === 'template' ? 'border-blue-600 text-blue-600' : 'border-transparent text-slate-400'}`}>从模板组调取</button>
              </div>
            </div>

            <div className="flex-1 overflow-y-auto px-8 py-6">
              {sourceMode === 'template' ? (
                <div className="rounded-lg border border-dashed border-slate-300 bg-slate-50 p-12 text-center text-sm font-semibold text-slate-500">
                  功能待后端接口就绪
                </div>
              ) : (
                <div className="space-y-7">
                  <div className="grid grid-cols-[110px_1fr] gap-6">
                    <div className="pt-8 text-sm font-black text-slate-500">医嘱项目</div>
                    <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                      <TextField label="医嘱名称" value={form.name} onChange={(value) => updateField('name', value)} />
                      <TextField label="医嘱分类" value={form.category} onChange={(value) => updateField('category', value)} />
                      <TextField label="规格" value={form.spec} onChange={(value) => updateField('spec', value)} />
                      <SelectField label="优先级" value={form.priority} options={['普通', '紧急']} onChange={(value) => updateField('priority', value)} />
                    </div>
                  </div>

                  <div className="grid grid-cols-[110px_1fr] gap-6 border-t border-dashed border-slate-200 pt-6">
                    <div className="pt-8 text-sm font-black text-slate-500">执行细节</div>
                    <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                      <TextField label="单次剂量" value={form.dose} onChange={(value) => updateField('dose', value)} />
                      <TextField label="单位" value={form.unit} onChange={(value) => updateField('unit', value)} />
                      <SelectField label="给药 / 执行途径" value={form.route} options={['', '静脉推注', '口服', '外用', '管路给药']} onChange={(value) => updateField('route', value)} />
                      <SelectField label="执行频次" value={form.frequency} options={['', '每周三次', '每次透析', '每日一次', '必要时']} onChange={(value) => updateField('frequency', value)} />
                      <TextField label="执行时机" value={form.execTiming} placeholder="例如：透析结束前" onChange={(value) => updateField('execTiming', value)} />
                      <TextField label="开立时机" value={form.timing} placeholder="首、中、末" onChange={(value) => updateField('timing', value)} />
                      <TextField label="开始时间" type="datetime-local" value={form.startTime} onChange={(value) => updateField('startTime', value)} />
                      <TextField label="结束时间" type="datetime-local" value={form.endTime} onChange={(value) => updateField('endTime', value)} />
                    </div>
                  </div>

                  <div className="grid grid-cols-[110px_1fr] gap-6 border-t border-dashed border-slate-200 pt-6">
                    <div className="pt-8 text-sm font-black text-slate-500">医嘱内容</div>
                    <div className="space-y-4">
                      <TextField label="医嘱内容" value={form.content} onChange={(value) => updateField('content', value)} />
                      <label className="block">
                        <span className="mb-2 block text-xs font-semibold text-slate-400">备注</span>
                        <textarea value={form.notes} onChange={(e) => updateField('notes', e.target.value)} rows={4} className="w-full resize-none rounded-lg border border-slate-200 px-4 py-3 text-sm font-semibold text-slate-700 outline-none transition focus:border-blue-400" />
                      </label>
                    </div>
                  </div>
                </div>
              )}
            </div>

            <div className="flex items-center justify-between border-t border-slate-100 px-8 py-5">
              <div className="rounded-lg border border-slate-100 bg-slate-50 px-4 py-3 text-xs font-bold text-slate-500">
                拟开嘱医生：{editingOrder?.doctorName || '待签名'}
                <div className="mt-1 text-xs tracking-widest text-slate-400">WAITING FOR SIGNATURE</div>
              </div>
              <div className="flex gap-3">
                <button type="button" onClick={closeModal} className="rounded-lg border border-slate-200 bg-white px-8 py-3 text-sm font-semibold text-slate-500">取消返回</button>
                <button type="button" onClick={() => void handleSave()} disabled={saving || sourceMode === 'template'} className="rounded-lg bg-blue-600 px-8 py-3 text-sm font-bold text-white disabled:cursor-not-allowed disabled:opacity-60">
                  {saving ? '保存中...' : editingOrder ? '保存修改' : '保存医嘱'}
                </button>
              </div>
            </div>
          </div>
        </div>
      ) : null}
    </div>
  )
}
