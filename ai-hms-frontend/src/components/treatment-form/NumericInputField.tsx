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
  warn?: boolean       // 异常高亮（红框）
  warnText?: string    // 异常说明
  hint?: string        // 普通提示（灰字）
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
  warn,
  warnText,
  hint,
}: NumericInputFieldProps) {
  const displayValue = value ?? ''
  const btnClass = disabled
    ? 'bg-slate-50 border-slate-200 text-slate-300 cursor-not-allowed'
    : warn
      ? 'bg-red-50 border-red-400 text-red-600 cursor-pointer'
      : 'bg-white border-slate-200 text-slate-700 hover:border-blue-300 focus:ring-2 focus:ring-blue-500/10 focus:border-blue-500 cursor-pointer'
  return (
    <div className="space-y-1">
      <label className={`text-[10px] font-black uppercase tracking-tighter ml-1 ${warn ? 'text-red-500' : 'text-slate-400'}`}>
        {required && <span className="text-red-500 mr-1">*</span>}
        {label}{warn ? ' ⚠' : ''}
      </label>
      <button
        type="button"
        onClick={() => openNumericPad({ label, value: displayValue, suffix, disabled, onConfirm })}
        className={`w-full h-10 px-3 border rounded-lg text-xs font-bold text-left relative outline-none transition-all ${btnClass}`}
      >
        <span className={displayValue ? (warn ? 'text-red-600' : 'text-slate-700') : 'text-slate-300 font-medium'}>
          {displayValue || placeholder || ''}
        </span>
        {suffix && (
          <span className={`absolute right-3 top-1/2 -translate-y-1/2 text-[9px] font-black uppercase ${warn ? 'text-red-400' : 'text-slate-300'}`}>
            {suffix}
          </span>
        )}
      </button>
      {warn && warnText ? <div className="text-[10px] font-semibold text-red-500 ml-1">{warnText}</div> : null}
      {!warn && hint ? <div className="text-[10px] font-medium text-slate-400 ml-1">{hint}</div> : null}
    </div>
  )
})
