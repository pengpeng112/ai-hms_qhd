import { Navigate } from 'react-router-dom'
import { getSelectedRole } from '@/services/role'

const ADMIN_ROLES = new Set(['ADMIN', '管理员', '安全管理员', '运维管理员'])

function isAdminRole(): boolean {
  const role = getSelectedRole()
  if (!role) return false
  return ADMIN_ROLES.has(role)
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
