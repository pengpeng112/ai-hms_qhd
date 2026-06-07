import { Navigate } from 'react-router-dom'
import { getSelectedRole } from '@/services/role'
import { getUserInfo, getToken } from '@/utils/token'

const ADMIN_ROLES = new Set(['ADMIN', '管理员', '安全管理员', '运维管理员'])

/**
 * 直接从 JWT token payload 解析 roles 数组
 * 完全绕过 localStorage 中可能过期的 userInfo 缓存
 */
function parseRolesFromToken(): string[] {
  const token = getToken()
  if (!token) return []
  try {
    const payload = token.split('.')[1]
    if (!payload) return []
    let b64 = payload.replace(/-/g, '+').replace(/_/g, '/')
    const pad = b64.length % 4
    if (pad) b64 += '='.repeat(4 - pad)
    const obj = JSON.parse(atob(b64))
    return Array.isArray(obj.roles) ? obj.roles : []
  } catch {
    return []
  }
}

function isAdminRole(): boolean {
  // 第一层：直接从当前 JWT token 解析 roles（最可靠，不受缓存影响）
  const tokenRoles = parseRolesFromToken()
  if (tokenRoles.length > 0) {
    const hasAdmin = tokenRoles.some((r) => ADMIN_ROLES.has(r))
    if (!hasAdmin) {
      console.warn('[PermissionGuard] JWT roles 中无管理员角色:', tokenRoles)
    }
    return hasAdmin
  }

  // 第二层：localStorage userInfo（兼容旧缓存）
  const userInfo = getUserInfo()
  if (userInfo?.roles && userInfo.roles.length > 0) {
    const hasAdmin = userInfo.roles.some((r) => ADMIN_ROLES.has(r))
    if (hasAdmin) return true
  }

  // 第三层：role 单字段
  if (userInfo?.role && ADMIN_ROLES.has(userInfo.role)) {
    return true
  }

  // 第四层：本地缓存的 selectedRole
  const selected = getSelectedRole()
  if (selected && ADMIN_ROLES.has(selected)) {
    return true
  }

  console.warn('[PermissionGuard] 无管理员权限。tokenRoles:', tokenRoles, 'userInfo:', userInfo, 'selectedRole:', selected)
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
