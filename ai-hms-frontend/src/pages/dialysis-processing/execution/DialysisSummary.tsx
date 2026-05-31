import { message } from 'antd'
import { CheckCircle2, FileText, NotebookPen, Save, ShieldCheck, TrendingUp, Waves } from 'lucide-react'
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

function MetricTile({ label, value, tone }: { label: string; value: string; tone?: string }) {
  return (
    <div className={`rounded-lg border px-5 py-4 ${tone || 'border-slate-100 bg-slate-50'}`}>
      <div className="text-xs font-bold text-slate-400">{label}</div>
      <div className="mt-3 text-xl font-black text-slate-900">{value}</div>
    </div>
  )
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

  const autoSummary = useMemo(() => {
    const symItems = treatment?.afterSymptomItems
    const dialyzerCoag = getSymptomItemValue(symItems, 'dialyzer_coag')
    const lineACoag = getSymptomItemValue(symItems, 'line_a_coag')
    const lineVCoag = getSymptomItemValue(symItems, 'line_v_coag')
    const coagText = dialyzerCoag || lineACoag || lineVCoag
      ? `透析器${dialyzerCoag}、A端${lineACoag}、V端${lineVCoag}`
      : '未记录'
    return `患者于 ${formatDateTime(treatment?.startTime)} 开始 ${treatment?.treatmentType || patient.treatmentPlan || '--'} 治疗，透前体重 ${toText(treatment?.beforeSigns?.weight)}kg，干体重 ${toText(patient.dryWeight)}kg，治疗过程中生命体征${monitoringRows.length > 0 ? '已有监测记录' : '无完整监测趋势'}，实际超滤量为 ${toText(treatment?.afterSigns?.realUfVolume)}L。下机观察凝血分级为 ${coagText}，并发症：${treatment?.complications || '无'}。最新备注：${getLatestSymptom(treatment)}。`
  }, [monitoringRows.length, patient.dryWeight, patient.treatmentPlan, treatment])

  const symItems = treatment?.afterSymptomItems
  const dialyzerCoag = getSymptomItemValue(symItems, 'dialyzer_coag')
  const lineACoag = getSymptomItemValue(symItems, 'line_a_coag')
  const lineVCoag = getSymptomItemValue(symItems, 'line_v_coag')

  return (
    <div className="space-y-5 pb-8">
      {treatmentLoading ? (
        <section className="rounded-lg border border-blue-100 bg-blue-50 px-6 py-4 text-sm font-semibold text-blue-700">
          正在加载新患者治疗数据，治疗小结与监测摘要已清空旧患者内容。
        </section>
      ) : null}

      <div className="grid grid-cols-1 gap-5 xl:grid-cols-3">
        <section className="rounded-lg border border-slate-200 bg-white p-6 shadow-sm">
          <div className="mb-5 flex items-center justify-between border-b border-slate-100 pb-4">
            <div className="flex items-center gap-2 text-sm font-black text-slate-900"><Waves size={16} className="text-emerald-600" />容量负荷评估</div>
            <span className="text-xs font-black tracking-widest text-slate-300">FLUID SUMMARY</span>
          </div>
          <div className="grid gap-4 md:grid-cols-2">
            <MetricTile label="透前/透后净重" value={`${toText(treatment?.beforeSigns?.weight)} → ${toText(treatment?.afterSigns?.weight)} kg`} />
            <MetricTile label="本次体重丢失" value={`${toText(treatment?.afterSigns?.lossWeight ?? treatment?.weightLossKg)} kg`} tone="border-emerald-100 bg-emerald-50" />
          </div>
          <div className="mt-4 rounded-lg border border-blue-100 bg-blue-50 p-5">
            <div className="text-xs font-bold text-slate-500">超滤总量（目标/实际）</div>
            <div className="mt-3 text-2xl font-black text-blue-700">-- L → {toText(treatment?.afterSigns?.realUfVolume)} L</div>
          </div>
        </section>

        <section className="rounded-lg border border-slate-200 bg-white p-6 shadow-sm">
          <div className="mb-5 flex items-center justify-between border-b border-slate-100 pb-4">
            <div className="flex items-center gap-2 text-sm font-black text-slate-900"><TrendingUp size={16} className="text-rose-500" />关键生命体征趋势</div>
            <div className="flex gap-3 text-xs font-bold"><span className="text-rose-500">SBP</span><span className="text-amber-500">DBP</span><span className="text-blue-500">HR</span></div>
          </div>
          <div className="flex h-44 items-center justify-center rounded-lg bg-slate-50 text-sm font-semibold text-slate-400">
            {monitoringRows.length > 0 ? `共有 ${monitoringRows.length} 个透中监测点` : '暂无生命体征趋势数据'}
          </div>
        </section>

        <section className="rounded-lg border border-slate-200 bg-white p-6 shadow-sm">
          <div className="mb-5 flex items-center justify-between border-b border-slate-100 pb-4">
            <div className="flex items-center gap-2 text-sm font-black text-slate-900"><ShieldCheck size={16} className="text-indigo-600" />质控与安全汇总</div>
            <span className="text-xs font-black tracking-widest text-slate-300">QUALITY ASSURANCE</span>
          </div>
          <div className="space-y-4">
            <MetricTile label="凝血分级评估" value={dialyzerCoag !== '未记录' ? `透析器:${dialyzerCoag} A端:${lineACoag} V端:${lineVCoag}` : '未记录'} tone="border-indigo-100 bg-indigo-50" />
            <MetricTile label="血管通路状态" value="未记录" />
            <MetricTile label="透析并发症" value={treatment?.complications || '无'} />
          </div>
        </section>
      </div>

      <section className="overflow-hidden rounded-lg border border-slate-200 bg-white shadow-sm">
          <div className="flex items-center justify-between border-b border-slate-100 px-6 py-4">
          <div className="flex items-center gap-2"><NotebookPen size={16} className="text-blue-600" /><div><h3 className="text-sm font-black text-slate-800">透析小结</h3><div className="mt-1 text-xs font-semibold text-slate-400">保存到老库字段 Treatment_Treatment.TreatmentSummary / NurseSummary</div></div></div>
          <button type="button" onClick={() => void handleSaveSummary()} disabled={!treatment || treatmentLoading || saving} className="inline-flex items-center gap-2 rounded-lg bg-blue-600 px-5 py-2 text-sm font-bold text-white disabled:opacity-50 hover:bg-blue-700 transition">
            <Save size={16} />{saving ? '保存中...' : '保存小结'}
          </button>
        </div>
        <div className="grid grid-cols-1 gap-5 p-6 xl:grid-cols-2">
          <label className="block"><span className="mb-3 block text-sm font-bold text-blue-700">医生小结</span><textarea rows={8} value={doctorSummary} onChange={(e) => setDoctorSummary(e.target.value)} placeholder="记录医生对本次透析治疗过程、并发症和后续处理意见的总结。" className="w-full resize-none rounded-lg border border-blue-200 bg-white px-4 py-3 text-sm font-semibold text-slate-700 outline-none focus:border-blue-400" /></label>
          <label className="block"><span className="mb-3 block text-sm font-bold text-emerald-700">治疗/护理小结</span><textarea rows={8} value={treatmentSummary} onChange={(e) => setTreatmentSummary(e.target.value)} placeholder="记录护士对上下机、监测、通路、凝血和患者宣教执行情况的总结。" className="w-full resize-none rounded-lg border border-emerald-200 bg-white px-4 py-3 text-sm font-semibold text-slate-700 outline-none focus:border-emerald-400" /></label>
        </div>
        <div className="border-t border-slate-100 bg-slate-50 px-6 py-5">
          <div className="mb-2 text-xs font-bold text-slate-400">自动汇总参考</div>
          <p className="text-sm font-semibold leading-7 text-slate-800">{autoSummary}</p>
        </div>
      </section>

      <div className="grid grid-cols-1 gap-5 xl:grid-cols-3">
        <section className="rounded-lg border border-slate-200 bg-white p-6 shadow-sm"><h3 className="mb-4 text-sm font-black text-slate-800">透前/透后摘要</h3><div className="space-y-3 text-sm font-semibold text-slate-600"><div>透前血压：{treatment?.beforeSigns?.sbp && treatment?.beforeSigns?.dbp ? `${treatment.beforeSigns.sbp}/${treatment.beforeSigns.dbp}` : '--'} mmHg</div><div>透后血压：{parseBp(treatment?.endBp)} mmHg</div><div>透前症状：--</div><div>透后情况：{treatment?.complications || '--'}</div></div></section>
        <section className="rounded-lg border border-slate-200 bg-white p-6 shadow-sm"><h3 className="mb-4 text-sm font-black text-slate-800">双人核对摘要</h3><div className="space-y-3 text-sm font-semibold text-slate-600"><div>首次核对：耗材、参数、通路、管路按当前记录展示</div><div>二次核对：透析方式、处方、抗凝剂、血管通路按当前记录展示</div></div></section>
        <section className="rounded-lg border border-slate-200 bg-white p-6 shadow-sm"><h3 className="mb-4 text-sm font-black text-slate-800">凝血与通路摘要</h3><div className="space-y-3 text-sm font-semibold text-slate-600"><div>透析器：{dialyzerCoag}</div><div>A端：{lineACoag}；V端：{lineVCoag}</div><div>血管通路：未记录</div><div>并发症：{treatment?.complications || '无'}</div></div></section>
      </div>

      <section className="overflow-hidden rounded-lg border border-slate-200 bg-white shadow-sm">
        <div className="flex items-center justify-between border-b border-slate-100 px-6 py-4">
          <div className="flex items-center gap-2 text-sm font-black text-slate-800"><FileText size={16} className="text-blue-600" />本次执行医嘱明细汇总</div>
          <span className="text-xs font-bold text-slate-400">Total: {orders.length} items executed</span>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full min-w-[900px] text-left">
            <thead className="bg-slate-50 text-xs text-slate-500"><tr><th className="px-6 py-3">类型</th><th className="px-6 py-3">项目名称</th><th className="px-6 py-3">执行编码</th><th className="px-6 py-3">执行时间</th><th className="px-6 py-3">执行人</th><th className="px-6 py-3">状态</th></tr></thead>
            <tbody>
              {orders.length > 0 ? orders.map((order) => (
                <tr key={order.id} className="border-t border-slate-100 text-sm"><td className="px-6 py-5"><span className="rounded-md bg-blue-50 px-2 py-1 text-xs font-bold text-blue-700">{order.type}</span></td><td className="px-6 py-5 font-black text-slate-900">{order.content || order.name || '--'}</td><td className="px-6 py-5 text-slate-600">{order.route || order.frequency || '--'}</td><td className="px-6 py-5 text-slate-500">{formatDateTime(order.executedAt)}</td><td className="px-6 py-5 text-slate-700">{order.executedBy || order.doctorName || '--'}</td><td className="px-6 py-5"><CheckCircle2 size={18} className="text-emerald-500" /></td></tr>
              )) : <tr><td colSpan={6} className="px-6 py-10 text-center text-sm text-slate-400">暂无本次执行医嘱</td></tr>}
            </tbody>
          </table>
        </div>
      </section>

      <section className="rounded-lg bg-blue-50 px-6 py-4 text-sm font-black text-blue-700">最新治疗时间：{formatDateTime(treatment?.endTime || treatment?.startTime)}，透中监测点：{monitoringRows.length} 个，最新备注：{getLatestSymptom(treatment)}。</section>
    </div>
  )
}
