import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { User as UserIcon, ArrowRightLeft, Settings, LogOut } from 'lucide-react'
import { getRoleLabel } from '@/services/role'
import type { AppRole } from '@/services/role'

interface HeaderUserMenuProps {
  username: string
  userRole: AppRole
  onLogout: () => void
}

export default function HeaderUserMenu({ username, userRole, onLogout }: HeaderUserMenuProps) {
  const { t } = useTranslation('nav')
  const navigate = useNavigate()

  const items = [
    {
      key: 'role',
      icon: <UserIcon size={16} />,
      label: `${t('header.currentRole')}: ${getRoleLabel(userRole)}`,
      disabled: true,
    },
    { type: 'divider' as const },
    {
      key: 'switchRole',
      icon: <ArrowRightLeft size={16} />,
      label: t('header.switchRole'),
      onClick: () => navigate('/role-select'),
    },
    {
      key: 'settings',
      icon: <Settings size={16} />,
      label: t('header.settings'),
      onClick: () => navigate('/settings'),
    },
    { type: 'divider' as const },
    {
      key: 'logout',
      icon: <LogOut size={16} />,
      label: t('header.logout'),
      danger: true,
      onClick: onLogout,
    },
  ]

  return (
    <div className="w-52 py-1">
      {/* 用户信息 */}
      <div className="px-4 py-3 border-b border-gray-100">
        <p className="text-sm font-bold text-gray-800">{username}</p>
        <p className="text-meta text-foreground-muted mt-0.5">{getRoleLabel(userRole)}</p>
      </div>

      {/* 菜单项 */}
      <div className="py-1">
        {items.map((item, idx) => {
          if ('type' in item && item.type === 'divider') {
            return <div key={idx} className="h-px bg-gray-100 my-1" />
          }
          if ('key' in item) {
            return (
              <button
                key={item.key}
                onClick={item.disabled ? undefined : item.onClick}
                disabled={item.disabled}
                className={`w-full flex items-center gap-3 px-4 py-2 text-sm transition-colors ${
                  item.disabled
                    ? 'text-gray-400 cursor-default'
                    : item.danger
                      ? 'text-red-600 hover:bg-red-50'
                      : 'text-gray-700 hover:bg-gray-50'
                }`}
              >
                {item.icon}
                {item.label}
              </button>
            )
          }
          return null
        })}
      </div>
    </div>
  )
}
