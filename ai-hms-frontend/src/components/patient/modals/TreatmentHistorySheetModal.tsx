// TreatmentHistorySheetModal - 历史治疗记录单弹窗

import { FileText, X, Stethoscope, ClipboardList } from 'lucide-react'
import type { ExtendedTreatmentHistory } from './types'

interface TreatmentHistorySheetModalProps {
  isOpen: boolean
  onClose: () => void
  records: ExtendedTreatmentHistory[]
  availableYears: string[]
  availableMonths: string[]
  availableDays: string[]
  filters: { year: string; month: string; day: string }
  onFilterChange: (key: 'year' | 'month' | 'day', val: string) => void
}

export default function TreatmentHistorySheetModal({
  isOpen,
  onClose,
  records,
  availableYears,
  availableMonths,
  availableDays,
  filters,
  onFilterChange
}: TreatmentHistorySheetModalProps) {
  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-[200] flex items-center justify-center bg-black/40 backdrop-blur-sm animate-fade-in">
      <div className="bg-white rounded-[32px] shadow-2xl w-full max-w-6xl overflow-hidden animate-scale-in border border-slate-100 flex flex-col max-h-[90vh]">
        <div className="bg-slate-50 px-10 py-6 flex items-center justify-between border-b border-slate-200 shrink-0">
          <h3 className="text-xl font-black text-slate-800 flex items-center gap-2">
            <FileText size={22} className="text-blue-600" /> 历史治疗记录单
          </h3>
          <div className="flex items-center gap-4">
            <div className="flex items-center gap-2 bg-white px-4 py-1.5 rounded-xl border border-slate-200">
              <span className="text-[10px] font-black text-slate-400 uppercase">时间筛选:</span>
              <select value={filters.year} onChange={(e) => onFilterChange('year', e.target.value)} className="text-xs font-bold outline-none border-none bg-transparent cursor-pointer">
                <option value="ALL">全部年份</option>
                {availableYears.map((y) => (
                  <option key={y} value={y}>
                    {y}年
                  </option>
                ))}
              </select>
              <select value={filters.month} onChange={(e) => onFilterChange('month', e.target.value)} className="text-xs font-bold outline-none border-none bg-transparent cursor-pointer">
                <option value="ALL">全部月份</option>
                {availableMonths.map((m) => (
                  <option key={m} value={m}>
                    {m}月
                  </option>
                ))}
              </select>
              <select value={filters.day} onChange={(e) => onFilterChange('day', e.target.value)} className="text-xs font-bold outline-none border-none bg-transparent cursor-pointer">
                <option value="ALL">全部日期</option>
                {availableDays.map((d) => (
                  <option key={d} value={d}>
                    {d}日
                  </option>
                ))}
              </select>
            </div>
            <button onClick={onClose} className="p-2 hover:bg-slate-200 rounded-full transition-all text-slate-400">
              <X size={20} />
            </button>
          </div>
        </div>

        <div className="flex-1 overflow-y-auto p-10 bg-slate-100/30 custom-scrollbar">
          <div className="space-y-12">
            {records.map((record) => (
              <div key={record.id} className="bg-white border border-slate-200 shadow-sm p-12 rounded-[40px] relative overflow-hidden group hover:shadow-xl transition-all duration-500">
                <div className="absolute top-0 right-0 w-32 h-32 bg-blue-50/50 rounded-bl-[80px] -z-0 opacity-0 group-hover:opacity-100 transition-opacity"></div>
                <div className="flex justify-between items-start mb-10 border-b border-slate-100 pb-6">
                  <div>
                    <h4 className="text-2xl font-black text-slate-900 tracking-tighter">
                      治疗记录单 <span className="text-slate-300 font-mono text-sm font-normal ml-3">#{record.id}</span>
                    </h4>
                    <p className="text-sm font-bold text-blue-600 mt-1 uppercase tracking-widest">{record.date}</p>
                  </div>
                  <div className="text-right">
                    <span className="px-4 py-1.5 bg-slate-900 text-white rounded-xl text-xs font-black shadow-lg uppercase">{record.mode}</span>
                  </div>
                </div>
                <div className="grid grid-cols-2 md:grid-cols-4 gap-8 mb-10">
                  <div>
                    <p className="text-[10px] text-slate-400 font-black uppercase mb-1">治疗时长</p>
                    <p className="font-bold text-slate-700">{record.duration}</p>
                  </div>
                  <div>
                    <p className="text-[10px] text-slate-400 font-black uppercase mb-1">超滤总量</p>
                    <p className="font-bold text-slate-700">{record.weightLoss} kg</p>
                  </div>
                  <div>
                    <p className="text-[10px] text-slate-400 font-black uppercase mb-1">开始/结束血压</p>
                    <p className="font-bold text-slate-700">
                      {record.startBP} / {record.endBP}
                    </p>
                  </div>
                  <div>
                    <p className="text-[10px] text-slate-400 font-black uppercase mb-1">责任医生</p>
                    <p className="font-bold text-slate-700">{record.doctor}</p>
                  </div>
                </div>
                <div className="space-y-6">
                  <div className="p-6 bg-slate-50 rounded-3xl border border-slate-100">
                    <p className="text-xs font-black text-slate-400 uppercase tracking-widest mb-2 flex items-center gap-2">
                      <Stethoscope size={14} className="text-blue-500" /> 医生小结
                    </p>
                    <p className="text-sm text-slate-700 font-bold leading-relaxed">{record.doctorSummary}</p>
                  </div>
                  <div className="p-6 bg-slate-50 rounded-3xl border border-slate-100">
                    <p className="text-xs font-black text-slate-400 uppercase tracking-widest mb-2 flex items-center gap-2">
                      <ClipboardList size={14} className="text-blue-500" /> 治疗小结
                    </p>
                    <p className="text-sm text-slate-700 font-bold leading-relaxed">{record.treatmentSummary}</p>
                  </div>
                </div>
              </div>
            ))}
            {records.length === 0 && (
              <div className="py-20 text-center text-slate-400 bg-white rounded-[40px] border border-dashed border-slate-200">
                <FileText size={48} className="mx-auto mb-4 opacity-20" />
                <p className="font-bold">该时间段内未查询到符合条件的治疗记录单</p>
              </div>
            )}
          </div>
        </div>
        <div className="p-8 bg-slate-50 border-t border-slate-200 flex justify-end shrink-0">
          <button onClick={onClose} className="px-10 py-3 bg-slate-900 text-white rounded-2xl text-sm font-black hover:bg-slate-800 transition-all shadow-xl shadow-slate-200">
            确定
          </button>
        </div>
      </div>
    </div>
  )
}
