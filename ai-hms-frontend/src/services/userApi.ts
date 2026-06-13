/**
 * 用户管理 REST API 模块
 */

import { restGet, restPost, restPut, restDelete } from './restRequest'

// ============ 类型定义 ============

export interface RestUser {
  id: string
  username: string
  realName: string
  gender?: string
  type?: string
  accountType?: string
  phone?: string
  email?: string
  birthdate?: string
  status: string
  sort?: number
  idNumber?: string
  icNumber?: string
  avatar?: string
  isCreateAccount?: boolean
  bindStatus?: 'bound' | 'unbound'
  isSyncCloud?: boolean
  syncStatus?: string
  createdAt?: string
  updatedAt?: string
  role: string
  roles?: string[]
  roleNames?: string[]
  hasSignature?: boolean
  signatureImageId?: string
  signatureImage?: string
  departmentId: number | null
}

export interface CreateUserRequest {
  username: string
  password?: string
  realName: string
  gender?: string
  type?: string
  sort?: number
  phone?: string
  email?: string
  birthdate?: string
  idNumber?: string
  icNumber?: string
  avatar?: string
  signatureImage?: string
  roles?: string[]
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

interface UserListResponse {
  items: RestUser[]
  total: number
  page: number
  pageSize: number
  totalPages: number
}

interface SignatureResponse {
  userId: string
  signatureId?: string
  signatureImage?: string
}

// ============ API 方法 ============

async function getList(params?: UserListParams): Promise<{ items: RestUser[]; total: number }> {
  const data = await restGet<RestUser[] | UserListResponse>('/api/v1/users', params as Record<string, string | number | boolean | undefined>)
  if (Array.isArray(data)) {
    return { items: data, total: data.length }
  }
  return { items: data.items, total: data.total }
}

async function getById(id: string): Promise<RestUser> {
  return restGet<RestUser>(`/api/v1/users/${id}`)
}

async function create(data: CreateUserRequest): Promise<RestUser> {
  return restPost<RestUser>('/api/v1/users', data)
}

async function update(id: string, data: UpdateUserRequest): Promise<RestUser> {
  return restPut<RestUser>(`/api/v1/users/${id}`, data)
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

async function getSignature(id: string): Promise<SignatureResponse> {
  return restGet<SignatureResponse>(`/api/v1/users/${id}/signature`)
}

async function updateSignature(id: string, signatureImage: string): Promise<SignatureResponse> {
  return restPut<SignatureResponse>(`/api/v1/users/${id}/signature`, { signatureImage })
}

async function deleteSignature(id: string): Promise<void> {
  return restDelete<void>(`/api/v1/users/${id}/signature`)
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
  getSignature,
  updateSignature,
  deleteSignature,
}
