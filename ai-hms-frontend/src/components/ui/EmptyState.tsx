/**
 * EmptyState - 空状态组件
 * 用于列表为空时的展示
 */

import { useTranslation } from 'react-i18next'
import { Filter, CheckCircle2, type LucideIcon } from 'lucide-react'

export interface EmptyStateProps {
    /** 图标 */
    icon?: LucideIcon
    /** 提示文字 */
    message?: string
    /** 子元素（如操作按钮） */
    children?: React.ReactNode
}

export default function EmptyState({
    icon: Icon = Filter,
    message,
    children
}: EmptyStateProps) {
    const { t } = useTranslation('common')
    const displayMessage = message ?? t('empty.default')

    return (
        <div className="flex flex-col items-center justify-center py-12 text-gray-400">
            <div className="bg-gray-50 p-4 rounded-full mb-3">
                <Icon size={24} className="text-gray-300" />
            </div>
            <p className="text-sm">{displayMessage}</p>
            {children && <div className="mt-4">{children}</div>}
        </div>
    )
}

/**
 * 任务完成空状态变体
 */
export function TasksCompletedState({ message }: { message?: string }) {
    const { t } = useTranslation('common')
    const displayMessage = message ?? t('empty.tasks')

    return (
        <EmptyState
            icon={CheckCircle2}
            message={displayMessage}
        />
    )
}
