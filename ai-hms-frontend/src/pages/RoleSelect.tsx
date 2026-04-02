/**
 * 角色选择页面
 * 
 * 登录后进入此页面选择角色，根据角色跳转到对应的初始页面
 */

import { useState, useEffect, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import {
  UserCheck,
  Stethoscope,
  Activity,
  ClipboardList,
  CalendarClock,
  Truck,
  UserCog,
  Wrench,
  LogOut,
  Loader2,
} from 'lucide-react'
import {
  getRoleUsersByGroup,
  saveSelectedRoleUser,
  getDefaultRouteByRole,
  type RoleUser,
  type RoleGroup,
  UserRole,
} from '@/services/role'
import { logout } from '@/services/auth'

// 角色图标映射
const RoleIcons: Record<UserRole, React.ComponentType<{ size?: number; className?: string }>> = {
  [UserRole.DOCTOR_CHIEF]: UserCheck,
  [UserRole.DOCTOR_SUPERVISOR]: Stethoscope,
  [UserRole.DOCTOR_DUTY]: Activity,
  [UserRole.NURSE_HEAD]: ClipboardList,
  [UserRole.NURSE_SCHEDULER]: CalendarClock,
  [UserRole.NURSE_MANAGER]: Truck,
  [UserRole.NURSE_RESPONSIBLE]: UserCog,
  [UserRole.ENGINEER]: Wrench,
}

// 角色颜色映射
const RoleColors: Record<UserRole, string> = {
  [UserRole.DOCTOR_CHIEF]: 'border-indigo-600',
  [UserRole.DOCTOR_SUPERVISOR]: 'border-blue-500',
  [UserRole.DOCTOR_DUTY]: 'border-blue-400',
  [UserRole.NURSE_HEAD]: 'border-teal-600',
  [UserRole.NURSE_SCHEDULER]: 'border-teal-500',
  [UserRole.NURSE_MANAGER]: 'border-teal-400',
  [UserRole.NURSE_RESPONSIBLE]: 'border-emerald-400',
  [UserRole.ENGINEER]: 'border-orange-500',
}

// 角色按钮组件
function RoleButton({
  user,
  onClick,
  t,
}: {
  user: RoleUser
  onClick: () => void
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  t: (key: any) => string
}) {
  const Icon = RoleIcons[user.role] || UserCheck
  const colorClass = RoleColors[user.role] || 'border-gray-400'

  return (
    <button
      onClick={onClick}
      className={`relative flex flex-col items-center justify-center p-5 bg-white rounded-xl shadow-sm hover:shadow-lg hover:-translate-y-1 transition-all duration-300 border-t-4 ${colorClass} group`}
    >
      <div className="p-3 rounded-full mb-3 bg-gray-50 group-hover:bg-gray-100 transition-colors">
        <Icon size={28} className="text-gray-700" />
      </div>
      <span className="font-bold text-gray-800 text-base">{user.name}</span>
      <span className="text-xs text-gray-400 mt-1">{t(user.subLabelKey)}</span>
    </button>
  )
}

// 分组分隔线
function GroupDivider({ label }: { label: string }) {
  return (
    <div className="col-span-full my-2 flex items-center">
      <div className="h-px bg-gray-300 flex-1"></div>
      <span className="px-4 text-gray-400 text-sm font-medium">{label}</span>
      <div className="h-px bg-gray-300 flex-1"></div>
    </div>
  )
}

export default function RoleSelect() {
  const navigate = useNavigate()
  const { t } = useTranslation(['common', 'role'])
  const [roleGroups, setRoleGroups] = useState<RoleGroup[]>([])
  const [loading, setLoading] = useState(true)

  const loadRoleUsers = useCallback(async () => {
    setLoading(true)
    try {
      const groups = await getRoleUsersByGroup()
      setRoleGroups(groups)
    } catch (error) {
      console.error(t('common:role.loadError'), error)
    } finally {
      setLoading(false)
    }
  }, [t])

  useEffect(() => {
    loadRoleUsers()
  }, [loadRoleUsers])

  const handleRoleSelect = (user: RoleUser) => {
    // 保存选中的角色
    saveSelectedRoleUser(user)
    // 根据角色获取默认路由并跳转
    const defaultRoute = getDefaultRouteByRole(user.role)
    navigate(defaultRoute)
  }

  const handleLogout = () => {
    logout()
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-100 flex items-center justify-center">
        <Loader2 className="w-8 h-8 animate-spin text-blue-500" />
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-100 flex flex-col items-center justify-center p-6">
      {/* 登出按钮 */}
      <button
        onClick={handleLogout}
        className="absolute top-6 right-6 flex items-center gap-2 px-4 py-2 text-gray-500 hover:text-gray-700 hover:bg-gray-200 rounded-lg transition-colors"
      >
        <LogOut size={18} />
        <span className="text-sm">{t('common:auth.logout')}</span>
      </button>

      {/* 标题 */}
      <div className="max-w-5xl w-full text-center mb-10">
        <h1 className="text-3xl font-extrabold text-slate-800 mb-2">
          {t('common:app.aiTitle')}
        </h1>
        <p className="text-slate-500">{t('common:app.aiSubtitle')}</p>
      </div>

      {/* 角色选择网格 */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-5 w-full max-w-6xl">
        {roleGroups.map((group) => (
          <div key={group.key} className="contents">
            {/* 分组标题 - 第一组不需要上边距 */}
            <GroupDivider label={t(group.labelKey as never)} />

            {/* 该分组的角色按钮 */}
            {group.roles.map((user) => (
              <RoleButton
                key={user.id}
                user={user}
                onClick={() => handleRoleSelect(user)}
                t={t}
              />
            ))}
          </div>
        ))}
      </div>

      {/* 版本信息 */}
      <div className="mt-10 text-gray-400 text-xs">
        {t('common:app.version')}
      </div>
    </div>
  )
}
