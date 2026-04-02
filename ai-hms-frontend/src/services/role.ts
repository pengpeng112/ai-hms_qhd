/**
 * 角色服务层
 * 
 * 提供角色相关的数据获取和操作
 * 当前使用 Mock 数据，后续可对接 API
 */

import { UserRole, UserRoleLabel, RoleGroups } from '@/types/original'

// ============ 类型定义 ============

export interface RoleUser {
  id: string
  name: string
  role: UserRole
  avatar?: string
  subLabelKey: string // 岗位副标题翻译 key
}

export interface RoleGroup {
  key: string
  labelKey: string // 分组标签翻译 key
  roles: RoleUser[]
}

// ============ 存储 Key ============

const SELECTED_ROLE_KEY = 'selected_role'
const SELECTED_USER_KEY = 'selected_role_user'

// ============ Mock 数据 ============
// 每个角色配置一个示例用户，后续从 API 获取

const MOCK_ROLE_USERS: RoleUser[] = [
  // 医生组
  {
    id: 'doctor-chief-1',
    name: '陈主任',
    role: UserRole.DOCTOR_CHIEF,
    subLabelKey: 'role:subLabel.doctorChief',
    avatar: undefined,
  },
  {
    id: 'doctor-supervisor-1',
    name: '王医生',
    role: UserRole.DOCTOR_SUPERVISOR,
    subLabelKey: 'role:subLabel.doctorSupervisor',
    avatar: undefined,
  },
  {
    id: 'doctor-duty-1',
    name: '李医生',
    role: UserRole.DOCTOR_DUTY,
    subLabelKey: 'role:subLabel.doctorDuty',
    avatar: undefined,
  },
  // 护士组
  {
    id: 'nurse-head-1',
    name: '刘护士长',
    role: UserRole.NURSE_HEAD,
    subLabelKey: 'role:subLabel.nurseHead',
    avatar: undefined,
  },
  {
    id: 'nurse-scheduler-1',
    name: '赵护士',
    role: UserRole.NURSE_SCHEDULER,
    subLabelKey: 'role:subLabel.nurseScheduler',
    avatar: undefined,
  },
  {
    id: 'nurse-manager-1',
    name: '孙护士',
    role: UserRole.NURSE_MANAGER,
    subLabelKey: 'role:subLabel.nurseManager',
    avatar: undefined,
  },
  {
    id: 'nurse-responsible-1',
    name: '周护士',
    role: UserRole.NURSE_RESPONSIBLE,
    subLabelKey: 'role:subLabel.nurseResponsible',
    avatar: undefined,
  },
  // 技术组
  {
    id: 'engineer-1',
    name: '吴工程师',
    role: UserRole.ENGINEER,
    subLabelKey: 'role:subLabel.engineer',
    avatar: undefined,
  },
]

// ============ 数据获取函数 ============

/**
 * 获取所有可选角色用户列表
 * TODO: 后续对接 API
 */
export async function getRoleUsers(): Promise<RoleUser[]> {
  // 模拟 API 延迟
  await new Promise(resolve => setTimeout(resolve, 100))
  return MOCK_ROLE_USERS
}

/**
 * 按分组获取角色用户
 * TODO: 后续对接 API
 */
export async function getRoleUsersByGroup(): Promise<RoleGroup[]> {
  const users = await getRoleUsers()
  
  // 类型断言以避免 readonly 数组类型限制
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

/**
 * 根据角色获取默认首页路由
 * 基于 v1.5 设计规范
 */
export function getDefaultRouteByRole(role: UserRole): string {
  switch (role) {
    // 主任/护士长 -> 病区概览
    case UserRole.DOCTOR_CHIEF:
    case UserRole.NURSE_HEAD:
      return '/ward-overview'
    // 医生 -> 患者列表
    case UserRole.DOCTOR_SUPERVISOR:
    case UserRole.DOCTOR_DUTY:
      return '/patients'
    // 排班护士 -> 排班管理
    case UserRole.NURSE_SCHEDULER:
      return '/schedule'
    // 库管护士 -> 耗材管理
    case UserRole.NURSE_MANAGER:
      return '/inventory'
    // 责任护士 -> 透析执行
    case UserRole.NURSE_RESPONSIBLE:
      return '/dialysis-processing'
    // 工程师 -> 设备管理
    case UserRole.ENGINEER:
      return '/device-binding'
    default:
      return '/dashboard'
  }
}

/**
 * 获取角色可访问的菜单项
 *
 * 菜单权限映射基于 v1.5 UI 设计:
 * - dashboard: 所有角色
 * - ward_overview: 科室主任、护士长
 * - patients: 主任、主治、值班、护士长、责任护士
 * - monitoring: 主任、主治、值班、责任护士、排班护士、工程师
 * - dialysis_processing: 责任护士、护士长
 * - schedule: 主任、主治、值班、护士长、排班护士
 * - inventory: 库管护士、护士长
 * - device_binding: 工程师
 * - statistics: 所有角色
 * - master_data: 护士长、工程师
 * - settings: 所有角色
 */
export function getMenusByRole(role: UserRole): string[] {
  switch (role) {
    // 科室主任 - 全局管理视角
    case UserRole.DOCTOR_CHIEF:
      return [
        'dashboard', 'ward_overview', 'patients', 'monitoring',
        'schedule', 'statistics', 'treatment_config', 'dict_config', 'settings'
      ]

    // 主管医生 - 患者诊疗 + 排班查看
    case UserRole.DOCTOR_SUPERVISOR:
      return ['dashboard', 'patients', 'monitoring', 'schedule', 'statistics', 'settings']

    // 值班医生 - 患者诊疗 + 排班查看
    case UserRole.DOCTOR_DUTY:
      return ['dashboard', 'patients', 'monitoring', 'schedule', 'statistics', 'settings']

    // 护士长 - 护理管理全权限
    case UserRole.NURSE_HEAD:
      return [
        'dashboard', 'ward_overview', 'patients',
        'dialysis_processing', 'schedule', 'inventory',
        'statistics', 'master_data', 'treatment_config', 'dict_config', 'settings'
      ]

    // 排班护士 - 排班管理 + 实时监控
    case UserRole.NURSE_SCHEDULER:
      return ['dashboard', 'monitoring', 'schedule', 'statistics', 'settings']

    // 库管护士 - 耗材管理为主
    case UserRole.NURSE_MANAGER:
      return ['dashboard', 'inventory', 'statistics', 'settings']

    // 责任护士 - 透析执行为主
    case UserRole.NURSE_RESPONSIBLE:
      return [
        'dashboard', 'patients', 'monitoring',
        'dialysis_processing', 'statistics', 'settings'
      ]

    // 工程师 - 设备相关
    case UserRole.ENGINEER:
      return [
        'dashboard', 'monitoring', 'device_binding',
        'statistics', 'master_data', 'dict_config', 'settings'
      ]

    default:
      return ['dashboard', 'statistics', 'settings']
  }
}

// ============ 本地存储操作 ============

/**
 * 保存选中的角色用户
 */
export function saveSelectedRoleUser(user: RoleUser): void {
  localStorage.setItem(SELECTED_ROLE_KEY, user.role)
  localStorage.setItem(SELECTED_USER_KEY, JSON.stringify(user))
}

/**
 * 获取已选中的角色用户
 */
export function getSelectedRoleUser(): RoleUser | null {
  const userStr = localStorage.getItem(SELECTED_USER_KEY)
  if (!userStr) return null
  try {
    return JSON.parse(userStr) as RoleUser
  } catch {
    return null
  }
}

/**
 * 获取已选中的角色
 */
export function getSelectedRole(): UserRole | null {
  const role = localStorage.getItem(SELECTED_ROLE_KEY)
  if (!role) return null
  // 验证是否是有效角色
  if (Object.values(UserRole).includes(role as UserRole)) {
    return role as UserRole
  }
  return null
}

/**
 * 清除角色选择
 */
export function clearSelectedRole(): void {
  localStorage.removeItem(SELECTED_ROLE_KEY)
  localStorage.removeItem(SELECTED_USER_KEY)
}

/**
 * 检查是否已选择角色
 */
export function hasSelectedRole(): boolean {
  return getSelectedRole() !== null
}

// ============ 导出角色常量 ============

export { UserRole, UserRoleLabel, RoleGroups }
