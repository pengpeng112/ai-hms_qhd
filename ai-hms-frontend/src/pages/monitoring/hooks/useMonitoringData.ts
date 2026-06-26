import { useState, useEffect, useCallback, useRef } from 'react'
import type { MonitorDevice } from '@/types/original'
import { restApi } from '@/services/restClient'
import { getMonitoringLiveData, type RestMonitoringLiveData } from '@/services/monitoringApi'
import { getMyDuties } from '@/services/smartScheduleApi'
import { buildDeviceAssignments, toMonitorDevice, ensureDeviceCache } from '../types'

function applyLiveData(devices: MonitorDevice[], liveData: RestMonitoringLiveData[], myWardIds: Set<number>): MonitorDevice[] {
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
      treatmentId: ld.treatmentId || device.treatmentId,
      wardId: ld.wardId || device.wardId,
      isMine: !!ld.wardId && myWardIds.has(ld.wardId),
      timeRemaining: ld.startTime && ld.estimatedDuration
        ? `${Math.max(0, Math.round(ld.estimatedDuration - (Date.now() - new Date(ld.startTime).getTime()) / 60000))}min`
        : device.timeRemaining,
      age: ld.age || device.age,
      dialysisNo: ld.dialysisNo || device.dialysisNo,
      accessType: ld.accessType,
      startTime: ld.startTime || device.startTime,
      estimatedDuration: ld.estimatedDuration || device.estimatedDuration,
      alarmLevel: (ld.alarmLevel as MonitorDevice['alarmLevel']) || undefined,
      alerts: ld.alerts ?? [],
      idhRisk: ld.idhRisk,
      vitalsSeries: ld.vitalsSeries ?? [],
      rnaCompletion: ld.rnaCompletion,
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
        vp: ld.venousPressure || 0,
      },
    }
  })
}

function todayStr(): string {
  const d = new Date()
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

async function fetchMyWardIds(): Promise<Set<number>> {
  try {
    const duties = await getMyDuties(todayStr())
    return new Set((duties || []).map((d) => d.wardId).filter((w): w is number => !!w))
  } catch {
    return new Set()
  }
}

export function useMonitoringData() {
  const [devices, setDevices] = useState<MonitorDevice[]>([])
  const [loading, setLoading] = useState(true)
  const [loadError, setLoadError] = useState<string | null>(null)
  const myWardIdsRef = useRef<Set<number>>(new Set())

  const loadAll = useCallback(async () => {
    setLoading(true)
    setLoadError(null)
    try {
      const [settled, myWardIds] = await Promise.all([
        Promise.allSettled([
          restApi.getDeviceList({ pageSize: 200 }),
          restApi.getPatientList({ page: 1, pageSize: 500 }),
        ]),
        fetchMyWardIds(),
      ])
      const [devicesResult, patientsResult] = settled
      myWardIdsRef.current = myWardIds

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
        setDevices(applyLiveData(mapped, liveData, myWardIdsRef.current))
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
        setDevices((prev) => applyLiveData(prev, liveData, myWardIdsRef.current))
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
