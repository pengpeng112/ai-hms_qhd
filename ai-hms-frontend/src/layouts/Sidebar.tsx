import { NavLink } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { UserRoleLabel } from '@/types/original'
import { getSelectedRoleUser, getMenusByRole } from '@/services/role'
import {
  LayoutDashboard,
  Users,
  Monitor,
  Calendar,
  BarChart3,
  Settings,
  HeartPulse,
  ClipboardCheck,
  Package,
  Server,
  Database,
  Layers,
  BookOpen,
} from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'

interface SidebarProps {
  isOpen: boolean
}

const menuItems = [
  { key: 'dashboard', path: '/', icon: LayoutDashboard },
  { key: 'wardOverview', path: '/ward-overview', icon: HeartPulse },
  { key: 'patients', path: '/patients', icon: Users },
  { key: 'monitoring', path: '/monitoring', icon: Monitor },
  { key: 'dialysisProcessing', path: '/dialysis-processing', icon: ClipboardCheck },
  { key: 'schedule', path: '/schedule', icon: Calendar },
  { key: 'inventory', path: '/inventory', icon: Package },
  { key: 'deviceBinding', path: '/device-binding', icon: Server },
  { key: 'statistics', path: '/statistics', icon: BarChart3 },
  { key: 'masterData', path: '/master-data', icon: Database },
  { key: 'treatmentConfig', path: '/treatment-config', icon: Layers },
  { key: 'dictConfig', path: '/dict-config', icon: BookOpen },
  { key: 'settings', path: '/settings', icon: Settings },
]

const roleMenuMap: Record<string, string[]> = {
  dashboard: ['dashboard'],
  ward_overview: ['wardOverview'],
  patients: ['patients'],
  monitoring: ['monitoring'],
  dialysis_processing: ['dialysisProcessing'],
  schedule: ['schedule'],
  inventory: ['inventory'],
  device_binding: ['deviceBinding'],
  statistics: ['statistics'],
  master_data: ['masterData'],
  treatment_config: ['treatmentConfig'],
  dict_config: ['dictConfig'],
  settings: ['settings'],
}

export default function Sidebar({ isOpen }: SidebarProps) {
  const { t } = useTranslation(['nav', 'role'])
  const roleUser = useMemo(() => getSelectedRoleUser(), [])
  const role = roleUser?.role
  const [allowedMenuKeys, setAllowedMenuKeys] = useState<string[]>([])

  useEffect(() => {
    if (!role) {
      setAllowedMenuKeys([])
      return
    }

    let alive = true
    getMenusByRole(role)
      .then(keys => {
        if (alive) {
          setAllowedMenuKeys(keys)
        }
      })
      .catch(() => {
        if (alive) {
          setAllowedMenuKeys([])
        }
      })

    return () => {
      alive = false
    }
  }, [role])

  const visibleMenuItems = useMemo(() => {
    return menuItems.filter(item => {
      const originalKey = Object.entries(roleMenuMap).find(([, values]) => values.includes(item.key))?.[0]
      return originalKey ? allowedMenuKeys.includes(originalKey) : false
    })
  }, [allowedMenuKeys])

  return (
    <aside className={`${isOpen ? 'w-64' : 'w-20'} bg-slate-900 flex flex-col transition-all duration-300 shadow-xl z-20 shrink-0`}>
      <div className="h-16 flex items-center justify-center border-b border-slate-800 text-white font-bold text-xl overflow-hidden">
        {isOpen ? (
          <span className="animate-fade-in whitespace-nowrap">{t('nav:brand.full')}</span>
        ) : (
          <span className="text-blue-500">{t('nav:brand.short')}</span>
        )}
      </div>

      <div className="flex-1 px-3 py-4 overflow-y-auto no-scrollbar">
        {visibleMenuItems.map(item => (
          <NavLink
            key={item.path}
            to={item.path}
            end={item.path === '/'}
            className={({ isActive }) => `
              w-full flex items-center p-3 my-1 rounded-xl transition-all
              ${isActive ? 'bg-blue-600 text-white shadow-lg shadow-blue-900/20' : 'text-slate-400 hover:bg-slate-800 hover:text-white'}
            `}
          >
            {({ isActive }) => (
              <>
                <item.icon size={20} className={isActive ? 'animate-pulse-slow' : ''} />
                {isOpen && (
                  <span className="ml-3 font-medium text-sm whitespace-nowrap">
                    {t(item.key as keyof typeof import('@/i18n/locales/zh-CN/nav.json'))}
                  </span>
                )}
              </>
            )}
          </NavLink>
        ))}
      </div>

      {isOpen && roleUser && (
        <div className="p-4 border-t border-slate-800">
          <div className="bg-slate-800/50 rounded-lg p-3">
            <p className="text-[10px] text-slate-500 uppercase font-bold tracking-wider">{t('role:label.current')}</p>
            <p className="text-sm text-white mt-1 font-medium">{roleUser.name}</p>
            <p className="text-xs text-blue-400 mt-0.5">{String(roleUser.role) === 'ADMIN' ? '系统管理员' : (UserRoleLabel[roleUser.role] || roleUser.role)}</p>
          </div>
        </div>
      )}
    </aside>
  )
}
