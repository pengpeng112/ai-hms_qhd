interface RadioGroupProps {
  label: string
  options: string[]
  defaultValue?: string
  disabled?: boolean
  onChange?: (value: string) => void
}

export default function RadioGroup({ label, options, defaultValue, disabled, onChange }: RadioGroupProps) {
  return (
    <div className="flex flex-col gap-2">
      <label className="text-[11px] font-black text-slate-500 uppercase tracking-tighter">{label}</label>
      <div className="flex items-center gap-4 flex-wrap">
        {options.map(opt => (
          <label key={opt} className={`flex items-center gap-1.5 cursor-pointer group ${disabled ? 'opacity-50 pointer-events-none' : ''}`}>
            <input
              type="radio"
              name={label}
              defaultChecked={opt === defaultValue}
              disabled={disabled}
              onChange={() => onChange?.(opt)}
              className="w-4 h-4 text-blue-600 border-slate-300 focus:ring-0"
            />
            <span className="text-xs font-bold text-slate-600 group-hover:text-blue-600 transition-colors">{opt}</span>
          </label>
        ))}
      </div>
    </div>
  )
}
