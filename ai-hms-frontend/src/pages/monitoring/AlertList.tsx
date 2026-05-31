import { AlertTriangle, AlertCircle, Wifi, Clock } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import type { MonitorDevice } from '@/types/original'
import type { ModalType } from './types'

interface AlertListProps {
  devices: MonitorDevice[]
  onOpenModal: (device: MonitorDevice, type: ModalType) => void
}

export default function AlertList({ devices, onOpenModal }: AlertListProps) {
  const { t: tRaw } = useTranslation('monitoring')
  const t = tRaw as (key: string) => string
  const alertDevices = devices.filter(d => d.status === 'alarm' || d.status === 'warning')

  if (alertDevices.length === 0) return null

  return (
    <div className="mt-4">
      <h3 className="text-sm font-bold text-foreground mb-3 flex items-center gap-2">
        <AlertTriangle size={16} className="text-state-alert" />
        {t('alert.title') || '报警列表'} ({alertDevices.length})
      </h3>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
        {alertDevices.map(device => (
          <div
            key={device.id}
            onClick={() => onOpenModal(device, 'COMPREHENSIVE')}
            className={`rounded-md border-l-4 p-3 cursor-pointer hover:shadow-md transition-shadow ${
              device.status === 'alarm'
                ? 'border-l-state-alert bg-state-alert-bg'
                : 'border-l-state-waiting bg-state-waiting-bg'
            }`}
          >
            <div className="flex items-center justify-between mb-1">
              <span className="font-bold text-sm text-foreground">
                <span className="mr-2 px-1.5 py-0.5 rounded text-white text-xs font-bold bg-slate-800">
                  {device.bedNumber}
                </span>
                {device.patientName || t('card.idle')}
              </span>
              {device.status === 'alarm' ? (
                <AlertTriangle size={16} className="text-state-alert" />
              ) : (
                <AlertCircle size={16} className="text-state-waiting" />
              )}
            </div>
            <div className="flex items-center text-meta text-foreground-muted gap-3">
              <span className="flex items-center gap-1">
                <Wifi size={10} /> {device.mode || '--'}
              </span>
              <span className="flex items-center gap-1">
                <Clock size={10} /> {device.timeRemaining}
              </span>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
