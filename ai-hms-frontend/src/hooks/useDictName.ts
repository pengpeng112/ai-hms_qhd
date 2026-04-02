/**
 * useDictName - 字典名称查询 Hook
 * 用于将字典代码转换为显示名称
 */

import { useEffect, useMemo, useState } from 'react'
import { dictCache } from '@/services/dictApi'

/**
 * 获取单个字典项的名称
 * @param typeCode 字典类型代码
 * @param code 字典项代码
 * @returns 字典项名称（加载中返回代码本身）
 */
export function useDictName(typeCode: string, code: string | undefined): string {
  const [resolved, setResolved] = useState<{ code: string; name: string } | null>(null)

  useEffect(() => {
    if (!typeCode || !code) return

    let cancelled = false

    dictCache.getNameByCode(typeCode, code)
      .then(result => {
        if (!cancelled) {
          setResolved({ code, name: result || code })
        }
      })
      .catch(() => {
        if (!cancelled) {
          setResolved({ code, name: code })
        }
      })

    return () => {
      cancelled = true
    }
  }, [typeCode, code])

  if (!typeCode || !code) return code || ''
  return resolved?.code === code ? resolved.name : code
}

/**
 * 获取字典名称映射表
 * @param typeCode 字典类型代码
 * @returns Map<code, name>
 */
export function useDictNameMap(typeCode: string): Map<string, string> {
  const [resolved, setResolved] = useState<{ typeCode: string; map: Map<string, string> } | null>(null)

  useEffect(() => {
    if (!typeCode) return

    let cancelled = false

    dictCache.getNameMap(typeCode)
      .then(map => {
        if (!cancelled) {
          setResolved({ typeCode, map })
        }
      })
      .catch(() => {
        if (!cancelled) {
          setResolved({ typeCode, map: new Map() })
        }
      })

    return () => {
      cancelled = true
    }
  }, [typeCode])

  if (!typeCode) return new Map()
  return resolved?.typeCode === typeCode ? resolved.map : new Map()
}

/**
 * 批量获取多个字典类型的名称映射
 * @param typeCodes 字典类型代码数组
 * @returns Record<typeCode, Map<code, name>>
 */
export function useDictNameMaps(typeCodes: string[]): Record<string, Map<string, string>> {
  const [nameMaps, setNameMaps] = useState<Record<string, Map<string, string>>>({})
  const stableKey = useMemo(
    () => Array.from(new Set(typeCodes.map(code => code?.trim()).filter(Boolean) as string[])).sort().join('|'),
    [typeCodes]
  )
  const normalizedTypeCodes = useMemo(() => stableKey ? stableKey.split('|') : [], [stableKey])

  useEffect(() => {
    if (!stableKey) return

    let cancelled = false

    Promise.all(
      normalizedTypeCodes.map(async typeCode => {
        try {
          const map = await dictCache.getNameMap(typeCode)
          return { typeCode, map }
        } catch {
          return { typeCode, map: new Map<string, string>() }
        }
      })
    ).then(results => {
      if (!cancelled) {
        const maps: Record<string, Map<string, string>> = {}
        results.forEach(({ typeCode, map }) => {
          maps[typeCode] = map
        })
        setNameMaps(maps)
      }
    })

    return () => {
      cancelled = true
    }
  }, [stableKey, normalizedTypeCodes])

  if (!normalizedTypeCodes.length) return {}
  const output: Record<string, Map<string, string>> = {}
  normalizedTypeCodes.forEach((typeCode) => {
    output[typeCode] = nameMaps[typeCode] || new Map()
  })
  return output
}

/**
 * 工具函数：从映射表中获取名称，如果不存在则返回代码本身
 */
export function getNameFromMap(map: Map<string, string>, code: string | undefined): string {
  if (!code) return ''
  return map.get(code) || code
}
