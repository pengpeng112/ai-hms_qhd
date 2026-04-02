// 数字输入字段 - 点击弹出数字键盘

import { memo } from 'react'
import type { OpenNumericPad } from './types'

interface NumericInputFieldProps {
  label: string
  placeholder?: string
  suffix?: string
  required?: boolean
  value?: string
  disabled?: boolean
  openNumericPad: OpenNumericPad
  onConfirm: (value: string) => void
}

export const NumericInputField = memo(function NumericInputField({
  label,
  placeholder,
  suffix,
  required,
  value,
  disabled,
  openNumericPad,
  onConfirm,
}: NumericInputFieldProps) {
  const displayValue = value ?? ''
  return (
    <div className="space-y-1">
      <label className="text-[10px] font-black text-slate-400 uppercase tracking-tighter ml-1">
        {required && <span className="text-red-500 mr-1">*</span>}
        {label}
      </label>
      <button
        type="button"
        onClick={() => openNumericPad({ label, value: displayValue, suffix, disabled, onConfirm })}
        className={`w-full h-10 px-3 border rounded-lg text-xs font-bold text-left relative outline-none transition-all ${
          disabled
            ? 'bg-slate-50 border-slate-200 text-slate-300 cursor-not-allowed'
            : 'bg-white border-slate-200 text-slate-700 hover:border-blue-300 focus:ring-2 focus:ring-blue-500/10 focus:border-blue-500 cursor-pointer'
        }`}
      >
        <span className={displayValue ? 'text-slate-700' : 'text-slate-300 font-medium'}>
          {displayValue || placeholder || ''}
        </span>
        {suffix && (
          <span className="absolute right-3 top-1/2 -translate-y-1/2 text-[9px] font-black text-slate-300 uppercase">
            {suffix}
          </span>
        )}
      </button>
    </div>
  )
})
