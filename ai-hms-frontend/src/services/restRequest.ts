/**
 * REST 请求通用 helper
 *
 * 封装 axios 实例的常用请求方法，统一错误处理。
 * 新增 REST API 模块时优先使用这些 helper，避免重复 boilerplate。
 */

import { apiClient, type ApiSuccessResponse, type ApiErrorResponse } from './restClient'

type RequestParams = Record<string, string | number | boolean | undefined>

function isSuccessResponse<T>(data: ApiSuccessResponse<T> | ApiErrorResponse): data is ApiSuccessResponse<T> {
  return data.success === true
}

export async function restGet<T>(url: string, params?: RequestParams): Promise<T> {
  const response = await apiClient.get<ApiSuccessResponse<T> | ApiErrorResponse>(url, { params })
  if (!isSuccessResponse(response.data)) {
    const err = response.data as ApiErrorResponse
    throw new Error(err.error?.message || 'API 请求失败')
  }
  return response.data.data
}

export async function restPost<T>(url: string, data?: unknown): Promise<T> {
  const response = await apiClient.post<ApiSuccessResponse<T> | ApiErrorResponse>(url, data)
  if (!isSuccessResponse(response.data)) {
    const err = response.data as ApiErrorResponse
    throw new Error(err.error?.message || 'API 请求失败')
  }
  return response.data.data
}

export async function restPut<T>(url: string, data?: unknown): Promise<T> {
  const response = await apiClient.put<ApiSuccessResponse<T> | ApiErrorResponse>(url, data)
  if (!isSuccessResponse(response.data)) {
    const err = response.data as ApiErrorResponse
    throw new Error(err.error?.message || 'API 请求失败')
  }
  return response.data.data
}

export async function restDelete<T = void>(url: string): Promise<T> {
  const response = await apiClient.delete<ApiSuccessResponse<T> | ApiErrorResponse>(url)
  if (response.status === 204) {
    return undefined as T
  }
  if (!isSuccessResponse(response.data)) {
    const err = response.data as ApiErrorResponse
    throw new Error(err.error?.message || 'API 请求失败')
  }
  return response.data.data
}
