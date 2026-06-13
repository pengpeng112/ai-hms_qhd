import { memo, useRef, useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import type { MonitorDevice } from '@/types/original'
import {
  Monitor, Wifi, Clock, ClipboardList, FileEdit, Droplet, AlertOctagon,
  Heart, Activity,
} from 'lucide-react'
import { AreaChart, Area, Line } from 'recharts'
import { cachedGraphData, formatPositive, formatBloodPressure, classifyBedStatus } from './types'
import type { ModalType } from './types'

const MiniVitalsChart = memo(({ deviceId }: { deviceId: string }) => {
  const data = cachedGraphData.get(deviceId) || []
  const containerRef = useRef<HTMLDivElement>(null)
  const [width, setWidth] = useState(180)
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => {
    const container = containerRef.current
    if (!container) return
    const updateWidth = () => { setWidth(container.clientWidth || 180) }
    updateWidth()
    const resizeObserver = new ResizeObserver(() => {
      if (timeoutRef.current) clearTimeout(timeoutRef.current)
      timeoutRef.current = setTimeout(updateWidth, 200)
    })
    resizeObserver.observe(container)
    return () => {
      resizeObserver.disconnect()
      if (timeoutRef.current) clearTimeout(timeoutRef.current)
    }
  }, [])

  return (
    <div ref={containerRef} className="h-10 w-full mt-1" style={{ contain: 'layout style paint' }}>
      <AreaChart data={data} width={width} height={40}>
        <Area type="monotone" dataKey="sbp" stroke="#ef4444" fill="#fef2f2" strokeWidth={1.5} dot={false} isAnimationActive={false} />
        <Line type="monotone" dataKey="hr" stroke="#10b981" strokeWidth={1.5} dot={false} isAnimationActive={false} />
      </AreaChart>
    </div>
  )
})
MiniVitalsChart.displayName = 'MiniVitalsChart'

interface StatusGridProps {
  devices: MonitorDevice[]
  loading: boolean
  loadError: string | null
  onOpenModal: (device: MonitorDevice, type: ModalType) => void
  onReload: () => void
}

const EmptyBedCard = memo(({ device, onOpenModal }: { device: MonitorDevice; onOpenModal: StatusGridProps['onOpenModal'] }) => {
  const { t } = useTranslation('monitoring')
  return (
    <div className="rounded-[14px] bg-white border border-slate-100 p-4 flex flex-col justify-between min-h-[110px] hover:shadow-md hover:border-blue-200 transition-all duration-200"
      style={{ boxShadow: '0 1px 3px rgba(15,23,42,0.04)' }}>
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-[12px] bg-slate-100 flex items-center justify-center text-slate-400 font-bold text-sm shrink-0">
            {device.bedNumber}
          </div>
          <div>
            <p className="font-bold text-sm text-slate-700">{device.bedNumber}</p>
            <p className="text-xs text-slate-400 mt-0.5">{t('card.idle')}</p>
          </div>
        </div>
      </div>
      <div className="flex items-center justify-between mt-2 pt-2 border-t border-slate-50">
        <span className="text-xs text-slate-400 flex items-center gap-1">
          <Wifi size={10} className={device.status === 'normal' ? 'text-green-400' : 'text-slate-300'} />
          在线
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
  const displayStatus = classifyBedStatus(device)
  const ufPercent = device.vitals.ufGoal > 0
    ? Math.round((device.vitals.ufVolume / device.vitals.ufGoal) * 100)
    : 0

  const borderColor = displayStatus === 'danger' ? 'border-red-300 bg-red-50/30'
    : displayStatus === 'warning' ? 'border-amber-300 bg-amber-50/30'
    : 'border-blue-100 bg-white'

  return (
    <div className={`rounded-[14px] border p-4 flex flex-col min-h-[220px] transition-all duration-200 hover:shadow-lg ${borderColor}`}
      style={{ borderLeftWidth: '4px', boxShadow: displayStatus === 'danger' ? '0 4px 16px rgba(239,68,68,0.10)' : displayStatus === 'warning' ? '0 4px 16px rgba(245,158,11,0.10)' : '0 1px 3px rgba(15,23,42,0.04)' }}>
      {/* Card header */}
      <div className="flex justify-between items-start mb-3">
        <div className="flex items-center min-w-0 flex-1 gap-3">
          <div className={`w-10 h-10 rounded-[12px] flex items-center justify-center font-bold text-sm shrink-0 shadow-sm
            ${displayStatus === 'danger' ? 'bg-red-500 text-white'
            : displayStatus === 'warning' ? 'bg-amber-500 text-white'
            : 'bg-slate-700 text-white'}`}>
            {device.bedNumber}
          </div>
          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2">
              <h4 className="font-bold text-sm text-slate-800 truncate">{device.patientName || '--'}</h4>
              {displayStatus === 'danger' && <span className="px-1.5 py-0.5 rounded text-xs font-bold bg-red-100 text-red-600 shrink-0">异常</span>}
              {displayStatus === 'warning' && <span className="px-1.5 py-0.5 rounded text-xs font-bold bg-amber-100 text-amber-600 shrink-0">预警</span>}
            </div>
            <div className="flex items-center text-xs text-slate-400 mt-0.5 gap-3">
              <span className="flex items-center gap-1"><Wifi size={10} className="text-green-400" /> {device.mode || 'HD'}</span>
              {device.timeRemaining !== '--' && <span className="flex items-center gap-1"><Clock size={10} /> {device.timeRemaining}</span>}
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

      {/* Vitals display */}
      <button onClick={() => device.status !== 'offline' && onOpenModal(device, 'COMPREHENSIVE')}
        className="w-full text-left bg-slate-50/80 rounded-[12px] p-3 mb-3 hover:bg-white hover:shadow-sm transition-colors border border-transparent hover:border-blue-200">
        <div className="flex justify-between items-end mb-1">
          <div className="flex flex-col">
            <span className="text-xs font-medium text-slate-400 uppercase flex items-center gap-1">
              <Heart size={10} className="text-red-400" /> {t('card.bp')}
            </span>
            <div className="flex items-baseline">
              <span className={`text-lg font-bold font-mono leading-none ${displayStatus === 'danger' ? 'text-red-600' : 'text-slate-800'}`}>
                {formatBloodPressure(device)}
              </span>
              <span className="text-xs text-slate-400 ml-1 font-medium">mmHg</span>
            </div>
          </div>
          <div className="flex flex-col items-end">
            <span className="text-xs font-medium text-slate-400 uppercase flex items-center gap-1">
              <Activity size={10} className="text-emerald-400" /> {t('card.hr')}
            </span>
            <div className="flex items-baseline">
              <span className="text-lg font-bold font-mono leading-none text-slate-800">{formatPositive(device.vitals.hr)}</span>
              <span className="text-xs text-slate-400 ml-1 font-medium">bpm</span>
            </div>
          </div>
        </div>
        <MiniVitalsChart deviceId={device.id} />
      </button>

      {/* UF Progress */}
      <button onClick={() => device.status !== 'offline' && onOpenModal(device, 'PRESCRIPTION')}
        className="w-full text-left mt-auto">
        <div className="flex justify-between items-center text-xs mb-1.5 font-medium">
          <span className="text-slate-500 flex items-center gap-1">
            <Droplet size={10} className="text-blue-400" />
            {device.vitals.ufGoal > 0
              ? `${device.vitals.ufVolume.toFixed(2)} / ${device.vitals.ufGoal.toFixed(1)} L`
              : '暂无超滤数据'}
          </span>
          <span className={`font-bold ${displayStatus === 'danger' ? 'text-red-500' : 'text-blue-600'}`}>{ufPercent}%</span>
        </div>
        <div className="w-full bg-slate-100 h-2 rounded-full overflow-hidden shadow-inner">
          <div className={`h-full rounded-full transition-all duration-500 ${displayStatus === 'danger' ? 'bg-red-500' : 'bg-blue-500'}`}
            style={{ width: `${Math.min(100, ufPercent)}%` }} />
        </div>
      </button>
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
      {/* 加载骨架 */}
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

      {/* 错误提示 */}
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

      {/* 空状态 */}
      {!loading && !loadError && devices.length === 0 && (
        <div className="flex flex-col items-center justify-center py-24 text-center">
          <div className="w-16 h-16 bg-slate-100 rounded-[14px] flex items-center justify-center mb-4">
            <Monitor size={32} className="text-slate-400" />
          </div>
          <p className="text-base font-bold text-slate-500 mb-1">暂无设备数据</p>
          <p className="text-sm text-slate-400">请先在设备管理中录入透析机信息</p>
        </div>
      )}

      {/* 设备网格 */}
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
