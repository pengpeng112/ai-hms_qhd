import { useCallback, useEffect, useState } from 'react'
import { message } from 'antd'
import {
  getThresholds, saveThresholds, resetThresholds,
} from '@/services/monitoringThresholdApi'
import type { ThresholdPayload } from '@/services/monitoringThresholdApi'

export function useThresholds() {
  const [data, setData] = useState<ThresholdPayload | null>(null)
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      setData(await getThresholds())
    } catch (e) {
      message.error(e instanceof Error ? e.message : '加载失败')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { void load() }, [load])

  const save = useCallback(async (payload: ThresholdPayload) => {
    setSaving(true)
    try {
      await saveThresholds(payload)
      message.success('已保存')
      await load()
    } catch (e) {
      message.error(e instanceof Error ? e.message : '保存失败')
    } finally {
      setSaving(false)
    }
  }, [load])

  const reset = useCallback(async () => {
    setSaving(true)
    try {
      await resetThresholds()
      message.success('已恢复默认')
      await load()
    } catch (e) {
      message.error(e instanceof Error ? e.message : '恢复默认失败')
    } finally {
      setSaving(false)
    }
  }, [load])

  return { data, loading, saving, load, save, reset }
}
