import { useState, useMemo, useCallback } from 'react'

// 状态过滤选项
export type StatusFilter = 'all' | 'enabled' | 'disabled'

interface UseSearchOptions<T> {
  data: T[]
  searchFields: (keyof T)[]
  statusField?: keyof T  // 状态字段名，如 'isEnabled'
}

interface UseSearchReturn<T> {
  keyword: string
  setKeyword: (value: string) => void
  statusFilter: StatusFilter
  setStatusFilter: (value: StatusFilter) => void
  filteredData: T[]
  clearSearch: () => void
  clearAll: () => void
  hasFilter: boolean
  hasStatusFilter: boolean
  resultCount: number
}

/**
 * 通用搜索 Hook
 * 支持多字段搜索 + 状态过滤
 */
export function useSearch<T extends object>({
  data,
  searchFields,
  statusField
}: UseSearchOptions<T>): UseSearchReturn<T> {
  const [keyword, setKeyword] = useState('')
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('all')

  const filteredData = useMemo(() => {
    let result = data

    // 状态过滤
    if (statusField && statusFilter !== 'all') {
      result = result.filter(item => {
        const isEnabled = item[statusField] as boolean
        return statusFilter === 'enabled' ? isEnabled : !isEnabled
      })
    }

    // 关键词搜索
    const trimmed = keyword.trim().toLowerCase()
    if (trimmed) {
      result = result.filter(item =>
        searchFields.some(field => {
          const value = item[field]
          if (value == null) return false
          return String(value).toLowerCase().includes(trimmed)
        })
      )
    }

    return result
  }, [data, keyword, searchFields, statusField, statusFilter])

  const clearSearch = useCallback(() => {
    setKeyword('')
  }, [])

  const clearAll = useCallback(() => {
    setKeyword('')
    setStatusFilter('all')
  }, [])

  const hasStatusFilter = statusFilter !== 'all'
  const hasKeywordFilter = keyword.trim().length > 0

  return {
    keyword,
    setKeyword,
    statusFilter,
    setStatusFilter,
    filteredData,
    clearSearch,
    clearAll,
    hasFilter: hasKeywordFilter || hasStatusFilter,
    hasStatusFilter,
    resultCount: filteredData.length
  }
}
