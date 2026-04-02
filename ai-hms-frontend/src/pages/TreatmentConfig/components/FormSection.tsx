// 表单区块组件

import React, { memo } from 'react'

interface FormSectionProps {
  title: string
  icon: React.ElementType
  children?: React.ReactNode
}

export const FormSection = memo(function FormSection({
  title,
  icon: Icon,
  children
}: FormSectionProps) {
  return (
    <div className="bg-slate-50/50 rounded-2xl border border-slate-100 p-4 space-y-3">
      <div className="flex items-center gap-2 mb-1">
        <div className="p-1.5 bg-white rounded-lg shadow-sm text-blue-600">
          <Icon size={16} />
        </div>
        <h4 className="text-[11px] font-black text-slate-800 uppercase tracking-widest">
          {title}
        </h4>
      </div>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-3">
        {children}
      </div>
    </div>
  )
})
