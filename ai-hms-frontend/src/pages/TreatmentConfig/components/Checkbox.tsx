// Checkbox.tsx - 复选框组件

import { Check } from 'lucide-react'

interface CheckboxProps {
  checked: boolean
  indeterminate?: boolean
  onChange: (checked: boolean) => void
  disabled?: boolean
}

export function Checkbox({ checked, indeterminate = false, onChange, disabled = false }: CheckboxProps) {
  const handleClick = () => {
    if (!disabled) {
      onChange(!checked)
    }
  }

  return (
    <button
      onClick={handleClick}
      disabled={disabled}
      className={`
        w-5 h-5 rounded border-2 flex items-center justify-center transition-all
        ${disabled ? 'opacity-40 cursor-not-allowed' : 'cursor-pointer'}
        ${checked || indeterminate
          ? 'bg-blue-600 border-blue-600'
          : 'border-slate-300 hover:border-blue-400 bg-white'
        }
      `}
      type="button"
    >
      {checked && !indeterminate && <Check size={12} className="text-white" />}
      {indeterminate && <div className="w-2.5 h-0.5 bg-white rounded" />}
    </button>
  )
}
