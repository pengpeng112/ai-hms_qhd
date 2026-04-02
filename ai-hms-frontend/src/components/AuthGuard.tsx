import { useEffect, useState, useTransition } from 'react'
import { Navigate, useLocation } from 'react-router-dom'
import { isLoggedIn } from '@/utils/token'
import { hasSelectedRole } from '@/services/role'
import { handleOAuthCallback } from '@/services/auth'
import { useAuth } from '@/contexts/AuthContext'
import { Loader2 } from 'lucide-react'

interface AuthGuardProps {
    children: React.ReactNode
}

// 检测是否有 OAuth 回调参数
function hasOAuthCallback(): boolean {
    const hash = window.location.hash
    return !!(hash && hash.includes('access_token'))
}

export default function AuthGuard({ children }: AuthGuardProps) {
    const location = useLocation()
    const { refreshAuth } = useAuth()
    const [, startTransition] = useTransition()
    // 初始状态：如果有 OAuth 回调参数，就先处理
    const [oauthHandled, setOauthHandled] = useState(!hasOAuthCallback())

    useEffect(() => {
        // 检测是否是 OAuth 回调（hash 中包含 access_token）
        if (!oauthHandled) {
            const result = handleOAuthCallback()

            if (result.success) {
                refreshAuth()
            }

            startTransition(() => {
                setOauthHandled(true)
            })
        }
    }, [oauthHandled, refreshAuth])

    // 正在处理 OAuth 回调
    if (!oauthHandled) {
        return (
            <div className="min-h-screen flex items-center justify-center">
                <Loader2 className="animate-spin text-teal-500" size={48} />
            </div>
        )
    }

    // 检查是否已登录
    if (!isLoggedIn()) {
        return <Navigate to="/login" state={{ from: location }} replace />
    }

    // 检查是否已选择角色
    if (!hasSelectedRole()) {
        return <Navigate to="/role-select" state={{ from: location }} replace />
    }

    return <>{children}</>
}
