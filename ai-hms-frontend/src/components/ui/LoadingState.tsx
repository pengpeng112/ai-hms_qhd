/**
 * LoadingState - 统一加载状态组件
 * 用于页面/区块的加载状态展示
 */

import { useTranslation } from 'react-i18next'
import { RefreshCw } from 'lucide-react'

export interface LoadingStateProps {
    /** 加载提示文字 */
    tip?: string
    /** 尺寸 */
    size?: 'sm' | 'md' | 'lg'
    /** 是否全屏居中 */
    fullscreen?: boolean
}

export default function LoadingState({
    tip,
    size = 'md',
    fullscreen = false
}: LoadingStateProps) {
    const { t } = useTranslation('common')
    const displayTip = tip ?? t('status.loading')

    const sizeMap = {
        sm: 16,
        md: 32,
        lg: 48
    }

    const containerClass = fullscreen
        ? 'flex justify-center items-center h-full py-20'
        : 'flex flex-col items-center justify-center py-8'

    return (
        <div className={containerClass}>
            <RefreshCw
                size={sizeMap[size]}
                className="animate-spin text-blue-500 mb-3"
            />
            {displayTip && (
                <p className="text-gray-500 text-sm">{displayTip}</p>
            )}
        </div>
    )
}
