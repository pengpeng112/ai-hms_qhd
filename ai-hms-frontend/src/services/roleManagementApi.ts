/**
 * 角色管理 REST API 模块
 *
 * 从 restClient.ts 中提取的角色相关接口，使用 restRequest 统一 helper。
 * 新代码优先导入此模块，restClient.ts 中的同名方法为历史兼容 facade。
 */

import { restGet, restPost, restPut, restDelete } from './restRequest'

// ============ 类型定义 ============

export interface AppRoleApi {
  code: string
  name: string
  description?: string
  isSystem?: boolean
  sortOrder?: number
}

export interface PermissionNodeApi {
  id: string
  name: string
  code: string
  children?: PermissionNodeApi[]
}

// ============ API 方法 ============

async function getRoleList(): Promise<AppRoleApi[]> {
  const data = await restGet<{ items: AppRoleApi[] }>('/api/v1/app-roles')
  return data.items
}

async function createRole(data: Partial<AppRoleApi>): Promise<AppRoleApi> {
  return restPost<AppRoleApi>('/api/v1/app-roles', data)
}

async function updateRole(code: string, data: Partial<AppRoleApi>): Promise<AppRoleApi> {
  return restPut<AppRoleApi>(`/api/v1/app-roles/${code}`, data)
}

async function deleteRole(code: string): Promise<void> {
  return restDelete<void>(`/api/v1/app-roles/${code}`)
}

async function getPermissionTree(): Promise<PermissionNodeApi[]> {
  const data = await restGet<{ items: PermissionNodeApi[] }>('/api/v1/app-permissions/tree')
  return data.items
}

async function getRolePermissions(role: string): Promise<{ role: string; permissionCodes: string[] }> {
  return restGet<{ role: string; permissionCodes: string[] }>(`/api/v1/role-permissions/${role}`)
}

async function setRolePermissions(role: string, permissionCodes: string[]): Promise<{ role: string; permissionCodes: string[] }> {
  return restPut<{ role: string; permissionCodes: string[] }>(`/api/v1/role-permissions/${role}`, { permissionCodes })
}

// ============ 统一导出 ============

export const roleManagementApi = {
  getRoleList,
  createRole,
  updateRole,
  deleteRole,
  getPermissionTree,
  getRolePermissions,
  setRolePermissions,
}
