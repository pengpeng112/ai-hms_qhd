import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { performLogin } from '@/services/restAuth'
import { useAuth } from '@/contexts/AuthContext'
import { LogIn, AlertCircle, Loader2, User, Lock } from 'lucide-react'

export default function Login() {
  const navigate = useNavigate()
  const { refreshAuth } = useAuth()
  const { t } = useTranslation(['common'])

  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  // 密码登录表单
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')

  // 密码登录
  const handlePasswordLogin = async () => {
    if (!username.trim() || !password.trim()) {
      setError('请输入用户名和密码')
      return
    }

    setLoading(true)
    setError('')

    try {
      await performLogin({
        username: username.trim(),
        password: password.trim(),
      })

      // 登录成功
      refreshAuth()
      navigate('/')
    } catch (err) {
      setError(err instanceof Error ? err.message : '登录失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 via-white to-teal-50 flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        {/* Logo & Title */}
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-16 h-16 bg-gradient-to-br from-blue-500 to-teal-500 rounded-2xl shadow-lg mb-4">
            <span className="text-white text-2xl font-bold">HD</span>
          </div>
          <h1 className="text-2xl font-bold bg-gradient-to-r from-blue-600 to-teal-600 bg-clip-text text-transparent">
            {t('common:app.title')}
          </h1>
          <p className="text-gray-600 mt-2">{t('common:app.subtitle')}</p>
        </div>

        {/* Login Card */}
        <div className="bg-white rounded-2xl shadow-xl border border-gray-100 p-8">
          <h2 className="text-xl font-semibold text-gray-800 mb-6">用户登录</h2>

          {/* 错误提示 */}
          {error && (
            <div data-testid="login-error" className="flex items-center gap-2 text-red-600 bg-red-50 px-4 py-3 rounded-lg text-sm mb-6">
              <AlertCircle size={16} />
              {error}
            </div>
          )}

          {/* 密码登录表单 */}
          <div className="space-y-5">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                用户名
              </label>
              <div className="relative">
                <User size={18} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
                <input
                  type="text"
                  data-testid="login-username"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  onKeyDown={(e) => e.key === 'Enter' && handlePasswordLogin()}
                  className="w-full pl-11 pr-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-teal-500 focus:border-transparent transition-all"
                  placeholder="请输入用户名"
                  disabled={loading}
                  autoComplete="username"
                  autoFocus
                />
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                密码
              </label>
              <div className="relative">
                <Lock size={18} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
                <input
                  type="password"
                  data-testid="login-password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  onKeyDown={(e) => e.key === 'Enter' && handlePasswordLogin()}
                  className="w-full pl-11 pr-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-teal-500 focus:border-transparent transition-all"
                  placeholder="请输入密码"
                  disabled={loading}
                  autoComplete="current-password"
                />
              </div>
            </div>

            <button
              onClick={handlePasswordLogin}
              disabled={loading}
              data-testid="login-submit"
              className="w-full py-3 bg-gradient-to-r from-blue-500 to-teal-500 text-white font-medium rounded-xl hover:from-blue-600 hover:to-teal-600 transition-all shadow-lg shadow-teal-500/25 flex items-center justify-center gap-2 disabled:opacity-50"
            >
              {loading ? (
                <>
                  <Loader2 size={20} className="animate-spin" />
                  登录中...
                </>
              ) : (
                <>
                  <LogIn size={20} />
                  登录
                </>
              )}
            </button>
          </div>
        </div>

        {/* Footer */}
        <p className="text-center text-gray-400 text-sm mt-6">
          {t('common:app.copyright')}
        </p>
      </div>
    </div>
  )
}
