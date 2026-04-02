// 数字键盘浮层组件

import { memo } from 'react'

interface NumericPadProps {
  open: boolean
  label: string
  value: string
  onKeyPress: (key: string) => void
  onConfirm: () => void
  onClear: () => void
  onClose: () => void
}

export const NumericPad = memo(function NumericPad({
  open,
  label,
  value,
  onKeyPress,
  onConfirm,
  onClear,
  onClose,
}: NumericPadProps) {
  if (!open) return null

  return (
    <div
      className="fixed inset-0 z-[120] flex items-center justify-center bg-slate-900/60 backdrop-blur-sm p-4"
      onClick={onClose}
    >
      <div
        className="bg-white rounded-[28px] shadow-2xl w-full max-w-sm p-6"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between mb-4">
          <div>
            <p className="text-[10px] text-slate-400 font-black uppercase tracking-widest">数值输入</p>
            <h4 className="text-base font-black text-slate-900">{label}</h4>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="px-3 py-1.5 rounded-full text-[10px] font-black text-slate-500 hover:text-slate-900 hover:bg-slate-100 transition-all"
          >
            取消
          </button>
        </div>
        <div className="relative mb-4">
          <div className="w-full px-3 py-3 border border-slate-200 rounded-xl text-right font-black text-slate-800 bg-slate-50">
            <span className={value ? 'text-slate-800' : 'text-slate-300'}>
              {value || '请输入'}
            </span>
          </div>
        </div>
        <div className="grid grid-cols-3 gap-3">
          {['1', '2', '3', '4', '5', '6', '7', '8', '9', '.', '0', '删除'].map((key) => (
            <button
              key={key}
              type="button"
              onClick={() => onKeyPress(key)}
              className={`py-3 rounded-xl text-sm font-black border transition-all ${
                key === '删除'
                  ? 'border-amber-200 text-amber-600 bg-amber-50 hover:bg-amber-100'
                  : 'border-slate-200 text-slate-700 hover:border-blue-200 hover:text-blue-600'
              }`}
            >
              {key}
            </button>
          ))}
        </div>
        <div className="grid grid-cols-2 gap-3 mt-4">
          <button
            type="button"
            onClick={onClear}
            className="py-3 rounded-xl text-sm font-black border border-slate-200 text-slate-500 hover:bg-slate-100 transition-all"
          >
            清空
          </button>
          <button
            type="button"
            onClick={onConfirm}
            className="py-3 rounded-xl text-sm font-black bg-blue-600 text-white hover:bg-blue-700 transition-all"
          >
            确定
          </button>
        </div>
      </div>
    </div>
  )
})
