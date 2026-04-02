// 输入框组件

import React, { memo } from 'react'

interface InputFieldProps {
  label: string
  placeholder?: string
  suffix?: string
  required?: boolean
  type?: string
  value?: string | number
  onChange?: (e: React.ChangeEvent<HTMLInputElement>) => void
  readOnly?: boolean
  inputMode?: React.HTMLAttributes<HTMLInputElement>['inputMode']
  pattern?: string
}

export const InputField = memo(function InputField({
  label,
  placeholder,
  suffix,
  required,
  type = 'text',
  value,
  onChange,
  readOnly,
  inputMode,
  pattern
}: InputFieldProps) {
  return (
    <div className="space-y-1">
      <label className="text-[10px] font-black text-slate-400 uppercase tracking-tighter ml-1">
        {required && <span className="text-red-500 mr-1">*</span>}
        {label}
      </label>
      <div className="relative group">
        <input
          type={type}
          value={value}
          onChange={onChange}
          readOnly={readOnly}
          placeholder={placeholder}
          inputMode={inputMode}
          pattern={pattern}
          className={`w-full px-3 py-2 bg-white border border-slate-200 rounded-lg text-xs font-bold text-slate-700 outline-none focus:ring-2 focus:ring-blue-500/10 focus:border-blue-500 transition-all placeholder:text-slate-300 placeholder:font-medium ${
            readOnly ? 'bg-slate-50 cursor-not-allowed' : ''
          }`}
        />
        {suffix && (
          <span className="absolute right-3 top-1/2 -translate-y-1/2 text-[9px] font-black text-slate-300 uppercase">
            {suffix}
          </span>
        )}
      </div>
    </div>
  )
})
