import { Clock } from 'lucide-react'
import { useTranslation } from 'react-i18next'

interface ShiftListCardProps {
  shifts: { Id: string; Name?: string; StartTime?: string; EndTime?: string; Status?: string; Type?: string }[]
}

export default function ShiftListCard({ shifts }: ShiftListCardProps) {
  const { t } = useTranslation('common')
  return (
    <div className="space-y-2">
      {shifts.length > 0 ? shifts.map(shift => (
        <div key={shift.Id} className="flex items-center justify-between p-3 bg-surface-sunken rounded-md border border-gray-100">
          <div className="flex items-center gap-3">
            <div className="w-8 h-8 rounded-md bg-state-treating-bg flex items-center justify-center">
              <Clock size={14} className="text-state-treating" />
            </div>
            <div>
              <p className="font-medium text-foreground text-sm">{shift.Name || `${t('shift')} ${shift.Id}`}</p>
              <p className="text-meta text-foreground-muted">
                {shift.StartTime && shift.EndTime ? `${shift.StartTime} - ${shift.EndTime}` : t('timeNotSet')}
              </p>
            </div>
          </div>
          <span className={`text-meta px-2 py-0.5 rounded-md ${shift.Status === '1' ? 'bg-state-finished-bg text-state-finished' : 'bg-gray-100 text-gray-500'}`}>
            {shift.Type || t('regular')}
          </span>
        </div>
      )) : (
        <div className="text-center py-8 text-foreground-muted text-sm">{t('noData.shift')}</div>
      )}
    </div>
  )
}
