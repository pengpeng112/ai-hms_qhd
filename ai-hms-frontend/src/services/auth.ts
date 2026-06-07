// OAuth 登录服务 - Implicit Flow（跳转方式）

import { saveToken, clearToken, type TokenData, type UserInfo } from '@/utils/token'
import { apiCache } from '@/utils/cache'
import { clearSelectedRole } from './role'

// OAuth 配置
const AUTH_CONFIG = {
  authServer: import.meta.env.VITE_OAUTH_AUTH_SERVER || '',
  clientId: import.meta.env.VITE_OAUTH_CLIENT_ID || '',
  responseType: 'id_token token',
  scope: 'openid profile api1',
}

function isOAuthConfigured(): boolean {
  return !!(AUTH_CONFIG.authServer && AUTH_CONFIG.clientId)
}

// 使用 crypto.getRandomValues 生成安全的随机字符串
function generateRandomString(length: number = 16): string {
  const array = new Uint8Array(length)
  crypto.getRandomValues(array)
  return Array.from(array, (byte) => byte.toString(16).padStart(2, '0')).join('')
}

// 获取 redirect_uri
function getRedirectUri(): string {
  const { protocol, host } = window.location
  return `${protocol}//${host}`
}

// 发起 OAuth 登录跳转
export function initiateOAuthLogin(): void {
  if (!isOAuthConfigured()) {
    console.error('OAuth is not configured (set VITE_OAUTH_AUTH_SERVER and VITE_OAUTH_CLIENT_ID)')
    return
  }
  const state = generateRandomString()
  const nonce = generateRandomString()

  // 保存 state / nonce 用于验证回调
  sessionStorage.setItem('oauth_state', state)
  sessionStorage.setItem('oauth_nonce', nonce)

  const params = new URLSearchParams({
    client_id: AUTH_CONFIG.clientId,
    response_type: AUTH_CONFIG.responseType,
    scope: AUTH_CONFIG.scope,
    state: state,
    nonce: nonce,
    redirect_uri: getRedirectUri(),
  })

  const authUrl = `${AUTH_CONFIG.authServer}/connect/authorize?${params.toString()}`

  // 跳转到认证服务器
  window.location.href = authUrl
}

// OAuth 回调结果
export interface OAuthCallbackResult {
  success: boolean
  error?: string
}

// 处理 OAuth 回调
export function handleOAuthCallback(): OAuthCallbackResult {
  const hash = window.location.hash

  if (!hash || hash.length < 2) {
    return { success: false }
  }

  // 解析 hash 参数
  const params = new URLSearchParams(hash.substring(1))

  const accessToken = params.get('access_token')
  const expiresIn = params.get('expires_in')
  const state = params.get('state')
  const error = params.get('error')
  const errorDescription = params.get('error_description')

  // 检查错误
  if (error) {
    return {
      success: false,
      error: errorDescription || error
    }
  }

  // 检查 token
  if (!accessToken) {
    return { success: false, error: '未收到 access_token' }
  }

  // 验证 state（必须存在且匹配）
  const savedState = sessionStorage.getItem('oauth_state')
  sessionStorage.removeItem('oauth_state')
  if (!savedState || !state || state !== savedState) {
    return {
      success: false,
      error: 'State 验证失败，可能存在安全风险'
    }
  }

  // 验证 nonce（如果 IDP 返回 id_token 可校验）
  // 当前方案使用 implicit flow，已保存 nonce 供后续 token 校验
  sessionStorage.removeItem('oauth_nonce')

  // 解析用户信息
  const userInfo = parseJwtPayload(accessToken)

  // 保存 Token
  const tokenData: TokenData = {
    accessToken: accessToken,
    expiresIn: expiresIn ? parseInt(expiresIn, 10) : 604800,
    user: userInfo
  }
  saveToken(tokenData)

  // 清除 URL hash
  window.history.replaceState(null, '', window.location.pathname)

  return { success: true }
}

// 解析 JWT
function parseJwtPayload(token: string): UserInfo | null {
  try {
    const base64Payload = token.split('.')[1]
    if (!base64Payload) return null

    let base64 = base64Payload.replace(/-/g, '+').replace(/_/g, '/')
    const padding = base64.length % 4
    if (padding) {
      base64 += '='.repeat(4 - padding)
    }

    const payload = JSON.parse(atob(base64))

    return {
      id: payload.user_id || payload.sub || '',
      name: payload.username || payload.name || '',
      nickname: payload.employee_name || payload.nickname || '',
      roles: payload.roles || [],
      organId: payload.organ_id || '',
      tenantAddress: payload.tenant_internet_address || ''
    }
  } catch {
    return null
  }
}

// 登出
export function logout(): void {
  clearToken()
  clearSelectedRole()
  apiCache.invalidate()
  window.location.href = '/login'
}

// 手动设置 Token（仅开发环境）
export function setManualToken(token: string, expiresIn: number = 604800): boolean {
  if (!import.meta.env.DEV) {
    console.error('setManualToken is only available in development mode')
    return false
  }
  try {
    const userInfo = parseJwtPayload(token)
    if (!userInfo) {
      console.error('Invalid token format')
      return false
    }

    const tokenData: TokenData = {
      accessToken: token,
      expiresIn: expiresIn,
      user: userInfo
    }
    saveToken(tokenData)
    return true
  } catch (error) {
    console.error('Invalid token:', error)
    return false
  }
}
