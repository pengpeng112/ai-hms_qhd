/**
 * 功能开发中占位页面
 *
 * 用于尚未实现的功能模块
 */

import { Construction, ArrowLeft } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'

interface UnderDevelopmentProps {
  title?: string
  description?: string
}

export default function UnderDevelopment({
  title,
  description
}: UnderDevelopmentProps) {
  const navigate = useNavigate()
  const { t } = useTranslation(['common'])

  const displayTitle = title || t('common:dev.title')
  const displayDescription = description || t('common:dev.description')

  return (
    <div className="flex flex-col items-center justify-center min-h-[60vh] text-center px-6">
      {/* 图标 */}
      <div className="w-24 h-24 bg-blue-50 rounded-full flex items-center justify-center mb-6">
        <Construction size={48} className="text-blue-500" />
      </div>

      {/* 标题 */}
      <h1 className="text-2xl font-bold text-gray-800 mb-3">
        {displayTitle}
      </h1>

      {/* 描述 */}
      <p className="text-gray-500 max-w-md mb-8">
        {displayDescription}
      </p>

      {/* 返回按钮 */}
      <button
        onClick={() => navigate(-1)}
        className="flex items-center gap-2 px-5 py-2.5 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
      >
        <ArrowLeft size={18} />
        {t('common:dev.backButton')}
      </button>

      {/* 版本信息 */}
      <p className="mt-12 text-xs text-gray-400">
        {t('common:dev.version')}
      </p>
    </div>
  )
}
