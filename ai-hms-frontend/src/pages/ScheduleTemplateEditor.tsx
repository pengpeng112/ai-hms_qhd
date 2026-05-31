import { message } from 'antd'
import { ArrowLeft, RefreshCw, Save } from 'lucide-react'
import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { getErrorMessage } from '@/services/restClient'
import { restApi } from '@/services'
import type { RestScheduleWeekShift } from '@/services/restClient'

export default function ScheduleTemplateEditor() {
  const navigate = useNavigate()
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [wardId, setWardId] = useState<number | undefined>(undefined)
  const [entries, setEntries] = useState<RestScheduleWeekShift[]>([])

  const loadEntries = async () => {
    setLoading(true)
    try {
      const res = await restApi.listScheduleTemplateEntries(wardId)
      setEntries(Array.isArray(res) ? res : [])
    } catch (error) {
      console.error('[ScheduleTemplateEditor] load failed', error)
      message.error(getErrorMessage(error))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { void loadEntries() }, [wardId])

  const handleSave = async () => {
    if (entries.length === 0) {
      message.warning('无可保存的模板数据')
      return
    }
    try {
      setSaving(true)
      await restApi.saveScheduleTemplate(
        entries.map((e) => ({
          patientId: e.patientId,
          shiftId: e.shiftId,
          wardId: e.wardId,
          bedId: e.bedId,
          patientPlanId: e.patientPlanId,
          weekday: 1,
        }))
      )
      message.success('模板已保存')
    } catch (error) {
      console.error('[ScheduleTemplateEditor] save failed', error)
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
          <h2 className="text-h2 font-bold text-foreground">编辑排班模板</h2>
        </div>
        <div className="flex items-center gap-3">
          <button onClick={loadEntries} disabled={loading} className="inline-flex items-center gap-2 rounded-lg border border-slate-200 bg-white px-4 py-2 text-sm font-semibold text-slate-700 disabled:opacity-50">
            <RefreshCw size={15} className={loading ? 'animate-spin' : ''} />刷新
          </button>
          <button onClick={() => void handleSave()} disabled={saving || entries.length === 0} className="inline-flex items-center gap-2 rounded-lg bg-blue-600 px-5 py-2 text-sm font-bold text-white disabled:opacity-50">
            <Save size={15} />{saving ? '保存中...' : '保存模板'}
          </button>
        </div>
      </div>

      <div className="mb-4 flex items-center gap-4">
        <label className="flex items-center gap-2 text-sm text-slate-500">
          病区过滤:
          <input type="number" value={wardId ?? ''} onChange={(e) => setWardId(e.target.value ? Number(e.target.value) : undefined)} placeholder="输入病区ID..." className="h-9 w-32 rounded-lg border border-slate-200 px-3 text-sm outline-none focus:border-blue-400" />
        </label>
        <span className="text-xs text-slate-400">共 {entries.length} 条排班条目</span>
      </div>

      {loading ? (
        <div className="rounded-lg border border-slate-200 bg-white p-10 text-center text-slate-500">正在加载模板...</div>
      ) : entries.length === 0 ? (
        <div className="rounded-lg border border-slate-200 bg-white p-10 text-center">
          <p className="text-slate-500 mb-2">暂未加载模板数据</p>
          <p className="text-xs text-slate-400">输入病区ID后点击刷新获取排班模板条目。</p>
        </div>
      ) : (
        <div className="overflow-x-auto rounded-lg border border-slate-200 bg-white shadow-sm">
          <table className="w-full min-w-[900px] text-left">
            <thead className="bg-slate-50 text-xs text-slate-500">
              <tr>
                <th className="px-4 py-3">#</th>
                <th className="px-4 py-3">患者</th>
                <th className="px-4 py-3">班次</th>
                <th className="px-4 py-3">病区</th>
                <th className="px-4 py-3">床位</th>
                <th className="px-4 py-3">透析模式</th>
                <th className="px-4 py-3">状态</th>
              </tr>
            </thead>
            <tbody>
              {entries.map((entry, idx) => (
                <tr key={entry.id || idx} className="border-t border-slate-100 text-sm">
                  <td className="px-4 py-3 text-slate-500">{idx + 1}</td>
                  <td className="px-4 py-3 font-semibold text-slate-800">{entry.patientName || `患者${entry.patientId}`}</td>
                  <td className="px-4 py-3 text-slate-600">班次 {entry.shiftId}</td>
                  <td className="px-4 py-3 text-slate-600">病区 {entry.wardId}</td>
                  <td className="px-4 py-3 text-slate-600">{entry.bedName || `床${entry.bedId}`}</td>
                  <td className="px-4 py-3 text-slate-600">{entry.dialysisMode || '--'}</td>
                  <td className="px-4 py-3"><span className="rounded-full bg-emerald-50 px-2 py-1 text-xs text-emerald-700">{entry.statusName || '待排'}</span></td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      <div className="mt-4 rounded-lg border border-slate-200 bg-slate-50 p-4 text-xs text-slate-500">
        <p className="font-semibold mb-1">使用说明</p>
        <p>数据来源：旧系统排班数据，可在此预览和重新保存为模板。</p>
      </div>
    </div>
  )
}
