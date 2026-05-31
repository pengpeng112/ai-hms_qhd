import { AlertTriangle, FilePlus } from 'lucide-react'
import { useTranslation } from 'react-i18next'

interface AlertItem {
  icon: 'alert' | 'file'
  titleKey: string
  descKey: string
  actionKey: string
  actionRoute: string
  color: 'orange' | 'blue'
}

interface AlertCardProps {
  items: AlertItem[]
  onNavigate: (route: string) => void
}

export default function AlertCard({ items, onNavigate }: AlertCardProps) {
  const { t: tRaw } = useTranslation(['dashboard', 'common'])
  const t = tRaw as (key: string) => string
  const IconMap = { alert: AlertTriangle, file: FilePlus }

  return (
    <div className="space-y-3">
      {items.map((item, idx) => {
        const Icon = IconMap[item.icon]
        const borderColor = item.color === 'orange' ? 'border-l-state-waiting' : 'border-l-state-treating'
        const bgColor = item.color === 'orange' ? 'bg-state-waiting-bg' : 'bg-state-treating-bg'
        const textColor = item.color === 'orange' ? 'text-state-waiting' : 'text-state-treating'

        return (
          <div key={idx} className={`p-3 bg-surface rounded-md border border-gray-100 border-l-4 ${borderColor}`}>
            <div className="flex justify-between items-start mb-1">
              <div className="flex items-center">
                <Icon size={16} className={`${textColor} mr-2`} />
                <h4 className="font-bold text-foreground text-sm">{t(item.titleKey)}</h4>
              </div>
            </div>
            <p className="text-xs text-gray-600 pl-6 mb-2">{t(item.descKey)}</p>
            <button
              onClick={(e) => { e.stopPropagation(); onNavigate(item.actionRoute) }}
              className={`ml-6 px-2 py-1 ${bgColor} ${textColor} text-xs rounded-md hover:opacity-80 transition-opacity`}
            >
              {t(item.actionKey)}
            </button>
          </div>
        )
      })}
    </div>
  )
}
