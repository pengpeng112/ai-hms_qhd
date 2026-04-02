// 下拉选择框组件

import React, { memo } from 'react'
import { ChevronDown } from 'lucide-react'

interface SelectFieldProps {
  label: string
  options: string[] | Array<{ value: string; label: string }>
  required?: boolean
  value?: string
  onChange?: (e: React.ChangeEvent<HTMLSelectElement>) => void
  disabled?: boolean
}

export const SelectField = memo(function SelectField({
  label,
  options,
  required,
  value,
  onChange,
  disabled
}: SelectFieldProps) {
  return (
    <div className="space-y-1">
      <label className="text-[10px] font-black text-slate-400 uppercase tracking-tighter ml-1">
        {required && <span className="text-red-500 mr-1">*</span>}
        {label}
      </label>
      <div className="relative">
        <select
          value={value}
          onChange={onChange}
          disabled={disabled}
          className={`w-full px-3 py-2 bg-white border border-slate-200 rounded-lg text-xs font-bold text-slate-700 outline-none appearance-none focus:ring-2 focus:ring-blue-500/10 focus:border-blue-500 transition-all cursor-pointer ${disabled ? 'bg-slate-50 cursor-not-allowed text-slate-400' : ''}`}
        >
          <option value="">请选择</option>
          {options.map((opt) => {
            const optValue = typeof opt === 'string' ? opt : opt.value
            const optLabel = typeof opt === 'string' ? opt : opt.label
            return (
              <option key={optValue} value={optValue}>
                {optLabel}
              </option>
            )
          })}
        </select>
        <ChevronDown
          size={12}
          className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 pointer-events-none"
        />
      </div>
    </div>
  )
})
