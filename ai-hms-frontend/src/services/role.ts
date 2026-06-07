import { UserRole, UserRoleLabel, RoleGroups } from '@/types/original'
import { restApi } from './restClient'
import { getUserInfo } from '@/utils/token'

export type AppRole = UserRole | 'ADMIN'

export interface RoleUser {
  id: string
  name: string
  role: AppRole
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
  ADMIN: '系统管理员',
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

// 从 GET /api/v1/me/roles 获取当前登录用户可选角色
export async function getMyRoles(): Promise<RoleUser[]> {
  try {
    const res = await restApi.getMyRoles()
    const data = res.data as { userId: string; username: string; realName: string; roles: string[] }
    if (Array.isArray(data.roles) && data.roles.length > 0) {
      return data.roles.map(roleCode => {
        const role = normalizeAppRole(roleCode)
        return {
          id: data.userId,
          name: data.realName || data.username,
          role,
          subLabelKey: ROLE_SUBLABEL_MAP[role] || '系统管理员',
        }
      })
    }
  } catch {
    // 接口异常，使用本地登录用户兜底
  }
  return buildFallbackFromCurrentUser()
}

// 从当前已登录用户信息构建兜底角色列表
function buildFallbackFromCurrentUser(): RoleUser[] {
  const userInfo = getUserInfo()
  if (!userInfo) return []

  const role = normalizeAppRole(userInfo.role)
  return [{
    id: userInfo.id || '0',
    name: userInfo.name || userInfo.nickname || '管理员',
    role,
    subLabelKey: ROLE_SUBLABEL_MAP[role] || '系统管理员',
  }]
}

function normalizeAppRole(role: string | null | undefined): AppRole {
  if (!role || role.toLowerCase() === 'admin') {
    return 'ADMIN'
  }
  if (role === '医生') {
    return UserRole.DOCTOR_DUTY
  }
  if (role === '护士') {
    return UserRole.NURSE_RESPONSIBLE
  }

  return role as UserRole
}

export async function getRoleUsers(): Promise<RoleUser[]> {
  try {
    const res = await restApi.getUserList()
    if (Array.isArray(res) && res.length > 0) {
      const mapped = (res as ApiUser[])
        .filter(u => u.role === 'ADMIN' || u.role === '医生' || u.role === '护士' || Object.values(UserRole).includes(u.role as UserRole))
        .map(u => {
          const role = normalizeAppRole(u.role)
          return {
            id: u.id,
            name: u.realName || u.username,
            role,
            subLabelKey: ROLE_SUBLABEL_MAP[role] || 'role:subLabel.doctorSupervisor',
          }
        })
      if (mapped.length > 0) return mapped
    }
  } catch {
    // 接口异常，使用本地登录用户兜底
  }
  // API 返回空或报错 → 用本地 localStorage 中的已登录用户
  return buildFallbackFromCurrentUser()
}

export async function getRoleUsersByGroup(): Promise<RoleGroup[]> {
  const users = await getMyRoles()

  const doctorRoles = RoleGroups.DOCTOR as readonly UserRole[]
  const nurseRoles = RoleGroups.NURSE as readonly UserRole[]
  const techRoles = RoleGroups.TECH as readonly UserRole[]

  return [
    {
      key: 'admin',
      labelKey: '系统管理组',
      roles: users.filter(u => u.role === 'ADMIN'),
    },
    {
      key: 'doctor',
      labelKey: 'role:group.doctor',
      roles: users.filter(u => u.role !== 'ADMIN' && doctorRoles.includes(u.role)),
    },
    {
      key: 'nurse',
      labelKey: 'role:group.nurse',
      roles: users.filter(u => u.role !== 'ADMIN' && nurseRoles.includes(u.role)),
    },
    {
      key: 'tech',
      labelKey: 'role:group.tech',
      roles: users.filter(u => u.role !== 'ADMIN' && techRoles.includes(u.role)),
    },
  ]
}

export function getDefaultRouteByRole(role: AppRole): string {
  switch (role) {
    case 'ADMIN':
      return '/dashboard'
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
  'ward_management',
  'bed_management',
  'education_management',
  'statistics',
  'master_data',
  'treatment_config',
  'dict_config',
  'settings',
  'user_management',
  'role_management',
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
  'menu.ward_management': 'ward_management',
  'menu.bed_management': 'bed_management',
  'menu.education_management': 'education_management',
  'menu.statistics': 'statistics',
  'menu.master_data': 'master_data',
  'menu.treatment_config': 'treatment_config',
  'menu.dict_config': 'dict_config',
  'menu.settings': 'settings',
  'menu.user_management': 'user_management',
  'menu.role_management': 'role_management',
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

export async function getRolePermissionCodes(role: AppRole): Promise<string[]> {
  if (role === 'ADMIN') {
    return []
  }

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

// 所有菜单 key（与 Sidebar.tsx roleMenuMap 保持一致）
const ALL_MENU_KEYS = [
  'dashboard', 'ward_overview', 'patients', 'monitoring',
  'dialysis_processing', 'schedule', 'inventory', 'device_binding',
  'ward_management', 'bed_management', 'education_management',
  'statistics', 'master_data', 'treatment_config', 'dict_config', 'settings',
  'user_management', 'role_management',
]

export async function getMenusByRole(role: AppRole): Promise<string[]> {
  // admin 直接返回全部菜单，不查接口
  if ((role as string) === 'ADMIN' || (role as string) === 'admin') {
    return ALL_MENU_KEYS
  }

  try {
    const items = await getRolePermissionCodes(role)
    if (items.length > 0) {
      const menuSet = new Set<string>()
      for (const code of items) {
        const menuKey = normalizePermissionToMenuKey(code)
        if (menuKey) menuSet.add(menuKey)
      }
      if (menuSet.size > 0) return Array.from(menuSet)
    }
  } catch {
    // 接口失败时不能扩大权限，交给页面展示无权限状态。
  }

  return []
}

export function getMenuKeyByPath(pathname: string): string | null {
  const path = pathname.split('?')[0].split('#')[0]
  if (path === '/' || path === '/dashboard') return 'dashboard'
  if (path === '/ward-overview') return 'ward_overview'
  if (path === '/monitoring') return 'monitoring'
  if (path === '/dialysis-processing') return 'dialysis_processing'
  if (path === '/schedule') return 'schedule'
  if (path === '/patients' || path.startsWith('/patients/')) return 'patients'
  if (path === '/education-management') return 'education_management'
  if (path === '/inventory') return 'inventory'
  if (path === '/device-binding') return 'device_binding'
  if (path === '/ward-management') return 'ward_management'
  if (path === '/bed-management') return 'bed_management'
  if (path === '/master-data') return 'master_data'
  if (path === '/treatment-config') return 'treatment_config'
  if (path === '/dict-config') return 'dict_config'
  if (path === '/user-management') return 'user_management'
  if (path === '/role-management') return 'role_management'
  if (path === '/settings') return 'settings'
  if (path === '/statistics') return 'statistics'
  if (path === '/schedule-templates' || path.startsWith('/schedule-templates/')) return 'schedule'
  if (path === '/shift-config') return 'schedule'
  return null
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

export function getSelectedRole(): AppRole | null {
  const role = localStorage.getItem(SELECTED_ROLE_KEY)
  if (!role) return null
  if (role === 'ADMIN') {
    return role
  }
  if (Object.values(UserRole).includes(role as UserRole)) {
    return role as UserRole
  }
  return null
}

export function getRoleLabel(role: AppRole): string {
  if (role === 'ADMIN') {
    return '系统管理员'
  }

  return UserRoleLabel[role]
}

export function clearSelectedRole(): void {
  localStorage.removeItem(SELECTED_ROLE_KEY)
  localStorage.removeItem(SELECTED_USER_KEY)
}

export function hasSelectedRole(): boolean {
  return getSelectedRole() !== null
}

export { UserRole, UserRoleLabel, RoleGroups }
