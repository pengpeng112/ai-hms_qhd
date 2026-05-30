import { ClipboardList, Lock, Unlock } from 'lucide-react'

interface TaskbarRailProps {
  taskCount: number
  highCount: number
  mediumCount: number
  lowCount: number
  locked: boolean
  onExpand: () => void
  onToggleLock: () => void
}

export default function TaskbarRail({
  taskCount,
  highCount,
  mediumCount,
  lowCount,
  locked,
  onExpand,
  onToggleLock,
}: TaskbarRailProps) {
  const severityItems = [
    { label: '紧急', count: highCount, color: 'bg-state-alert text-white' },
    { label: '提醒', count: mediumCount, color: 'bg-state-waiting text-white' },
    { label: '普通', count: lowCount, color: 'bg-state-treating text-white' },
  ]

  return (
    <div className="w-14 bg-surface border-l border-gray-200 flex flex-col items-center py-3 shadow-xl">
      {/* 顶部：展开按钮 */}
      <button
        onClick={onExpand}
        className="relative p-2 rounded-lg text-gray-500 hover:bg-gray-100 transition-colors"
        title="展开任务栏"
      >
        <ClipboardList size={20} />
        {taskCount > 0 && (
          <span className="absolute -top-1 -right-1 min-w-[14px] h-3.5 bg-red-500 text-white text-[9px] flex items-center justify-center rounded-full px-0.5">
            {taskCount > 99 ? '99+' : taskCount}
          </span>
        )}
      </button>

      {/* 中部：severity 分类 */}
      <div className="flex-1 flex flex-col items-center justify-center gap-3 my-4">
        {severityItems.map(item => (
          item.count > 0 && (
            <button
              key={item.label}
              onClick={onExpand}
              className="flex flex-col items-center gap-0.5 group"
              title={`${item.label}: ${item.count}`}
            >
              {/* eslint-disable-next-line no-restricted-syntax -- density:strict 故意小字（severity 计数角标） */}
              <span className={`min-w-[20px] h-5 text-[10px] font-bold flex items-center justify-center rounded-full px-1 ${item.color}`}>
                {item.count}
              </span>
              <span className="text-[9px] text-gray-400 group-hover:text-gray-600">{item.label}</span>
            </button>
          )
        ))}
      </div>

      {/* 底部：锁定按钮 */}
      <button
        onClick={onToggleLock}
        className={`p-2 rounded-lg transition-colors ${
          locked ? 'text-blue-600 bg-blue-50' : 'text-gray-400 hover:bg-gray-100'
        }`}
        title={locked ? '解锁任务栏' : '锁定任务栏'}
      >
        {locked ? <Lock size={16} /> : <Unlock size={16} />}
      </button>
    </div>
  )
}
