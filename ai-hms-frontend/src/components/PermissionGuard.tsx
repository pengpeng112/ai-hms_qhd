import { Navigate } from 'react-router-dom'
import { getSelectedRole } from '@/services/role'
import { getUserInfo } from '@/utils/token'

const ADMIN_ROLES = new Set(['ADMIN', '管理员', '安全管理员', '运维管理员'])

function isAdminRole(): boolean {
  const selected = getSelectedRole()
  if (!selected) return false

  // 后端返回的真实角色列表优先；若不可用则 fallback 到本地 selectedRole
  const userInfo = getUserInfo()
  if (userInfo) {
    const roles = (userInfo as unknown as Record<string, unknown>).roles as string[] | undefined
    if (roles) {
      return roles.some((r) => ADMIN_ROLES.has(r))
    }
  }

  return ADMIN_ROLES.has(selected)
}

interface PermissionGuardProps {
  children: React.ReactNode
}

export default function PermissionGuard({ children }: PermissionGuardProps) {
  if (!isAdminRole()) {
    return <Navigate to="/dashboard" replace />
  }
  return <>{children}</>
}
