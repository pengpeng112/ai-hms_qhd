/**
 * REST 登录服务
 * 支持用户名/密码登录
 */

import { restApi } from './restClient'
import * as tokenStorage from '@/utils/token'

/**
 * 执行登录
 */
export async function performLogin(credentials: { username: string; password: string }) {
  try {
    // 调用后端登录接口
    const loginResult = await restApi.login(credentials)

    // 保存 token
    tokenStorage.saveToken({
      accessToken: loginResult.token,
      expiresIn: 86400, // 24 小时（后端 JWT 设置）
      user: {
        id: loginResult.userId,
        name: loginResult.realName || loginResult.username,
        nickname: loginResult.username,
        role: loginResult.role,
        organId: '',
        tenantAddress: '',
      },
    })

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

/**
 * 登出
 */
export function performLogout(): void {
  // 清除本地存储
  localStorage.removeItem('hdis_access_token')
  localStorage.removeItem('hdis_user_info')
  localStorage.removeItem('hdis_token_expiry')

  // 清除会话存储
  sessionStorage.removeItem('oauth_state')
  sessionStorage.removeItem('oauth_nonce')
}
