/**
 * 登录守卫
 * 
 * 只检查是否已登录，不检查角色选择
 * 用于角色选择页面
 */

import { Navigate, useLocation } from 'react-router-dom'
import { isLoggedIn } from '@/utils/token'

interface LoginGuardProps {
    children: React.ReactNode
}

export default function LoginGuard({ children }: LoginGuardProps) {
    const location = useLocation()

    // 检查是否已登录
    if (!isLoggedIn()) {
        return <Navigate to="/login" state={{ from: location }} replace />
    }

    return <>{children}</>
}
