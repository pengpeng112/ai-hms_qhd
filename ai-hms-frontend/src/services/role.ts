import { UserRole, UserRoleLabel, RoleGroups } from '@/types/original'
import { restApi } from './restClient'

export interface RoleUser {
  id: string
  name: string
  role: UserRole
  avatar?: string
  subLabelKey: string
}

export interface RoleGroup {
  key: string
  labelKey: string
  roles: RoleUser[]
}

const SELECTED_ROLE_KEY = 'selected_role'
const SELECTED_USER_KEY = 'selected_role_user'

const ROLE_SUBLABEL_MAP: Record<string, string> = {
  [UserRole.DOCTOR_CHIEF]: 'role:subLabel.doctorChief',
  [UserRole.DOCTOR_SUPERVISOR]: 'role:subLabel.doctorSupervisor',
  [UserRole.DOCTOR_DUTY]: 'role:subLabel.doctorDuty',
  [UserRole.NURSE_HEAD]: 'role:subLabel.nurseHead',
  [UserRole.NURSE_SCHEDULER]: 'role:subLabel.nurseScheduler',
  [UserRole.NURSE_MANAGER]: 'role:subLabel.nurseManager',
  [UserRole.NURSE_RESPONSIBLE]: 'role:subLabel.nurseResponsible',
  [UserRole.ENGINEER]: 'role:subLabel.engineer',
}

interface ApiUser {
  id: string
  username: string
  realName: string
  role: string
  status: string
}

export async function getRoleUsers(): Promise<RoleUser[]> {
  try {
    const res = await restApi.getUserList()
    if (!Array.isArray(res) || res.length === 0) {
      return []
    }
    return (res as ApiUser[])
      .filter(u => Object.values(UserRole).includes(u.role as UserRole))
      .map(u => ({
        id: u.id,
        name: u.realName || u.username,
        role: u.role as UserRole,
        subLabelKey: ROLE_SUBLABEL_MAP[u.role] || 'role:subLabel.doctorSupervisor',
      }))
  } catch {
    return []
  }
}

export async function getRoleUsersByGroup(): Promise<RoleGroup[]> {
  const users = await getRoleUsers()

  const doctorRoles = RoleGroups.DOCTOR as readonly UserRole[]
  const nurseRoles = RoleGroups.NURSE as readonly UserRole[]
  const techRoles = RoleGroups.TECH as readonly UserRole[]

  return [
    {
      key: 'doctor',
      labelKey: 'role:group.doctor',
      roles: users.filter(u => doctorRoles.includes(u.role)),
    },
    {
      key: 'nurse',
      labelKey: 'role:group.nurse',
      roles: users.filter(u => nurseRoles.includes(u.role)),
    },
    {
      key: 'tech',
      labelKey: 'role:group.tech',
      roles: users.filter(u => techRoles.includes(u.role)),
    },
  ]
}

export function getDefaultRouteByRole(role: UserRole): string {
  switch (role) {
    case UserRole.DOCTOR_CHIEF:
    case UserRole.NURSE_HEAD:
      return '/ward-overview'
    case UserRole.DOCTOR_SUPERVISOR:
    case UserRole.DOCTOR_DUTY:
      return '/patients'
    case UserRole.NURSE_SCHEDULER:
      return '/schedule'
    case UserRole.NURSE_MANAGER:
      return '/inventory'
    case UserRole.NURSE_RESPONSIBLE:
      return '/dialysis-processing'
    case UserRole.ENGINEER:
      return '/device-binding'
    default:
      return '/dashboard'
  }
}

const SUPPORTED_MENU_KEYS = new Set([
  'dashboard',
  'ward_overview',
  'patients',
  'monitoring',
  'dialysis_processing',
  'schedule',
  'inventory',
  'device_binding',
  'statistics',
  'master_data',
  'treatment_config',
  'dict_config',
  'settings',
])

const PERMISSION_TO_MENU_KEY: Record<string, string> = {
  'menu.dashboard': 'dashboard',
  'menu.ward_overview': 'ward_overview',
  'menu.patients': 'patients',
  'menu.monitoring': 'monitoring',
  'menu.dialysis_processing': 'dialysis_processing',
  'menu.schedule': 'schedule',
  'menu.inventory': 'inventory',
  'menu.device_binding': 'device_binding',
  'menu.statistics': 'statistics',
  'menu.master_data': 'master_data',
  'menu.treatment_config': 'treatment_config',
  'menu.dict_config': 'dict_config',
  'menu.settings': 'settings',
}

const normalizePermissionToMenuKey = (code: string): string | null => {
  const normalizedCode = code.trim().toLowerCase()
  if (!normalizedCode) {
    return null
  }

  if (PERMISSION_TO_MENU_KEY[normalizedCode]) {
    return PERMISSION_TO_MENU_KEY[normalizedCode]
  }

  if (SUPPORTED_MENU_KEYS.has(normalizedCode)) {
    return normalizedCode
  }

  return null
}

export async function getRolePermissionCodes(role: UserRole): Promise<string[]> {
  try {
    const res = await restApi.getRolePermissions(role)
    const items = Array.isArray(res.data.permissionCodes) ? res.data.permissionCodes : []
    return items
      .map(item => item.trim().toLowerCase())
      .filter(Boolean)
  } catch {
    return []
  }
}

export async function getMenusByRole(role: UserRole): Promise<string[]> {
  const items = await getRolePermissionCodes(role)
  const menuSet = new Set<string>()
  for (const code of items) {
    const menuKey = normalizePermissionToMenuKey(code)
    if (menuKey) {
      menuSet.add(menuKey)
    }
  }
  return Array.from(menuSet)
}

export function saveSelectedRoleUser(user: RoleUser): void {
  localStorage.setItem(SELECTED_ROLE_KEY, user.role)
  localStorage.setItem(SELECTED_USER_KEY, JSON.stringify(user))
}

export function getSelectedRoleUser(): RoleUser | null {
  const userStr = localStorage.getItem(SELECTED_USER_KEY)
  if (!userStr) return null
  try {
    return JSON.parse(userStr) as RoleUser
  } catch {
    return null
  }
}

export function getSelectedRole(): UserRole | null {
  const role = localStorage.getItem(SELECTED_ROLE_KEY)
  if (!role) return null
  if (Object.values(UserRole).includes(role as UserRole)) {
    return role as UserRole
  }
  return null
}

export function clearSelectedRole(): void {
  localStorage.removeItem(SELECTED_ROLE_KEY)
  localStorage.removeItem(SELECTED_USER_KEY)
}

export function hasSelectedRole(): boolean {
  return getSelectedRole() !== null
}

export { UserRole, UserRoleLabel, RoleGroups }
