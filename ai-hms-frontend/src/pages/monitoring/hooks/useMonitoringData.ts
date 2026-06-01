import { useState, useEffect, useCallback } from 'react'
import type { MonitorDevice } from '@/types/original'
import { restApi } from '@/services/restClient'
import { getMonitoringLiveData, type RestMonitoringLiveData } from '@/services/monitoringApi'
import { buildDeviceAssignments, toMonitorDevice, ensureDeviceCache } from '../types'

function applyLiveData(devices: MonitorDevice[], liveData: RestMonitoringLiveData[]): MonitorDevice[] {
  const byBedName = new Map<string, RestMonitoringLiveData>()
  liveData.forEach((d) => {
    if (d.bedName) byBedName.set(d.bedName, d)
  })
  return devices.map((device) => {
    const ld = byBedName.get(device.bedNumber)
    if (!ld) return device
    return {
      ...device,
      patientName: ld.patientName || device.patientName,
      mode: ld.dialysisMode || device.mode,
      timeRemaining: ld.startTime && ld.estimatedDuration
        ? `${Math.max(0, Math.round(ld.estimatedDuration - (Date.now() - new Date(ld.startTime).getTime()) / 60000))}min`
        : device.timeRemaining,
      vitals: {
        sbp: ld.sbp || 0,
        dbp: ld.dbp || 0,
        hr: ld.heartRate || 0,
        bf: ld.bf || 0,
        tmp: ld.tmp || 0,
        ufGoal: ld.ufGoal || 0,
        ufVolume: ld.ufVolume || 0,
        conductivity: ld.conductivity || 0,
        temp: ld.machineTmp || 0,
      },
    }
  })
}

export function useMonitoringData() {
  const [devices, setDevices] = useState<MonitorDevice[]>([])
  const [loading, setLoading] = useState(true)
  const [loadError, setLoadError] = useState<string | null>(null)

  const loadAll = useCallback(async () => {
    setLoading(true)
    setLoadError(null)
    try {
      const [devicesResult, patientsResult] = await Promise.allSettled([
        restApi.getDeviceList({ pageSize: 200 }),
        restApi.getPatientList({ page: 1, pageSize: 500 }),
      ])

      if (devicesResult.status !== 'fulfilled') {
        setLoadError('Device list load failed')
        setDevices([])
        return
      }

      const assignments =
        patientsResult.status === 'fulfilled'
          ? buildDeviceAssignments(patientsResult.value.data.items || [])
          : new Map()

      const mapped = devicesResult.value.map((item) =>
        toMonitorDevice(item, assignments.get(item.bedNumber || item.name)),
      )
      mapped.forEach(ensureDeviceCache)

      try {
        const liveData = await getMonitoringLiveData()
        setDevices(applyLiveData(mapped, liveData))
      } catch {
        setDevices(mapped)
      }
    } catch (err) {
      setLoadError('Data load error')
      console.error('[Monitoring]', err)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    const interval = setInterval(async () => {
      try {
        const liveData = await getMonitoringLiveData()
        setDevices((prev) => applyLiveData(prev, liveData))
      } catch {
        // ignore polling errors
      }
    }, 30000)

    return () => clearInterval(interval)
  }, [])

  useEffect(() => {
    void loadAll()
  }, [loadAll])

  return { devices, loading, loadError }
}
