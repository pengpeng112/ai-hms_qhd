import { useState, useEffect } from 'react'
import type { MonitorDevice } from '@/types/original'
import { restApi } from '@/services/restClient'
import { buildDeviceAssignments, toMonitorDevice, ensureDeviceCache } from '../types'

export function useMonitoringData() {
  const [devices, setDevices] = useState<MonitorDevice[]>([])
  const [loading, setLoading] = useState(true)
  const [loadError, setLoadError] = useState<string | null>(null)

  useEffect(() => {
    const loadMonitoringData = async () => {
      setLoading(true)
      setLoadError(null)
      try {
        const [devicesResult, patientsResult] = await Promise.allSettled([
          restApi.getDeviceList({ pageSize: 200 }),
          restApi.getPatientList({ page: 1, pageSize: 500 }),
        ])

        if (devicesResult.status !== 'fulfilled') {
          setLoadError('设备列表加载失败，请检查网络连接或联系管理员')
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
        setDevices(mapped)
      } catch (err) {
        setLoadError('数据加载异常，请刷新页面重试')
        console.error('[Monitoring] 加载失败', err)
      } finally {
        setLoading(false)
      }
    }

    void loadMonitoringData()
  }, [])

  return { devices, loading, loadError }
}
