import { useEffect, useMemo, useState } from 'react'
import { dictCache } from '@/services/dictApi'

export interface DictOption {
  value: string
  label: string
}

function normalizeOptions(options: DictOption[]): DictOption[] {
  const seen = new Set<string>()
  return options.filter((option) => {
    if (!option.value || seen.has(option.value)) return false
    seen.add(option.value)
    return true
  })
}

/**
 * 加载字典下拉选项；字典接口失败时保留 fallback，避免业务表单出现空下拉。
 */
export function useDictOptions(
  typeCode: string,
  fallback: DictOption[] = [],
  currentValues: Array<string | undefined> = []
): DictOption[] {
  const [loaded, setLoaded] = useState<{ typeCode: string; options: DictOption[] } | null>(null)

  useEffect(() => {
    if (!typeCode) return

    let cancelled = false
    dictCache.getOptions(typeCode)
      .then((options) => {
        if (!cancelled) {
          setLoaded({ typeCode, options })
        }
      })
      .catch(() => {
        if (!cancelled) {
          setLoaded({ typeCode, options: [] })
        }
      })

    return () => {
      cancelled = true
    }
  }, [typeCode])

  return useMemo(() => {
    const options = loaded?.typeCode === typeCode && loaded.options.length > 0 ? loaded.options : fallback
    const merged = [...options]
    currentValues.forEach((value) => {
      if (value && !merged.some((option) => option.value === value)) {
        merged.unshift({ value, label: value })
      }
    })
    return normalizeOptions(merged)
  }, [currentValues, fallback, loaded, typeCode])
}
