import { useLocation } from 'react-router-dom'
import { getRouteMeta } from './routeMeta'
import { ChevronRight } from 'lucide-react'

export default function PageBreadcrumb() {
  const { pathname } = useLocation()
  const meta = getRouteMeta(pathname)

  return (
    <div className="flex items-center text-sm text-gray-500 mb-4">
      {meta.breadcrumb.map((crumb, idx) => (
        <span key={idx} className="flex items-center">
          {idx > 0 && <ChevronRight size={14} className="mx-2 text-gray-300" />}
          <span className={idx === meta.breadcrumb.length - 1 ? 'text-gray-800 font-semibold' : ''}>
            {crumb}
          </span>
        </span>
      ))}
    </div>
  )
}
