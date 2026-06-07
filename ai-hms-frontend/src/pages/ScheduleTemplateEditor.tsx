import { Button, Input, Select, message } from 'antd'
import { ArrowLeft, Save } from 'lucide-react'
import { useEffect, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { getErrorMessage } from '@/services/restClient'
import { restApi } from '@/services'
import type { ScheduleTemplateResponse, ScheduleTemplateItemRequest } from '@/services/restClient'

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

  const handleSave = async () => {
    if (!name.trim()) {
      message.warning('请输入模板名称')
      return
    }
    if (items.length === 0) {
      message.warning('模板项不能为空')
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
        <div className="flex items-center gap-4">
          <label className="text-sm text-slate-500 w-20">模板项:</label>
          <span className="text-sm text-slate-400">{items.length} 项</span>
        </div>
      </div>

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
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {!loading && items.length === 0 && (
        <div className="rounded-lg border border-slate-200 bg-white p-10 text-center">
          <p className="text-slate-500 mb-2">暂未添加模板项</p>
          <p className="text-xs text-slate-400">请在排班管理页面将患者添加到模板。</p>
        </div>
      )}
    </div>
  )
}
