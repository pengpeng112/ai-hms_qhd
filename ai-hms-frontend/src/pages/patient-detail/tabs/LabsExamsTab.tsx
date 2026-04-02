// Labs & Exams Tab - 检查检验报告

import { useState, useEffect, useMemo, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { Beaker, Microscope, ExternalLink, BarChart2, Loader2 } from 'lucide-react'
import { message } from 'antd'
import { SectionHeader, DetailCard } from '@/components/ui'
import { LabTrendModal, LabHistoryModal } from '@/components/patient/modals'
import { restApi, type LabReportApi, type ExamReportApi, type KeyIndicatorApi, type LabReportSyncResult } from '@/services/restClient'
import type { Patient } from '@/types/original'
import type { LabResultItem } from '../types'

interface LabsExamsTabProps {
  patient: Patient
}

function toDateOnly(value?: string | null): string {
  if (!value) return '-'
  if (value.length >= 10) return value.slice(0, 10)
  return value
}

function parseTime(value: string): number {
  const ts = new Date(value).getTime()
  return Number.isNaN(ts) ? 0 : ts
}

function convertReportsToLabResultItems(reports: LabReportApi[]): LabResultItem[] {
  const rows: LabResultItem[] = []

  reports.forEach((report) => {
    const reportDate = toDateOnly(report.reportedAt || report.sampledAt || report.createdAt)
    const reportItems = report.items || []

    if (reportItems.length === 0) {
      rows.push({
        id: `header-${report.id}`,
        code: report.itemCode || '-',
        name: report.itemName || report.reportNo || '-',
        value: '--',
        unit: '--',
        reference: '--',
        date: reportDate,
        isAbnormal: false,
        pendingSync: true,
      })
      return
    }

    reportItems.forEach((item) => {
      const flag = (item.abnormalFlag || 'N').toUpperCase()
      const isAbnormal = flag === 'H' || flag === 'L'
      const abnormalType = flag === 'H' ? 'high' : flag === 'L' ? 'low' : undefined
      const date = toDateOnly(item.testedAt || report.reportedAt || report.sampledAt || report.createdAt)

      rows.push({
        id: item.id,
        code: item.itemCode || report.itemCode || '-',
        name: item.itemName || report.itemName || '-',
        value: item.resultValue || '-',
        unit: item.unit || '-',
        reference: item.referenceRange || '-',
        date,
        isAbnormal,
        abnormalType,
      })
    })
  })

  return rows.sort((a, b) => parseTime(b.date) - parseTime(a.date))
}

function convertKeyIndicatorsToLabResultItems(indicators: KeyIndicatorApi[]): LabResultItem[] {
  const rows = indicators.map((indicator) => {
    const sign = (indicator.resultSign || '').toUpperCase()
    const isAbnormal = sign === 'H' || sign === 'L'
    const abnormalType = sign === 'H' ? 'high' : sign === 'L' ? 'low' : undefined
    const date = toDateOnly(indicator.testTime || indicator.createdAt)

    return {
      id: `record-${indicator.id}`,
      code: indicator.indexCode || '-',
      name: indicator.indexName || '-',
      value: indicator.result || '-',
      unit: indicator.unit || '-',
      reference: indicator.reference || '-',
      date,
      isAbnormal,
      abnormalType,
      pendingSync: false,
    } as LabResultItem
  })

  return rows.sort((a, b) => parseTime(b.date) - parseTime(a.date))
}

function normalizeExamDate(report: ExamReportApi): string {
  return toDateOnly(report.examDate || report.createdAt)
}

export default function LabsExamsTab({ patient }: LabsExamsTabProps) {
  const { t } = useTranslation('patient')

  const [selectedTrendLab, setSelectedTrendLab] = useState<LabResultItem | null>(null)
  const [isLabHistoryModalOpen, setIsLabHistoryModalOpen] = useState(false)

  const [keyLoading, setKeyLoading] = useState(false)
  const [labReportLoading, setLabReportLoading] = useState(false)
  const [examLoading, setExamLoading] = useState(false)
  const [keyIndicators, setKeyIndicators] = useState<KeyIndicatorApi[]>([])
  const [labReports, setLabReports] = useState<LabReportApi[]>([])
  const [examReports, setExamReports] = useState<ExamReportApi[]>([])
  const [examSyncSummary, setExamSyncSummary] = useState<LabReportSyncResult | null>(null)

  const loadLabReports = useCallback(async () => {
    if (!patient.id) return

    setLabReportLoading(true)
    try {
      const result = await restApi.getLabReports(patient.id, { page: 1, pageSize: 200 })
      setLabReports(result.items || [])
    } catch (error) {
      console.error('加载检验报告失败:', error)
      message.error('加载检验报告失败')
      setLabReports([])
    } finally {
      setLabReportLoading(false)
    }
  }, [patient.id])

  const loadKeyIndicators = useCallback(async () => {
    if (!patient.id) return

    setKeyLoading(true)
    try {
      const result = await restApi.getKeyIndicators(patient.id, { page: 1, pageSize: 200 })
      setKeyIndicators(result.items || [])
    } catch (error) {
      console.error('加载关键指标失败:', error)
      setKeyIndicators([])
    } finally {
      setKeyLoading(false)
    }
  }, [patient.id])

  const loadExamReports = useCallback(async () => {
    if (!patient.id) return

    setExamLoading(true)
    try {
      const result = await restApi.getExamReports(patient.id, { page: 1, pageSize: 100 })
      setExamReports(result.items || [])
    } catch (error) {
      console.error('加载检查报告失败:', error)
      message.error('加载检查报告失败')
      setExamReports([])
    } finally {
      setExamLoading(false)
    }
  }, [patient.id])

  useEffect(() => {
    if (!patient.id) return

    const bootstrap = async () => {
      const [labSync, keySync, examSync] = await Promise.allSettled([
        restApi.syncLabReports(patient.id),
        restApi.syncKeyIndicators(patient.id),
        restApi.syncExamReports(patient.id),
      ])

      if (labSync.status === 'rejected') {
        console.warn('syncLabReports failed:', labSync.reason)
      }
      if (keySync.status === 'rejected') {
        console.warn('syncKeyIndicators failed:', keySync.reason)
      }
      if (examSync.status === 'rejected') {
        console.warn('syncExamReports failed:', examSync.reason)
        setExamSyncSummary(null)
      } else {
        setExamSyncSummary(examSync.value)
      }

      await Promise.all([
        loadKeyIndicators(),
        loadLabReports(),
        loadExamReports(),
      ])
    }

    bootstrap()
  }, [patient.id, loadKeyIndicators, loadLabReports, loadExamReports])

  const keyIndicatorResults = useMemo<LabResultItem[]>(
    () => convertKeyIndicatorsToLabResultItems(keyIndicators),
    [keyIndicators]
  )
  const fallbackLabResults = useMemo<LabResultItem[]>(
    () => convertReportsToLabResultItems(labReports),
    [labReports]
  )
  const labResults = keyIndicatorResults.length > 0 ? keyIndicatorResults : fallbackLabResults
  const labLoading = keyLoading || labReportLoading
  const sortedExamReports = useMemo<ExamReportApi[]>(
    () => [...examReports].sort((a, b) => parseTime(normalizeExamDate(b)) - parseTime(normalizeExamDate(a))),
    [examReports]
  )
  const examEmptyText = useMemo(() => {
    if (!examSyncSummary) return '暂无检查报告'
    if (examSyncSummary.created === 0 && examSyncSummary.updated === 0 && examSyncSummary.errors > 0) {
      return `暂无检查报告（同步失败 ${examSyncSummary.errors} 条）`
    }
    if (examSyncSummary.created === 0 && examSyncSummary.updated === 0 && examSyncSummary.skipped > 0) {
      return `暂无检查报告（同步完成，未发现新增数据）`
    }
    return '暂无检查报告'
  }, [examSyncSummary])

  const getDaysFromNow = (dateStr: string) => {
    const reportDate = new Date(dateStr)
    if (Number.isNaN(reportDate.getTime())) return '-'

    const now = new Date()
    const diff = now.getTime() - reportDate.getTime()
    const days = Math.floor(diff / (1000 * 60 * 60 * 24))
    if (days < 1) return t('label.today')
    return t('label.daysAgo', { count: days })
  }

  return (
    <div className="space-y-6 animate-fade-in pb-10">
      <div className="grid grid-cols-12 gap-6 items-start">
        {/* 左侧：重要指标结果 */}
        <div className="col-span-12 lg:col-span-8">
          <DetailCard>
            <SectionHeader
              icon={Beaker}
              title={t('section.keyLabResults')}
              action={
                <button
                  onClick={() => setIsLabHistoryModalOpen(true)}
                  className="text-xs font-black text-blue-600 hover:underline flex items-center"
                >
                  <ExternalLink size={14} className="mr-1" /> {t('action.viewLabReport')}
                </button>
              }
            />
            <div className="mt-4 border border-slate-100 rounded-2xl overflow-hidden">
              <table className="w-full text-left text-sm border-collapse table-fixed">
                <thead className="bg-slate-50 text-slate-400 font-black text-[10px] border-b border-slate-100 uppercase tracking-widest">
                  <tr>
                    <th className="py-4 px-4 w-12 text-center">{t('label.seqNo')}</th>
                    <th className="py-4 w-20">{t('label.indicator')}</th>
                    <th className="py-4">{t('label.name')}</th>
                    <th className="py-4 w-24">{t('label.result')}</th>
                    <th className="py-4 w-16">{t('label.unit')}</th>
                    <th className="py-4 w-28">{t('label.referenceRange')}</th>
                    <th className="py-4 w-24">{t('label.testTime')}</th>
                    <th className="py-4 text-right px-6 w-20">{t('label.daysFromNow')}</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-50">
                  {labLoading && (
                    <tr>
                      <td colSpan={8} className="py-10 text-center">
                        <div className="inline-flex items-center text-slate-400 font-bold">
                          <Loader2 size={16} className="animate-spin mr-2" />
                          加载中...
                        </div>
                      </td>
                    </tr>
                  )}

                  {!labLoading && labResults.length === 0 && (
                    <tr>
                      <td colSpan={8} className="py-10 text-center text-slate-300 font-bold italic">
                        暂无检验数据
                      </td>
                    </tr>
                  )}

                  {!labLoading && labResults.map((lab, idx) => (
                    <tr
                      key={lab.id}
                      className={`hover:bg-slate-50/50 group transition-colors ${lab.pendingSync ? '' : 'cursor-pointer'}`}
                      onClick={() => {
                        if (!lab.pendingSync) {
                          setSelectedTrendLab(lab)
                        }
                      }}
                    >
                      <td className="py-4 px-4 text-center text-slate-300 font-mono text-xs">{idx + 1}</td>
                      <td className="py-4 font-black text-blue-600 uppercase text-xs truncate" title={lab.code}>{lab.code}</td>
                      <td className="py-4 text-slate-800 font-bold text-xs">
                        <div className="flex items-center gap-2 min-w-0">
                          <span className="truncate" title={lab.name}>{lab.name}</span>
                          <BarChart2 size={12} className="text-slate-300 group-hover:text-blue-400 transition-colors shrink-0" />
                        </div>
                      </td>
                      <td className="py-4">
                        {lab.pendingSync ? (
                          <span className="text-xs font-bold px-2 py-1 rounded-lg border bg-slate-50 text-slate-500 border-slate-200">
                            待同步
                          </span>
                        ) : (
                          <span
                            className={`text-sm font-black font-mono px-2 py-0.5 rounded-lg border flex items-center w-fit gap-1 ${
                              !lab.isAbnormal
                                ? 'bg-yellow-50 text-yellow-600 border-yellow-100'
                                : lab.abnormalType === 'high'
                                  ? 'bg-red-50 text-red-600 border-red-100'
                                  : 'bg-green-50 text-green-600 border-green-100'
                            }`}
                          >
                            {lab.value}
                            {lab.isAbnormal && lab.abnormalType === 'high' && '↑'}
                            {lab.isAbnormal && lab.abnormalType === 'low' && '↓'}
                          </span>
                        )}
                      </td>
                      <td className="py-4 text-slate-400 text-[10px] font-bold">{lab.pendingSync ? '--' : lab.unit}</td>
                      <td className="py-4 text-slate-500 font-mono text-[10px]">{lab.pendingSync ? '待同步明细' : lab.reference}</td>
                      <td className="py-4 text-slate-400 font-mono text-xs">{lab.date}</td>
                      <td className="py-4 text-right px-6 text-slate-300 font-bold text-[10px] uppercase group-hover:text-blue-500">
                        {getDaysFromNow(lab.date)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </DetailCard>
        </div>

        {/* 右侧：检查报告 */}
        <div className="col-span-12 lg:col-span-4">
          <DetailCard>
            <SectionHeader icon={Microscope} title={t('section.examReports')} />
            <div className="space-y-4 mt-4">
              {examLoading && (
                <div className="p-6 bg-slate-50 rounded-2xl border border-slate-100 text-center text-slate-400 text-sm font-bold">
                  <span className="inline-flex items-center">
                    <Loader2 size={16} className="animate-spin mr-2" />
                    加载中...
                  </span>
                </div>
              )}

              {!examLoading && sortedExamReports.length === 0 && (
                <div className="p-6 bg-slate-50 rounded-2xl border border-slate-100 text-center text-slate-300 text-sm font-bold italic">
                  {examEmptyText}
                </div>
              )}

              {!examLoading && sortedExamReports.map((report) => (
                <div
                  key={report.id}
                  className="p-5 bg-slate-50 rounded-2xl border border-slate-100 hover:border-blue-200 group cursor-pointer transition-all"
                >
                  <div className="flex justify-between items-start mb-3">
                    <div>
                      <p className="text-[10px] font-black text-slate-400 uppercase tracking-widest">{report.department || '检查科室'}</p>
                      <h5 className="text-sm font-black text-slate-800 mt-1">{report.title || '-'}</h5>
                    </div>
                    <span className="text-[10px] font-mono text-slate-400 bg-white px-2 py-1 rounded shadow-sm shrink-0">
                      {normalizeExamDate(report)}
                    </span>
                  </div>
                  <p className="text-xs font-bold text-slate-700 leading-relaxed group-hover:text-slate-800">
                    {report.conclusion || '暂无检查结论'}
                  </p>
                </div>
              ))}
            </div>
          </DetailCard>
        </div>
      </div>

      <LabHistoryModal
        isOpen={isLabHistoryModalOpen}
        onClose={() => setIsLabHistoryModalOpen(false)}
        patientId={patient.id}
        patientName={patient.name}
      />

      <LabTrendModal
        isOpen={selectedTrendLab !== null}
        onClose={() => setSelectedTrendLab(null)}
        lab={selectedTrendLab}
      />
    </div>
  )
}
