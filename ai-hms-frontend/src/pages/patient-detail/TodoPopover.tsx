import { useEffect, useRef } from 'react'
import { ClipboardList, AlertCircle, FileText, Shield } from 'lucide-react'

interface TodoItem {
  id: string
  title: string
  description: string
  time: string
  severity: 'high' | 'medium' | 'low'
  icon: 'alert' | 'prescription' | 'shield'
}

interface TodoPopoverProps {
  open: boolean
  onClose: () => void
  anchorRef: React.RefObject<HTMLElement | null>
}

const MOCK_TASKS: TodoItem[] = [
  {
    id: '1',
    title: '保存后同步处方',
    description: '当前方案已修改，建议同步到处方',
    time: '刚刚',
    severity: 'medium',
    icon: 'prescription',
  },
  {
    id: '2',
    title: '目标超滤待复核',
    description: '目标 2.5L，建议与干体重联动校验',
    time: '5分',
    severity: 'high',
    icon: 'alert',
  },
  {
    id: '3',
    title: '院感数据待同步',
    description: '上次同步 15:40:13',
    time: '15分',
    severity: 'low',
    icon: 'shield',
  },
]

const iconMap = {
  alert: AlertCircle,
  prescription: FileText,
  shield: Shield,
}

const severityColorMap = {
  high: 'text-[#ef4444] bg-red-50',
  medium: 'text-[#f59e0b] bg-amber-50',
  low: 'text-[#10b981] bg-emerald-50',
}

export default function TodoPopover({ open, onClose, anchorRef }: TodoPopoverProps) {
  const popoverRef = useRef<HTMLDivElement>(null)

  // 点击外部关闭
  useEffect(() => {
    if (!open) return
    const handleClickOutside = (e: MouseEvent) => {
      const target = e.target as Node
      if (
        popoverRef.current &&
        !popoverRef.current.contains(target) &&
        anchorRef.current &&
        !anchorRef.current.contains(target)
      ) {
        onClose()
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [open, onClose, anchorRef])

  // Esc 键关闭
  useEffect(() => {
    if (!open) return
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose()
    }
    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [open, onClose])

  if (!open) return null

  return (
    <div
      ref={popoverRef}
      className="absolute right-0 top-full mt-2 w-[340px] bg-white rounded-2xl shadow-xl border border-[#e6ebf3] z-50 overflow-hidden"
      style={{ animation: 'fadeInUp 0.2s ease-out' }}
    >
      {/* 头部 */}
      <div className="flex items-center justify-between px-5 py-4 border-b border-[#e6ebf3]">
        <h4 className="text-sm font-bold text-[#0f1f3d] flex items-center gap-2">
          <ClipboardList size={16} className="text-[#1f63ff]" />
          实时待办任务
        </h4>
        <span className="text-xs text-[#6f7f99] bg-[#f5f8fc] px-2.5 py-1 rounded-full font-medium">
          {MOCK_TASKS.length}项
        </span>
      </div>

      {/* 任务列表 */}
      <div className="max-h-[400px] overflow-y-auto">
        {MOCK_TASKS.map((task) => {
          const Icon = iconMap[task.icon]
          return (
            <div
              key={task.id}
              className="flex items-start gap-3 px-5 py-4 hover:bg-[#f5f8fc] transition-colors cursor-pointer border-b border-[#e6ebf3] last:border-b-0"
            >
              <div className={`w-8 h-8 rounded-xl flex items-center justify-center shrink-0 ${severityColorMap[task.severity]}`}>
                <Icon size={14} />
              </div>
              <div className="flex-1 min-w-0">
                <div className="text-sm font-medium text-[#0f1f3d]">{task.title}</div>
                <div className="text-xs text-[#6f7f99] mt-0.5">{task.description}</div>
              </div>
              <span className="text-xs text-[#6f7f99] shrink-0 mt-0.5">{task.time}</span>
            </div>
          )
        })}
      </div>
    </div>
  )
}
