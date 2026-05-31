import { useState } from 'react'
import { Drawer } from 'antd'
import { ShieldAlert, Clock } from 'lucide-react'
import type { Patient } from '@/types/original'

interface FocusPanelProps {
  patient: Patient
}

export default function FocusPanel({ patient }: FocusPanelProps) {
  const [drawerOpen, setDrawerOpen] = useState(false)
  const isNarrow = typeof window !== 'undefined' && window.innerWidth < 1280

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

  const content = (
    <div className="space-y-6">
      {/* 关键风险列表 */}
      <div>
        <h4 className="text-meta font-bold text-foreground-muted uppercase tracking-wider mb-3 flex items-center gap-2">
          <ShieldAlert size={14} className="text-state-alert" /> 关键风险
        </h4>
        <div className="space-y-2">
          {riskItems.map(item => (
            <div key={item.label} className="flex items-center justify-between p-3 bg-surface-sunken rounded-md">
              <span className="text-meta text-foreground-muted">{item.label}</span>
              <span className="text-sm font-medium text-foreground">{item.value}</span>
            </div>
          ))}
        </div>
      </div>

      {/* 最近透析摘要 */}
      <div>
        <h4 className="text-meta font-bold text-foreground-muted uppercase tracking-wider mb-3 flex items-center gap-2">
          <Clock size={14} /> 透析参数
        </h4>
        <div className="space-y-2">
          {dialysisSummary.map(item => (
            <div key={item.label} className="flex items-center justify-between p-3 bg-surface-sunken rounded-md">
              <span className="text-meta text-foreground-muted">{item.label}</span>
              <span className="text-sm font-medium text-foreground">{item.value}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  )

  // 窄屏：用 Drawer
  if (isNarrow) {
    return (
      <>
        <button
          onClick={() => setDrawerOpen(true)}
          className="p-2 rounded-lg bg-slate-100 text-slate-500 hover:bg-slate-200 transition-colors relative"
          title="查看风险"
        >
          <ShieldAlert size={20} />
          <span className="absolute top-1 right-1 w-2 h-2 bg-state-alert rounded-full border border-white"></span>
        </button>
        <Drawer
          title="临床焦点"
          placement="right"
          width={340}
          open={drawerOpen}
          onClose={() => setDrawerOpen(false)}
        >
          {content}
        </Drawer>
      </>
    )
  }

  // 宽屏：固定卡
  return (
    <div className="w-[320px] bg-white border-l border-slate-200 flex flex-col shrink-0 shadow-lg overflow-y-auto">
      <div className="p-4 border-b border-slate-100 flex items-center justify-between sticky top-0 bg-white z-10">
        <h3 className="text-sm font-bold text-foreground flex items-center gap-2">
          <ShieldAlert size={16} className="text-state-alert" /> 临床焦点
        </h3>
      </div>
      <div className="p-4 flex-1">
        {content}
      </div>
      <div className="p-4 border-t border-slate-100 shrink-0">
        <div className="flex items-center text-foreground-muted text-meta">
          <Clock size={12} className="mr-2" /> 最后同步: {new Date().toLocaleTimeString()}
        </div>
      </div>
    </div>
  )
}
