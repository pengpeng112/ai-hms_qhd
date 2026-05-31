import { useState, useMemo } from 'react'
import type { MonitorDevice } from '@/types/original'

export function useDeviceFilter(devices: MonitorDevice[]) {
  const [activeZone, setActiveZone] = useState('ALL')
  const [searchTerm, setSearchTerm] = useState('')

  const filteredDevices = useMemo(() => {
    return devices.filter((d) => {
      const zoneMatch = activeZone === 'ALL' || d.bedNumber.startsWith(activeZone)
      const searchMatch =
        (d.patientName || '').includes(searchTerm) || d.bedNumber.includes(searchTerm)
      return zoneMatch && searchMatch
    })
  }, [devices, activeZone, searchTerm])

  return { filteredDevices, activeZone, setActiveZone, searchTerm, setSearchTerm }
}
