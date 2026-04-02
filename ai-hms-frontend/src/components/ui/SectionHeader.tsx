/**
 * SectionHeader - 区块标题组件
 * 用于各种卡片/区块的标题展示
 */

import type { LucideIcon } from 'lucide-react'

export interface SectionHeaderProps {
    /** 标题图标 */
    icon: LucideIcon
    /** 标题文字 */
    title: string
    /** 右侧操作区域 */
    action?: React.ReactNode
    /** 图标颜色 class */
    iconColor?: string
    /** 深色模式 */
    dark?: boolean
}

export default function SectionHeader({
    icon: Icon,
    title,
    action,
    iconColor,
    dark = false
}: SectionHeaderProps) {
    const textColor = dark ? 'text-slate-200' : 'text-slate-800'
    const defaultIconColor = dark ? 'text-blue-400' : 'text-blue-600'

    return (
        <div className="flex items-center justify-between mb-4 px-1">
            <h3 className={`text-sm font-black uppercase tracking-wider flex items-center ${textColor}`}>
                <Icon size={18} className={`mr-2 ${iconColor || defaultIconColor}`} /> {title}
            </h3>
            {action && action}
        </div>
    )
}
