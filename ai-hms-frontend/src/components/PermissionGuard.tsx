import { Navigate } from 'react-router-dom'
import { getSelectedRole } from '@/services/role'
import { getUserInfo } from '@/utils/token'

const ADMIN_ROLES = new Set(['ADMIN', '管理员', '安全管理员', '运维管理员'])

function isAdminRole(): boolean {
  // JWT 解析后的 roles 数组优先
  const userInfo = getUserInfo()
  if (userInfo?.roles && userInfo.roles.length > 0) {
    const hasAdmin = userInfo.roles.some((r) => ADMIN_ROLES.has(r))
    if (!hasAdmin) {
      console.warn('[PermissionGuard] roles 数组中无管理员角色:', userInfo.roles)
    }
    return hasAdmin
  }

  // 兼容旧格式：role 单字段
  if (userInfo?.role && ADMIN_ROLES.has(userInfo.role)) {
    return true
  }

  // fallback 到本地缓存的 selectedRole
  const selected = getSelectedRole()
  if (selected && ADMIN_ROLES.has(selected)) {
    return true
  }

  console.warn('[PermissionGuard] 无管理员权限。userInfo:', userInfo, 'selectedRole:', selected)
  return false
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
