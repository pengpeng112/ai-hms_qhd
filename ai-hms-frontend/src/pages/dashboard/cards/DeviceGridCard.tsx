import { Monitor } from 'lucide-react'

interface DeviceGridCardProps {
  devices: { Id: string; Name?: string; IDNo?: string; Status?: string }[]
  maxItems?: number
}

export default function DeviceGridCard({ devices, maxItems = 24 }: DeviceGridCardProps) {
  const display = devices.slice(0, maxItems)

  return (
    <div className="grid grid-cols-4 lg:grid-cols-6 gap-2">
      {display.length > 0 ? display.map((eq, i) => {
        const status = (eq.Status || '').toLowerCase()
        const isAlarm = status === 'alarm' || status === 'error' || status === '报警'
        const isOffline = status === 'offline' || status === 'inactive' || status === '离线'

        return (
          <div key={eq.Id} className={`p-2 rounded-md border flex flex-col items-center text-center relative ${
            isAlarm ? 'bg-state-alert-bg border-state-alert' :
            isOffline ? 'bg-state-offline-bg border-gray-200' :
            'bg-surface border-gray-100'
          }`}>
            {/* eslint-disable-next-line no-restricted-syntax -- density:strict 故意小字（设备序号） */}
            <span className="text-[10px] font-bold text-gray-500 absolute top-1 left-1">{i + 1}</span>
            <div className={`mt-3 mb-1 ${isAlarm ? 'text-state-alert' : isOffline ? 'text-gray-400' : 'text-state-finished'}`}>
              <Monitor size={16} />
            </div>
            <span className="text-[8px] text-gray-400 truncate w-full">{eq.Name || eq.IDNo}</span>
          </div>
        )
      }) : (
        Array.from({ length: 12 }).map((_, i) => (
          <div key={i} className="p-2 rounded-md border border-gray-100 bg-surface-sunken flex flex-col items-center text-center relative">
            {/* eslint-disable-next-line no-restricted-syntax -- density:strict 故意小字（设备序号） */}
            <span className="text-[10px] font-bold text-gray-300 absolute top-1 left-1">{i + 1}</span>
            <div className="mt-3 mb-1 text-gray-300">
              <Monitor size={16} />
            </div>
          </div>
        ))
      )}
    </div>
  )
}
