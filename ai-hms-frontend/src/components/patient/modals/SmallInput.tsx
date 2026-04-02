// SmallInput - 小型输入框组件

interface SmallInputProps {
  label: string
  suffix?: React.ReactNode
  defaultValue?: string
  required?: boolean
  placeholder?: string
  readOnly?: boolean
}

export default function SmallInput({
  label,
  suffix,
  defaultValue,
  required,
  placeholder,
  readOnly
}: SmallInputProps) {
  return (
    <div className="flex flex-col gap-2">
      <label className="text-[11px] font-black text-slate-500 uppercase tracking-tighter">
        {required && <span className="text-red-500 mr-1">*</span>}
        {label}
      </label>
      <div className="relative">
        <input
          type="text"
          readOnly={readOnly}
          defaultValue={defaultValue}
          placeholder={placeholder}
          className={`w-full h-9 px-3 border rounded-lg text-xs font-bold text-slate-700 outline-none transition-all ${
            !readOnly
              ? 'bg-white border-slate-300 focus:ring-1 focus:ring-blue-500'
              : 'bg-slate-50 border-slate-200 text-slate-400'
          }`}
        />
        {suffix && (
          <span className="absolute right-3 top-1/2 -translate-y-1/2 text-[10px] text-slate-400 font-bold">
            {suffix}
          </span>
        )}
      </div>
    </div>
  )
}
