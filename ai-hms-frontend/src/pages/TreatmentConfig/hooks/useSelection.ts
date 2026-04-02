// useSelection.ts - 批量选择 hook

import { useState, useCallback, useMemo } from 'react'

interface UseSelectionProps<T> {
  data: T[]
  keyField?: string
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function useSelection<T extends Record<string, any>>({
  data,
  keyField = 'id'
}: UseSelectionProps<T>) {
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set())

  // 当前页所有项是否都被选中
  const allSelected = useMemo(() => {
    if (data.length === 0) return false
    const currentPageIds = new Set(data.map(item => String(item[keyField])))
    return Array.from(currentPageIds).every(id => selectedIds.has(id))
  }, [data, selectedIds, keyField])

  // 当前页是否有部分选中
  const someSelected = useMemo(() => {
    const currentPageIds = new Set(data.map(item => String(item[keyField])))
    return Array.from(currentPageIds).some(id => selectedIds.has(id))
  }, [data, selectedIds, keyField])

  // 当前页选中的项
  const currentPageSelected = useMemo(() => {
    return data.filter(item => selectedIds.has(String(item[keyField])))
  }, [data, selectedIds, keyField])

  // 切换选中状态 - 支持 string 或 number 类型的 ID
  const toggleSelection = useCallback((id: string | number) => {
    const idStr = String(id)
    setSelectedIds(prev => {
      const newSet = new Set(prev)
      if (newSet.has(idStr)) {
        newSet.delete(idStr)
      } else {
        newSet.add(idStr)
      }
      return newSet
    })
  }, [])

  // 检查是否选中 - 支持 string 或 number 类型的 ID
  const isSelected = useCallback((id: string | number) => {
    return selectedIds.has(String(id))
  }, [selectedIds])

  // 全选/取消全选当前页
  const toggleSelectAll = useCallback(() => {
    setSelectedIds(prev => {
      const newSet = new Set(prev)
      const currentPageIds = data.map(item => String(item[keyField]))

      if (allSelected) {
        // 取消全选当前页
        currentPageIds.forEach(id => newSet.delete(id))
      } else {
        // 全选当前页
        currentPageIds.forEach(id => newSet.add(id))
      }
      return newSet
    })
  }, [data, allSelected, keyField])

  // 清空所有选择
  const clearSelection = useCallback(() => {
    setSelectedIds(new Set())
  }, [])

  // 获取选中的完整数据
  const getSelectedItems = useCallback(() => {
    return data.filter(item => selectedIds.has(String(item[keyField])))
  }, [data, selectedIds, keyField])

  // 选中数量
  const selectedCount = selectedIds.size

  return {
    selectedIds,
    selectedCount,
    allSelected,
    someSelected,
    currentPageSelected,
    toggleSelection,
    isSelected,
    toggleSelectAll,
    clearSelection,
    getSelectedItems
  }
}
