// usePagination.ts - 分页逻辑 hook

import { useState, useCallback, useMemo } from 'react'

interface UsePaginationProps<T> {
  data: T[]
  pageSize?: number
}

interface PaginationState {
  currentPage: number
  pageSize: number
}

export function usePagination<T>({ data, pageSize = 10 }: UsePaginationProps<T>) {
  const [state, setState] = useState<PaginationState>({
    currentPage: 1,
    pageSize
  })

  // 总页数
  const totalPages = useMemo(() => {
    return Math.ceil(data.length / state.pageSize) || 1
  }, [data.length, state.pageSize])

  // 当前页数据
  const currentPageData = useMemo(() => {
    const startIndex = (state.currentPage - 1) * state.pageSize
    const endIndex = startIndex + state.pageSize
    return data.slice(startIndex, endIndex)
  }, [data, state.currentPage, state.pageSize])

  // 显示信息
  const displayInfo = useMemo(() => {
    const startIndex = (state.currentPage - 1) * state.pageSize
    const endIndex = Math.min(startIndex + state.pageSize, data.length)
    return {
      startIndex: startIndex + 1,
      endIndex: endIndex,
      total: data.length
    }
  }, [data.length, state.currentPage, state.pageSize])

  // 翻页函数
  const goToPage = useCallback((page: number) => {
    setState(prev => ({ ...prev, currentPage: Math.max(1, Math.min(page, totalPages)) }))
  }, [totalPages])

  const nextPage = useCallback(() => {
    setState(prev => ({ ...prev, currentPage: Math.min(prev.currentPage + 1, totalPages) }))
  }, [totalPages])

  const prevPage = useCallback(() => {
    setState(prev => ({ ...prev, currentPage: Math.max(prev.currentPage - 1, 1) }))
  }, [])

  const firstPage = useCallback(() => {
    setState(prev => ({ ...prev, currentPage: 1 }))
  }, [])

  const lastPage = useCallback(() => {
    setState(prev => ({ ...prev, currentPage: totalPages }))
  }, [totalPages])

  // 重置到第一页（当数据变化时）
  const reset = useCallback(() => {
    setState(prev => ({ ...prev, currentPage: 1 }))
  }, [])

  return {
    currentPage: state.currentPage,
    pageSize: state.pageSize,
    totalPages,
    currentPageData,
    displayInfo,
    goToPage,
    nextPage,
    prevPage,
    firstPage,
    lastPage,
    reset
  }
}
