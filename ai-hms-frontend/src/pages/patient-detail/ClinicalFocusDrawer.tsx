import { useEffect } from 'react'
import { ShieldAlert, Clock, X } from 'lucide-react'
import type { Patient } from '@/types/original'

interface ClinicalFocusDrawerProps {
  patient: Patient
  open: boolean
  onClose: () => void
}

export default function ClinicalFocusDrawer({ patient, open, onClose }: ClinicalFocusDrawerProps) {
  // Esc 键关闭
  useEffect(() => {
    if (!open) return
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose()
    }
    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [open, onClose])

  // 阻止背景滚动
  useEffect(() => {
    if (open) {
      document.body.style.overflow = 'hidden'
    } else {
      document.body.style.overflow = ''
    }
    return () => { document.body.style.overflow = '' }
  }, [open])

  if (!open) return null

  const riskItems = [
    { label: '诊断', value: patient.diagnosis || '--' },
    { label: '风险等级', value: patient.riskLevel || '--' },
    { label: '院感', value: patient.infection?.hbsag === '阳性' ? 'HBV(+)' : '安全' },
    { label: '过敏', value: patient.medicalHistory?.allergies?.[0] || '无记录' },
  ]

  const dialysisSummary = [
    { label: '当前模式', value: patient.dialysisParams?.mode || '--' },
    { label: '目标超滤', value: patient.dialysisParams?.targetUf ? `${patient.dialysisParams.targetUf}L` : '--' },
    { label: '血流量', value: patient.dialysisParams?.bloodFlow ? `${patient.dialysisParams.bloodFlow} ml/min` : '--' },
    { label: '透析液流量', value: patient.dialysisParams?.dialysateFlow ? `${patient.dialysisParams.dialysateFlow} ml/min` : '--' },
  ]

  return (
    <div className="fixed inset-0 z-50 flex justify-end">
      {/* 遮罩 */}
      <div
        className="absolute inset-0 bg-slate-900/30 backdrop-blur-sm transition-opacity"
        onClick={onClose}
      />

      {/* 抽屉内容 */}
      <div className="relative w-[380px] max-w-[90vw] bg-white shadow-2xl flex flex-col animate-slide-in-right">
        {/* 头部 */}
        <div className="flex items-center justify-between px-6 py-5 border-b border-slate-100">
          <h3 className="text-base font-bold text-[#0f1f3d] flex items-center gap-2.5">
            <div className="w-8 h-8 rounded-xl bg-blue-50 flex items-center justify-center">
              <ShieldAlert size={16} className="text-[#1f63ff]" />
            </div>
            临床焦点
          </h3>
          <button
            onClick={onClose}
            className="w-8 h-8 rounded-lg flex items-center justify-center text-slate-400 hover:text-slate-600 hover:bg-slate-100 transition-colors"
          >
            <X size={18} />
          </button>
        </div>

        {/* 内容区 */}
        <div className="flex-1 overflow-y-auto px-6 py-5 space-y-6">
          {/* 关键风险 */}
          <div>
            <h4 className="text-xs font-bold text-[#6f7f99] uppercase tracking-wider mb-3 flex items-center gap-2">
              <ShieldAlert size={13} className="text-[#ef4444]" />
              关键风险
            </h4>
            <div className="space-y-2">
              {riskItems.map(item => (
                <div
                  key={item.label}
                  className="flex items-center justify-between px-4 py-3 rounded-xl bg-[#f5f8fc]"
                >
                  <span className="text-xs text-[#6f7f99]">{item.label}</span>
                  <span className="text-sm font-medium text-[#0f1f3d]">{item.value}</span>
                </div>
              ))}
            </div>
          </div>

          {/* 透析参数 */}
          <div>
            <h4 className="text-xs font-bold text-[#6f7f99] uppercase tracking-wider mb-3 flex items-center gap-2">
              <Clock size={13} className="text-[#1f63ff]" />
              透析参数
            </h4>
            <div className="space-y-2">
              {dialysisSummary.map(item => (
                <div
                  key={item.label}
                  className="flex items-center justify-between px-4 py-3 rounded-xl bg-[#f5f8fc]"
                >
                  <span className="text-xs text-[#6f7f99]">{item.label}</span>
                  <span className="text-sm font-medium text-[#0f1f3d]">{item.value}</span>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* 底部同步时间 */}
        <div className="px-6 py-4 border-t border-slate-100">
          <div className="flex items-center text-xs text-[#6f7f99]">
            <Clock size={12} className="mr-2" />
            最后同步: {new Date().toLocaleTimeString()}
          </div>
        </div>
      </div>
    </div>
  )
}
