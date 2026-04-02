// Overview Tab - 核心诊疗概览

import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Activity, ShieldCheck, Microscope, ClipboardList, ChevronRight, FileText, AlertCircle } from 'lucide-react'
import { LineChart, Line, XAxis, YAxis, Tooltip, Legend, ResponsiveContainer } from 'recharts'
import { SectionHeader, DetailCard, LabelValue } from '@/components/ui'
import { useDictNameMaps, getNameFromMap } from '@/hooks/useDictName'
import { DICT_TYPES } from '@/services/dictApi'
import type { Patient } from '@/types/original'
import type { TrendDataItem, TabID } from '../types'

// 生成伪随机数（基于索引的确定性序列）
const pseudoRandom = (index: number, seed: number): number => {
  const x = Math.sin(seed + index) * 10000
  return x - Math.floor(x)
}

// 空状态组件
const EmptyState = ({ icon: Icon, title, description, action }: {
  icon: React.ElementType
  title: string
  description: string
  action?: { label: string; onClick: () => void }
}) => (
  <div className="flex flex-col items-center justify-center py-8 px-4 text-center">
    <div className="w-16 h-16 bg-slate-100 rounded-2xl flex items-center justify-center mb-4">
      <Icon size={32} className="text-slate-400" />
    </div>
    <p className="text-sm font-bold text-slate-600 mb-2">{title}</p>
    <p className="text-xs text-slate-400 mb-4 max-w-[200px]">{description}</p>
    {action && (
      <button
        onClick={action.onClick}
        className="text-xs font-black text-blue-600 hover:underline px-4 py-2 bg-blue-50 rounded-xl"
      >
        {action.label}
      </button>
    )}
  </div>
)

interface OverviewTabProps {
  patient: Patient
  onNavigate?: (tab: TabID) => void
}

export default function OverviewTab({ patient, onNavigate }: OverviewTabProps) {
  const { t } = useTranslation('patient')
  // 使用种子确保每次渲染时随机序列一致
  const [randomSeed] = useState(() => Date.now())

  // 字典名称映射
  const dictTypeCodes = useMemo(() => [DICT_TYPES.DIALYSIS_MODE], [])
  const dictNameMaps = useDictNameMaps(dictTypeCodes)

  // 检查是否有检验数据（如果有 labTrends 且不为空，说明有数据）
  const hasLabData = patient.recentLabs && patient.recentLabs.length > 0

  // 趋势数据 - 如果没有检验数据，显示空状态
  const overviewTrendData = useMemo(() => {
    // 如果没有检验数据，返回空数组
    if (!hasLabData) {
      return []
    }

    // NOTE: Placeholder trend data — will be replaced when lab-reports API integration is completed in Phase 2+
    const months = [
      t('chart.month.aug'),
      t('chart.month.sep'),
      t('chart.month.oct'),
      t('chart.month.nov'),
      t('chart.month.dec'),
      t('chart.month.jan')
    ]
    const ref = {
      hgb: { min: 130, max: 175 },
      ca: { min: 2.0, max: 2.75 },
      p: { min: 0.8, max: 1.62 }
    }

    const getColor = (val: number, range: { min: number; max: number }) => {
      if (val > range.max) return '#EF4444'
      if (val < range.min) return '#22C55E'
      return '#EAB308'
    }

    return months.map((m, index) => {
      // 使用确定性伪随机数替代 Math.random()
      const hgbVal = 100 + pseudoRandom(index * 3, randomSeed) * 90
      const caVal = 1.6 + pseudoRandom(index * 3 + 1, randomSeed) * 1.6
      const pVal = 0.5 + pseudoRandom(index * 3 + 2, randomSeed) * 1.8

      return {
        month: m,
        hgb: parseFloat(hgbVal.toFixed(1)),
        hgbColor: getColor(hgbVal, ref.hgb),
        ca: parseFloat(caVal.toFixed(2)),
        caColor: getColor(caVal, ref.ca),
        p: parseFloat(pVal.toFixed(2)),
        pColor: getColor(pVal, ref.p)
      }
    })
  }, [t, hasLabData, randomSeed])

  // 渐变色停止点生成器
  const getGradientStops = (data: TrendDataItem[], key: string, colorKey: keyof TrendDataItem) => {
    if (!data.length) return null
    const n = data.length
    const stops = []
    for (let i = 1; i < n; i++) {
      const startOffset = ((i - 1) / (n - 1)) * 100
      const endOffset = (i / (n - 1)) * 100
      const color = data[i][colorKey] as string
      stops.push(<stop key={`start-${key}-${i}`} offset={`${startOffset}%`} stopColor={color} />)
      stops.push(<stop key={`end-${key}-${i}`} offset={`${endOffset}%`} stopColor={color} />)
    }
    return stops
  }

  return (
    <div className="space-y-6 animate-fade-in">
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <DetailCard className="lg:col-span-2 bg-gradient-to-br from-slate-900 via-slate-800 to-blue-900 text-white border-none relative overflow-hidden">
          <div className="absolute top-0 right-0 w-32 h-32 bg-blue-500 opacity-10 rounded-full -translate-y-1/2 translate-x-1/2"></div>
          <SectionHeader icon={Activity} title={t('section.corePlan')} dark />
          {patient.treatmentPlan ? (
            <>
              <div className="grid grid-cols-4 gap-6 mt-4">
                <LabelValue label={t('label.dialysisMode')} value={getNameFromMap(dictNameMaps[DICT_TYPES.DIALYSIS_MODE] || new Map(), patient.defaultMode)} color="text-blue-300 text-2xl font-black" />
                <LabelValue label={t('info.dryWeight')} value={`${patient.dryWeight} kg`} color="text-white text-2xl font-black" />
                <LabelValue label={t('label.bloodFlow')} value="250 ml/min" color="text-blue-200 font-bold" />
                <LabelValue label={t('label.anticoagulant')} value={t('label.anticoagulantDefault')} color="text-blue-200 font-bold" />
              </div>
              <div className="mt-8 p-4 bg-white/5 backdrop-blur-sm rounded-2xl border border-white/10 flex justify-between items-center group">
                <div>
                  <p className="text-[10px] text-slate-400 uppercase font-black">{t('label.lastTreatment')}</p>
                  <p className="text-sm text-slate-200 mt-1">{t('label.lastTreatmentNote')}</p>
                </div>
                <button onClick={() => onNavigate?.('treatment_plan')} className="p-2 bg-white/10 rounded-xl hover:bg-white/20 transition-all group-hover:translate-x-1">
                  <ChevronRight size={18} />
                </button>
              </div>
            </>
          ) : (
            <EmptyState
              icon={FileText}
              title={t('section.corePlan')}
              description="暂无核心诊疗方案，请在治疗方案模块中创建"
              action={{ label: '前往创建', onClick: () => onNavigate?.('treatment_plan') }}
            />
          )}
        </DetailCard>
        <DetailCard className="border-l-8 border-l-red-500 bg-red-50/20">
          <SectionHeader icon={ShieldCheck} title={t('section.infectionMonitor')} />
          {patient.infection && (patient.infection.hbsag || patient.infection.hcvab || patient.infection.hivab || patient.infection.tpab) ? (
            <div className="space-y-3 mt-4">
              {[
                { label: t('label.hbsag'), value: patient.infection?.hbsag, positiveText: t('status.positive'), negativeText: t('status.negative') },
                { label: t('label.hcvab'), value: patient.infection?.hcvab, positiveText: t('status.positive'), negativeText: t('status.negative') },
                { label: t('label.tpab'), value: patient.infection?.tpab, positiveText: t('status.positive'), negativeText: t('status.negative') },
                { label: t('label.hivab'), value: patient.infection?.hivab, positiveText: t('status.positive'), negativeText: t('status.negative') },
                { label: t('label.tb'), value: patient.infection?.tb, positiveText: t('status.positive'), negativeText: t('status.negative') },
              ].map(inf => {
                const isPositive = inf.value === '阳性' || String(inf.value) === 'Positive'
                const displayValue = isPositive ? inf.positiveText : (inf.value || inf.negativeText)
                return (
                  <div key={inf.label} className="flex justify-between items-center py-2.5 border-b border-slate-100 last:border-0">
                    <span className="text-sm text-slate-500 font-bold">{inf.label}</span>
                    <span className={`text-xs font-black px-3 py-1 rounded-full ${isPositive ? 'bg-red-100 text-red-600' : 'bg-green-100 text-green-600'}`}>
                      {displayValue}
                    </span>
                  </div>
                )
              })}
            </div>
          ) : (
            <EmptyState
              icon={AlertCircle}
              title={t('section.infectionMonitor')}
              description="暂无院感/传染病监控数据，后续将从检验报告中自动获取"
            />
          )}
        </DetailCard>
      </div>
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <DetailCard>
          <SectionHeader icon={Microscope} title={t('section.labsTrend')} />
          {overviewTrendData.length > 0 ? (
            <div className="h-44 mt-4">
              <ResponsiveContainer width="100%" height="100%">
                <LineChart data={overviewTrendData} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
                  <defs>
                    <linearGradient id="hgbOverGrad" x1="0" y1="0" x2="1" y2="0">{getGradientStops(overviewTrendData, 'hgb', 'hgbColor')}</linearGradient>
                    <linearGradient id="caOverGrad" x1="0" y1="0" x2="1" y2="0">{getGradientStops(overviewTrendData, 'ca', 'caColor')}</linearGradient>
                    <linearGradient id="pOverGrad" x1="0" y1="0" x2="1" y2="0">{getGradientStops(overviewTrendData, 'p', 'pColor')}</linearGradient>
                  </defs>
                  <XAxis dataKey="month" axisLine={false} tickLine={false} tick={{ fill: '#94a3b8', fontSize: 10, fontWeight: 700 }} />
                  <YAxis yAxisId="hgb" hide domain={['auto', 'auto']} />
                  <YAxis yAxisId="ca" hide domain={['auto', 'auto']} />
                  <YAxis yAxisId="p" hide domain={['auto', 'auto']} />
                  <Tooltip
                    contentStyle={{ borderRadius: '12px', border: 'none', boxShadow: '0 4px 6px -1px rgba(0,0,0,0.1)' }}
                    content={({ active, payload }) => {
                      if (active && payload && payload.length) {
                        const data = payload[0].payload as TrendDataItem
                        return (
                          <div className="bg-white p-3 rounded-xl shadow-xl border border-slate-100 text-[10px] font-bold">
                            <p className="text-slate-400 mb-1">{data.month}</p>
                            <div className="space-y-1">
                              <div className="flex justify-between gap-4"><span>{t('chart.hgb')}:</span> <span style={{ color: data.hgbColor }}>{data.hgb}</span></div>
                              <div className="flex justify-between gap-4"><span>{t('chart.calcium')}:</span> <span style={{ color: data.caColor }}>{data.ca}</span></div>
                              <div className="flex justify-between gap-4"><span>{t('chart.phosphorus')}:</span> <span style={{ color: data.pColor }}>{data.p}</span></div>
                            </div>
                          </div>
                        )
                      }
                      return null
                    }}
                  />
                  <Legend verticalAlign="top" height={24} iconSize={12} wrapperStyle={{ fontSize: '10px', fontWeight: 'bold' }} />
                  <Line
                    yAxisId="hgb"
                    type="monotone"
                    dataKey="hgb"
                    name={t('chart.hemoglobin')}
                    stroke="url(#hgbOverGrad)"
                    strokeWidth={3}
                    dot={(props: { cx?: number; cy?: number; payload?: TrendDataItem }) => {
                      const { cx = 0, cy = 0, payload } = props
                      return <circle cx={cx} cy={cy} r={4} fill={payload?.hgbColor || '#000'} stroke="white" strokeWidth={1} />
                    }}
                  />
                  <Line
                    yAxisId="ca"
                    type="monotone"
                    dataKey="ca"
                    name={t('chart.calcium')}
                    stroke="url(#caOverGrad)"
                    strokeWidth={3}
                    dot={(props: { cx?: number; cy?: number; payload?: TrendDataItem }) => {
                      const { cx = 0, cy = 0, payload } = props
                      return <path d={`M ${cx} ${cy - 6} L ${cx + 6} ${cy + 4} L ${cx - 6} ${cy + 4} Z`} fill={payload?.caColor || '#000'} stroke="white" strokeWidth={1} />
                    }}
                  />
                  <Line
                    yAxisId="p"
                    type="monotone"
                    dataKey="p"
                    name={t('chart.phosphorus')}
                    stroke="url(#pOverGrad)"
                    strokeWidth={3}
                    dot={(props: { cx?: number; cy?: number; payload?: TrendDataItem }) => {
                      const { cx = 0, cy = 0, payload } = props
                      return <rect x={cx - 4} y={cy - 4} width={8} height={8} fill={payload?.pColor || '#000'} stroke="white" strokeWidth={1} />
                    }}
                  />
                </LineChart>
              </ResponsiveContainer>
            </div>
          ) : (
            <EmptyState
              icon={Microscope}
              title={t('section.labsTrend')}
              description="暂无关键检验指标趋势数据，后续将从检验报告接口自动获取"
            />
          )}
        </DetailCard>
        <DetailCard>
          <SectionHeader icon={ClipboardList} title={t('section.activeOrders')} action={<button className="text-[10px] font-black text-blue-600 hover:underline">{t('action.viewExecution')}</button>} />
          <div className="space-y-2 mt-2">
            {patient.orders.slice(0, 3).map(o => {
              const isLongTerm = o.type === '长期' || String(o.type) === 'Long-term'
              return (
                <div key={o.id} className="flex items-center p-3 bg-slate-50 rounded-2xl border border-slate-100 group hover:border-blue-200 transition-all">
                  <div className={`w-1 h-8 rounded-full mr-3 ${isLongTerm ? 'bg-blue-500' : 'bg-orange-500'}`}></div>
                  <div className="flex-1">
                    <p className="text-sm font-black text-slate-800">{o.content}</p>
                    <p className="text-[10px] text-slate-400 mt-0.5">{t('label.orderMeta', { startTime: o.startTime, doctor: o.doctor })}</p>
                  </div>
                  <span className="text-[10px] bg-white px-2 py-1 rounded-lg shadow-sm font-black text-green-600 border border-slate-100 uppercase tracking-tighter">{t('status.executed')}</span>
                </div>
              )
            })}
          </div>
        </DetailCard>
      </div>
    </div>
  )
}
