/**
 * REST 登录服务
 * 支持用户名/密码登录
 */

import { restApi } from './restClient'
import * as tokenStorage from '@/utils/token'
import { clearSelectedRole, type RoleUser } from './role'

export function isLocalPreviewLoginEnabled() {
  return import.meta.env.DEV || import.meta.env.VITE_ENABLE_LOCAL_PREVIEW_LOGIN === 'true'
}

function normalizeRole(role: string | undefined): RoleUser['role'] {
  if (!role) {
    return 'ADMIN' as RoleUser['role']
  }

  if (role.toLowerCase() === 'admin') {
    return 'ADMIN' as RoleUser['role']
  }

  return role as RoleUser['role']
}

/**
 * 执行登录
 */
export async function performLogin(credentials: { username: string; password: string }) {
  try {
    // 调用后端登录接口
    const loginResult = await restApi.login(credentials)
    const normalizedRole = normalizeRole(loginResult.role)

    clearSelectedRole()

    // 保存 token
    const jwtRoles = parseJwtRoles(loginResult.token)
    tokenStorage.saveToken({
      accessToken: loginResult.token,
      expiresIn: 86400, // 24 小时（后端 JWT 设置）
      user: {
        id: loginResult.userId,
        name: loginResult.realName || loginResult.username,
        nickname: loginResult.username,
        role: normalizedRole,
        roles: jwtRoles,
        organId: '',
        tenantAddress: '',
      },
    })

    // 登录成功后不自动保存角色，跳转角色选择页

    return loginResult
  } catch (error) {
    console.error('Login failed:', error)

    const fallbackMessage = '登录失败，请检查用户名和密码后重试'

    if (error && typeof error === 'object') {
      const responseData = (error as { response?: { data?: unknown } }).response?.data
      if (responseData && typeof responseData === 'object') {
        const data = responseData as {
          error?: { message?: string }
          message?: string
          msg?: string
          errorMessage?: string
        }

        const message = data.error?.message || data.message || data.msg || data.errorMessage
        if (message) {
          throw new Error(message)
        }
      }
    }

    throw new Error(error instanceof Error && error.message ? error.message : fallbackMessage)
  }
}

export function performLocalPreviewLogin() {
  if (!isLocalPreviewLoginEnabled()) {
    throw new Error('本地预览登录未启用')
  }

  const previewUser: RoleUser = {
    id: 'local-preview',
    name: '本地预览管理员',
    role: 'ADMIN',
    subLabelKey: '系统管理员',
  }

  clearSelectedRole()

  tokenStorage.saveToken({
    accessToken: 'local-preview-token',
    expiresIn: 86400,
    user: {
      id: previewUser.id,
      name: previewUser.name,
      nickname: 'local-preview',
      role: previewUser.role,
      roles: [previewUser.role],
      organId: '',
      tenantAddress: '',
    },
  })

  // 本地预览登录也不自动保存角色
  // saveSelectedRoleUser(previewUser)

  return previewUser
}

/**
 * 登出
 */
export function performLogout(): void {
  tokenStorage.clearToken()
  clearSelectedRole()
  sessionStorage.removeItem('oauth_state')
  sessionStorage.removeItem('oauth_nonce')
}

/**
 * 从 JWT token 中解析 roles 数组
 */
function parseJwtRoles(token: string): string[] {
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
