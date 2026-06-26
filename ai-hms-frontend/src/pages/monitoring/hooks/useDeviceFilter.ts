import { useState, useMemo } from 'react'
import type { MonitorDevice } from '@/types/original'
import { type StatusFilter, classifyBedStatus } from '../types'

export function useDeviceFilter(devices: MonitorDevice[]) {
  const [activeZone, setActiveZone] = useState('ALL')
  const [searchTerm, setSearchTerm] = useState('')
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('ALL')

  const filteredDevices = useMemo(() => {
    return devices.filter((d) => {
      const zoneMatch = activeZone === 'ALL' || d.bedNumber.startsWith(activeZone)
      const searchMatch =
        (d.patientName || '').includes(searchTerm) || d.bedNumber.includes(searchTerm)
      const s = classifyBedStatus(d)
      const statusMatch =
        statusFilter === 'ALL' ||
        (statusFilter === 'alerts' ? s === 'warning' || s === 'danger' : s === statusFilter)
      return zoneMatch && searchMatch && statusMatch
    })
  }, [devices, activeZone, searchTerm, statusFilter])

  return { filteredDevices, activeZone, setActiveZone, searchTerm, setSearchTerm, statusFilter, setStatusFilter }
}
