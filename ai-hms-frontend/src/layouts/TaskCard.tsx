import { Tooltip } from 'antd'
import { AlertCircle, FileEdit, Zap, CheckCircle2, Check, UserCheck, ChevronRight, ClipboardList } from 'lucide-react'
import type { RestClinicalTask } from '@/services/restClient'

interface TaskCardProps {
  task: RestClinicalTask
  canHandle: boolean
  onClick?: () => void
}

const severityBarMap: Record<string, string> = {
  high: 'bg-state-alert',
  medium: 'bg-state-waiting',
  low: 'bg-state-treating',
}

const getTaskIcon = (type: string) => {
  switch (type) {
    case 'ALERT': return <AlertCircle size={16} />
    case 'PRESCRIPTION': return <FileEdit size={16} />
    case 'ORDER': return <Zap size={16} />
    case 'ASSESSMENT': return <CheckCircle2 size={16} />
    default: return <ClipboardList size={16} />
  }
}

export default function TaskCard({ task, canHandle, onClick }: TaskCardProps) {
  const barColor = severityBarMap[task.severity] || 'bg-gray-300'

  return (
    <div
      onClick={canHandle ? onClick : undefined}
      className={`group relative p-4 rounded-md bg-surface border border-gray-100 shadow-sm transition-all ${
        canHandle ? 'cursor-pointer hover:bg-surface-sunken active:scale-[0.98]' : 'cursor-not-allowed opacity-70'
      }`}
    >
      {/* 左侧 severity 色条 */}
      <div className={`absolute left-0 top-2 bottom-2 w-1 rounded-full ${barColor}`} />

      <div className="ml-3">
        {/* 标题行 */}
        <div className="flex justify-between items-start mb-1">
          <div className="flex items-center font-bold text-sm text-gray-800">
            <span className="mr-2 p-1.5 bg-gray-50 rounded-lg shrink-0">{getTaskIcon(task.type)}</span>
            {task.title}
          </div>
          <div className="flex items-center gap-1 shrink-0">
            <Tooltip title="即将上线">
              <button disabled className="p-1 text-gray-300 cursor-not-allowed">
                <Check size={14} />
              </button>
            </Tooltip>
            <Tooltip title="即将上线">
              <button disabled className="p-1 text-gray-300 cursor-not-allowed">
                <UserCheck size={14} />
              </button>
            </Tooltip>
            {canHandle && (
              <button onClick={onClick} className="p-1 text-gray-400 hover:text-blue-600 transition-colors">
                <ChevronRight size={14} />
              </button>
            )}
          </div>
        </div>

        {/* 内容 */}
        <p className="text-xs text-gray-600 mb-0.5">{task.patientName || '--'}</p>
        <p className="text-xs text-gray-400 leading-relaxed">{task.description || ''}</p>

        {/* 底部 */}
        <div className={`flex items-center justify-between mt-2 text-meta text-gray-400 transition-transform ${canHandle ? 'group-hover:translate-x-1' : ''}`}>
          {/* eslint-disable-next-line no-restricted-syntax -- density:strict 故意小字（床号角标） */}
          <span className="text-[10px] opacity-60">{task.bedNumber || '--'}</span>
          <span className="uppercase tracking-wider font-bold">
            {canHandle ? '去处理' : '无处理权限'}
          </span>
        </div>
      </div>
    </div>
  )
}
