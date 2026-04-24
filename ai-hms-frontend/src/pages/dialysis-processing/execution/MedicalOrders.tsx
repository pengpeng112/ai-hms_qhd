import { message } from 'antd'
import { Edit3, PauseCircle, Plus } from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'
import { orderApi } from '@/services/orderApi'
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
  priority: '',
  startTime: '',
  endTime: '',
  notes: '',
}

function formatDateTime(value?: string) {
  if (!value) return '-'
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
  return parts.join(' / ') || '-'
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
    priority: order.priority || '',
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

function FormField({
  label,
  value,
  onChange,
  type = 'text',
}: {
  label: string
  value: string
  onChange: (value: string) => void
  type?: 'text' | 'datetime-local'
}) {
  return (
    <label className="block">
      <div className="mb-2 text-[11px] font-semibold uppercase tracking-wide text-slate-400">{label}</div>
      <input
        type={type}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="h-11 w-full rounded-2xl border border-slate-200 px-4 text-sm font-semibold text-slate-700 outline-none transition focus:border-blue-400"
      />
    </label>
  )
}

export default function MedicalOrders({ patient }: Props) {
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [orders, setOrders] = useState<Order[]>([])
  const [modalOpen, setModalOpen] = useState(false)
  const [editingOrder, setEditingOrder] = useState<Order | null>(null)
  const [form, setForm] = useState<OrderFormState>(EMPTY_FORM)

  useEffect(() => {
    const loadOrders = async () => {
      setLoading(true)
      try {
        const data = await orderApi.list(patient.id, { includeExpired: false })
        setOrders(data)
      } catch (error) {
        console.error('[MedicalOrders] load failed', error)
        message.error('透析医嘱加载失败')
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

  const openCreate = () => {
    setEditingOrder(null)
    setForm(EMPTY_FORM)
    setModalOpen(true)
  }

  const openEdit = (order: Order) => {
    setEditingOrder(order)
    setForm(mapOrderToForm(order))
    setModalOpen(true)
  }

  const handleSave = async () => {
    if (!form.content.trim() && !form.name.trim()) {
      message.warning('请至少填写医嘱内容或名称')
      return
    }

    try {
      setSaving(true)
      if (editingOrder) {
        const updated = await orderApi.update(
          patient.id,
          editingOrder.id,
          buildPayload(form) as OrderUpdateRequest
        )
        setOrders((items) => items.map((item) => (item.id === updated.id ? updated : item)))
        message.success('医嘱已更新')
      } else {
        const created = await orderApi.create(
          patient.id,
          buildPayload(form) as OrderCreateRequest
        )
        setOrders((items) => [created, ...items])
        message.success('医嘱已新增')
      }
      setModalOpen(false)
      setEditingOrder(null)
      setForm(EMPTY_FORM)
    } catch (error) {
      console.error('[MedicalOrders] save failed', error)
      message.error(editingOrder ? '医嘱更新失败' : '医嘱新增失败')
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
      message.error('医嘱停用失败')
    }
  }

  return (
    <div className="space-y-6 pb-8">
      <div className="rounded-3xl border border-slate-200 bg-white px-6 py-5 shadow-sm">
        <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
          <div>
            <div className="text-xs font-semibold uppercase tracking-wide text-slate-400">当前患者</div>
            <div className="mt-1 text-xl font-black text-slate-800">{patient.name}</div>
            <div className="mt-2 flex flex-wrap gap-2 text-xs text-slate-500">
              <span className="rounded-full bg-slate-100 px-3 py-1">长期医嘱 {stats.longTerm}</span>
              <span className="rounded-full bg-slate-100 px-3 py-1">临时医嘱 {stats.temporary}</span>
              <span className="rounded-full bg-blue-50 px-3 py-1 text-blue-700">有效医嘱 {stats.active}</span>
            </div>
          </div>
          <button
            type="button"
            onClick={openCreate}
            className="inline-flex items-center justify-center gap-2 rounded-2xl bg-blue-600 px-5 py-3 text-sm font-semibold text-white shadow-sm"
          >
            <Plus size={16} />
            新增透析医嘱
          </button>
        </div>
      </div>

      <section className="overflow-hidden rounded-3xl border border-slate-200 bg-white shadow-sm">
        {loading ? (
          <div className="p-10 text-center text-slate-500">正在加载透析医嘱...</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full min-w-[1180px] text-left">
              <thead className="bg-slate-50 text-xs uppercase tracking-wide text-slate-500">
                <tr>
                  <th className="px-6 py-3">类型</th>
                  <th className="px-6 py-3">医嘱内容</th>
                  <th className="px-6 py-3">给药/执行</th>
                  <th className="px-6 py-3">频次</th>
                  <th className="px-6 py-3">开立医生</th>
                  <th className="px-6 py-3">开始时间</th>
                  <th className="px-6 py-3">状态</th>
                  <th className="px-6 py-3">备注</th>
                  <th className="px-6 py-3 text-right">操作</th>
                </tr>
              </thead>
              <tbody>
                {sortedOrders.length > 0 ? (
                  sortedOrders.map((order) => (
                    <tr key={order.id} className="border-t border-slate-100 text-sm">
                      <td className="px-6 py-4">
                        <span
                          className={`inline-flex rounded-md px-2 py-1 text-xs font-semibold ${
                            order.type === '长期'
                              ? 'bg-indigo-100 text-indigo-700'
                              : 'bg-amber-100 text-amber-700'
                          }`}
                        >
                          {order.type}
                        </span>
                      </td>
                      <td className="px-6 py-4">
                        <div className="font-semibold text-slate-800">{buildOrderSummary(order)}</div>
                        <div className="mt-1 text-xs text-slate-400">
                          {order.category || '未分类'}
                          {order.spec ? ` / ${order.spec}` : ''}
                        </div>
                      </td>
                      <td className="px-6 py-4 text-slate-600">
                        <div>{order.route || '-'}</div>
                        <div className="mt-1 text-xs text-slate-400">{order.execTiming || order.timing || '-'}</div>
                      </td>
                      <td className="px-6 py-4 text-slate-600">{order.frequency || '-'}</td>
                      <td className="px-6 py-4 text-slate-600">{order.doctorName || '-'}</td>
                      <td className="px-6 py-4 text-slate-600">{formatDateTime(order.startTime || order.createdAt)}</td>
                      <td className="px-6 py-4">
                        <span className="rounded-full bg-slate-100 px-3 py-1 text-xs font-semibold text-slate-700">
                          {order.status || '-'}
                        </span>
                      </td>
                      <td className="px-6 py-4 text-slate-500">{order.notes || '-'}</td>
                      <td className="px-6 py-4">
                        <div className="flex items-center justify-end gap-2">
                          <button
                            type="button"
                            onClick={() => openEdit(order)}
                            className="flex h-9 w-9 items-center justify-center rounded-xl border border-slate-200 text-slate-500 transition hover:bg-slate-50"
                          >
                            <Edit3 size={14} />
                          </button>
                          <button
                            type="button"
                            onClick={() => void handleStop(order)}
                            disabled={order.status === '已停止'}
                            className="flex h-9 w-9 items-center justify-center rounded-xl border border-slate-200 text-slate-500 transition hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-50"
                          >
                            <PauseCircle size={14} />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))
                ) : (
                  <tr>
                    <td colSpan={9} className="px-6 py-10 text-center text-sm text-slate-400">
                      当前患者暂无透析医嘱
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        )}
      </section>

      {modalOpen ? (
        <div className="fixed inset-0 z-[100] flex items-center justify-center bg-slate-900/50 px-4 backdrop-blur-sm">
          <div className="w-full max-w-4xl rounded-[32px] border border-slate-200 bg-white p-8 shadow-2xl">
            <div className="mb-6 flex items-center justify-between">
              <div>
                <h3 className="text-xl font-black text-slate-800">
                  {editingOrder ? '编辑透析医嘱' : '新增透析医嘱'}
                </h3>
                <div className="mt-1 text-sm text-slate-400">{patient.name}</div>
              </div>
              <button
                type="button"
                onClick={() => {
                  setModalOpen(false)
                  setEditingOrder(null)
                  setForm(EMPTY_FORM)
                }}
                className="text-sm font-semibold text-slate-500"
              >
                关闭
              </button>
            </div>

            <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
              <label className="block">
                <div className="mb-2 text-[11px] font-semibold uppercase tracking-wide text-slate-400">医嘱类型</div>
                <select
                  value={form.type}
                  onChange={(e) => updateField('type', e.target.value as OrderFormState['type'])}
                  className="h-11 w-full rounded-2xl border border-slate-200 px-4 text-sm font-semibold text-slate-700 outline-none transition focus:border-blue-400"
                >
                  <option value="长期">长期</option>
                  <option value="临时">临时</option>
                </select>
              </label>
              <FormField label="医嘱分类" value={form.category} onChange={(value) => updateField('category', value)} />
              <FormField label="名称" value={form.name} onChange={(value) => updateField('name', value)} />
              <FormField label="规格" value={form.spec} onChange={(value) => updateField('spec', value)} />
              <FormField label="剂量" value={form.dose} onChange={(value) => updateField('dose', value)} />
              <FormField label="单位" value={form.unit} onChange={(value) => updateField('unit', value)} />
              <FormField label="用药途径" value={form.route} onChange={(value) => updateField('route', value)} />
              <FormField label="执行时机" value={form.execTiming} onChange={(value) => updateField('execTiming', value)} />
              <FormField label="频次" value={form.frequency} onChange={(value) => updateField('frequency', value)} />
              <FormField label="优先级" value={form.priority} onChange={(value) => updateField('priority', value)} />
              <FormField label="开始时间" type="datetime-local" value={form.startTime} onChange={(value) => updateField('startTime', value)} />
              <FormField label="结束时间" type="datetime-local" value={form.endTime} onChange={(value) => updateField('endTime', value)} />
              <FormField label="开立时机" value={form.timing} onChange={(value) => updateField('timing', value)} />
              <label className="block md:col-span-2">
                <div className="mb-2 text-[11px] font-semibold uppercase tracking-wide text-slate-400">医嘱内容</div>
                <input
                  value={form.content}
                  onChange={(e) => updateField('content', e.target.value)}
                  className="h-11 w-full rounded-2xl border border-slate-200 px-4 text-sm font-semibold text-slate-700 outline-none transition focus:border-blue-400"
                />
              </label>
              <label className="block md:col-span-2">
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
                  setEditingOrder(null)
                  setForm(EMPTY_FORM)
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
                {saving ? '保存中...' : editingOrder ? '保存修改' : '保存医嘱'}
              </button>
            </div>
          </div>
        </div>
      ) : null}
    </div>
  )
}
