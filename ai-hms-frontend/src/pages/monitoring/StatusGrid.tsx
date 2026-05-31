import { memo, useRef, useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import type { MonitorDevice } from '@/types/original'
import {
  Monitor, Wifi, Clock, ClipboardList, FileEdit, Droplet, AlertOctagon, AlertTriangle,
} from 'lucide-react'
import { AreaChart, Area, Line } from 'recharts'
import { cachedGraphData, formatPositive, formatBloodPressure } from './types'
import type { ModalType } from './types'

// Mini 图表组件
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
  getStatusColor: (status: string) => string
  onOpenModal: (device: MonitorDevice, type: ModalType) => void
  onReload: () => void
}

export default function StatusGrid({ devices, loading, loadError, getStatusColor, onOpenModal, onReload }: StatusGridProps) {
  const { t } = useTranslation('monitoring')

  return (
    <div className="flex-1 overflow-y-auto pr-2 pb-10">
      {/* 加载骨架 */}
      {loading && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5 gap-4">
          {Array.from({ length: 10 }).map((_, i) => (
            <div key={i} className="rounded-lg border-2 border-gray-200 bg-gray-50 p-3 h-48 animate-pulse">
              <div className="flex items-center gap-2 mb-3">
                <div className="w-8 h-8 bg-gray-200 rounded-lg shrink-0" />
                <div className="flex-1 space-y-1.5">
                  <div className="h-3 bg-gray-200 rounded w-3/4" />
                  <div className="h-2 bg-gray-200 rounded w-1/2" />
                </div>
              </div>
              <div className="grid grid-cols-3 gap-1 py-2 border-t border-b border-gray-200 mb-3">
                {[1,2,3].map(j => <div key={j} className="h-8 bg-gray-200 rounded" />)}
              </div>
              <div className="h-16 bg-gray-200 rounded-lg" />
            </div>
          ))}
        </div>
      )}

      {/* 错误提示 */}
      {!loading && loadError && (
        <div className="flex flex-col items-center justify-center py-24 text-center">
          <div className="w-16 h-16 bg-state-alert-bg rounded-lg flex items-center justify-center mb-4">
            <AlertOctagon size={32} className="text-state-alert" />
          </div>
          <p className="text-base font-bold text-gray-700 mb-1">{loadError}</p>
          <button onClick={onReload} className="mt-4 px-4 py-2 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700">
            刷新页面
          </button>
        </div>
      )}

      {/* 空状态 */}
      {!loading && !loadError && devices.length === 0 && (
        <div className="flex flex-col items-center justify-center py-24 text-center">
          <div className="w-16 h-16 bg-gray-100 rounded-lg flex items-center justify-center mb-4">
            <Monitor size={32} className="text-gray-400" />
          </div>
          <p className="text-base font-bold text-gray-600 mb-1">暂无设备数据</p>
          <p className="text-sm text-gray-400">请先在设备管理中录入透析机信息</p>
        </div>
      )}

      {/* 设备网格 */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5 gap-4" style={{ contain: 'layout style' }}>
        {!loading && !loadError && devices.map((device) => {
          const ufPercent = device.vitals.ufGoal > 0
            ? Math.round((device.vitals.ufVolume / device.vitals.ufGoal) * 100)
            : 0

          return (
            <div
              key={device.id}
              className={`rounded-lg border-2 p-3 flex flex-col shadow-sm relative group ${getStatusColor(device.status)}`}
              style={{ contain: 'layout style paint' }}
            >
              <div className="flex justify-between items-start mb-2 gap-1">
                <div className="flex items-center min-w-0 flex-1">
                  <div className={`w-8 h-8 rounded-lg flex items-center justify-center font-bold text-sm mr-2 shadow-sm shrink-0 ${
                    device.status === 'alarm' ? 'bg-state-alert text-white'
                    : device.status === 'warning' ? 'bg-state-waiting text-white'
                    : 'bg-slate-800 text-white'
                  }`}>
                    {device.bedNumber}
                  </div>
                  <div className="min-w-0 flex-1">
                    <h4 className="font-bold text-gray-900 text-sm flex items-center truncate">
                      <span className="truncate">{device.patientName || t('card.idle')}</span>
                    </h4>
                    <div className="flex items-center text-meta text-gray-500 font-medium">
                      <Wifi size={10} className={`mr-1 shrink-0 ${device.status === 'normal' ? 'text-green-500' : 'text-gray-300'}`} />
                      <span className="text-blue-600">{device.mode || '--'}</span>
                      <span className="mx-1 opacity-40">路</span>
                      <Clock size={10} className="mr-0.5 shrink-0" /> {device.timeRemaining}
                    </div>
                  </div>
                </div>
                <div className="flex shrink-0">
                  <button onClick={() => onOpenModal(device, 'ORDERS')} className="p-1.5 hover:bg-white rounded text-gray-400 hover:text-blue-600 transition-colors" title={t('action.viewOrders')}>
                    <ClipboardList size={16} />
                  </button>
                  {device.patientName && (
                    <button onClick={() => onOpenModal(device, 'SUMMARY')} className="p-1.5 hover:bg-white rounded text-gray-400 hover:text-indigo-600 transition-colors" title={t('action.writeSummary')}>
                      <FileEdit size={16} />
                    </button>
                  )}
                </div>
              </div>

              <div className="grid grid-cols-3 gap-1 mb-2 py-1.5 border-t border-b border-gray-100 text-meta font-medium text-gray-600">
                <div className="flex flex-col">
                  <span className="text-gray-400 scale-90 origin-left">{t('card.dryWeight')}</span>
                  <span className="text-gray-800 font-bold">--</span>
                </div>
                <div className="flex flex-col">
                  <span className="text-gray-400 scale-90 origin-left">{t('card.gainPct')}</span>
                  <span className="text-blue-600 font-bold">--</span>
                </div>
                <div className="flex flex-col items-end">
                  <span className="text-gray-400 scale-90 origin-right">{t('card.vascularAccess')}</span>
                  <span className="text-gray-800 font-bold">--</span>
                </div>
              </div>

              <div onClick={() => device.status !== 'offline' && onOpenModal(device, 'COMPREHENSIVE')} className="bg-white/80 rounded-lg p-2 mb-2 cursor-pointer hover:bg-white hover:shadow-md transition-colors border border-transparent hover:border-blue-200">
                <div className="flex justify-between items-end mb-1 px-1">
                  <div className="flex flex-col">
                    <span className="text-[9px] font-bold text-gray-400 uppercase">{t('card.bp')}</span>
                    <div className="flex items-baseline">
                        <span className={`text-base font-bold font-mono leading-none ${device.status === 'alarm' ? 'text-state-alert' : 'text-gray-900'}`}>
                        {formatBloodPressure(device)}
                      </span>
                      <span className="text-[9px] text-gray-400 ml-0.5 font-bold">mmHg</span>
                    </div>
                  </div>
                  <div className="flex flex-col items-end">
                    <span className="text-[9px] font-bold text-gray-400 uppercase">{t('card.hr')}</span>
                    <div className="flex items-baseline">
                      <span className="text-base font-bold font-mono leading-none text-gray-900">{formatPositive(device.vitals.hr)}</span>
                      <span className="text-[9px] text-gray-400 ml-0.5 font-bold">bpm</span>
                    </div>
                  </div>
                </div>
                <MiniVitalsChart deviceId={device.id} />
              </div>

              <div onClick={() => device.status !== 'offline' && onOpenModal(device, 'PRESCRIPTION')} className="flex-1 cursor-pointer group/pres pt-1">
                <div className="flex justify-between items-center text-meta mb-1 px-1 font-bold">
                  <span className="text-gray-500 flex items-center">
                    <Droplet size={10} className="mr-1 text-blue-500" />
                    {device.vitals.ufGoal > 0 ? `${device.vitals.ufVolume.toFixed(2)}L / ${device.vitals.ufGoal.toFixed(1)}L` : '暂无超滤数据'}
                  </span>
                  <span className="text-blue-700">{ufPercent}%</span>
                </div>
                <div className="w-full bg-gray-200 h-1.5 rounded-full overflow-hidden mb-2 relative shadow-inner">
                  <div className={`h-full rounded-full ${device.status === 'alarm' ? 'bg-state-alert' : 'bg-state-treating'}`} style={{ width: `${Math.min(100, ufPercent)}%` }}></div>
                </div>
              </div>
              {device.status === 'alarm' && (
                <div className="absolute bottom-1 right-1 text-state-alert flex items-center bg-state-alert-bg p-0.5 rounded-full">
                    <AlertTriangle size={14} />
                </div>
              )}
            </div>
          )
        })}
      </div>
    </div>
  )
}
