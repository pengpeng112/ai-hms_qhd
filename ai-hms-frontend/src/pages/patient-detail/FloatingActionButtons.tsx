import { forwardRef } from 'react'
import { ShieldAlert, ClipboardList } from 'lucide-react'

interface FloatingActionButtonsProps {
  onClinicalFocusClick: () => void
  onTodoClick: () => void
  hasRisk?: boolean
  todoCount?: number
  todoButtonRef: React.RefObject<HTMLButtonElement | null>
}

const FloatingActionButtons = forwardRef<HTMLDivElement, FloatingActionButtonsProps>(
  ({ onClinicalFocusClick, onTodoClick, hasRisk = false, todoCount = 3, todoButtonRef }, ref) => {
    return (
      <div ref={ref} className="flex items-center gap-2">
        {/* 临床焦点按钮 */}
        <button
          onClick={onClinicalFocusClick}
          className="relative w-11 h-11 rounded-xl bg-white border border-[#e6ebf3] flex items-center justify-center text-[#6f7f99] hover:text-[#1f63ff] hover:border-[#1f63ff] hover:shadow-md transition-all group"
          title="临床焦点"
        >
          <ShieldAlert size={18} />
          {hasRisk && (
            <span className="absolute -top-1 -right-1 w-4 h-4 bg-[#ef4444] rounded-full border-2 border-white flex items-center justify-center">
              <span className="text-[8px] text-white font-bold">!</span>
            </span>
          )}
          {/* 悬停提示 */}
          <span className="absolute bottom-full mb-2 px-3 py-1.5 bg-[#0f1f3d] text-white text-xs rounded-lg opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap pointer-events-none">
            临床焦点
          </span>
        </button>

        {/* 待办任务按钮 */}
        <button
          ref={todoButtonRef}
          onClick={onTodoClick}
          className="relative w-11 h-11 rounded-xl bg-white border border-[#e6ebf3] flex items-center justify-center text-[#6f7f99] hover:text-[#1f63ff] hover:border-[#1f63ff] hover:shadow-md transition-all group"
          title="实时待办任务"
        >
          <ClipboardList size={18} />
          {todoCount > 0 && (
            <span className="absolute -top-1 -right-1 min-w-[18px] h-[18px] bg-[#ef4444] rounded-full border-2 border-white flex items-center justify-center px-1">
              <span className="text-[10px] text-white font-bold">{todoCount}</span>
            </span>
          )}
          {/* 悬停提示 */}
          <span className="absolute bottom-full mb-2 px-3 py-1.5 bg-[#0f1f3d] text-white text-xs rounded-lg opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap pointer-events-none">
            实时待办任务
          </span>
        </button>
      </div>
    )
  }
)

FloatingActionButtons.displayName = 'FloatingActionButtons'

export default FloatingActionButtons
