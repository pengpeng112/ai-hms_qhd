/**
 * LabelValue - 标签-值对展示组件
 * 用于展示 "标签: 值" 格式的数据
 */

export interface LabelValueProps {
    /** 标签文字 */
    label: string
    /** 值内容 */
    value: React.ReactNode
    /** 值的颜色 class */
    color?: string
}

export default function LabelValue({
    label,
    value,
    color = 'text-slate-900'
}: LabelValueProps) {
    return (
        <div className="flex flex-col">
            <span className="text-[10px] text-slate-400 font-black uppercase mb-1 tracking-tighter">
                {label}
            </span>
            <div className={`text-sm font-bold ${color}`}>
                {value || <span className="text-slate-300 font-normal">暂无记录</span>}
            </div>
        </div>
    )
}
