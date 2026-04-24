import { FileText, NotebookPen, TrendingUp } from 'lucide-react'
import type { RestTreatment } from '@/services'
import type { Patient } from '../types'

interface Props {
  patient: Patient
  treatment: RestTreatment | null
}

function toText(value?: string | number | null) {
  if (value === undefined || value === null || value === '') return '-'
  return String(value)
}

function formatDateTime(value?: string) {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString('zh-CN', { hour12: false })
}

function parseBp(bp?: string) {
  if (!bp) return '-'
  return bp
}

function getLatestSymptom(treatment: RestTreatment | null) {
  const latest = treatment?.duringParams?.[0]
  if (!latest?.notes) return '-'
  return latest.notes
}

export default function DialysisSummary({ patient, treatment }: Props) {
  const monitoringRows = [...(treatment?.duringParams || [])].sort(
    (a, b) => new Date(b.recordTime).getTime() - new Date(a.recordTime).getTime()
  )

  return (
    <div className="space-y-6 pb-8">
      <div className="grid grid-cols-1 gap-4 md:grid-cols-4">
        {[
          ['患者', patient.name, ''],
          ['治疗方式', treatment?.treatmentType || patient.treatmentPlan, ''],
          ['目标干体重', toText(patient.dryWeight), 'kg'],
          ['本次超滤', toText(treatment?.weightLossKg), 'kg'],
        ].map(([label, value, unit]) => (
          <div key={label} className="rounded-2xl border border-slate-200 bg-white px-4 py-4 shadow-sm">
            <div className="text-[11px] font-semibold uppercase tracking-wide text-slate-400">{label}</div>
            <div className="mt-2 text-2xl font-black text-slate-800">
              {value}
              {unit ? <span className="ml-1 text-xs text-slate-400">{unit}</span> : null}
            </div>
          </div>
        ))}
      </div>

      <section className="overflow-hidden rounded-3xl border border-slate-200 bg-white shadow-sm">
        <div className="flex items-center gap-2 border-b border-slate-200 px-6 py-4">
          <NotebookPen size={16} className="text-blue-600" />
          <h3 className="text-sm font-black text-slate-800">治疗小结</h3>
        </div>
        <div className="grid grid-cols-1 gap-4 p-6 lg:grid-cols-2">
          <label className="block">
            <div className="mb-2 text-[11px] font-semibold uppercase tracking-wide text-slate-400">医生小结</div>
            <textarea
              rows={7}
              value={treatment?.doctorSummary || ''}
              readOnly
              className="w-full resize-none rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm font-semibold text-slate-700 outline-none"
            />
          </label>
          <label className="block">
            <div className="mb-2 text-[11px] font-semibold uppercase tracking-wide text-slate-400">护士小结</div>
            <textarea
              rows={7}
              value={treatment?.treatmentSummary || treatment?.notes || ''}
              readOnly
              className="w-full resize-none rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm font-semibold text-slate-700 outline-none"
            />
          </label>
        </div>
      </section>

      <section className="overflow-hidden rounded-3xl border border-slate-200 bg-white shadow-sm">
        <div className="flex items-center gap-2 border-b border-slate-200 px-6 py-4">
          <TrendingUp size={16} className="text-emerald-600" />
          <h3 className="text-sm font-black text-slate-800">监测摘要</h3>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full min-w-[920px] text-left">
            <thead className="bg-slate-50 text-xs uppercase tracking-wide text-slate-500">
              <tr>
                <th className="px-6 py-3">时间</th>
                <th className="px-6 py-3">血压</th>
                <th className="px-6 py-3">心率</th>
                <th className="px-6 py-3">累计超滤</th>
                <th className="px-6 py-3">血流量</th>
                <th className="px-6 py-3">备注</th>
              </tr>
            </thead>
            <tbody>
              {monitoringRows.length > 0 ? (
                monitoringRows.map((item) => (
                  <tr key={item.id} className="border-t border-slate-100 text-sm">
                    <td className="px-6 py-4 font-semibold text-slate-800">{formatDateTime(item.recordTime)}</td>
                    <td className="px-6 py-4 text-slate-600">
                      {item.sbp !== undefined && item.dbp !== undefined ? `${item.sbp}/${item.dbp}` : '-'}
                    </td>
                    <td className="px-6 py-4 text-slate-600">{toText(item.heartRate)}</td>
                    <td className="px-6 py-4 text-slate-600">
                      {item.ufVolume !== undefined ? `${item.ufVolume} L` : '-'}
                    </td>
                    <td className="px-6 py-4 text-slate-600">{toText(item.bloodFlow)}</td>
                    <td className="px-6 py-4 text-slate-600">{item.notes || '-'}</td>
                  </tr>
                ))
              ) : (
                <tr>
                  <td colSpan={6} className="px-6 py-10 text-center text-sm text-slate-400">
                    当前治疗暂无透中监测数据
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </section>

      <section className="rounded-3xl bg-slate-900 px-6 py-5 text-white shadow-lg">
        <div className="flex items-center justify-between gap-4">
          <div className="flex items-center gap-3">
            <FileText size={18} className="text-slate-300" />
            <span className="text-sm font-semibold">
              最新治疗时间：{formatDateTime(treatment?.endTime || treatment?.startTime)}，最新备注：{getLatestSymptom(treatment)}
            </span>
          </div>
          <div className="text-sm font-semibold text-slate-300">
            末次血压：{parseBp(treatment?.endBp)}
          </div>
        </div>
      </section>
    </div>
  )
}
