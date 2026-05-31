/**
 * 用户管理 REST API 模块
 *
 * 从 restClient.ts 中提取的用户相关接口，使用 restRequest 统一 helper。
 * 新代码优先导入此模块，restClient.ts 中的同名方法为历史兼容 facade。
 */

import { restGet, restPost, restPut, restDelete } from './restRequest'

// ============ 类型定义 ============

export interface RestUser {
  id: string
  username: string
  realName: string
  role: string
  roles?: string[]
  roleNames?: string[]
  gender?: string
  type?: string
  accountType?: string
  phone?: string
  email?: string
  birthdate?: string
  syncStatus?: string
  updatedAt?: string
  status: string
  departmentId: number | null
}

export interface CreateUserRequest {
  username: string
  password?: string
  realName: string
  role?: string
  roles?: string[]
  gender?: string
  phone?: string
  email?: string
  type?: string
  accountType?: string
  departmentId?: number | null
}

export type UpdateUserRequest = Partial<CreateUserRequest>

export interface UserListParams {
  keyword?: string
  type?: string
  role?: string
  status?: string
  syncStatus?: string
  page?: number
  pageSize?: number
}

interface RestUserListResponse {
  items: RestUser[]
  total: number
  page: number
  pageSize: number
  totalPages: number
}

// ============ API 方法 ============

async function getList(params?: UserListParams): Promise<{ items: RestUser[]; total: number }> {
  const data = await restGet<RestUser[] | RestUserListResponse>('/api/v1/users', params as Record<string, string | number | boolean | undefined>)
  if (Array.isArray(data)) {
    return { items: data, total: data.length }
  }
  return { items: data.items, total: data.total }
}

async function getById(id: string): Promise<unknown> {
  return restGet<unknown>(`/api/v1/users/${id}`)
}

async function create(data: CreateUserRequest): Promise<unknown> {
  return restPost<unknown>('/api/v1/users', data)
}

async function update(id: string, data: UpdateUserRequest): Promise<unknown> {
  return restPut<unknown>(`/api/v1/users/${id}`, data)
}

async function updateStatus(id: string, status: string): Promise<void> {
  return restPut<void>(`/api/v1/users/${id}/status`, { status })
}

async function remove(id: string): Promise<void> {
  return restDelete<void>(`/api/v1/users/${id}`)
}

async function resetPassword(id: string, newPassword: string): Promise<void> {
  return restPut<void>(`/api/v1/users/${id}/password`, { newPassword })
}

async function getRoles(id: string): Promise<string[]> {
  const data = await restGet<{ roles: string[] }>(`/api/v1/users/${id}/roles`)
  return data.roles
}

async function setRoles(id: string, roleCodes: string[]): Promise<void> {
  return restPut<void>(`/api/v1/users/${id}/roles`, { roleCodes })
}

async function getMyRoles(): Promise<{ userId: string; username: string; realName: string; roles: string[] }> {
  return restGet<{ userId: string; username: string; realName: string; roles: string[] }>('/api/v1/me/roles')
}

// ============ 统一导出 ============

export const userApi = {
  getList,
  getById,
  create,
  update,
  updateStatus,
  remove,
  resetPassword,
  getRoles,
  setRoles,
  getMyRoles,
}
