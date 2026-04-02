/**
 * useOutcomeDict - 转归字典 Hook
 * 提供转归类型和原因的联动查询功能
 * 后端使用单一 OUTCOME 字典类型，通过 parentCode 区分一级（类型）和二级（原因）
 */

import { useCallback, useMemo, useState } from 'react'
import { dictCache, DICT_TYPES, type DictItem } from '@/services/dictApi'

export function useOutcomeDict() {
  const [types, setTypes] = useState<DictItem[]>([])
  const [reasons, setReasons] = useState<DictItem[]>([])
  const [loading, setLoading] = useState(true)

  // 初始化加载数据
  const loadDicts = useCallback(async () => {
    setLoading(true)
    try {
      const allItems = await dictCache.getItems(DICT_TYPES.OUTCOME)
      // 一级分类：parentCode 为空 → 转归类型（在科/转出）
      const typeItems = allItems.filter(item => !item.parentCode)
      // 二级分类：parentCode 非空 → 转归原因（具体原因）
      const reasonItems = allItems.filter(item => !!item.parentCode)

      setTypes(typeItems.map(item => ({
        ...item,
        code: String(item.code ?? ''),
      })))
      setReasons(reasonItems.map(item => ({
        ...item,
        code: String(item.code ?? ''),
        parentCode: item.parentCode != null ? String(item.parentCode) : '',
      })))
    } catch (error) {
      console.error('[useOutcomeDict] 加载转归字典失败:', error)
      setTypes([])
      setReasons([])
    } finally {
      setLoading(false)
    }
  }, [])

  // 根据转归类型 code 获取对应的原因列表
  const getReasonsByType = useMemo(() => {
    return (typeCode: string): DictItem[] => {
      if (!typeCode) return []
      return reasons.filter(r => r.parentCode === typeCode)
    }
  }, [reasons])

  // 根据 code 获取转归类型名称
  const getTypeName = useMemo(() => {
    return (code: string): string => {
      const item = types.find(t => t.code === code)
      return item?.name || code
    }
  }, [types])

  // 根据 code 获取转归原因名称
  const getReasonName = useMemo(() => {
    return (code: string): string => {
      const item = reasons.find(r => r.code === code)
      return item?.name || code
    }
  }, [reasons])

  // 获取转归类型选项（用于 Select）
  const typeOptions = useMemo(() => {
    return types.map(t => ({ value: t.code, label: t.name }))
  }, [types])

  // 获取转归原因选项（用于 Select，按类型过滤）
  const getReasonOptions = useMemo(() => {
    return (typeCode: string) => {
      return getReasonsByType(typeCode).map(r => ({ value: r.code, label: r.name }))
    }
  }, [getReasonsByType])

  return {
    types,
    reasons,
    loading,
    loadDicts,
    getReasonsByType,
    getTypeName,
    getReasonName,
    typeOptions,
    getReasonOptions,
  }
}

export default useOutcomeDict
