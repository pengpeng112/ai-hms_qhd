import type { ReactNode } from 'react'

interface SmallInputProps {
  label: string
  defaultValue?: string
  placeholder?: string
  suffix?: string | ReactNode
  readOnly?: boolean
  onChange?: (value: string) => void
}

export default function SmallInput({ label, defaultValue, placeholder, suffix, readOnly, onChange }: SmallInputProps) {
  return (
    <div className="flex flex-col gap-2">
      <label className="text-[11px] font-black text-slate-500 uppercase tracking-tighter">{label}</label>
      <div className="relative">
        <input
          type="text"
          defaultValue={defaultValue}
          placeholder={placeholder}
          readOnly={readOnly}
          onChange={(e) => onChange?.(e.target.value)}
          className={`w-full h-9 px-3 border rounded-lg text-sm font-bold outline-none transition-all ${
            readOnly
              ? 'bg-slate-50 border-slate-100 text-slate-700 cursor-default'
              : 'bg-white border-slate-200 focus:ring-1 focus:ring-blue-500'
          } ${suffix ? 'pr-12' : ''}`}
        />
        {suffix && (
          <span className="absolute right-3 top-1/2 -translate-y-1/2 text-xs text-slate-400 font-bold">
            {suffix}
          </span>
        )}
      </div>
    </div>
  )
}
