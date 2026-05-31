import { TrendingUp } from 'lucide-react'
import { useTranslation } from 'react-i18next'

interface StatItem {
  labelKey: string
  value: number | string
  color: string
  trend?: string
}

interface StatKpiCardProps {
  items: StatItem[]
}

export default function StatKpiCard({ items }: StatKpiCardProps) {
  const { t: tRaw } = useTranslation()
  const t = tRaw as (key: string) => string
  return (
    <div className="grid grid-cols-2 md:grid-cols-4 gap-3 h-full">
      {items.map(item => (
        <div key={item.labelKey} className={`p-3 rounded-md border border-gray-100 bg-surface-sunken flex flex-col justify-center`}>
          <p className="text-meta text-foreground-muted uppercase mb-1">{t(item.labelKey)}</p>
          <div className="flex items-baseline">
            <p className="text-h2 font-semibold text-foreground">{item.value}</p>
            {item.trend && (
              <span className="ml-2 text-meta text-state-finished flex items-center">
                <TrendingUp size={10} className="mr-0.5" /> {item.trend}
              </span>
            )}
          </div>
        </div>
      ))}
    </div>
  )
}
