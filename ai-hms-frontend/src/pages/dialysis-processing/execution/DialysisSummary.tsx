import { message } from 'antd'
import { CheckCircle2, Copy, FileText, NotebookPen, Save, ShieldCheck, TrendingUp, Waves } from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'
import { restApi } from '@/services'
import type { RestTreatment } from '@/services'
import { orderApi } from '@/services/orderApi'
import type { Order } from '@/services/orderApi'
import { getErrorMessage } from '@/services/restClient'
import type { Patient } from '../types'

interface Props {
  patient: Patient
  treatment: RestTreatment | null
  treatmentLoading?: boolean
  onTreatmentUpdated: (treatment: RestTreatment) => void
}

function toText(value?: string | number | null) {
  if (value === undefined || value === null || value === '') return '--'
  return String(value)
}

function formatDateTime(value?: string) {
  if (!value) return '--'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString('zh-CN', { hour12: false })
}

function parseBp(bp?: string) {
  return bp || '--'
}

function getLatestSymptom(treatment: RestTreatment | null) {
  const latest = treatment?.duringParams?.[0]
  return latest?.notes || '--'
}

function getSymptomItemValue(items: Array<{ code: string; value: string }> | undefined, code: string, fallback?: string) {
  const v = items?.find((item) => item.code === code)?.value
  return v || fallback || '未记录'
}

export default function DialysisSummary({ patient, treatment, treatmentLoading = false, onTreatmentUpdated }: Props) {
  const [orders, setOrders] = useState<Order[]>([])
  const [doctorSummary, setDoctorSummary] = useState('')
  const [treatmentSummary, setTreatmentSummary] = useState('')
  const [saving, setSaving] = useState(false)
  const [lastTreatmentId, setLastTreatmentId] = useState<number | null>(null)

  useEffect(() => {
    if (treatment && treatment.id !== lastTreatmentId) {
      setDoctorSummary(treatment.doctorSummary || '')
      setTreatmentSummary(treatment.treatmentSummary || treatment.notes || '')
      setLastTreatmentId(treatment.id)
    } else if (!treatment) {
      setDoctorSummary('')
      setTreatmentSummary('')
      setLastTreatmentId(null)
    }
  }, [treatment?.id, treatment, lastTreatmentId])

  useEffect(() => {
    const loadOrders = async () => {
      try {
        const data = await orderApi.list(patient.id, { includeExpired: false })
        setOrders(data)
      } catch (error) {
        console.error('[DialysisSummary] load orders failed', error)
        message.error(getErrorMessage(error))
      }
    }
    void loadOrders()
  }, [patient.id])

  const handleSaveSummary = async () => {
    if (!treatment) return
    try {
      setSaving(true)
      const updated = await restApi.updateTreatmentSummary(treatment.id, {
        doctorSummary: doctorSummary.trim(),
        treatmentSummary: treatmentSummary.trim(),
      })
      onTreatmentUpdated(updated.data)
      message.success('透析小结已保存')
    } catch (error) {
      console.error('[DialysisSummary] save failed', error)
      message.error(getErrorMessage(error))
    } finally {
      setSaving(false)
    }
  }

  const monitoringRows = useMemo(
    () => [...(treatment?.duringParams || [])].sort((a, b) => new Date(b.recordTime).getTime() - new Date(a.recordTime).getTime()),
    [treatment]
  )

  const handleCopyAutoSummary = async () => {
    try {
      await navigator.clipboard.writeText(autoSummary)
      message.success('自动汇总已复制')
    } catch {
      message.error('复制失败')
    }
  }

  const symItems = treatment?.afterSymptomItems
  const dialyzerCoag = getSymptomItemValue(symItems, 'dialyzer_coag')
  const lineACoag = getSymptomItemValue(symItems, 'line_a_coag')
  const lineVCoag = getSymptomItemValue(symItems, 'line_v_coag')
  const hasCoag = dialyzerCoag !== '未记录' || lineACoag !== '未记录' || lineVCoag !== '未记录'

  const autoSummary = useMemo(() => {
    const coagText = hasCoag
      ? `透析器${dialyzerCoag}、A端${lineACoag}、V端${lineVCoag}`
      : '未记录'
    return `患者于 ${formatDateTime(treatment?.startTime)} 开始 ${treatment?.treatmentType || patient.treatmentPlan || '--'} 治疗，透前体重 ${toText(treatment?.beforeSigns?.weight)}kg，干体重 ${toText(patient.dryWeight)}kg，治疗过程中生命体征${monitoringRows.length > 0 ? '已有监测记录' : '无完整监测趋势'}，实际超滤量为 ${toText(treatment?.afterSigns?.realUfVolume)}L。下机观察凝血分级为 ${coagText}，并发症：${treatment?.complications || '无'}。最新备注：${getLatestSymptom(treatment)}。`
  }, [hasCoag, dialyzerCoag, lineACoag, lineVCoag, monitoringRows.length, patient.dryWeight, patient.treatmentPlan, treatment])

  const missingSummaryFields = useMemo(() => [
    !doctorSummary.trim() ? '医生小结' : '',
    !treatmentSummary.trim() ? '护理小结' : '',
  ].filter(Boolean), [doctorSummary, treatmentSummary])

  const overviewTiles = [
    { label: '容量负荷', value: `${toText(treatment?.afterSigns?.realUfVolume)} L`, sub: '实际超滤量', icon: <Waves size={16} className="text-blue-500" />, tone: 'border-blue-100 bg-blue-50/60' },
    { label: '生命体征', value: monitoringRows.length > 0 ? `${monitoringRows.length} 点` : '暂无趋势', sub: monitoringRows.length > 0 ? '透中监测点' : '等待监测', icon: <TrendingUp size={16} className="text-rose-500" />, tone: 'border-rose-100 bg-rose-50/60' },
    { label: '安全质控', value: hasCoag ? `${dialyzerCoag}` : '未记录', sub: hasCoag ? `A:${lineACoag} V:${lineVCoag}` : '凝血、通路、并发症', icon: <ShieldCheck size={16} className="text-indigo-500" />, tone: 'border-indigo-100 bg-indigo-50/60' },
    { label: '医嘱执行', value: `${orders.length} 项`, sub: '本次执行医嘱', icon: <FileText size={16} className="text-emerald-500" />, tone: 'border-emerald-100 bg-emerald-50/60' },
  ]

  return (
    <div className="space-y-4 pb-8">
      {treatmentLoading ? (
        <section className="rounded-lg border border-blue-100 bg-blue-50 px-4 py-3 text-sm font-semibold text-blue-700">
          正在加载新患者治疗数据，治疗小结与监测摘要已清空旧患者内容。
        </section>
      ) : null}

      <div className="flex flex-wrap items-center gap-3 text-xs text-slate-500">
        <span className="text-lg font-black text-slate-900">{patient.name}</span>
        <span>ID: {patient.id}</span>
        <span>{patient.gender} / {patient.age}岁</span>
        <span>方案: {patient.treatmentPlan || '--'}</span>
        <span className="rounded-md bg-slate-100 px-2 py-0.5 font-semibold">干体重 {patient.dryWeight || 0}kg</span>
      </div>

      <section className="rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
        <h3 className="mb-3 text-sm font-bold text-slate-800">结项总览</h3>
        <div className="grid grid-cols-2 gap-3 md:grid-cols-4">
          {overviewTiles.map((tile) => (
            <div key={tile.label} className={`flex items-start gap-3 rounded-xl border p-3.5 ${tile.tone}`}>
              <span className="mt-0.5">{tile.icon}</span>
              <div>
                <div className="text-[11px] font-semibold text-slate-400">{tile.label}</div>
                <div className="mt-0.5 text-base font-black text-slate-900">{tile.value}</div>
                <div className="text-[10px] text-slate-400">{tile.sub}</div>
              </div>
            </div>
          ))}
        </div>
      </section>

      <div className="grid grid-cols-1 gap-4 xl:grid-cols-2">
        <section className="overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm">
          <div className="flex items-center justify-between border-b border-slate-100 px-4 py-3">
            <div className="flex items-center gap-2">
              <NotebookPen size={15} className="text-blue-600" />
              <div>
                <h3 className="text-sm font-bold text-slate-800">透析小结</h3>
                <div className="text-[10px] text-slate-400">保存到老库 Treatment_Treatment</div>
              </div>
            </div>
            <button type="button" onClick={() => void handleSaveSummary()} disabled={!treatment || treatmentLoading || saving} className="inline-flex items-center gap-1.5 rounded-lg bg-blue-600 px-4 py-1.5 text-xs font-bold text-white transition hover:bg-blue-700 disabled:opacity-50">
              <Save size={14} />{saving ? '保存中...' : '保存小结'}
            </button>
          </div>
          <div className="space-y-3 p-4">
            <label className="block">
              <span className="mb-1.5 block text-xs font-bold text-blue-700">医生小结</span>
              <textarea rows={5} value={doctorSummary} onChange={(e) => setDoctorSummary(e.target.value)} placeholder="记录医生对本次透析治疗过程、并发症和后续处理意见的总结。" className="w-full resize-none rounded-lg border border-blue-200 bg-white px-3 py-2.5 text-sm font-semibold text-slate-700 outline-none focus:border-blue-400" />
            </label>
            <label className="block">
              <span className="mb-1.5 block text-xs font-bold text-emerald-700">治疗/护理小结</span>
              <textarea rows={5} value={treatmentSummary} onChange={(e) => setTreatmentSummary(e.target.value)} placeholder="记录护士对上下机、监测、通路、凝血和患者宣教执行情况的总结。" className="w-full resize-none rounded-lg border border-emerald-200 bg-white px-3 py-2.5 text-sm font-semibold text-slate-700 outline-none focus:border-emerald-400" />
            </label>
          </div>
        </section>

        <div className="overflow-hidden rounded-xl border border-amber-100 bg-amber-50/40 shadow-sm">
          <div className="flex items-center justify-between border-b border-amber-100 px-4 py-3">
            <div className="flex items-center gap-2 text-xs font-bold text-amber-800">
              <FileText size={14} />
              自动汇总参考
            </div>
            <button type="button" onClick={() => void handleCopyAutoSummary()} className="inline-flex items-center gap-1 rounded-lg border border-amber-200 bg-white px-3 py-1 text-[11px] font-semibold text-amber-700 transition hover:bg-amber-100">
              <Copy size={12} />复制自动汇总
            </button>
          </div>
          <div className="space-y-2 p-4">
            <p className="text-sm leading-6 text-slate-700">{autoSummary}</p>
            <div className="flex flex-wrap gap-2 text-[11px] text-amber-600">
              <span>操作提示：</span>
              <span>复制容量和超滤要点到小结</span>
              <span>·</span>
              <span>补充生命体征、通路、并发症和医嘱</span>
            </div>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 gap-3 md:grid-cols-3">
        <div className="rounded-xl border border-slate-200 bg-white p-4">
          <h3 className="mb-3 text-xs font-bold text-slate-800">透前/透后摘要</h3>
          <div className="space-y-1.5 text-xs text-slate-500">
            <div className="flex justify-between"><span>透前血压</span><span className="font-semibold text-slate-700">{treatment?.beforeSigns?.sbp && treatment?.beforeSigns?.dbp ? `${treatment.beforeSigns.sbp}/${treatment.beforeSigns.dbp}` : '--'} mmHg</span></div>
            <div className="flex justify-between"><span>透后血压</span><span className="font-semibold text-slate-700">{parseBp(treatment?.endBp)} mmHg</span></div>
            <div className="flex justify-between"><span>透后情况</span><span className="font-semibold text-slate-700">{treatment?.complications || '--'}</span></div>
          </div>
        </div>
        <div className="rounded-xl border border-slate-200 bg-white p-4">
          <h3 className="mb-3 text-xs font-bold text-slate-800">双人核对摘要</h3>
          <div className="space-y-1.5 text-xs text-slate-500">
            <div>首次核对：{treatment?.firstCheck?.operatorId ? '已核对' : '未核对'}</div>
            <div>二次核对：{treatment?.secondCheck?.operatorId ? '已核对' : '未核对'}</div>
          </div>
        </div>
        <div className="rounded-xl border border-slate-200 bg-white p-4">
          <h3 className="mb-3 text-xs font-bold text-slate-800">凝血与通路摘要</h3>
          <div className="space-y-1.5 text-xs text-slate-500">
            <div className="flex justify-between"><span>透析器</span><span className="font-semibold text-slate-700">{dialyzerCoag}</span></div>
            <div className="flex justify-between"><span>A端 / V端</span><span className="font-semibold text-slate-700">{lineACoag} / {lineVCoag}</span></div>
            <div className="flex justify-between"><span>并发症</span><span className="font-semibold text-slate-700">{treatment?.complications || '无'}</span></div>
          </div>
        </div>
      </div>

      <section className="overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm">
        <div className="flex items-center justify-between border-b border-slate-100 px-4 py-3">
          <div className="flex items-center gap-2 text-sm font-bold text-slate-800">
            <FileText size={15} className="text-blue-600" />
            本次执行医嘱明细汇总
          </div>
          <span className="text-[11px] font-semibold text-slate-400">共 {orders.length} 项</span>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full min-w-[700px] text-left">
            <thead className="bg-slate-50 text-[11px] font-semibold text-slate-500">
              <tr>
                <th className="px-4 py-2.5">类型</th>
                <th className="px-4 py-2.5">项目名称</th>
                <th className="px-4 py-2.5">执行编码</th>
                <th className="px-4 py-2.5">执行时间</th>
                <th className="px-4 py-2.5">执行人</th>
                <th className="px-4 py-2.5">状态</th>
              </tr>
            </thead>
            <tbody>
              {orders.length > 0 ? orders.map((order) => (
                <tr key={order.id} className="border-t border-slate-50 text-xs">
                  <td className="px-4 py-2.5"><span className="rounded-md bg-blue-50 px-2 py-0.5 text-[11px] font-bold text-blue-700">{order.type}</span></td>
                  <td className="px-4 py-2.5 font-semibold text-slate-900">{order.content || order.name || '--'}</td>
                  <td className="px-4 py-2.5 text-slate-500">{order.route || order.frequency || '--'}</td>
                  <td className="px-4 py-2.5 text-slate-500">{formatDateTime(order.executedAt)}</td>
                  <td className="px-4 py-2.5 text-slate-600">{order.executedBy || order.doctorName || '--'}</td>
                  <td className="px-4 py-2.5"><CheckCircle2 size={14} className="text-emerald-500" /></td>
                </tr>
              )) : (
                <tr><td colSpan={6} className="px-4 py-12 text-center text-xs text-slate-400">暂无本次执行医嘱</td></tr>
              )}
            </tbody>
          </table>
        </div>
      </section>

      <div className="sticky bottom-0 z-10 flex flex-wrap items-center justify-between gap-3 rounded-xl border border-slate-200 bg-white/95 px-4 py-3.5 shadow-lg backdrop-blur">
        <div className="flex flex-wrap items-center gap-x-5 gap-y-1 text-xs text-slate-500">
          <span>最新治疗时间：{formatDateTime(treatment?.endTime || treatment?.startTime)}</span>
          <span>透中监测点：{monitoringRows.length} 个</span>
          <span>最新备注：{getLatestSymptom(treatment)}</span>
          {missingSummaryFields.length > 0 && (
            <span className="font-bold text-rose-500">缺项：{missingSummaryFields.join('、')}未填写</span>
          )}
        </div>
        <div className="flex gap-3">
          <button type="button" onClick={() => void handleCopyAutoSummary()} className="inline-flex items-center gap-1.5 rounded-lg border border-slate-200 bg-white px-4 py-2 text-xs font-semibold text-slate-600 transition hover:bg-slate-50">
            <Copy size={13} />复制汇总
          </button>
          <button type="button" onClick={() => void handleSaveSummary()} disabled={!treatment || treatmentLoading || saving} className="inline-flex items-center gap-1.5 rounded-lg bg-blue-600 px-5 py-2 text-xs font-bold text-white shadow-sm shadow-blue-900/20 transition hover:bg-blue-700 disabled:opacity-50">
            <Save size={14} />{saving ? '保存中...' : '保存小结'}
          </button>
        </div>
      </div>
    </div>
  )
}
