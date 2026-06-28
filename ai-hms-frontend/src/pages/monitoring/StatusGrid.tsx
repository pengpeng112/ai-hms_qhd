import { memo, useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { LineChart, Line, XAxis, YAxis, Tooltip, ResponsiveContainer } from 'recharts'
import type { MonitorDevice, VitalSample } from '@/types/original'
import {
  Monitor, Wifi, Clock, ClipboardList, FileEdit, Droplet, AlertOctagon,
  Heart, Activity, Gauge, Beaker, Bell, AlertTriangle,
} from 'lucide-react'
import {
  classifyBedStatus, computeMAP, computeNa, alertLevelFor, formatTimeProgress,
  formatPositive, MONITOR_METRIC_LABELS,
} from './types'
import type { BedDisplayStatus, ModalType } from './types'

interface StatusGridProps {
  devices: MonitorDevice[]
  loading: boolean
  loadError: string | null
  onOpenModal: (device: MonitorDevice, type: ModalType) => void
  onReload: () => void
}

function cardShell(status: BedDisplayStatus): { border: string; shadow: string } {
  if (status === 'danger') {
    return { border: 'border-red-400 bg-red-50/40', shadow: '0 4px 16px rgba(239,68,68,0.16)' }
  }
  if (status === 'warning') {
    return { border: 'border-amber-400 bg-amber-50/40', shadow: '0 4px 16px rgba(245,158,11,0.12)' }
  }
  return { border: 'border-emerald-400 bg-white', shadow: '0 2px 10px rgba(16,185,129,0.10)' }
}

function valueClass(level?: 'warning' | 'danger'): string {
  if (level === 'danger') return 'text-red-600'
  if (level === 'warning') return 'text-amber-600'
  return 'text-slate-800'
}

type MiniRow = { ts: number; map: number | null; mapPred: number | null; hr: number | null; hrPred: number | null; sbp: number; dbp: number }
function miniRows(vitals?: VitalSample[]): MiniRow[] {
  return (vitals || [])
    .map((v): MiniRow => {
      const pred = v.kind === 'predicted'
      const map = v.map > 0 ? Math.round(v.map) : null
      const hr = v.hr > 0 ? v.hr : null
      return {
        ts: new Date(v.t).getTime(),
        map: pred ? null : map,
        mapPred: pred ? map : null,
        hr: pred ? null : hr,
        hrPred: pred ? hr : null,
        sbp: v.sbp || 0,
        dbp: v.dbp || 0,
      }
    })
    .filter((r) => !Number.isNaN(r.ts))
    .sort((a, b) => a.ts - b.ts)
}

const TrendTooltip = ({ active, payload }: { active?: boolean; payload?: { payload: MiniRow }[] }) => {
  if (!active || !payload?.length) return null
  const r = payload[0].payload
  const map = r.map ?? r.mapPred
  const hr = r.hr ?? r.hrPred
  const time = new Date(r.ts).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
  return (
    <div className="bg-white/95 border border-slate-200 rounded-lg shadow-lg px-2.5 py-1.5 text-xs leading-tight">
      <div className="font-bold text-slate-600 mb-0.5">{time}</div>
      {r.sbp > 0 && r.dbp > 0 && <div className="text-slate-500">血压 <b className="text-slate-700 font-mono">{r.sbp}/{r.dbp}</b></div>}
      {map != null && <div className="text-red-500">MAP <b className="font-mono">{map}</b></div>}
      {hr != null && <div className="text-emerald-500">心率 <b className="font-mono">{hr}</b></div>}
    </div>
  )
}

const MiniTrend = memo(({ device }: { device: MonitorDevice }) => {
  const rows = useMemo(() => miniRows(device.vitalsSeries), [device.vitalsSeries])
  const startMs = device.startTime ? new Date(device.startTime).getTime() : undefined
  const endMs = startMs && device.estimatedDuration ? startMs + device.estimatedDuration * 60000 : undefined
  if (rows.length === 0) {
    return <div className="h-12 flex items-center justify-center text-meta text-slate-300">暂无趋势</div>
  }
  return (
    <div className="h-12 w-full">
      <ResponsiveContainer width="100%" height="100%">
        <LineChart data={rows} margin={{ top: 4, right: 2, bottom: 0, left: 2 }}>
          <XAxis dataKey="ts" type="number" domain={startMs && endMs ? [startMs, endMs] : ['dataMin', 'dataMax']} hide />
          <YAxis hide domain={['auto', 'auto']} />
          <Tooltip content={<TrendTooltip />} />
          <Line dataKey="map" stroke="#ef4444" strokeWidth={1.5} dot={false} connectNulls isAnimationActive={false} />
          <Line dataKey="mapPred" stroke="#ef4444" strokeWidth={1.5} strokeDasharray="4 3" dot={false} connectNulls isAnimationActive={false} />
          <Line dataKey="hr" stroke="#10b981" strokeWidth={1.5} dot={false} connectNulls isAnimationActive={false} />
          <Line dataKey="hrPred" stroke="#10b981" strokeWidth={1.5} strokeDasharray="4 3" dot={false} connectNulls isAnimationActive={false} />
        </LineChart>
      </ResponsiveContainer>
    </div>
  )
})
MiniTrend.displayName = 'MiniTrend'

const EmptyBedCard = memo(({ device, onOpenModal }: { device: MonitorDevice; onOpenModal: StatusGridProps['onOpenModal'] }) => {
  const { t } = useTranslation('monitoring')
  return (
    <div className="rounded-[14px] bg-slate-50/60 border border-dashed border-slate-200 p-4 flex flex-col justify-between min-h-[110px] opacity-70 hover:opacity-100 transition-all duration-200">
      <div className="flex items-center gap-3">
        <div className="w-10 h-10 rounded-[12px] bg-slate-100 flex items-center justify-center text-slate-400 font-bold text-sm shrink-0">
          {device.bedNumber}
        </div>
        <div>
          <p className="font-bold text-sm text-slate-500">{device.bedNumber}</p>
          <p className="text-xs text-slate-400 mt-0.5">{t('card.idle')}</p>
        </div>
      </div>
      <div className="flex items-center justify-between mt-2 pt-2 border-t border-slate-100">
        <span className="text-xs text-slate-400 flex items-center gap-1">
          <Wifi size={10} className={device.status === 'normal' ? 'text-green-400' : 'text-slate-300'} /> 在线
        </span>
        <button onClick={() => device.status !== 'offline' && onOpenModal(device, 'COMPREHENSIVE')}
          className="text-xs text-blue-500 hover:text-blue-700 font-medium px-2 py-1 rounded hover:bg-blue-50 transition-colors">
          详情
        </button>
      </div>
    </div>
  )
})
EmptyBedCard.displayName = 'EmptyBedCard'

const ActiveBedCard = memo(({ device, onOpenModal }: { device: MonitorDevice; onOpenModal: StatusGridProps['onOpenModal'] }) => {
  const { t } = useTranslation('monitoring')
  const status = classifyBedStatus(device)
  const shell = cardShell(status)
  const ufPercent = device.vitals.ufGoal > 0
    ? Math.round((device.vitals.ufVolume / device.vitals.ufGoal) * 100)
    : 0
  const mapVal = computeMAP(device)
  const naVal = computeNa(device)
  const vp = device.vitals.vp || 0
  const time = formatTimeProgress(device)
  const badge = status === 'danger' ? 'bg-red-500 text-white'
    : status === 'warning' ? 'bg-amber-500 text-white'
    : 'bg-slate-700 text-white'

  return (
    <div
      className={`rounded-[14px] border p-4 flex flex-col min-h-[230px] transition-all duration-200 hover:shadow-lg ${shell.border} ${device.isMine ? 'ring-2 ring-blue-300' : ''} ${status === 'danger' ? 'animate-pulse-slow' : ''}`}
      style={{ borderLeftWidth: '4px', boxShadow: shell.shadow }}
    >
      <div className="flex justify-between items-start mb-3">
        <div className="flex items-center min-w-0 flex-1 gap-3">
          <div className={`w-10 h-10 rounded-[12px] flex items-center justify-center font-bold text-sm shrink-0 shadow-sm ${badge}`}>
            {device.bedNumber}
          </div>
          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-1.5">
              <h4 className="font-bold text-sm text-slate-800 truncate">{device.patientName || '--'}</h4>
              {device.age ? <span className="text-xs text-slate-400 shrink-0">{device.age}岁</span> : null}
              {device.isMine && <span className="px-1 py-0.5 rounded text-meta font-bold bg-blue-100 text-blue-600 shrink-0">我</span>}
            </div>
            <div className="flex items-center text-xs text-slate-400 mt-0.5 gap-2">
              <span className="inline-flex items-center gap-1"><span className="w-1.5 h-1.5 rounded-full bg-emerald-500" />在机</span>
              {device.dialysisNo && <span className="truncate">透{device.dialysisNo}</span>}
              <span className="shrink-0">{device.mode || 'HD'}</span>
            </div>
          </div>
        </div>
        <div className="flex shrink-0 gap-0.5">
          <button onClick={() => onOpenModal(device, 'ORDERS')} className="p-1.5 rounded-lg text-slate-300 hover:text-blue-500 hover:bg-blue-50 transition-colors" title={t('action.viewOrders')}>
            <ClipboardList size={15} />
          </button>
          <button onClick={() => onOpenModal(device, 'SUMMARY')} className="p-1.5 rounded-lg text-slate-300 hover:text-indigo-500 hover:bg-indigo-50 transition-colors" title={t('action.writeSummary')}>
            <FileEdit size={15} />
          </button>
        </div>
      </div>

      <button onClick={() => device.status !== 'offline' && onOpenModal(device, 'COMPREHENSIVE')}
        className="w-full text-left bg-slate-50/80 rounded-[12px] px-3 py-2 mb-2 hover:bg-white hover:shadow-sm transition-colors border border-transparent hover:border-blue-200">
        <div className="flex justify-between items-center text-xs mb-0.5">
          <span className="flex items-center gap-1 font-medium"><Heart size={11} className="text-red-500" /><span className="text-red-500">MAP</span>
            <b className={`font-mono ${valueClass(alertLevelFor(device, 'map')) || 'text-slate-700'}`}>{mapVal > 0 ? mapVal : '--'}</b></span>
          <span className="flex items-center gap-1 font-medium"><Activity size={11} className="text-emerald-500" /><span className="text-emerald-500">{t('card.hr')}</span>
            <b className={`font-mono ${valueClass(alertLevelFor(device, 'heartRate')) || 'text-slate-700'}`}>{formatPositive(device.vitals.hr)}</b></span>
        </div>
        <MiniTrend device={device} />
        <div className="flex justify-between text-[9px] text-slate-300 leading-none"><span>上机</span><span>计划下机</span></div>
      </button>

      <button onClick={() => device.status !== 'offline' && onOpenModal(device, 'PRESCRIPTION')} className="w-full text-left mb-2">
        <div className="flex justify-between items-center text-xs mb-1 font-medium">
          <span className="text-slate-500 flex items-center gap-1">
            <Droplet size={10} className="text-blue-400" />
            {device.vitals.ufGoal > 0 ? `${device.vitals.ufVolume.toFixed(2)} / ${device.vitals.ufGoal.toFixed(1)} L` : '暂无超滤数据'}
          </span>
          <span className={`font-bold ${valueClass(alertLevelFor(device, 'ufr')) || 'text-blue-600'}`}>{ufPercent}%</span>
        </div>
        <div className="w-full bg-slate-100 h-2 rounded-full overflow-hidden shadow-inner">
          <div className={`h-full rounded-full transition-all duration-500 ${status === 'danger' ? 'bg-red-500' : 'bg-blue-500'}`}
            style={{ width: `${Math.min(100, ufPercent)}%` }} />
        </div>
      </button>

      {device.rnaCompletion?.available ? (
        <div className="mb-2">
          <div className="flex justify-between items-center text-xs mb-1 font-medium">
            <span className="text-cyan-600 flex items-center gap-1"><Beaker size={10} className="text-cyan-500" />钠清除 RNa</span>
            <span className="font-bold text-cyan-600">{device.rnaCompletion.percent}%</span>
          </div>
          <div className="w-full bg-slate-100 h-2 rounded-full overflow-hidden shadow-inner">
            <div className="h-full rounded-full bg-cyan-500 transition-all duration-500" style={{ width: `${Math.min(100, device.rnaCompletion.percent)}%` }} />
          </div>
        </div>
      ) : (
        <div className="text-meta text-slate-300 mb-2 flex items-center gap-1"><Beaker size={10} />钠清除 RNa 待数据</div>
      )}

      <div className="flex items-center justify-between text-xs text-slate-500 mb-2 gap-2">
        <span className="inline-flex items-center gap-1 shrink-0">
          <Clock size={10} className="text-slate-400" />
          {time ? `${time.elapsed}/${time.planned}` : '--'}
        </span>
        <span className="inline-flex items-center gap-1">
          <Gauge size={10} className="text-slate-400" /> VP
          <b className={`font-mono ${valueClass(alertLevelFor(device, 'vp'))}`}>{vp > 0 ? Math.round(vp) : '--'}</b>
        </span>
        <span className="inline-flex items-center gap-1">
          <Beaker size={10} className="text-slate-400" /> Na
          <b className={`font-mono ${valueClass(alertLevelFor(device, 'dialysateNa'))}`}>{naVal > 0 ? naVal : '--'}</b>
        </span>
      </div>

      <div className="flex items-center flex-wrap gap-1.5 mt-auto pt-2 border-t border-slate-100 min-h-[26px]">
        {device.doubleChecked === false && (
          <span
            className="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-meta font-bold bg-orange-100 text-orange-600"
            title={`上机前双人核对：首次核对${device.firstChecked ? '已完成' : '未完成'} · 二次核对${device.secondChecked ? '已完成' : '未完成'}`}
          >
            <AlertTriangle size={11} /> 未双核
          </span>
        )}
        {device.idhRisk?.available && device.idhRisk.level !== 'low' && (
          <span className={`inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-meta font-bold ${device.idhRisk.level === 'high' ? 'bg-red-100 text-red-600' : 'bg-amber-100 text-amber-600'}`}>
            <AlertTriangle size={11} /> IDH{device.idhRisk.level === 'high' ? '高' : '中'}
          </span>
        )}
        {(device.alerts || []).map((a, i) => (
          <span key={i} className={`inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-meta font-medium ${a.level === 'danger' ? 'bg-red-100 text-red-600' : 'bg-amber-100 text-amber-600'}`}>
            <Bell size={10} /> {MONITOR_METRIC_LABELS[a.metric] || a.metric}
          </span>
        ))}
        {status === 'active' && device.doubleChecked !== false && !(device.idhRisk?.available && device.idhRisk.level !== 'low') && (device.alerts || []).length === 0 && (
          <span className="text-meta text-slate-300">指标正常</span>
        )}
      </div>
    </div>
  )
})
ActiveBedCard.displayName = 'ActiveBedCard'

const OfflineBedCard = memo(({ device }: { device: MonitorDevice }) => (
  <div className="rounded-[14px] bg-white border border-slate-100 p-4 flex flex-col justify-between min-h-[110px] opacity-50 grayscale-[0.2]">
    <div className="flex items-center gap-3">
      <div className="w-10 h-10 rounded-[12px] bg-slate-200 flex items-center justify-center text-slate-400 font-bold text-sm shrink-0">
        {device.bedNumber}
      </div>
      <div>
        <p className="font-bold text-sm text-slate-500">{device.bedNumber}</p>
        <p className="text-xs text-slate-400 mt-0.5">离线</p>
      </div>
    </div>
    <div className="flex items-center mt-2 pt-2 border-t border-slate-50">
      <span className="text-xs text-slate-300 flex items-center gap-1">
        <Wifi size={10} className="text-slate-300" /> 离线
      </span>
    </div>
  </div>
))
OfflineBedCard.displayName = 'OfflineBedCard'

export default function StatusGrid({ devices, loading, loadError, onOpenModal, onReload }: StatusGridProps) {
  return (
    <div className="flex-1 overflow-y-auto pb-10" style={{ scrollbarWidth: 'thin' }}>
      {loading && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5 gap-4">
          {Array.from({ length: 10 }).map((_, i) => (
            <div key={i} className="rounded-[14px] border border-slate-100 bg-white p-4 h-48 animate-pulse">
              <div className="flex items-center gap-3 mb-3">
                <div className="w-10 h-10 bg-slate-100 rounded-[12px] shrink-0" />
                <div className="flex-1 space-y-1.5">
                  <div className="h-3 bg-slate-100 rounded w-3/4" />
                  <div className="h-2 bg-slate-100 rounded w-1/2" />
                </div>
              </div>
              <div className="grid grid-cols-2 gap-3 py-2 mb-3">
                {[1, 2].map(j => <div key={j} className="h-12 bg-slate-50 rounded-[10px]" />)}
              </div>
              <div className="h-16 bg-slate-50 rounded-[12px]" />
            </div>
          ))}
        </div>
      )}

      {!loading && loadError && (
        <div className="flex flex-col items-center justify-center py-24 text-center">
          <div className="w-16 h-16 bg-red-50 rounded-[14px] flex items-center justify-center mb-4">
            <AlertOctagon size={32} className="text-red-500" />
          </div>
          <p className="text-base font-bold text-slate-700 mb-1">{loadError}</p>
          <button onClick={onReload} className="mt-4 px-4 py-2 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700">
            刷新页面
          </button>
        </div>
      )}

      {!loading && !loadError && devices.length === 0 && (
        <div className="flex flex-col items-center justify-center py-24 text-center">
          <div className="w-16 h-16 bg-slate-100 rounded-[14px] flex items-center justify-center mb-4">
            <Monitor size={32} className="text-slate-400" />
          </div>
          <p className="text-base font-bold text-slate-500 mb-1">暂无设备数据</p>
          <p className="text-sm text-slate-400">请先在设备管理中录入透析机信息</p>
        </div>
      )}

      {/* 设备网格（分区固定布局，不动态重排 · Q3） */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5 gap-4" style={{ contain: 'layout style' }}>
        {!loading && !loadError && devices.map((device) => {
          const bedStatus = classifyBedStatus(device)
          if (bedStatus === 'empty') return <EmptyBedCard key={device.id} device={device} onOpenModal={onOpenModal} />
          if (bedStatus === 'offline') return <OfflineBedCard key={device.id} device={device} />
          return <ActiveBedCard key={device.id} device={device} onOpenModal={onOpenModal} />
        })}
      </div>
    </div>
  )
}
