import { NavLink } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Popover } from 'antd'
import { getSelectedRoleUser, getMenusByRole, getRoleLabel } from '@/services/role'
import {
  LayoutDashboard,
  type LucideIcon,
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
  Building2,
  Bed,
  GraduationCap,
  ChevronDown,
} from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'

interface SidebarProps {
  isOpen: boolean
}

type MenuItem = {
  key: string
  path: string
  icon: LucideIcon
  hidden?: boolean
}

type MenuGroup = {
  key: string
  itemKeys: string[]
}

const menuItems: MenuItem[] = [
  { key: 'dashboard', path: '/', icon: LayoutDashboard },
  { key: 'wardOverview', path: '/ward-overview', icon: HeartPulse },
  { key: 'patients', path: '/patients', icon: Users },
  { key: 'monitoring', path: '/monitoring', icon: Monitor },
  { key: 'dialysisProcessing', path: '/dialysis-processing', icon: ClipboardCheck },
  { key: 'schedule', path: '/schedule', icon: Calendar },
  { key: 'inventory', path: '/inventory', icon: Package },
  { key: 'deviceBinding', path: '/device-binding', icon: Server },
  { key: 'wardManagement', path: '/ward-management', icon: Building2 },
  { key: 'bedManagement', path: '/bed-management', icon: Bed },
  { key: 'educationManagement', path: '/education-management', icon: GraduationCap, hidden: true },
  { key: 'statistics', path: '/statistics', icon: BarChart3 },
  { key: 'masterData', path: '/master-data', icon: Database },
  { key: 'treatmentConfig', path: '/treatment-config', icon: Layers },
  { key: 'dictConfig', path: '/dict-config', icon: BookOpen },
  { key: 'settings', path: '/settings', icon: Settings },
  { key: 'userManagement', path: '/user-management', icon: Users },
  { key: 'roleManagement', path: '/role-management', icon: Settings },
]

const menuGroups: MenuGroup[] = [
  { key: 'dailyWork', itemKeys: ['dashboard', 'wardOverview', 'monitoring', 'dialysisProcessing'] },
  { key: 'patientCenter', itemKeys: ['patients', 'educationManagement'] },
  { key: 'schedule', itemKeys: ['schedule'] },
  { key: 'resource', itemKeys: ['inventory', 'deviceBinding', 'wardManagement', 'bedManagement'] },
  { key: 'systemConfig', itemKeys: ['masterData', 'treatmentConfig', 'dictConfig', 'userManagement', 'roleManagement', 'settings', 'statistics'] },
]

const defaultExpandedGroups = Object.fromEntries(menuGroups.map(group => [group.key, true])) as Record<string, boolean>

const roleMenuMap: Record<string, string[]> = {
  dashboard: ['dashboard'],
  ward_overview: ['wardOverview'],
  patients: ['patients'],
  monitoring: ['monitoring'],
  dialysis_processing: ['dialysisProcessing'],
  schedule: ['schedule'],
  inventory: ['inventory'],
  device_binding: ['deviceBinding'],
  ward_management: ['wardManagement'],
  bed_management: ['bedManagement'],
  education_management: ['educationManagement'],
  statistics: ['statistics'],
  master_data: ['masterData'],
  treatment_config: ['treatmentConfig'],
  dict_config: ['dictConfig'],
  settings: ['settings'],
  user_management: ['userManagement'],
  role_management: ['roleManagement'],
}

export default function Sidebar({ isOpen }: SidebarProps) {
  const { t } = useTranslation(['nav', 'role'])
  const roleUser = useMemo(() => getSelectedRoleUser(), [])
  const role = roleUser?.role
  const [allowedMenuKeys, setAllowedMenuKeys] = useState<string[]>([])
  const [expandedGroups, setExpandedGroups] = useState<Record<string, boolean>>(defaultExpandedGroups)

  useEffect(() => {
    if (!role) {
      queueMicrotask(() => setAllowedMenuKeys([]))
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
      if ('hidden' in item && item.hidden) return false
      const originalKey = Object.entries(roleMenuMap).find(([, values]) => values.includes(item.key))?.[0]
      return originalKey ? allowedMenuKeys.includes(originalKey) : false
    })
  }, [allowedMenuKeys])

  const visibleMenuGroups = useMemo(() => {
    const visibleItemsByKey = new Map(visibleMenuItems.map(item => [item.key, item]))
    return menuGroups
      .map(group => ({
        ...group,
        items: group.itemKeys
          .map(key => visibleItemsByKey.get(key))
          .filter((item): item is MenuItem => Boolean(item)),
      }))
      .filter(group => group.items.length > 0)
  }, [visibleMenuItems])

  const toggleGroup = (groupKey: string) => {
    setExpandedGroups(prev => ({
      ...prev,
      [groupKey]: !prev[groupKey],
    }))
  }

  const translateNav = (key: string) => t(`nav:${key}` as never)

  return (
    <aside className={`${isOpen ? 'w-56' : 'w-16'} bg-[var(--color-surface-sidebar)] flex flex-col transition-all duration-300 shadow-xl z-20 shrink-0`}>
      {/* 品牌区 */}
      <div className="h-14 flex items-center justify-center border-b border-white/10 text-white overflow-hidden px-3">
        {isOpen ? (
          <div className="text-center">
            <p className="font-bold text-sm whitespace-nowrap animate-fade-in">{t('nav:brand.full')}</p>
            {roleUser && (
              <p className="text-meta text-foreground-muted mt-0.5">{t('role:label.current')}: {getRoleLabel(roleUser.role)}</p>
            )}
          </div>
        ) : (
          <span className="text-blue-500 font-bold text-lg">{t('nav:brand.short')}</span>
        )}
      </div>

      {/* 菜单区 */}
      <div className="flex-1 px-2 py-2 overflow-y-auto no-scrollbar">
        {visibleMenuGroups.map((group, groupIndex) => {
          const expanded = expandedGroups[group.key] ?? true

          // 折叠态：用 Popover 弹出子菜单
          if (!isOpen) {
            const popoverContent = (
              <div className="py-1 min-w-[160px]">
                <p className="px-3 py-1.5 text-meta font-bold text-foreground-muted uppercase tracking-wider">{translateNav(`group.${group.key}`)}</p>
                {group.items.map(item => (
                  <NavLink
                    key={item.path}
                    to={item.path}
                    end={item.path === '/'}
                    className={({ isActive }) =>
                      `flex items-center gap-2 px-3 py-2 text-sm rounded transition-colors ${
                        isActive ? 'bg-blue-50 text-blue-700 font-medium' : 'text-gray-700 hover:bg-gray-50'
                      }`
                    }
                  >
                    <item.icon size={16} />
                    {translateNav(item.key)}
                  </NavLink>
                ))}
              </div>
            )

            return (
              <div key={group.key} className={groupIndex > 0 ? 'mt-1' : ''}>
                {group.items.map(item => (
                  <Popover
                    key={item.key}
                    content={popoverContent}
                    trigger="hover"
                    placement="right"
                    mouseEnterDelay={0.2}
                    arrow={false}
                  >
                    <NavLink
                      to={item.path}
                      end={item.path === '/'}
                      title={translateNav(item.key)}
                      className={({ isActive }) =>
                        `group relative w-full flex items-center justify-center py-2 my-0.5 rounded-lg transition-all ${
                          isActive ? 'bg-blue-600 text-white shadow-lg shadow-blue-900/20' : 'text-slate-400 hover:bg-slate-800 hover:text-white'
                        }`
                      }
                    >
                      {({ isActive }) => (
                        <item.icon size={18} className={isActive ? 'animate-pulse-slow' : ''} />
                      )}
                    </NavLink>
                  </Popover>
                ))}
              </div>
            )
          }

          // 展开态：正常渲染
          return (
            <div key={group.key} className={groupIndex > 0 ? 'mt-2 pt-2 border-t border-white/10' : ''}>
              <button
                type="button"
                onClick={() => toggleGroup(group.key)}
                className="w-full flex items-center justify-between px-3 py-1.5 text-xs font-semibold tracking-wide text-slate-300 hover:text-white transition-colors"
              >
                <span>{translateNav(`group.${group.key}`)}</span>
                <ChevronDown size={13} className={`text-slate-500 transition-transform ${expanded ? '' : '-rotate-90'}`} />
              </button>

              {expanded && group.items.map(item => (
                <NavLink
                  key={item.path}
                  to={item.path}
                  end={item.path === '/'}
                  className={({ isActive }) =>
                    `group relative w-full flex items-center px-2.5 py-2 my-0.5 rounded-lg transition-all ${
                      isActive ? 'bg-blue-600 text-white shadow-lg shadow-blue-900/20' : 'text-slate-400 hover:bg-slate-800 hover:text-white'
                    }`
                  }
                >
                  {({ isActive }) => (
                    <>
                      <item.icon size={18} className={`shrink-0 ${isActive ? 'animate-pulse-slow' : ''}`} />
                      <span className="ml-2.5 font-medium text-[13px] leading-5 whitespace-nowrap">
                        {translateNav(item.key)}
                      </span>
                    </>
                  )}
                </NavLink>
              ))}
            </div>
          )
        })}
      </div>

      {/* 底部留白（版本号预留） */}
      <div className="p-3 border-t border-white/10">
        {/* TODO U5 版本号 */}
      </div>
    </aside>
  )
}
