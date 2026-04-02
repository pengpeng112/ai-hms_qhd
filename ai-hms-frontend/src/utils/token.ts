// Token 存储和管理工具

const TOKEN_KEY = 'hdis_access_token'
const USER_KEY = 'hdis_user_info'
const TOKEN_EXPIRY_KEY = 'hdis_token_expiry'

export interface UserInfo {
  id: string
  name: string
  nickname: string
  role?: string
  organId: string
  tenantAddress: string
}

export interface TokenData {
  accessToken: string
  expiresIn: number
  user: UserInfo | null
}

// 保存 Token
export function saveToken(data: TokenData): void {
  localStorage.setItem(TOKEN_KEY, data.accessToken)

  if (data.user) {
    localStorage.setItem(USER_KEY, JSON.stringify(data.user))
  }

  // 计算过期时间
  const expiryTime = Date.now() + data.expiresIn * 1000
  localStorage.setItem(TOKEN_EXPIRY_KEY, expiryTime.toString())
}

// 获取 Token
export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

// 获取用户信息
export function getUserInfo(): UserInfo | null {
  const userStr = localStorage.getItem(USER_KEY)
  if (!userStr) return null

  try {
    return JSON.parse(userStr)
  } catch {
    return null
  }
}

// 检查 Token 是否过期
export function isTokenExpired(): boolean {
  const expiryStr = localStorage.getItem(TOKEN_EXPIRY_KEY)
  if (!expiryStr) return true

  const expiryTime = parseInt(expiryStr, 10)
  // 提前 5 分钟认为过期
  return Date.now() > expiryTime - 5 * 60 * 1000
}

// 检查是否已登录
export function isLoggedIn(): boolean {
  const token = getToken()
  return !!token && !isTokenExpired()
}

// 清除 Token
export function clearToken(): void {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(USER_KEY)
  localStorage.removeItem(TOKEN_EXPIRY_KEY)
}

// 获取 Token 剩余有效时间（秒）
export function getTokenRemainingTime(): number {
  const expiryStr = localStorage.getItem(TOKEN_EXPIRY_KEY)
  if (!expiryStr) return 0

  const expiryTime = parseInt(expiryStr, 10)
  const remaining = Math.max(0, expiryTime - Date.now())
  return Math.floor(remaining / 1000)
}

// 解析 JWT Payload
export function parseJwtPayload(token: string): Record<string, unknown> | null {
  try {
    const base64Payload = token.split('.')[1]
    const payload = atob(base64Payload)
    return JSON.parse(payload)
  } catch {
    return null
  }
}
