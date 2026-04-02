/**
 * FormField - 表单字段组件
 * 统一的表单输入字段样式
 */

export interface FormFieldProps {
    /** 字段标签 */
    label: string
    /** 后缀单位 */
    suffix?: string
    /** 默认值 */
    defaultValue?: string | number
    /** 宽度 class */
    width?: string
    /** 是否只读 */
    readOnly?: boolean
    /** 是否必填 */
    required?: boolean
    /** 占位符 */
    placeholder?: string
    /** 值变更回调 */
    onChange?: (value: string) => void
    /** 深色模式 */
    dark?: boolean
    /** 字段名称（用于表单提交） */
    name?: string
    /** 字段标识（用于获取 DOM 值） */
    dataField?: string
}

export default function FormField({
    label,
    suffix,
    defaultValue,
    width = 'w-full',
    readOnly,
    required,
    placeholder,
    onChange,
    dark = false,
    name,
    dataField
}: FormFieldProps) {
    const labelColor = dark ? 'text-slate-400' : 'text-slate-500'
    const suffixColor = dark ? 'text-slate-500' : 'text-slate-400'

    const getInputStyles = () => {
        if (dark) {
            return readOnly
                ? 'bg-slate-700 border-slate-600 text-slate-400'
                : 'bg-slate-800 border-slate-700 text-white focus:ring-1 focus:ring-blue-400 focus:border-blue-400'
        }
        return readOnly
            ? 'bg-slate-50 text-slate-400 border-slate-200'
            : 'bg-white border-slate-300 focus:ring-1 focus:ring-blue-500 focus:border-blue-500'
    }

    return (
        <div className="flex flex-col gap-1.5">
            <label className={`text-[11px] font-bold flex items-center ${labelColor}`}>
                {required && <span className="text-red-500 mr-0.5">*</span>}
                {label}
            </label>
            <div className={`relative ${width}`}>
                <input
                    type="text"
                    name={name}
                    data-field={dataField}
                    defaultValue={defaultValue}
                    placeholder={placeholder}
                    readOnly={readOnly}
                    onChange={(e) => onChange?.(e.target.value)}
                    className={`w-full h-10 px-3 border rounded-lg text-sm outline-none transition-all ${getInputStyles()}`}
                />
                {suffix && (
                    <span className={`absolute right-3 top-1/2 -translate-y-1/2 text-[10px] pointer-events-none ${suffixColor}`}>
                        {suffix}
                    </span>
                )}
            </div>
        </div>
    )
}
