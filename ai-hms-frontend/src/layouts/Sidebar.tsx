import { NavLink } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Popover } from 'antd'
import { getSelectedRoleUser, getMenusByRole, getRoleLabel } from '@/services/role'
import {
  LayoutDashboard,
  Gauge,
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
  RefreshCw,
  FileSpreadsheet,
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
  { key: 'cockpit', path: '/cockpit', icon: Gauge },
  { key: 'wardOverview', path: '/ward-overview', icon: HeartPulse },
  { key: 'patients', path: '/patients', icon: Users },
  { key: 'monitoring', path: '/monitoring', icon: Monitor },
  { key: 'dialysisProcessing', path: '/dialysis-processing', icon: ClipboardCheck },
  { key: 'schedule', path: '/schedule', icon: Calendar },
  { key: 'staffSchedule', path: '/staff-schedule', icon: Calendar },
  { key: 'inventory', path: '/inventory', icon: Package },
  { key: 'deviceBinding', path: '/device-binding', icon: Server },
  { key: 'wardManagement', path: '/ward-management', icon: Building2, hidden: true },
  { key: 'bedManagement', path: '/bed-management', icon: Bed, hidden: true },
  { key: 'wardBedManagement', path: '/ward-bed-management', icon: Building2 },
  { key: 'educationManagement', path: '/education-management', icon: GraduationCap, hidden: true },
  { key: 'statistics', path: '/statistics', icon: BarChart3 },
  { key: 'qcScoring', path: '/qc-scoring', icon: BarChart3 },
  { key: 'masterData', path: '/master-data', icon: Database },
  { key: 'treatmentConfig', path: '/treatment-config', icon: Layers },
  { key: 'dictConfig', path: '/dict-config', icon: BookOpen },
  { key: 'settings', path: '/settings', icon: Settings },
  { key: 'syncCenter', path: '/sync-center', icon: RefreshCw },
  { key: 'cnrdsReport', path: '/cnrds-report', icon: FileSpreadsheet },
  { key: 'userManagement', path: '/user-management', icon: Users },
  { key: 'roleManagement', path: '/role-management', icon: Settings },
]

const menuGroups: MenuGroup[] = [
  { key: 'dailyWork', itemKeys: ['dashboard', 'cockpit', 'wardOverview', 'monitoring', 'dialysisProcessing'] },
  { key: 'patientCenter', itemKeys: ['patients', 'educationManagement'] },
  { key: 'schedule', itemKeys: ['schedule', 'staffSchedule'] },
  { key: 'resource', itemKeys: ['inventory', 'deviceBinding', 'wardBedManagement'] },
  { key: 'systemConfig', itemKeys: ['masterData', 'treatmentConfig', 'dictConfig', 'userManagement', 'roleManagement', 'settings', 'statistics', 'qcScoring', 'syncCenter', 'cnrdsReport'] },
]

const defaultExpandedGroups = Object.fromEntries(menuGroups.map(group => [group.key, true])) as Record<string, boolean>

const roleMenuMap: Record<string, string[]> = {
  dashboard: ['dashboard', 'cockpit'],
  ward_overview: ['wardOverview'],
  patients: ['patients'],
  monitoring: ['monitoring'],
  dialysis_processing: ['dialysisProcessing'],
  schedule: ['schedule', 'staffSchedule'],
  inventory: ['inventory'],
  device_binding: ['deviceBinding'],
  ward_management: ['wardManagement', 'wardBedManagement'],
  bed_management: ['bedManagement', 'wardBedManagement'],
  education_management: ['educationManagement'],
  statistics: ['statistics', 'qcScoring'],
  cnrds_report: ['cnrdsReport'],
  master_data: ['masterData'],
  treatment_config: ['treatmentConfig'],
  dict_config: ['dictConfig'],
  settings: ['settings', 'syncCenter'],
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
      return Object.entries(roleMenuMap).some(([permKey, values]) =>
        values.includes(item.key) && allowedMenuKeys.includes(permKey)
      )
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
    <aside className={`${isOpen ? 'w-60' : 'w-16'} bg-[var(--color-surface-sidebar)] flex flex-col transition-all duration-300 shadow-xl z-20 shrink-0`}>
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
                        `flex items-center justify-center py-2.5 my-1 rounded-[10px] transition-all ${
                          isActive ? 'bg-blue-600 text-white shadow-[0_8px_20px_rgba(29,99,255,0.28)]' : 'text-slate-400 hover:bg-slate-800 hover:text-white'
                        }`
                      }
                    >
                      {() => (
                        <item.icon size={17} />
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
                className="w-full flex items-center justify-between px-2 h-[30px] text-[14px] font-extrabold tracking-[0.2px] text-slate-300 hover:text-white transition-colors"
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
                    `group relative my-1 flex h-[38px] w-full items-center gap-3 rounded-[10px] px-3 text-[15px] font-bold transition-all ${
                      isActive
                        ? 'bg-blue-600 text-white shadow-[0_8px_20px_rgba(29,99,255,0.28)] before:absolute before:left-0 before:top-[9px] before:h-5 before:w-[3px] before:rounded before:bg-blue-200'
                        : 'text-slate-300 hover:bg-white/5 hover:text-white'
                    }`
                  }
                >
                  {() => (
                    <>
                      <item.icon size={17} className="shrink-0" />
                      <span className="whitespace-nowrap leading-5">
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

      {/* 底部身份信息 */}
      <div className="p-3 border-t border-white/10">
        {roleUser && (
          <div className="text-center">
            <p className="text-meta text-slate-400">{t('role:label.current')}</p>
            <p className="text-sm font-medium text-white mt-0.5 truncate">{roleUser.name || roleUser.role}</p>
            <p className="text-meta text-slate-500 mt-0.5">{getRoleLabel(roleUser.role)}</p>
          </div>
        )}
      </div>
    </aside>
  )
}
