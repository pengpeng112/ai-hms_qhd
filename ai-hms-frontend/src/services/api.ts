/**
 * GraphQL API 客户端
 * 用于与后端 HDIS API 通信
 */

import type { PaginatedItem, PaginatedResponse, QueryParams } from './types/api'

// localStorage keys
const API_URL_KEY = 'hms_api_url'          // Settings 页面配置
const API_TOKEN_KEY = 'hms_api_token'      // Settings 页面配置
const AUTH_TOKEN_KEY = 'hdis_access_token' // 登录系统 token
const AUTH_USER_KEY = 'hdis_user_info'     // 登录系统用户信息

// 获取 API 配置：优先使用登录系统的 token，其次使用 Settings 配置
function getApiConfig() {
    // Token 优先级：登录系统 > Settings 配置 > 环境变量
    const authToken = localStorage.getItem(AUTH_TOKEN_KEY)
    const settingsToken = localStorage.getItem(API_TOKEN_KEY)
    const envToken = import.meta.env.VITE_API_TOKEN || ''

    // API URL 优先级：登录系统用户的 tenantAddress > Settings 配置 > 环境变量
    let apiUrl = localStorage.getItem(API_URL_KEY) || import.meta.env.VITE_API_URL || ''

    // 尝试从登录用户信息中获取 API URL
    const userInfoStr = localStorage.getItem(AUTH_USER_KEY)
    if (userInfoStr) {
        try {
            const userInfo = JSON.parse(userInfoStr)
            if (userInfo.tenantAddress) {
                // tenantAddress 格式: https://hdis.ingatek.com:7777/api/python
                // 需要转换为 GraphQL 端点: https://hdis.ingatek.com:7777/api/python/pygql
                const baseUrl = userInfo.tenantAddress.replace(/\/$/, '')
                apiUrl = baseUrl.includes('/pygql') ? baseUrl : `${baseUrl}/pygql`
            }
        } catch {
            // 解析失败，使用默认配置
        }
    }

    return {
        apiUrl,
        apiToken: authToken || settingsToken || envToken,
    }
}

/**
 * 执行 GraphQL 查询
 * @param query GraphQL 查询字符串
 * @returns 查询结果
 */
export async function graphqlQuery<T>(query: string): Promise<T> {
    const { apiUrl, apiToken } = getApiConfig()

    if (!apiUrl || !apiToken) {
        throw new Error('API 未配置，请先在设置页面配置 API 地址和 Token')
    }

    const formData = new FormData()
    formData.append('query', query)

    const response = await fetch(apiUrl, {
        method: 'POST',
        headers: {
            'Authorization': `Bearer ${apiToken}`,
        },
        body: formData,
    })

    if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
    }

    const result = await response.json()

    // HDIS API 直接返回数据对象，不包裹在 data 中
    // 检查是否返回 null（表示该实体不存在或无数据）
    if (result === null) {
        return {} as T
    }

    // 检查是否有错误
    if (result.errors && result.errors.length > 0) {
        throw new Error(result.errors[0].message)
    }

    // 直接返回结果（API返回格式为 { PatientInfomation: [...] }）
    return result as T
}

/**
 * 构建带分页参数的查询
 * 注意：API 要求参数使用双引号格式
 */
export function buildPaginatedQuery(
    entityName: string,
    fields: string[],
    page: number = 1,
    pageSize: number = 50
): string {
    const fieldsStr = fields.join('\n  ')
    const params = { pageSize, page }
    const paramsStr = JSON.stringify(params).replace(/"/g, '\\"')
    return `{${entityName}(parameters:"${paramsStr}"){
  ${fieldsStr}
  RowCount
}}`
}

// ============ 扩展查询构建器 ============

/**
 * 检查 API 配置是否有效
 */
export function isApiConfigured(): boolean {
    const { apiUrl, apiToken } = getApiConfig()
    return !!(apiUrl && apiToken)
}

/**
 * 构建带过滤条件的查询
 * 注意：API 要求参数使用双引号，且过滤字段名使用小驼峰（如 patientId）
 */
export function buildFilteredQuery(
    entityName: string,
    fields: string[],
    filters: Record<string, string | number | boolean>,
    page: number = 1,
    pageSize: number = 50
): string {
    const fieldsStr = fields.join('\n  ')
    const params: Record<string, string> = {
        pageSize: String(pageSize),
        page: String(page),
    }
    // 将过滤条件的 key 转换为小驼峰格式（首字母小写）
    Object.entries(filters).forEach(([key, value]) => {
        const camelKey = key.charAt(0).toLowerCase() + key.slice(1)
        params[camelKey] = String(value)
    })
    // 使用双引号并转义
    const paramsStr = JSON.stringify(params).replace(/"/g, '\\"')
    return `{${entityName}(parameters:"${paramsStr}"){
  ${fieldsStr}
  RowCount
}}`
}

/**
 * 构建简单查询（无分页）
 */
export function buildSimpleQuery(
    entityName: string,
    fields: string[]
): string {
    const fieldsStr = fields.join('\n  ')
    return `{${entityName}{
  ${fieldsStr}
}}`
}

// ============ 通用数据获取工具 ============

/**
 * 通用分页数据获取函数
 */
export async function fetchPaginatedData<T extends PaginatedItem>(
    entityName: string,
    fields: string[],
    options: QueryParams = {}
): Promise<PaginatedResponse<T>> {
    const { page = 1, pageSize = 50, filter } = options

    const query = filter
        ? buildFilteredQuery(entityName, fields, filter as Record<string, string | number | boolean>, page, pageSize)
        : buildPaginatedQuery(entityName, fields, page, pageSize)

    const result = await graphqlQuery<Record<string, T[]>>(query)
    const items = result[entityName] || []
    const total = items.length > 0 ? (items[0].RowCount || items.length) : 0

    return { data: items, total }
}

/**
 * 通用列表数据获取函数（无分页）
 */
export async function fetchListData<T>(
    entityName: string,
    fields: string[]
): Promise<T[]> {
    const query = buildSimpleQuery(entityName, fields)
    const result = await graphqlQuery<Record<string, T[]>>(query)
    return result[entityName] || []
}

/**
 * 通用带过滤的数据获取
 */
export async function fetchFilteredData<T extends PaginatedItem>(
    entityName: string,
    fields: string[],
    filters: Record<string, string | number | boolean>,
    page: number = 1,
    pageSize: number = 100
): Promise<PaginatedResponse<T>> {
    const query = buildFilteredQuery(entityName, fields, filters, page, pageSize)
    const result = await graphqlQuery<Record<string, T[]>>(query)
    const items = result[entityName] || []
    const total = items.length > 0 ? (items[0].RowCount || items.length) : 0

    return { data: items, total }
}

// ============ 日期工具 ============

/**
 * 获取今日日期字符串 (YYYY-MM-DD)
 */
export function getTodayString(): string {
    return new Date().toISOString().split('T')[0]
}

/**
 * 格式化日期为 API 需要的格式
 */
export function formatDateForApi(date: Date | string): string {
    if (typeof date === 'string') return date
    return date.toISOString().split('T')[0]
}
