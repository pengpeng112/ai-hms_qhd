// LabTrendModal - 检验趋势弹窗

import { useState, useMemo } from 'react'
import { Beaker, X } from 'lucide-react'
import { ResponsiveContainer, LineChart, CartesianGrid, XAxis, YAxis, Tooltip, ReferenceArea, Line } from 'recharts'
import type { LabResult } from '@/types/original'

interface LabTrendModalProps {
  isOpen: boolean
  onClose: () => void
  lab: LabResult | null
}

// 生成伪随机数（基于索引的确定性序列）
const pseudoRandom = (index: number, seed: number): number => {
  const x = Math.sin(seed + index) * 10000
  return x - Math.floor(x)
}

export default function LabTrendModal({ isOpen, onClose, lab }: LabTrendModalProps) {
  const [range, setRange] = useState<'6M' | '1Y' | 'ALL'>('6M')
  const [randomSeed] = useState(() => Date.now())

  const trendData = useMemo(() => {
    if (!lab) return []
    const points = range === '6M' ? 8 : range === '1Y' ? 14 : 20
    const res = []
    const baseVal = parseFloat(lab.value.replace(/[^0-9.]/g, ''))

    // 解析参考范围
    const refParts = (lab.reference || '0--100').split('--').map((p) => parseFloat(p.trim()))
    const min = refParts[0] || 0
    const max = refParts[1] || 9999

    for (let i = points - 1; i >= 0; i--) {
      const date = new Date(lab.date)
      date.setMonth(date.getMonth() - i)
      // 使用确定性伪随机数替代 Math.random()
      const randomOffset = pseudoRandom(i, randomSeed)
      const val = i === 0 ? baseVal : baseVal + randomOffset * (max - min) * 0.4 - (max - min) * 0.2

      let color = '#EAB308' // 默认黄色 (正常)
      if (val > max) color = '#EF4444' // 红色 (高)
      if (val < min) color = '#22C55E' // 绿色 (低)

      res.push({
        date: `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}`,
        value: parseFloat(val.toFixed(2)),
        color,
        min,
        max
      })
    }
    return res
  }, [lab, range, randomSeed])

  // 计算渐变停止点
  const gradientStops = useMemo(() => {
    if (!trendData.length) return null
    const n = trendData.length
    const stops = []
    for (let i = 1; i < n; i++) {
      const startOffset = ((i - 1) / (n - 1)) * 100
      const endOffset = (i / (n - 1)) * 100
      const color = trendData[i].color
      stops.push(<stop key={`start-${i}`} offset={`${startOffset}%`} stopColor={color} />)
      stops.push(<stop key={`end-${i}`} offset={`${endOffset}%`} stopColor={color} />)
    }
    return stops
  }, [trendData])

  if (!isOpen || !lab) return null

  return (
    <div className="fixed inset-0 z-[130] flex items-center justify-center bg-black/60 backdrop-blur-sm animate-fade-in p-4">
      <div className="bg-white rounded-[32px] shadow-2xl w-full max-w-4xl overflow-hidden animate-scale-in flex flex-col ring-1 ring-black/5">
        <div className="px-10 py-6 border-b border-slate-100 flex justify-between items-center bg-slate-50/50">
          <div>
            <h3 className="text-xl font-black text-slate-800 flex items-center gap-3">
              <Beaker className="text-blue-600" size={24} /> {lab.name} 趋势分析
            </h3>
            <p className="text-xs text-slate-400 font-bold mt-1 uppercase tracking-widest">HISTORICAL INDICATOR TREND</p>
          </div>
          <div className="flex items-center gap-4">
            <div className="flex bg-slate-100 p-1 rounded-xl shadow-inner shrink-0">
              {[
                { id: '6M', label: '近半年' },
                { id: '1Y', label: '近一年' },
                { id: 'ALL', label: '全部历史' }
              ].map((r) => (
                <button
                  key={r.id}
                  onClick={() => setRange(r.id as '6M' | '1Y' | 'ALL')}
                  className={`px-4 py-1.5 rounded-lg text-xs font-black transition-all ${range === r.id ? 'bg-white text-blue-600 shadow-sm' : 'text-slate-400 hover:text-slate-600'}`}
                >
                  {r.label}
                </button>
              ))}
            </div>
            <button onClick={onClose} className="p-2 text-slate-300 hover:text-slate-600 hover:bg-white rounded-xl transition-all shadow-sm border border-transparent hover:border-slate-100">
              <X size={20} />
            </button>
          </div>
        </div>

        <div className="p-10 flex-1 overflow-hidden">
          <div className="p-2 border-2 border-blue-500 rounded-xl bg-white shadow-inner">
            <div className="h-[400px] w-full relative p-4">
              <ResponsiveContainer width="100%" height="100%">
                <LineChart data={trendData} margin={{ top: 20, right: 30, left: 10, bottom: 20 }}>
                  <defs>
                    <linearGradient id="trendGradient" x1="0" y1="0" x2="1" y2="0">
                      {gradientStops}
                    </linearGradient>
                  </defs>
                  <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#f0f0f0" />
                  <XAxis dataKey="date" axisLine={false} tickLine={false} tick={{ fill: '#94a3b8', fontSize: 10, fontWeight: 700 }} dy={10} />
                  <YAxis axisLine={false} tickLine={false} tick={{ fill: '#94a3b8', fontSize: 10, fontWeight: 700 }} />
                  <Tooltip
                    content={({ active, payload }) => {
                      if (active && payload && payload.length) {
                        const data = payload[0].payload
                        return (
                          <div className="bg-white p-4 rounded-2xl shadow-2xl border border-slate-100 ring-1 ring-black/5 animate-scale-in">
                            <p className="text-[10px] font-black text-slate-400 mb-1 uppercase tracking-widest">{data.date}</p>
                            <div className="flex items-center gap-3">
                              <span className="text-2xl font-black" style={{ color: data.color }}>
                                {data.value}
                              </span>
                              <span className="text-[10px] font-bold text-slate-400 uppercase">{lab.unit}</span>
                            </div>
                            <div className="mt-2 pt-2 border-t border-slate-50">
                              <span className="text-[10px] font-bold text-slate-500 bg-slate-100 px-2 py-0.5 rounded">参考: {lab.reference}</span>
                            </div>
                          </div>
                        )
                      }
                      return null
                    }}
                  />
                  {trendData.length > 0 && <ReferenceArea y1={trendData[0].min} y2={trendData[0].max} fill="#EAB308" fillOpacity={0.03} />}
                  <Line
                    type="monotone"
                    dataKey="value"
                    stroke="url(#trendGradient)"
                    strokeWidth={4}
                    dot={(props: { cx?: number; cy?: number; payload: { date: string; color: string } }) => {
                      const { cx, cy, payload } = props
                      if (cx === undefined || cy === undefined) return <></>
                      return <circle key={payload.date} cx={cx} cy={cy} r={6} fill="white" stroke={payload.color} strokeWidth={3} />
                    }}
                    activeDot={{ r: 8, strokeWidth: 4 }}
                  />
                </LineChart>
              </ResponsiveContainer>
            </div>
          </div>

          <div className="mt-10 grid grid-cols-3 gap-8">
            <div className="p-8 rounded-[24px] bg-slate-50/80 border border-slate-100">
              <p className="text-[11px] font-black text-slate-400 uppercase mb-2 tracking-widest">指标名称</p>
              <p className="text-xl font-black text-slate-800">{lab.name}</p>
            </div>
            <div className="p-8 rounded-[24px] bg-slate-50/80 border border-slate-100">
              <p className="text-[11px] font-black text-slate-400 uppercase mb-2 tracking-widest">参考范围</p>
              <p className="text-xl font-black text-slate-800 font-mono">{lab.reference}</p>
            </div>
            <div className="p-8 rounded-[24px] bg-slate-50/80 border border-slate-100">
              <p className="text-[11px] font-black text-slate-400 uppercase mb-2 tracking-widest">当前状态</p>
              <p className={`text-xl font-black ${!lab.isAbnormal ? 'text-yellow-500' : 'text-red-500'}`}>{!lab.isAbnormal ? '正常' : '异常'}</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
