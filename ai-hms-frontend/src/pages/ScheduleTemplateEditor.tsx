import { Button } from 'antd'
import { ArrowLeft } from 'lucide-react'
import { useNavigate } from 'react-router-dom'

export default function ScheduleTemplateEditor() {
  const navigate = useNavigate()

  return (
    <div className="max-w-[1200px] mx-auto">
      <div className="flex items-center gap-4 mb-6">
        <Button onClick={() => navigate('/schedule-templates')} icon={<ArrowLeft size={16} />}>返回列表</Button>
        <h2 className="text-h2 font-bold text-foreground">编辑排班模板</h2>
      </div>
      <div className="bg-surface rounded-lg border border-gray-200 p-8 text-center text-foreground-muted">
        <p>排班模板编辑器待实现</p>
        <p className="text-meta mt-2">TODO: 拖拽排班 + 模板保存（待 Schedule 业务重构方案）</p>
      </div>
    </div>
  )
}
