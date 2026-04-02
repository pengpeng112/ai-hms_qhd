/**
 * DetailCard - 详情卡片容器
 * 统一的卡片样式容器组件
 */

export interface DetailCardProps {
    children?: React.ReactNode
    className?: string
}

export default function DetailCard({ children, className = '' }: DetailCardProps) {
    return (
        <div className={`bg-white rounded-3xl border border-slate-200 shadow-sm p-6 hover:shadow-md transition-all duration-300 ${className}`}>
            {children}
        </div>
    )
}
