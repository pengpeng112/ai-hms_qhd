import { useState, useEffect, useCallback } from 'react'
import { message, DatePicker, Modal } from 'antd'
import dayjs, { type Dayjs } from 'dayjs'
import { BarChart3, ChevronRight } from 'lucide-react'
import { getQCDoctors, getQCDoctorDetail, QC_ITEM_LABELS, QC_ITEM_ORDER, QC_NOT_CONNECTED, type QCDoctorScore, type QCPatientRow } from '@/services/qcApi'

const pct = (v: number) => `${Math.round((v || 0) * 100)}%`
const rateColor = (v: number) => (v >= 0.8 ? 'text-emerald-600' : v >= 0.5 ? 'text-amber-600' : 'text-rose-600')

export default function QCScoringPage() {
  const [month, setMonth] = useState<Dayjs>(dayjs())
  const [doctors, setDoctors] = useState<QCDoctorScore[]>([])
  const [loading, setLoading] = useState(false)
  const [detail, setDetail] = useState<{ doctor: QCDoctorScore; patients: QCPatientRow[] } | null>(null)
  const [detailOpen, setDetailOpen] = useState(false)

  const monthStr = month.format('YYYY-MM')

  const load = useCallback(() => {
    setLoading(true)
    getQCDoctors(monthStr)
      .then((d) => setDoctors([...d].sort((a, b) => b.totalScore - a.totalScore)))
      .catch(() => message.error('加载赋分失败（质控表/数据未就绪？）'))
      .finally(() => setLoading(false))
  }, [monthStr])
  useEffect(() => { load() }, [load])

  const openDetail = async (doctorId: string) => {
    try {
      const d = await getQCDoctorDetail(doctorId, monthStr)
      setDetail(d)
      setDetailOpen(true)
    } catch {
      message.error('加载下钻失败')
    }
  }

  return (
    <div className="h-full overflow-y-auto bg-slate-50 p-6">
      <div className="flex items-center justify-between mb-5">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-blue-600 flex items-center justify-center text-white"><BarChart3 size={20} /></div>
          <div>
            <h1 className="text-lg font-black text-slate-800">医疗质控赋分</h1>
            <p className="text-xs text-slate-400">按责任医生(ResponsibilityDrId)聚合 · 规则对齐 V4.8 · 数据自动提取 · 达标=该项满分</p>
          </div>
        </div>
        <DatePicker picker="month" value={month} onChange={(v) => v && setMonth(v)} allowClear={false} />
      </div>

      <div className="bg-white rounded-2xl border border-slate-200 overflow-x-auto">
        <table className="w-full text-sm whitespace-nowrap">
          <thead>
            <tr className="bg-slate-50 border-b border-slate-200 text-slate-500">
              <th className="py-2.5 px-3 text-left font-black">医生</th>
              <th className="py-2.5 px-3 text-right font-black">病人数</th>
              <th className="py-2.5 px-3 text-right font-black">数量分</th>
              <th className="py-2.5 px-3 text-right font-black">质量分</th>
              <th className="py-2.5 px-3 text-right font-black text-blue-600">总分</th>
              {QC_ITEM_ORDER.map((k) => (
                <th key={k} className="py-2.5 px-2 text-right font-bold text-[12px]">{QC_ITEM_LABELS[k]}{QC_NOT_CONNECTED.has(k) ? '*' : ''}</th>
              ))}
              <th className="py-2.5 px-3"></th>
            </tr>
          </thead>
          <tbody>
            {doctors.map((d) => (
              <tr key={d.doctorId} className="border-b border-slate-50 hover:bg-sky-50/40 cursor-pointer" onClick={() => openDetail(d.doctorId)}>
                <td className="py-2.5 px-3 font-black text-slate-700">{d.doctorName || `医生 #${d.doctorId}`}</td>
                <td className="py-2.5 px-3 text-right font-mono">{d.patientCount}</td>
                <td className="py-2.5 px-3 text-right font-mono text-slate-500">{d.quantityScore}</td>
                <td className="py-2.5 px-3 text-right font-mono text-slate-500">{d.qualityScore}</td>
                <td className="py-2.5 px-3 text-right font-mono font-black text-blue-600">{d.totalScore}</td>
                {QC_ITEM_ORDER.map((k) => (
                  QC_NOT_CONNECTED.has(k)
                    ? <td key={k} className="py-2.5 px-2 text-right font-mono text-[12px] text-slate-300" title="数据源待接入，暂不计入达标率">待接</td>
                    : <td key={k} className={`py-2.5 px-2 text-right font-mono text-[12px] font-bold ${rateColor(d.onTargetRate?.[k] ?? 0)}`}>{pct(d.onTargetRate?.[k] ?? 0)}</td>
                ))}
                <td className="py-2.5 px-3 text-slate-300"><ChevronRight size={16} /></td>
              </tr>
            ))}
          </tbody>
        </table>
        {loading && <div className="py-6 text-center text-slate-400">加载中…</div>}
        {!loading && doctors.length === 0 && <div className="py-12 text-center text-slate-400">本月暂无赋分数据（需病人已设责任医生 + 有治疗/检验数据）</div>}
      </div>
      <p className="mt-3 text-[12px] text-slate-400">* CTR 影像数据源（ACTRS/胸片）尚未接入，当前按缺测计 0；最终计分口径待质控负责人确认。</p>

      <Modal open={detailOpen} onCancel={() => setDetailOpen(false)} footer={null} width={860}
        title={detail ? `${detail.doctor?.doctorName || `医生 #${detail.doctor?.doctorId}`} · 病人赋分下钻（${monthStr}）` : ''}>
        {detail && (
          <div className="max-h-[60vh] overflow-y-auto">
            <table className="w-full text-sm whitespace-nowrap">
              <thead>
                <tr className="bg-slate-50 text-slate-500 sticky top-0">
                  <th className="py-2 px-2 text-left font-black">病人</th>
                  <th className="py-2 px-2 text-right font-black text-blue-600">总分</th>
                  {QC_ITEM_ORDER.map((k) => <th key={k} className="py-2 px-1.5 text-right font-bold text-[11px]">{QC_ITEM_LABELS[k]}{QC_NOT_CONNECTED.has(k) ? '*' : ''}</th>)}
                </tr>
              </thead>
              <tbody>
                {detail.patients.map((p) => (
                  <tr key={p.patientId} className="border-b border-slate-50">
                    <td className="py-2 px-2 font-bold text-slate-700">{p.patientName || `#${p.patientId}`}</td>
                    <td className="py-2 px-2 text-right font-mono font-black text-blue-600">{p.score.total}</td>
                    {QC_ITEM_ORDER.map((k) => {
                      if (QC_NOT_CONNECTED.has(k)) {
                        return <td key={k} className="py-2 px-1.5 text-right font-mono text-[12px] text-slate-300" title="数据源待接入">待接</td>
                      }
                      const v = p.score.items?.[k] ?? 0
                      return <td key={k} className={`py-2 px-1.5 text-right font-mono text-[12px] ${v > 0 ? 'text-emerald-600' : v < 0 ? 'text-rose-600' : 'text-slate-400'}`}>{v > 0 ? `+${v}` : v}</td>
                    })}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Modal>
    </div>
  )
}
