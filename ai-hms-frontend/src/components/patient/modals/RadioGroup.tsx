// RadioGroup - 单选按钮组组件

interface RadioGroupProps {
  label?: string
  options: string[]
  defaultValue?: string
  required?: boolean
  disabled?: boolean
  name?: string
}

export default function RadioGroup({
  label,
  options,
  defaultValue,
  required,
  disabled,
  name
}: RadioGroupProps) {
  return (
    <div className="flex flex-col gap-2">
      {label && (
        <label className="text-[11px] font-black text-slate-500 uppercase tracking-tighter">
          {required && <span className="text-red-500 mr-1">*</span>}
          {label}
        </label>
      )}
      <div className="flex flex-wrap items-center gap-x-4 gap-y-2 h-9">
        {options.map((opt) => (
          <label key={opt} className="flex items-center gap-1.5 cursor-pointer group">
            <input
              type="radio"
              name={name || label}
              disabled={disabled}
              defaultChecked={opt === defaultValue}
              className="w-4 h-4 text-blue-600 focus:ring-0 border-slate-300"
            />
            <span className="text-xs font-bold text-slate-600 group-hover:text-blue-600 transition-colors">
              {opt}
            </span>
          </label>
        ))}
      </div>
    </div>
  )
}
