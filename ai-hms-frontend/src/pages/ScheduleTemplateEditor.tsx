import { Button, Input, Select, message } from 'antd'
import { ArrowLeft, Plus, Save, X } from 'lucide-react'
import { useEffect, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { getErrorMessage } from '@/services/restClient'
import { restApi } from '@/services'
import type { ScheduleTemplateResponse, ScheduleTemplateItemRequest } from '@/services/restClient'

const FREQ_OPTIONS = [
  { value: 10, label: '每周1次' },
  { value: 20, label: '每周2次' },
  { value: 30, label: '每两周3次' },
  { value: 40, label: '每周4次' },
  { value: 90, label: '每周9次' },
]

const ZONE_OPTIONS = [
  { value: 'A', label: 'A区' },
  { value: 'B', label: 'B区' },
  { value: 'C', label: 'C区' },
]

const emptyItem = (): ScheduleTemplateItemRequest => ({
  patientId: 0,
  zoneTag: 'A',
  wardId: undefined,
  shiftId: undefined,
  freqPattern: 20,
  fixedHdBedId: undefined,
  fixedHdfBedId: undefined,
  hdfEnabled: false,
  hdfWeekday: undefined,
  hdfWeekParity: undefined,
})

export default function ScheduleTemplateEditor() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const editId = searchParams.get('id') ? Number(searchParams.get('id')) : undefined

  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [name, setName] = useState('')
  const [scope, setScope] = useState('A')
  const [wardId, setWardId] = useState<number | undefined>(undefined)
  const [items, setItems] = useState<ScheduleTemplateItemRequest[]>([])
  const [showAddForm, setShowAddForm] = useState(false)
  const [newItem, setNewItem] = useState<ScheduleTemplateItemRequest>(emptyItem())

  const loadTemplate = async () => {
    if (!editId) return
    setLoading(true)
    try {
      const data = await restApi.listScheduleTemplates()
      const found = data.find((t: ScheduleTemplateResponse) => t.template.id === editId)
      if (found) {
        setName(found.template.name)
        setScope(found.template.scope || 'A')
        setWardId(found.template.wardId ?? undefined)
        setItems(found.items.map((it) => ({
          patientId: it.patientId,
          zoneTag: it.zoneTag,
          wardId: it.wardId,
          shiftId: it.shiftId,
          freqPattern: it.freqPattern,
          fixedHdBedId: it.fixedHdBedId,
          fixedHdfBedId: it.fixedHdfBedId,
          hdfEnabled: it.hdfEnabled,
          hdfWeekday: it.hdfWeekday,
          hdfWeekParity: it.hdfWeekParity,
        })))
      } else {
        message.error('模板不存在')
        navigate('/schedule-templates')
      }
    } catch (e) {
      message.error(getErrorMessage(e))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { void loadTemplate() }, [editId]) // eslint-disable-line react-hooks/exhaustive-deps

  const handleAddItem = () => {
    if (!newItem.patientId || newItem.patientId <= 0) {
      message.warning('请输入患者ID')
      return
    }
    setItems((prev) => [...prev, { ...newItem }])
    setNewItem(emptyItem())
    setShowAddForm(false)
  }

  const handleRemoveItem = (idx: number) => {
    setItems((prev) => prev.filter((_, i) => i !== idx))
  }

  const handleSave = async () => {
    if (!name.trim()) {
      message.warning('请输入模板名称')
      return
    }
    if (items.length === 0) {
      message.warning('请至少添加一个模板项')
      return
    }
    try {
      setSaving(true)
      await restApi.saveScheduleTemplate({
        id: editId,
        name: name.trim(),
        scope,
        wardId: wardId ?? null,
        items,
      })
      message.success('模板已保存')
      navigate('/schedule-templates')
    } catch (error) {
      message.error(getErrorMessage(error))
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="max-w-[1400px] mx-auto pb-10">
      <div className="flex items-center justify-between gap-4 mb-6">
        <div className="flex items-center gap-4">
          <button onClick={() => navigate('/schedule-templates')} className="p-2 rounded-lg hover:bg-slate-100 text-slate-500 transition-colors">
            <ArrowLeft size={18} />
          </button>
          <h2 className="text-h2 font-bold text-foreground">{editId ? '编辑排班模板' : '新建排班模板'}</h2>
        </div>
        <Button onClick={() => void handleSave()} loading={saving} icon={<Save size={15} />} type="primary">{saving ? '保存中...' : '保存模板'}</Button>
      </div>

      <div className="rounded-lg border border-slate-200 bg-white p-6 mb-6 space-y-4">
        <div className="flex items-center gap-4">
          <label className="text-sm text-slate-500 w-20">模板名称:</label>
          <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="输入模板名称" className="max-w-[300px]" />
        </div>
        <div className="flex items-center gap-4">
          <label className="text-sm text-slate-500 w-20">范围:</label>
          <Select value={scope} onChange={(v) => setScope(v)} options={[
            { value: 'ALL', label: '全局' },
            { value: 'A', label: 'A区' },
            { value: 'B', label: 'B区' },
            { value: 'C', label: 'C区' },
          ]} className="w-32" />
        </div>
        <div className="flex items-center gap-4">
          <label className="text-sm text-slate-500 w-20">病区ID:</label>
          <Input type="number" value={wardId ?? ''} onChange={(e) => setWardId(e.target.value ? Number(e.target.value) : undefined)} placeholder="全局留空" className="max-w-[200px]" />
        </div>
      </div>

      {/* 添加模板项按钮 */}
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-sm font-bold text-slate-700">模板项 ({items.length})</h3>
        <Button
          size="small"
          icon={<Plus size={14} />}
          onClick={() => { setNewItem(emptyItem()); setShowAddForm(true) }}
        >
          添加模板项
        </Button>
      </div>

      {/* 添加模板项表单 */}
      {showAddForm && (
        <div className="rounded-lg border border-teal-200 bg-teal-50 p-4 mb-4">
          <div className="flex items-center justify-between mb-3">
            <span className="text-sm font-bold text-teal-700">新模板项</span>
            <button onClick={() => setShowAddForm(false)} className="text-slate-400 hover:text-slate-600"><X size={16} /></button>
          </div>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
            <div>
              <label className="text-xs text-slate-500">患者ID *</label>
              <Input type="number" value={newItem.patientId || ''} onChange={(e) => setNewItem({ ...newItem, patientId: Number(e.target.value) || 0 })} size="small" />
            </div>
            <div>
              <label className="text-xs text-slate-500">分区</label>
              <Select value={newItem.zoneTag} onChange={(v) => setNewItem({ ...newItem, zoneTag: v })} options={ZONE_OPTIONS} size="small" className="w-full" />
            </div>
            <div>
              <label className="text-xs text-slate-500">病区ID</label>
              <Input type="number" value={newItem.wardId ?? ''} onChange={(e) => setNewItem({ ...newItem, wardId: e.target.value ? Number(e.target.value) : undefined })} size="small" />
            </div>
            <div>
              <label className="text-xs text-slate-500">班次ID</label>
              <Input type="number" value={newItem.shiftId ?? ''} onChange={(e) => setNewItem({ ...newItem, shiftId: e.target.value ? Number(e.target.value) : undefined })} size="small" />
            </div>
            <div>
              <label className="text-xs text-slate-500">频率</label>
              <Select value={newItem.freqPattern || 20} onChange={(v) => setNewItem({ ...newItem, freqPattern: v })} options={FREQ_OPTIONS} size="small" className="w-full" />
            </div>
            <div>
              <label className="text-xs text-slate-500">固定HD床位</label>
              <Input type="number" value={newItem.fixedHdBedId ?? ''} onChange={(e) => setNewItem({ ...newItem, fixedHdBedId: e.target.value ? Number(e.target.value) : undefined })} size="small" />
            </div>
            <div>
              <label className="text-xs text-slate-500">HDF床位</label>
              <Input type="number" value={newItem.fixedHdfBedId ?? ''} onChange={(e) => setNewItem({ ...newItem, fixedHdfBedId: e.target.value ? Number(e.target.value) : undefined })} size="small" />
            </div>
            <div className="flex items-end gap-2">
              <label className="flex items-center gap-1 text-xs cursor-pointer">
                <input type="checkbox" checked={newItem.hdfEnabled} onChange={(e) => setNewItem({ ...newItem, hdfEnabled: e.target.checked })} />
                HDF
              </label>
            </div>
          </div>
          <div className="mt-3 flex justify-end">
            <Button size="small" type="primary" icon={<Plus size={14} />} onClick={handleAddItem}>确认添加</Button>
          </div>
        </div>
      )}

      {items.length > 0 && (
        <div className="overflow-x-auto rounded-lg border border-slate-200 bg-white shadow-sm">
          <table className="w-full min-w-[900px] text-left">
            <thead className="bg-slate-50 text-xs text-slate-500">
              <tr>
                <th className="px-4 py-3">#</th>
                <th className="px-4 py-3">患者ID</th>
                <th className="px-4 py-3">分区</th>
                <th className="px-4 py-3">病区</th>
                <th className="px-4 py-3">班次</th>
                <th className="px-4 py-3">频率</th>
                <th className="px-4 py-3">固定HD床</th>
                <th className="px-4 py-3">HDF</th>
                <th className="px-4 py-3">操作</th>
              </tr>
            </thead>
            <tbody>
              {items.map((it, idx) => (
                <tr key={idx} className="border-t border-slate-100 text-sm">
                  <td className="px-4 py-3 text-slate-500">{idx + 1}</td>
                  <td className="px-4 py-3 font-semibold text-slate-800">{it.patientId}</td>
                  <td className="px-4 py-3 text-slate-600">{it.zoneTag}</td>
                  <td className="px-4 py-3 text-slate-600">{it.wardId ?? '--'}</td>
                  <td className="px-4 py-3 text-slate-600">{it.shiftId ?? '--'}</td>
                  <td className="px-4 py-3 text-slate-600">{it.freqPattern || 10}</td>
                  <td className="px-4 py-3 text-slate-600">{it.fixedHdBedId ?? '--'}</td>
                  <td className="px-4 py-3 text-slate-600">{it.hdfEnabled ? `是(床${it.fixedHdfBedId ?? '--'})` : '否'}</td>
                  <td className="px-4 py-3">
                    <Button size="small" danger icon={<X size={12} />} onClick={() => handleRemoveItem(idx)}>删除</Button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {!loading && items.length === 0 && (
        <div className="rounded-lg border border-slate-200 bg-white p-10 text-center">
          <p className="text-slate-500 mb-2">暂未添加模板项</p>
          <p className="text-xs text-slate-400">点击上方"添加模板项"按钮添加患者排班配置。</p>
        </div>
      )}
    </div>
  )
}
