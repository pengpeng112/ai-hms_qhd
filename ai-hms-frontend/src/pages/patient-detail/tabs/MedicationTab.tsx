import { useState, useEffect, useCallback } from 'react'
import { Pill, Activity, Loader2 } from 'lucide-react'
import { Card, Tag, Space } from 'antd'
import { SectionHeader } from '@/components/ui'
import { medicationApi, type MedicationAdmin, type MedSuggestion } from '@/services/medicationApi'
import type { TabProps } from '../types'
import dayjs from 'dayjs'

const STATUS_BADGE: Record<string, { color: string; text: string }> = {
  low: { color: 'orange', text: '偏低' },
  high: { color: 'red', text: '偏高' },
  normal: { color: 'green', text: '正常' },
  no_data: { color: 'default', text: '无数据' },
}

export default function MedicationTab({ patient }: TabProps) {
  const [suggestions, setSuggestions] = useState<MedSuggestion[]>([])
  const [history, setHistory] = useState<MedicationAdmin[]>([])
  const [loading, setLoading] = useState(true)

  const fetchData = useCallback(async () => {
    setLoading(true)
    try {
      const [sugs, rows] = await Promise.all([
        medicationApi.suggestions(patient.id).catch(() => [] as MedSuggestion[]),
        medicationApi.list({ patientId: Number(patient.id) }).catch(() => [] as MedicationAdmin[]),
      ])
      setSuggestions(sugs)
      setHistory(rows)
    } catch { /* ignore */ }
    setLoading(false)
  }, [patient.id])

  useEffect(() => { fetchData() }, [fetchData])

  if (loading) {
    return <div className="py-12 text-center text-slate-400"><Loader2 size={20} className="inline animate-spin" /> 加载中…</div>
  }

  return (
    <div className="space-y-4">
      <SectionHeader icon={Activity} title="用药与给药" />

      {/* Suggestions */}
      {suggestions.length > 0 && (
        <div className="space-y-2">
          <div className="text-[13px] font-bold text-slate-600">指标驱动调药建议</div>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-2">
            {suggestions.map((s) => {
              const badge = STATUS_BADGE[s.status] || STATUS_BADGE.no_data
              const isActionable = s.status === 'low' || s.status === 'high'
              return (
                <Card key={s.indicator} size="small" bordered
                  className={isActionable ? 'rounded-xl border-orange-200 bg-orange-50' : 'rounded-xl'}>
                  <div className="flex items-center justify-between mb-1">
                    <span className="font-bold text-[13px]">{s.label}</span>
                    <Tag color={badge.color} className="text-[11px]">{badge.text}</Tag>
                  </div>
                  {s.value !== undefined && (
                    <div className="text-[12px] text-slate-500">{s.value} {s.unit}</div>
                  )}
                  {isActionable && (
                    <div className="mt-2 text-[12px] text-slate-700">
                      <Tag className="text-[11px] mr-1">{s.drugLabel}</Tag>
                      {s.advice}
                    </div>
                  )}
                </Card>
              )
            })}
          </div>
        </div>
      )}

      {/* History */}
      <div className="space-y-2">
        <div className="text-[13px] font-bold text-slate-600">给药执行历史（最近）</div>
        {history.length === 0 ? (
          <div className="text-[12px] text-slate-400 py-2">暂无给药记录</div>
        ) : (
          <Space direction="vertical" style={{ width: '100%' }} size={8}>
            {history.slice(0, 20).map((r) => (
              <Card key={r.id} size="small" bordered
                className={r.status === 'verified' ? 'rounded-xl border-emerald-200 bg-emerald-50' : 'rounded-xl border-orange-200 bg-orange-50'}>
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <Pill size={14} className={r.status === 'verified' ? 'text-emerald-500' : 'text-orange-500'} />
                    <span className="font-bold text-[13px]">{r.drugName}</span>
                    {r.dose && <span className="text-[12px]">{r.dose} {r.route}</span>}
                    <Tag color={r.status === 'verified' ? 'green' : 'orange'} className="text-[11px]">
                      {r.status === 'verified' ? '已双核' : '待核对'}
                    </Tag>
                  </div>
                  <span className="text-[11px] text-slate-400">{dayjs(r.administeredAt).format('MM-DD HH:mm')}</span>
                </div>
                <div className="text-[11px] text-slate-400 mt-1">
                  执行：{r.administeredName || r.administeredBy}
                  {r.secondCheckBy && <> · 核对：{r.secondCheckName || r.secondCheckBy}</>}
                </div>
              </Card>
            ))}
          </Space>
        )}
      </div>
    </div>
  )
}
