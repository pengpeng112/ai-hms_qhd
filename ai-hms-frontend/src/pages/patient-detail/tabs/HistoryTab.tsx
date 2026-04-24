// History Tab - 治疗历史

import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { History, Printer, FileText, Clock, Stethoscope, ClipboardList, ChevronLeft, ChevronRight } from 'lucide-react'
import type { Patient } from '@/types/original'
import { restApi } from '@/services'
import { TreatmentHistorySheetModal } from '@/components/patient/modals'
import type { ExtendedTreatmentHistory } from '@/components/patient/modals/types'
import type { TreatmentHistoryItem } from '../types'

interface HistoryTabProps {
  patient: Patient
}

type HistoryFilterType = 'month' | 'halfyear' | 'year'

export default function HistoryTab({ patient }: HistoryTabProps) {
  const { t } = useTranslation('patient')

  const [historyFilter, setHistoryFilter] = useState<HistoryFilterType>('month')
  const [historyDateRange, setHistoryDateRange] = useState({ start: '2024-01-01', end: '2024-12-31' })
  const [historyPage, setHistoryPage] = useState(1)
  const [historyPageSize, setHistoryPageSize] = useState(20)
  const [selectedHistoryIds, setSelectedHistoryIds] = useState<string[]>([])
  const [historyList, setHistoryList] = useState<TreatmentHistoryItem[]>([])
  const [loading, setLoading] = useState(false)
  const [isSheetOpen, setIsSheetOpen] = useState(false)
  const [sheetFilters, setSheetFilters] = useState({ year: 'ALL', month: 'ALL', day: 'ALL' })

  useEffect(() => {
    const patientId = patient?.id
    if (!patientId) {
      setHistoryList([])
      setLoading(false)
      return
    }

    setLoading(true)
    restApi
      .getTreatments({ patientId, pageSize: 25 })
      .then((res) => {
        setHistoryList(
          res.data.items.map((treatment) => ({
            id: String(treatment.id),
            date: treatment.treatmentDate?.slice(0, 10) ?? '',
            timeRange: treatment.timeRange ?? (
              treatment.startTime && treatment.endTime
                ? `${new Date(treatment.startTime).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })}-${new Date(treatment.endTime).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })}`
                : ''
            ),
            mode: treatment.treatmentType ?? '',
            duration: treatment.durationMinutes ? `${(treatment.durationMinutes / 60).toFixed(1)} 小时` : '',
            weightLoss: treatment.weightLossKg ?? 0,
            startBP: treatment.startBp ?? '',
            endBP: treatment.endBp ?? '',
            complications: treatment.complications ?? '',
            doctor: treatment.doctorName ?? '',
            doctorSummary: treatment.doctorSummary ?? treatment.notes ?? '',
            treatmentSummary: treatment.treatmentSummary ?? '',
          }))
        )
      })
      .catch(() => setHistoryList([]))
      .finally(() => setLoading(false))
  }, [patient?.id])

  const filteredHistory = historyList
  const totalPages = Math.ceil(filteredHistory.length / historyPageSize) || 1
  const paginatedHistory = filteredHistory.slice((historyPage - 1) * historyPageSize, historyPage * historyPageSize)
  const baseSheetRecords = (selectedHistoryIds.length > 0
    ? historyList.filter(item => selectedHistoryIds.includes(item.id))
    : historyList
  ).map<ExtendedTreatmentHistory>((item) => ({
    id: item.id,
    date: item.date,
    mode: item.mode,
    duration: item.duration || '-',
    timeRange: item.timeRange,
    weightLoss: item.weightLoss || 0,
    startBP: item.startBP || '-',
    endBP: item.endBP || '-',
    complications: item.complications || '',
    doctor: item.doctor || '-',
    doctorSummary: item.doctorSummary,
    treatmentSummary: item.treatmentSummary,
  }))
  const selectedSheetRecords = baseSheetRecords.filter((item) => {
    const year = item.date.slice(0, 4)
    const month = item.date.slice(5, 7)
    const day = item.date.slice(8, 10)
    if (sheetFilters.year !== 'ALL' && sheetFilters.year !== year) return false
    if (sheetFilters.month !== 'ALL' && sheetFilters.month !== month) return false
    if (sheetFilters.day !== 'ALL' && sheetFilters.day !== day) return false
    return true
  })
  const availableYears = Array.from(new Set(historyList.map(item => item.date.slice(0, 4)).filter(Boolean))).sort()
  const availableMonths = Array.from(new Set(historyList.map(item => item.date.slice(5, 7)).filter(Boolean))).sort()
  const availableDays = Array.from(new Set(historyList.map(item => item.date.slice(8, 10)).filter(Boolean))).sort()

  return (
    <div className="space-y-6 animate-fade-in pb-10 flex flex-col h-full">
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center gap-4">
          <h3 className="text-sm font-black uppercase tracking-wider flex items-center text-slate-800">
            <History size={18} className="mr-2 text-blue-600" /> {t('section.treatmentHistory')}
            {selectedHistoryIds.length > 0 && (
              <span className="ml-4 px-3 py-1 bg-blue-100 text-blue-600 rounded-lg text-xs font-black border border-blue-200 animate-fade-in">
                {t('history.selectedCount', { count: selectedHistoryIds.length })}
              </span>
            )}
          </h3>
          <div className="flex items-center gap-2 bg-white px-4 py-2 rounded-2xl border border-slate-200 shadow-sm">
            <input
              type="date"
              value={historyDateRange.start}
              onChange={(e) => setHistoryDateRange(prev => ({ ...prev, start: e.target.value }))}
              className="text-xs font-bold outline-none border-none cursor-pointer"
            />
            <span className="text-slate-300">-</span>
            <input
              type="date"
              value={historyDateRange.end}
              onChange={(e) => setHistoryDateRange(prev => ({ ...prev, end: e.target.value }))}
              className="text-xs font-bold outline-none border-none cursor-pointer"
            />
          </div>
          <div className="flex gap-1.5 bg-slate-100 p-1 rounded-2xl shadow-inner">
            {([{ id: 'month' as const, labelKey: 'history.filter.month' as const }, { id: 'halfyear' as const, labelKey: 'history.filter.halfYear' as const }, { id: 'year' as const, labelKey: 'history.filter.year' as const }]).map(btn => (
              <button
                key={btn.id}
                onClick={() => setHistoryFilter(btn.id)}
                className={`px-4 py-1.5 rounded-xl text-[11px] font-black transition-all ${historyFilter === btn.id ? 'bg-white text-blue-600 shadow-md' : 'text-slate-400 hover:text-slate-600'}`}
              >
                {t(btn.labelKey)}
              </button>
            ))}
          </div>
        </div>
        <div className="flex gap-3">
          <button onClick={() => window.print()} className="px-8 py-2.5 bg-white border border-slate-200 text-slate-700 text-xs font-black rounded-2xl hover:bg-slate-50 shadow-sm flex items-center gap-2">
            <Printer size={16} /> {t('action.print')}
          </button>
          <button
            onClick={() => setIsSheetOpen(true)}
            className="px-8 py-2.5 bg-blue-600 text-white text-xs font-black rounded-2xl hover:bg-blue-700 shadow-xl shadow-blue-100 flex items-center gap-2"
          >
            <FileText size={16} /> {t('history.treatmentRecordSheet')}
          </button>
        </div>
      </div>

      <div className="flex-1 space-y-4">
        {loading ? (
          <div className="py-40 text-center text-slate-300 flex flex-col items-center gap-4 bg-white rounded-[40px] border border-dashed border-slate-200">
            <History size={64} className="opacity-10" />
            <p className="font-bold">{t('loading' as never)}</p>
          </div>
        ) : paginatedHistory.length > 0 ? paginatedHistory.map((h, idx) => (
          <div key={h.id} className="flex items-center gap-4 group/row">
            <div className="flex flex-col items-center gap-3 shrink-0">
              <span className="text-[11px] font-black text-slate-300 font-mono">#{String((historyPage - 1) * historyPageSize + idx + 1).padStart(2, '0')}</span>
              <input
                type="checkbox"
                checked={selectedHistoryIds.includes(h.id)}
                onChange={() => setSelectedHistoryIds(prev => prev.includes(h.id) ? prev.filter(id => id !== h.id) : [...prev, h.id])}
                className="w-5 h-5 rounded-[6px] border-slate-200 text-blue-600 focus:ring-0 cursor-pointer transition-all hover:scale-110"
              />
            </div>
            <div className="flex-1 bg-white rounded-[32px] border border-slate-100 shadow-sm p-8 hover:border-blue-300 transition-all group hover:shadow-xl hover:shadow-blue-50/50 flex flex-col gap-6">
              <div className="flex flex-wrap items-center gap-x-12 gap-y-2">
                <span className="text-lg font-black text-slate-900 tracking-tighter">{patient?.name}</span>
                <div className="px-4 py-1 bg-blue-50 text-blue-600 rounded-xl font-black text-xs border border-blue-100 shadow-sm">{t('history.bedLabel', { bed: patient?.bedNumber })}</div>
                <div className="px-4 py-1 bg-slate-900 text-white rounded-xl font-black text-[11px] uppercase tracking-wider shadow-md">{h.mode}</div>
                <div className="flex items-center gap-2 text-sm font-black text-slate-400 font-mono">
                  <Clock size={14} className="text-slate-300" /> {h.date} <span className="opacity-40 font-normal">|</span> {h.timeRange}
                </div>
              </div>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-10">
                <div className="p-5 bg-slate-50/50 rounded-2xl border border-slate-50 transition-colors group-hover:bg-slate-50 group-hover:border-slate-100">
                  <p className="text-[10px] text-slate-400 font-black uppercase mb-2 flex items-center gap-2 tracking-widest">
                    <Stethoscope size={12} className="text-blue-500" /> {t('history.doctorSummary')}
                  </p>
                  <p className="text-xs font-bold text-slate-800 leading-relaxed">{h.doctorSummary}</p>
                </div>
                <div className="p-5 bg-slate-50/50 rounded-2xl border border-slate-50 transition-colors group-hover:bg-slate-50 group-hover:border-slate-100">
                  <p className="text-[10px] text-slate-400 font-black uppercase mb-2 flex items-center gap-2 tracking-widest">
                    <ClipboardList size={12} className="text-blue-500" /> {t('history.treatmentSummary')}
                  </p>
                  <p className="text-xs font-bold text-slate-800 leading-relaxed">{h.treatmentSummary}</p>
                </div>
              </div>
            </div>
          </div>
        )) : (
          <div className="py-40 text-center text-slate-300 flex flex-col items-center gap-4 bg-white rounded-[40px] border border-dashed border-slate-200">
            <History size={64} className="opacity-10" />
            <p className="font-bold">暂无治疗历史记录</p>
          </div>
        )}
      </div>

      <div className="flex items-center justify-between mt-auto pt-8 border-t border-slate-100 relative">
        <button
          disabled={historyPage === 1}
          onClick={() => setHistoryPage(p => Math.max(1, p - 1))}
          className="p-4 bg-white border border-slate-200 rounded-[24px] text-slate-400 hover:text-blue-600 hover:border-blue-200 transition-all shadow-sm hover:shadow-lg disabled:opacity-30 disabled:pointer-events-none active:scale-95"
        >
          <ChevronLeft size={24} strokeWidth={3} />
        </button>
        <div className="flex items-center gap-12 bg-white/50 backdrop-blur-md px-10 py-3 rounded-full border border-slate-100 shadow-sm">
          <div className="text-center group">
            <p className="text-[9px] text-slate-400 font-black uppercase tracking-widest mb-0.5 group-hover:text-blue-500 transition-colors">{t('history.pagination.total')}</p>
            <p className="text-sm font-black text-slate-800">{filteredHistory.length}</p>
          </div>
          <div className="w-px h-6 bg-slate-100"></div>
          <div className="text-center group">
            <p className="text-[9px] text-slate-400 font-black uppercase tracking-widest mb-0.5 group-hover:text-blue-500 transition-colors">{t('history.pagination.totalPages')}</p>
            <p className="text-sm font-black text-slate-800">{totalPages}</p>
          </div>
          <div className="w-px h-6 bg-slate-100"></div>
          <div className="text-center group">
            <p className="text-[9px] text-slate-400 font-black uppercase tracking-widest mb-0.5 group-hover:text-blue-500 transition-colors">{t('history.pagination.currentPage')}</p>
            <p className="text-sm font-black text-blue-600">{historyPage}</p>
          </div>
          <div className="w-px h-6 bg-slate-100"></div>
          <div className="text-center group">
            <p className="text-[9px] text-slate-400 font-black uppercase tracking-widest mb-0.5 group-hover:text-blue-500 transition-colors">{t('history.pagination.pageSize')}</p>
            <select
              value={historyPageSize}
              onChange={(e) => {
                setHistoryPageSize(Number(e.target.value))
                setHistoryPage(1)
              }}
              className="text-sm font-black text-slate-800 bg-transparent outline-none cursor-pointer border-none p-0 focus:ring-0"
            >
              <option value={20}>20</option>
              <option value={50}>50</option>
              <option value={100}>100</option>
              <option value={200}>200</option>
            </select>
          </div>
        </div>
        <button
          disabled={historyPage === totalPages || totalPages === 0}
          onClick={() => setHistoryPage(p => Math.min(totalPages, p + 1))}
          className="p-4 bg-white border border-slate-200 rounded-[24px] text-slate-400 hover:text-blue-600 hover:border-blue-200 transition-all shadow-sm hover:shadow-lg disabled:opacity-30 disabled:pointer-events-none active:scale-95"
        >
          <ChevronRight size={24} strokeWidth={3} />
        </button>
      </div>

      <TreatmentHistorySheetModal
        isOpen={isSheetOpen}
        onClose={() => setIsSheetOpen(false)}
        records={selectedSheetRecords}
        availableYears={availableYears}
        availableMonths={availableMonths}
        availableDays={availableDays}
        filters={sheetFilters}
        onFilterChange={(key, val) => setSheetFilters(prev => ({ ...prev, [key]: val }))}
      />
    </div>
  )
}
