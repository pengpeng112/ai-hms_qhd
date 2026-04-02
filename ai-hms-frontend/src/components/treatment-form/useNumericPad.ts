// 数字键盘状态管理 Hook

import { useState, useCallback } from 'react'
import type { OpenNumericPad } from './types'

interface NumericPadState {
  open: boolean
  label: string
  value: string
  suffix?: string
  initialValue: string
  onConfirm?: (value: string) => void
}

const sanitizeDecimalInput = (value: string) => {
  if (value === '') return ''
  const normalized = value
    .replace(/[０-９]/g, (d) => String.fromCharCode(d.charCodeAt(0) - 0xff10 + 0x30))
    .replace(/[．。]/g, '.')
  let cleaned = normalized.replace(/[^0-9.]/g, '')
  const firstDot = cleaned.indexOf('.')
  if (firstDot !== -1) {
    cleaned = cleaned.slice(0, firstDot + 1) + cleaned.slice(firstDot + 1).replace(/\./g, '')
  }
  return cleaned
}

export function useNumericPad() {
  const [padState, setPadState] = useState<NumericPadState>({
    open: false,
    label: '',
    value: '',
    suffix: '',
    initialValue: '',
    onConfirm: undefined,
  })

  const openNumericPad: OpenNumericPad = useCallback((payload) => {
    if (payload.disabled) return
    setPadState({
      open: true,
      label: payload.label,
      value: payload.value,
      suffix: payload.suffix,
      initialValue: payload.value,
      onConfirm: payload.onConfirm,
    })
  }, [])

  const closeNumericPad = useCallback(() => {
    setPadState((prev) => ({
      ...prev,
      open: false,
      onConfirm: undefined,
    }))
  }, [])

  const handleKeyPress = useCallback((key: string) => {
    setPadState((prev) => {
      if (key === '删除') {
        return { ...prev, value: prev.value.slice(0, -1) }
      }
      if (key === '.') {
        if (prev.value.includes('.')) return prev
        return { ...prev, value: sanitizeDecimalInput(`${prev.value}.`) }
      }
      return { ...prev, value: sanitizeDecimalInput(`${prev.value}${key}`) }
    })
  }, [])

  const handleClear = useCallback(() => {
    setPadState((prev) => ({ ...prev, value: '' }))
  }, [])

  const handleConfirm = useCallback(() => {
    setPadState((prev) => {
      const sanitized = sanitizeDecimalInput(prev.value)
      const normalized = sanitized.endsWith('.') ? sanitized.slice(0, -1) : sanitized
      if (prev.onConfirm) {
        prev.onConfirm(normalized)
      }
      return { ...prev, open: false, onConfirm: undefined }
    })
  }, [])

  return {
    padState,
    openNumericPad,
    closeNumericPad,
    handleKeyPress,
    handleClear,
    handleConfirm,
  }
}
