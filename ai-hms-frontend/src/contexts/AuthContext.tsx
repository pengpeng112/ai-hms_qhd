import { createContext, useContext, useState, useEffect, useTransition, type ReactNode } from 'react'
import { isLoggedIn, getToken, getUserInfo, clearToken, type UserInfo } from '@/utils/token'

interface AuthContextType {
  isAuthenticated: boolean
  user: UserInfo | null
  token: string | null
  logout: () => void
  refreshAuth: () => void
}

const AuthContext = createContext<AuthContextType | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [isAuthenticated, setIsAuthenticated] = useState(false)
  const [user, setUser] = useState<UserInfo | null>(null)
  const [token, setToken] = useState<string | null>(null)
  const [, startTransition] = useTransition()

  const refreshAuth = () => {
    const loggedIn = isLoggedIn()
    setIsAuthenticated(loggedIn)
    setUser(loggedIn ? getUserInfo() : null)
    setToken(loggedIn ? getToken() : null)
  }

  const logout = () => {
    clearToken()
    setIsAuthenticated(false)
    setUser(null)
    setToken(null)
    window.location.href = '/login'
  }

  useEffect(() => {
    // 使用 startTransition 避免直接在 effect 中调用 setState
    startTransition(() => {
      refreshAuth()
    })

    // 监听 storage 变化（多标签页同步）
    const handleStorageChange = (e: StorageEvent) => {
      if (e.key?.startsWith('hdis_')) {
        startTransition(() => {
          refreshAuth()
        })
      }
    }

    window.addEventListener('storage', handleStorageChange)
    return () => window.removeEventListener('storage', handleStorageChange)
  }, [])

  return (
    <AuthContext.Provider value={{ isAuthenticated, user, token, logout, refreshAuth }}>
      {children}
    </AuthContext.Provider>
  )
}

/* eslint-disable react-refresh/only-export-components */
export function useAuth() {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider')
  }
  return context
}
/* eslint-enable react-refresh/only-export-components */
