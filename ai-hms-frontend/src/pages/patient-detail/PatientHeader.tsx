import { User, AlertTriangle } from 'lucide-react'
import type { Patient } from '@/types/original'

interface PatientHeaderProps {
  patient: Patient
  avatarFailed: boolean
  onAvatarError: () => void
}

const riskStyles: Record<string, { bg: string; text: string; label: string }> = {
  '高危': { bg: 'bg-state-alert/10', text: 'text-state-alert', label: '高危' },
  '中危': { bg: 'bg-state-waiting/10', text: 'text-state-waiting', label: '中危' },
  '低危': { bg: 'bg-state-finished/10', text: 'text-state-finished', label: '低危' },
}

export default function PatientHeader({ patient, avatarFailed, onAvatarError }: PatientHeaderProps) {
  const risk = riskStyles[patient.riskLevel] || riskStyles['低危']

  return (
    <div className="flex items-center gap-6">
      {/* 头像 */}
      <div className="relative shrink-0">
        {patient.avatar && !avatarFailed ? (
          <img
            src={patient.avatar}
            alt={`${patient.name} avatar`}
            className="w-16 h-16 rounded-lg object-cover bg-slate-100 shadow-lg"
            onError={onAvatarError}
          />
        ) : (
          <div className="w-16 h-16 rounded-lg bg-blue-600 flex items-center justify-center text-white shadow-lg">
            <User size={32} strokeWidth={2.5} />
          </div>
        )}
        {patient.riskLevel === '高危' && (
          <div className="absolute -top-1 -right-1 w-5 h-5 bg-state-alert border-2 border-white rounded-full flex items-center justify-center">
            <AlertTriangle size={10} className="text-white" />
          </div>
        )}
      </div>

      {/* 信息区 */}
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-3">
          <h1 data-testid="patient-detail-name" className="text-h2 font-bold text-foreground truncate">
            {patient.name}
          </h1>
          <span className={`px-2 py-0.5 rounded-md text-meta font-bold ${risk.bg} ${risk.text}`}>
            {risk.label}
          </span>
          <span className="px-2 py-0.5 bg-slate-900 text-white rounded-md text-meta font-bold">
            {patient.bedNumber} 床
          </span>
        </div>
        <p className="text-meta text-foreground-muted mt-1">
          {patient.gender} · {patient.age}岁 · {patient.insuranceType} · 主治: {patient.doctorName || '--'} · {patient.defaultMode} · {patient.status}
        </p>
      </div>
    </div>
  )
}
