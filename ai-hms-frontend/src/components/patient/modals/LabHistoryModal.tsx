// LabHistoryModal - 检验历史弹窗

import { Fragment, useEffect, useMemo, useState } from 'react'
import { Beaker, ChevronDown, Loader2, X } from 'lucide-react'
import { message } from 'antd'
import { restApi, type LabReportApi } from '@/services/restClient'

interface LabHistoryModalProps {
  isOpen: boolean
  onClose: () => void
  patientId: string
  patientName: string
}

interface LabHistoryGroup {
  date: string
  items: LabReportApi[]
}

function toDateOnly(value?: string | null): string {
  if (!value) return '-'
  if (value.length >= 10) return value.slice(0, 10)
  return value
}

function toDateTime(value?: string | null): string {
  if (!value) return '-'
  if (value.length >= 19) return value.slice(0, 19).replace('T', ' ')
  return value.replace('T', ' ')
}

function getGroupDate(report: LabReportApi): string {
  return toDateOnly(report.reportedAt || report.sampledAt || report.createdAt)
}

export default function LabHistoryModal({ isOpen, onClose, patientId, patientName }: LabHistoryModalProps) {
  const [loading, setLoading] = useState(false)
  const [reports, setReports] = useState<LabReportApi[]>([])
  const [expandedId, setExpandedId] = useState<string | null>(null)

  useEffect(() => {
    if (!isOpen || !patientId) return

    setLoading(true)
    restApi.getLabReports(patientId, { page: 1, pageSize: 200 })
      .then((result) => {
        setReports(result.items || [])
      })
      .catch((error) => {
        console.error('加载检验历史失败:', error)
        message.error('加载检验历史失败')
        setReports([])
      })
      .finally(() => {
        setLoading(false)
      })
  }, [isOpen, patientId])

  useEffect(() => {
    if (!isOpen) {
      setExpandedId(null)
    }
  }, [isOpen])

  const groupedReports = useMemo<LabHistoryGroup[]>(() => {
    const groupMap = new Map<string, LabReportApi[]>()

    reports.forEach((report) => {
      const date = getGroupDate(report)
      if (!groupMap.has(date)) {
        groupMap.set(date, [])
      }
      groupMap.get(date)?.push(report)
    })

    return Array.from(groupMap.entries())
      .sort((a, b) => {
        const aTime = Number.isNaN(new Date(a[0]).getTime()) ? 0 : new Date(a[0]).getTime()
        const bTime = Number.isNaN(new Date(b[0]).getTime()) ? 0 : new Date(b[0]).getTime()
        return bTime - aTime
      })
      .map(([date, items]) => ({ date, items }))
  }, [reports])

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-[120] flex items-center justify-center bg-black/40 backdrop-blur-sm animate-fade-in">
      <div className="bg-white rounded-[12px] shadow-2xl w-[98vw] max-w-[1800px] overflow-hidden animate-scale-in flex flex-col ring-1 ring-black/5 max-h-[92vh]">
        <div className="px-6 py-4 border-b border-slate-100 flex justify-between items-center bg-[#f8faff] shrink-0">
          <h3 className="text-lg font-black text-slate-800 flex items-center">
            <Beaker className="mr-3 text-blue-600" size={20} /> {patientName} - 检验报告历史列表
          </h3>
          <button onClick={onClose} className="p-1.5 text-slate-400 hover:text-slate-600 rounded-full hover:bg-slate-100 transition-all">
            <X size={20} />
          </button>
        </div>

        <div className="flex-1 overflow-auto p-4 bg-white custom-scrollbar">
          <div className="border border-slate-200 rounded-lg overflow-hidden min-w-[1400px]">
            <table className="w-full text-left text-sm border-collapse">
              <thead className="bg-[#e2f0ff] text-slate-700 font-bold text-[12px] border-b border-slate-200 sticky top-0 z-10">
                <tr>
                  <th className="py-2.5 px-4 border-r border-slate-200 w-28 font-medium"></th>
                  <th className="py-2.5 px-4 border-r border-slate-200 w-16 text-center font-medium">序号</th>
                  <th className="py-2.5 px-4 border-r border-slate-200 w-32 font-medium">申请单号</th>
                  <th className="py-2.5 px-4 border-r border-slate-200 w-56 font-medium">项目名称</th>
                  <th className="py-2.5 px-4 border-r border-slate-200 w-32 font-medium">项目编号</th>
                  <th className="py-2.5 px-4 border-r border-slate-200 w-24 font-medium">临床诊断</th>
                  <th className="py-2.5 px-4 border-r border-slate-200 w-20 font-medium text-center">标本</th>
                  <th className="py-2.5 px-4 border-r border-slate-200 w-20 font-medium text-center">紧急</th>
                  <th className="py-2.5 px-4 border-r border-slate-200 w-24 font-medium text-center">申请医生</th>
                  <th className="py-2.5 px-4 border-r border-slate-200 w-36 font-medium">申请日期</th>
                  <th className="py-2.5 px-4 border-r border-slate-200 w-36 font-medium">采样日期</th>
                  <th className="py-2.5 px-4 border-r border-slate-200 w-36 font-medium">接收日期</th>
                  <th className="py-2.5 px-4 w-36 font-medium">报告日期</th>
                </tr>
              </thead>
              <tbody className="text-[12px] text-slate-600 font-medium">
                {loading && (
                  <tr>
                    <td colSpan={13} className="py-12 text-center text-slate-400">
                      <span className="inline-flex items-center">
                        <Loader2 className="animate-spin mr-2" size={14} />
                        加载中...
                      </span>
                    </td>
                  </tr>
                )}

                {!loading && groupedReports.length === 0 && (
                  <tr>
                    <td colSpan={13} className="py-12 text-center text-slate-300">
                      暂无检验历史数据
                    </td>
                  </tr>
                )}

                {!loading && groupedReports.map((group) =>
                  group.items.map((item, index) => (
                    <Fragment key={item.id}>
                      <tr className="hover:bg-slate-50 transition-colors border-b border-slate-100 last:border-b-0">
                        <td className={`py-3 px-4 border-r border-slate-200 bg-white font-mono align-top ${index > 0 ? 'text-transparent' : ''}`}>
                          {group.date}
                        </td>
                        <td className="py-3 px-4 border-r border-slate-200 text-center">{index + 1}</td>
                        <td className="py-3 px-4 border-r border-slate-200">{item.reportNo || '-'}</td>
                        <td className="py-3 px-4 border-r border-slate-200">
                          <button
                            type="button"
                            onClick={(e) => {
                              e.stopPropagation()
                              setExpandedId(expandedId === item.id ? null : item.id)
                            }}
                            className="inline-flex items-center gap-1 text-blue-600 hover:underline text-left"
                          >
                            <span>{item.itemName || '-'}</span>
                            <ChevronDown
                              size={12}
                              className={`transition-transform ${expandedId === item.id ? 'rotate-180' : ''}`}
                            />
                          </button>
                        </td>
                        <td className="py-3 px-4 border-r border-slate-200 font-mono text-slate-400">{item.itemCode || '-'}</td>
                        <td className="py-3 px-4 border-r border-slate-200">{item.clinicalDiagnosis || '-'}</td>
                        <td className="py-3 px-4 border-r border-slate-200 text-center">{item.specimenType || '-'}</td>
                        <td className="py-3 px-4 border-r border-slate-200 text-center">{item.urgency || '-'}</td>
                        <td className="py-3 px-4 border-r border-slate-200 text-center font-bold text-slate-700">{item.requestDoctor || '-'}</td>
                        <td className="py-3 px-4 border-r border-slate-200 font-mono">{toDateTime(item.requestedAt)}</td>
                        <td className="py-3 px-4 border-r border-slate-200 font-mono">{toDateTime(item.sampledAt)}</td>
                        <td className="py-3 px-4 border-r border-slate-200 font-mono">{toDateTime(item.receivedAt)}</td>
                        <td className="py-3 px-4 font-mono">{toDateTime(item.reportedAt)}</td>
                      </tr>

                      {expandedId === item.id && (
                        <tr>
                          <td colSpan={13} className="p-0 border-b border-slate-200">
                            <div className="bg-blue-50/50 px-8 py-4">
                              {(!item.items || item.items.length === 0) ? (
                                <p className="text-xs text-slate-400 italic py-2">明细数据待同步</p>
                              ) : (
                                <table className="w-full text-left text-xs border-collapse table-fixed">
                                  <thead>
                                    <tr className="text-slate-500 text-[10px] uppercase tracking-widest">
                                      <th className="py-2 pr-4 w-16">序号</th>
                                      <th className="py-2 pr-4">项目名称</th>
                                      <th className="py-2 pr-4 w-20">代码</th>
                                      <th className="py-2 pr-4 w-20">结果</th>
                                      <th className="py-2 pr-4 w-16">单位</th>
                                      <th className="py-2 pr-4 w-24">参考范围</th>
                                      <th className="py-2 w-16">异常</th>
                                    </tr>
                                  </thead>
                                  <tbody>
                                    {item.items.map((detail, dIdx) => {
                                      const flag = (detail.abnormalFlag || 'N').toUpperCase()
                                      const isAbnormal = flag === 'H' || flag === 'L'
                                      return (
                                        <tr key={detail.id} className="border-t border-blue-100/50">
                                          <td className="py-2 pr-4 text-slate-400">{dIdx + 1}</td>
                                          <td className="py-2 pr-4 font-bold text-slate-700 truncate" title={detail.itemName}>
                                            {detail.itemName}
                                          </td>
                                          <td className="py-2 pr-4 font-mono text-slate-400 truncate">{detail.itemCode}</td>
                                          <td className={`py-2 pr-4 font-mono font-bold ${isAbnormal ? 'text-red-600' : 'text-slate-800'}`}>
                                            {detail.resultValue}{flag === 'H' ? ' ↑' : flag === 'L' ? ' ↓' : ''}
                                          </td>
                                          <td className="py-2 pr-4 text-slate-400">{detail.unit || '-'}</td>
                                          <td className="py-2 pr-4 font-mono text-slate-500">{detail.referenceRange || '-'}</td>
                                          <td className="py-2">
                                            {isAbnormal ? (
                                              <span className={`text-[10px] font-bold px-1.5 py-0.5 rounded ${flag === 'H' ? 'bg-red-100 text-red-600' : 'bg-green-100 text-green-600'}`}>
                                                {flag === 'H' ? '偏高' : '偏低'}
                                              </span>
                                            ) : (
                                              <span className="text-[10px] text-slate-400">正常</span>
                                            )}
                                          </td>
                                        </tr>
                                      )
                                    })}
                                  </tbody>
                                </table>
                              )}
                            </div>
                          </td>
                        </tr>
                      )}
                    </Fragment>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>
        <div className="p-4 bg-[#f8faff] border-t border-slate-100 flex justify-end shrink-0 gap-2">
          <button onClick={onClose} className="px-6 py-2 bg-slate-800 text-white rounded-[6px] text-xs font-black shadow-lg hover:bg-slate-700">
            关闭
          </button>
        </div>
      </div>
    </div>
  )
}
