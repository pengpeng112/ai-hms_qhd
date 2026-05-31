import { ArrowRight } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { PatientListItem } from '@/components'
import type { Patient } from '@/types/original'

interface PatientListCardProps {
  patients: Partial<Patient>[]
  onSelect: (id: string) => void
  onViewAll: () => void
  maxItems?: number
}

export default function PatientListCard({ patients, onSelect, onViewAll, maxItems = 5 }: PatientListCardProps) {
  const { t } = useTranslation(['dashboard', 'common'])
  return (
    <div className="space-y-2">
      {patients.slice(0, maxItems).map(patient => (
        <div key={patient.id} onClick={(e) => e.stopPropagation()}>
          <PatientListItem
            patient={patient as Patient}
            variant="compact"
            onClick={(p) => onSelect(p.id)}
          />
        </div>
      ))}
      {patients.length === 0 && (
        <div className="text-center py-4 text-foreground-muted text-sm">{t('common:noData.patient') || '暂无患者数据'}</div>
      )}
      <button
        onClick={(e) => { e.stopPropagation(); onViewAll() }}
        className="w-full py-2 text-xs text-center text-foreground-muted hover:text-state-treating transition-colors border-t border-gray-50 mt-1"
      >
        {t('common:action.viewAll')} <ArrowRight size={10} className="inline ml-1" />
      </button>
    </div>
  )
}
